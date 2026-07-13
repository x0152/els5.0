package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/els/backend/internal/config"
	"github.com/els/backend/internal/infrastructure/postgres"
	cfgutil "github.com/els/backend/internal/utils/config"
	"github.com/els/backend/internal/utils/logger"
)

type Config struct {
	App      config.App      `envPrefix:"APP_"`
	Logger   config.Logger   `envPrefix:"LOG_"`
	Postgres config.Postgres `envPrefix:"POSTGRES_"`
}

func (c Config) Validate() error {
	return errors.Join(
		c.App.Validate(),
		c.Logger.Validate(),
		c.Postgres.Validate(),
	)
}

func main() {
	direction := flag.String("direction", "up", "migration direction: up | down")
	flag.Parse()

	var cfg Config
	cfgutil.MustLoad("migrator", &cfg)

	log := logger.New(logger.Config{
		Level:     cfg.Logger.Level,
		Format:    cfg.Logger.Format,
		Module:    "migrator",
		AddSource: cfg.Logger.AddSource,
	})
	slog.SetDefault(log)

	var dir postgres.MigrateDirection
	switch *direction {
	case "up":
		dir = postgres.MigrateUp
	case "down":
		dir = postgres.MigrateDown
	default:
		fmt.Fprintf(os.Stderr, "unknown direction %q (expected up|down)\n", *direction)
		os.Exit(2)
	}

	log.Info("migrate start",
		slog.String("direction", *direction),
		slog.String("host", cfg.Postgres.Host),
		slog.Int("port", cfg.Postgres.Port),
		slog.String("database", cfg.Postgres.Database),
	)

	res, err := postgres.Migrate(cfg.Postgres.DSN().Raw(), dir)
	if err != nil {
		log.Error("migrate failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("migrate done",
		slog.Uint64("version", uint64(res.Version)),
		slog.Bool("dirty", res.Dirty),
	)
}
