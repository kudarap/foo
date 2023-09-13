package config

import (
	"errors"
	"os"

	"github.com/kudarap/foo/postgres"
	"github.com/kudarap/foo/server"
	"github.com/kudarap/foo/telemetry"
	"github.com/spf13/viper"
)

const DefaultFile = ".env"

// Config represents application configuration.
type Config struct {
	Server                       server.Config
	WorkerQueueSize              int
	Telemetry                    telemetry.Config
	GoogleApplicationCredentials string
	Postgres                     postgres.Config
}

// Load loads config from environment variables and file.
func Load(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	c := &Config{
		Server: server.Config{
			Addr:         viper.GetString("SERVER_ADDR"),
			ReadTimeout:  viper.GetDuration("SERVER_READ_TIMEOUT"),
			WriteTimeout: viper.GetDuration("SERVER_WRITE_TIMEOUT"),
		},
		WorkerQueueSize: viper.GetInt("WORKER_QUEUE_SIZE"),
		Telemetry: telemetry.Config{
			Enabled:      viper.GetBool("TELEMETRY_ENABLED"),
			CollectorURL: viper.GetString("TELEMETRY_COLLECTOR_URL"),
			ServiceName:  viper.GetString("TELEMETRY_SERVICE_NAME"),
			Env:          viper.GetString("TELEMETRY_ENV"),
		},
		Postgres: postgres.Config{
			URL:             viper.GetString("POSTGRES_URL"),
			MaxConns:        viper.GetInt("POSTGRES_MAX_CONNS"),
			MaxConnIdleTime: viper.GetDuration("POSTGRES_MAX_IDLE_TIME"),
			MaxConnLifetime: viper.GetDuration("POSTGRES_MAX_LIFE_TIME"),
		},
		GoogleApplicationCredentials: viper.GetString("GOOGLE_APPLICATION_CREDENTIALS"),
	}
	return c, nil
}

// LoadDefault loads config from environment variables and .env file.
func LoadDefault() (*Config, error) {
	return Load(DefaultFile)
}
