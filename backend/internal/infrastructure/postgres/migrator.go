package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const (
	migrationsDir = "migrations"

	// migrationAdvisoryLockKey is a stable cluster-wide key used by Migrate to
	// serialize concurrent migrators (CI, replicas, manual runs) under
	// pg_advisory_lock. Keep it stable across releases.
	migrationAdvisoryLockKey int64 = 0x73617069656e735F // "els_"

	advisoryLockTimeout = 2 * time.Minute
)

type MigrateDirection int

const (
	MigrateUp MigrateDirection = iota
	MigrateDown
)

type MigrateResult struct {
	Version uint
	Dirty   bool
}

func Migrate(dsn string, dir MigrateDirection) (MigrateResult, error) {
	mDSN := toMigrateDSN(dsn)
	connDSN := toConnDSN(dsn)

	ctx, cancel := context.WithTimeout(context.Background(), advisoryLockTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, connDSN)
	if err != nil {
		return MigrateResult{}, fmt.Errorf("advisory lock: connect: %w", err)
	}
	defer func() {
		_ = conn.Close(context.Background())
	}()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockKey); err != nil {
		return MigrateResult{}, fmt.Errorf("advisory lock: acquire: %w", err)
	}
	defer func() {
		relCtx, relCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer relCancel()
		_, _ = conn.Exec(relCtx, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockKey)
	}()

	src, err := iofs.New(migrationsFS, migrationsDir)
	if err != nil {
		return MigrateResult{}, fmt.Errorf("iofs source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, mDSN)
	if err != nil {
		return MigrateResult{}, fmt.Errorf("new migrate: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	var runErr error
	switch dir {
	case MigrateUp:
		runErr = m.Up()
	case MigrateDown:
		runErr = m.Steps(-1)
	default:
		return MigrateResult{}, fmt.Errorf("unknown direction: %d", dir)
	}
	if runErr != nil && !errors.Is(runErr, migrate.ErrNoChange) {
		return MigrateResult{}, fmt.Errorf("apply migrations: %w", runErr)
	}

	version, dirty, verr := m.Version()
	if verr != nil {
		if errors.Is(verr, migrate.ErrNilVersion) {
			return MigrateResult{Version: 0, Dirty: false}, nil
		}
		return MigrateResult{}, fmt.Errorf("read version: %w", verr)
	}
	return MigrateResult{Version: version, Dirty: dirty}, nil
}

func toConnDSN(dsn string) string {
	switch {
	case strings.HasPrefix(dsn, "pgx5://"):
		return "postgres://" + strings.TrimPrefix(dsn, "pgx5://")
	default:
		return dsn
	}
}

func toMigrateDSN(dsn string) string {
	switch {
	case strings.HasPrefix(dsn, "pgx5://"):
		return dsn
	case strings.HasPrefix(dsn, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgresql://")
	case strings.HasPrefix(dsn, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgres://")
	default:
		return dsn
	}
}
