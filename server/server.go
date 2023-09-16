package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Server represents application server.
type Server struct {
	*http.Server
	config Config

	service         service
	authenticator   authenticator
	databaseChecker databasePinger

	tracing tracing
	logger  *slog.Logger
	Version Version
}

// New creates new instance of Server.
func New(
	config Config,
	service service,
	authenticator authenticator,
	databaseChecker databasePinger,
	tracing tracing,
	version Version,
	logger *slog.Logger,
) *Server {
	c := config.setDefaults()

	l := logger.With("pkg", "server")
	l.Info("config",
		"addr", c.Addr,
		"read-timeout", c.ReadTimeout.String(),
		"write-timeout", c.WriteTimeout.String(),
		"shutdown-timeout", c.ShutdownTimeout.String(),
	)

	s := &Server{
		service:         service,
		authenticator:   authenticator,
		databaseChecker: databaseChecker,
		tracing:         tracing,
		Version:         version,
		logger:          l,
	}
	s.Server = &http.Server{
		Addr:         c.Addr,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		Handler:      s.Routes(),
	}
	return s
}

// Routes setups middlewares and route endpoints.
func (s *Server) Routes() http.Handler {
	// Add CORS middleware here
	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedOrigins([]string{"*"}), // Allow requests from any origin
	)

	r := mux.NewRouter()
	r.Use(
		s.tracing.Middleware(),
		authentication(s.authenticator),
		requestIDMiddleware,
		s.loggingMiddleware,
		s.recoveryMiddleware,
		cors,
	)

	// Public endpoints
	r.HandleFunc("/version", GetVersion(s.Version)).Methods(http.MethodGet)
	r.HandleFunc("/healthcheck", Healthcheck(s.databaseChecker)).Methods(http.MethodGet)
	r.HandleFunc("/fighters/{id}", GetFighterByID(s.service)).Methods(http.MethodGet)
	r.HandleFunc("/archers/{id}", GetArcherByID(s.service)).Methods(http.MethodGet)
	r.NotFoundHandler = s.noMatchHandler(http.StatusNotFound)
	r.MethodNotAllowedHandler = s.noMatchHandler(http.StatusMethodNotAllowed)

	// Private endpoints
	pr := r.PathPrefix("/").Subrouter()
	//pr.Use(authorizedMiddleware)
	pr.HandleFunc("/fighters", ListFighters(s.service)).Methods(http.MethodGet)
	return r
}

// Stop shuts down server gracefully with deadline of shutdownTimeout.
func (s *Server) Stop() error {
	timeout := s.config.ShutdownTimeout
	done := make(chan error, 1)
	go func() {
		ctx := context.Background()
		var cancel context.CancelFunc
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		s.logger.Info("shutting down gracefully...")
		done <- s.Shutdown(ctx)
		s.logger.Info("shutdown")
	}()
	return <-done
}

// Run starts serving and listening http server with graceful shutdown.
func (s *Server) Run() error {
	s.logger.Info(fmt.Sprintf("running on %s", s.Addr))
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// noMatchHandler handler with logging middlewares since these handlers not being hit
// on router middleware chain and must be instrumented separately.
func (s *Server) noMatchHandler(status int) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		e := errors.New(http.StatusText(status))
		encodeJSONError(w, e, status)
	}
	return requestIDMiddleware(s.loggingMiddleware(http.HandlerFunc(h)))
}

type tracing interface {
	Middleware() func(next http.Handler) http.Handler
}

// default server config values.
const (
	defaultAddr            = ":8000"
	defaultShutdownTimeout = time.Second * 5
)

// Config represents server config.
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func (c Config) setDefaults() Config {
	if strings.TrimSpace(c.Addr) == "" {
		c.Addr = defaultAddr
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = defaultShutdownTimeout
	}
	return c
}
