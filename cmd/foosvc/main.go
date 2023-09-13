package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kudarap/foo"
	"github.com/kudarap/foo/config"
	"github.com/kudarap/foo/fakeproducer"
	"github.com/kudarap/foo/firebase"
	"github.com/kudarap/foo/logging"
	"github.com/kudarap/foo/postgres"
	"github.com/kudarap/foo/server"
	"github.com/kudarap/foo/telemetry"
	"github.com/kudarap/foo/worker"
)

const (
	modeServer = "server"
	modeWorker = "worker"
)

type App struct {
	config   *config.Config
	server   *server.Server
	worker   *worker.Worker
	logger   *slog.Logger
	version  server.Version
	closerFn func() error
}

func (a *App) Setup() error {
	postgresClient, err := postgres.NewClient(a.config.Postgres, a.logger)
	if err != nil {
		return fmt.Errorf("could not setup postgres: %s", err)
	}
	firebaseClient, err := firebase.NewClient(a.config.GoogleApplicationCredentials)
	if err != nil {
		return fmt.Errorf("could not setup firebase: %s", err)
	}

	service := telemetry.TraceFooService(foo.NewService(postgresClient, a.logger))

	tsi := telemetry.NewServerInstrumentation(a.config.Telemetry.ServiceName)
	a.server = server.New(a.config.Server, service, firebaseClient, postgresClient, tsi, a.version, a.logger)

	fp := fakeproducer.New(time.Second)
	a.worker = worker.New(fp, a.config.WorkerQueueSize, a.logger)
	a.worker.Use(worker.LoggingMiddleware(a.logger), telemetry.TraceWorker)
	a.worker.HandleFunc("demo", worker.FakeFighterConsumer(a.logger))

	a.closerFn = func() error {
		if err = postgresClient.Close(); err != nil {
			return fmt.Errorf("could not close postgres: %s", err)
		}
		if err = a.server.Close(); err != nil {
			return fmt.Errorf("could not close server: %s", err)
		}
		return nil
	}
	return nil
}

func (a *App) Run(mode string) error {
	switch mode {
	case modeServer:
		return appRunner(a.server)
	case modeWorker:
		return appRunner(a.worker)
	default:
		return fmt.Errorf("app mode not supported: %s", mode)
	}
}

func main() {
	log := logging.New()

	mode, err := appMode()
	if err != nil {
		log.Error("app mode", "err", err)
		return
	}

	log.Info("app: loading config...")
	c, err := config.LoadDefault()
	if err != nil {
		log.Error("could not load config", "err", err)
		return
	}

	version := buildVer()
	log.Info(fmt.Sprintf("telemetry enabled:%v url:%s", c.Telemetry.Enabled, c.Telemetry.CollectorURL))
	shutdown, err := telemetry.InitProvider(c.Telemetry, mode, version.Tag)
	if err != nil {
		log.Error("could not init telemetry provider", "err", err)
		return
	}
	defer func() {
		if err = shutdown(context.Background()); err != nil {
			log.Error("failed to shutdown TracerProvider: %w", err)
		}
	}()
	log = telemetry.TraceLogger(log)

	app := &App{config: c, logger: log, version: version}
	if err = app.Setup(); err != nil {
		log.Error("could not setup app", "err", err)
		return
	}
	log.Info("running", "mode", mode, "version", version)
	if err = app.Run(mode); err != nil {
		log.Error("could not run app", "err", err)
		return
	}
	if err = app.closerFn(); err != nil {
		log.Error("error occurred when closing app", "err", err)
	}
	log.Info("app: exited!")
}

// appMode returns application mode base on arguments.
func appMode() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("app mode required")
	}
	return strings.ToLower(os.Args[1]), nil
}

func appRunner(app runner) error {
	done := make(chan error, 1)
	// Waits for CTRL-C or os SIGINT for server shutdown.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		done <- app.Stop()
	}()

	if err := app.Run(); err != nil {
		return err
	}
	return <-done
}

type runner interface {
	Run() error
	Stop() error
}

// version data will set during build time using go build -ldflags.
var vTag, vCommit, vBuilt string

func buildVer() server.Version {
	ts, _ := strconv.Atoi(vBuilt)
	bt := time.Unix(int64(ts), 0)
	return server.Version{
		Tag:    vTag,
		Commit: vCommit,
		Built:  bt,
	}
}
