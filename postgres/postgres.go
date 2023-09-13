package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client represents postgres database client.
type Client struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

// NewClient creates new instance of postgres client.
func NewClient(conf Config, log *slog.Logger) (*Client, error) {
	log = log.With("name", "postgres")

	// Validates database connection string and setup config setDefaults.
	poolConf, err := pgxpool.ParseConfig(conf.URL)
	if err != nil {
		return nil, fmt.Errorf("could not parse database URL: %s", err)
	}
	poolConf.MaxConns = int32(conf.MaxConns)
	poolConf.MaxConnLifetime = conf.MaxConnLifetime
	poolConf.MaxConnIdleTime = conf.MaxConnIdleTime

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		return nil, fmt.Errorf("could not connection to database: %s", err)
	}
	if err = pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	ml := &migrationLogger{log}
	if err = autoMigrate(conf.URL, ml); err != nil {
		return nil, fmt.Errorf("auto-migration failed: %s", err)
	}

	c := &Client{db: pool, logger: log}
	return c, nil
}

// Ping checks connection that sends empty sql statement.
func (c *Client) Ping() (ok bool, err error) {
	if err = c.db.Ping(context.Background()); err != nil {
		return
	}
	return true, nil
}

// Close closes all connection.
func (c *Client) Close() error {
	c.db.Close()
	return nil
}

type Config struct {
	URL             string
	MaxConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

//go:embed migrations/*.sql
var fs embed.FS

func autoMigrate(databaseURL string, log migrate.Logger) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, databaseURL)
	if err != nil {
		return err
	}
	if log != nil {
		m.Log = log
	}
	if err = m.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

type migrationLogger struct {
	logger *slog.Logger
}

func (m *migrationLogger) Printf(format string, v ...interface{}) {
	m.logger.Info(fmt.Sprintf(format, v...))
}

func (m *migrationLogger) Verbose() bool {
	return m.logger.Enabled(context.Background(), slog.LevelDebug)
}
