package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/config"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/passwordhash"
	"github.com/els/backend/internal/infrastructure/adapters/rediscode"
	"github.com/els/backend/internal/infrastructure/postgres"
	adminrepo "github.com/els/backend/internal/infrastructure/repositories/admin"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	cfgutil "github.com/els/backend/internal/utils/config"
	"github.com/els/backend/internal/utils/logger"
	"github.com/els/backend/internal/utils/redisx"
)

type Config struct {
	App      config.App      `envPrefix:"APP_"`
	Logger   config.Logger   `envPrefix:"LOG_"`
	Postgres config.Postgres `envPrefix:"POSTGRES_"`
	Redis    config.Redis    `envPrefix:"REDIS_"`
	Invite   config.Invite   `envPrefix:"INVITE_"`
	Admin    AdminConfig     `envPrefix:"BOOTSTRAP_ADMIN_"`
	Password PasswordConfig  `envPrefix:"AUTH_PASSWORD_"`
}

type AdminConfig struct {
	Email      string `env:"EMAIL,required"`
	Password   string `env:"PASSWORD" secret:"true"`
	FirstName  string `env:"FIRST_NAME,required"`
	LastName   string `env:"LAST_NAME,required"`
	UseInvite  bool   `env:"USE_INVITE" envDefault:"false"`
}

type PasswordConfig struct {
	MinLength int `env:"MIN_LENGTH" envDefault:"8"`
	MaxLength int `env:"MAX_LENGTH" envDefault:"128"`

	Argon2Memory  uint32 `env:"ARGON2_MEMORY_KIB" envDefault:"65536"`
	Argon2Time    uint32 `env:"ARGON2_TIME" envDefault:"3"`
	Argon2Threads uint8  `env:"ARGON2_THREADS" envDefault:"2"`
	Argon2SaltLen uint32 `env:"ARGON2_SALT_LEN" envDefault:"16"`
	Argon2KeyLen  uint32 `env:"ARGON2_KEY_LEN" envDefault:"32"`
}

func (c Config) Validate() error {
	errs := []error{
		c.App.Validate(),
		c.Logger.Validate(),
		c.Postgres.Validate(),
		c.Admin.Validate(),
		c.Password.Validate(),
	}
	if c.Admin.UseInvite {
		errs = append(errs, c.Redis.Validate(), c.Invite.Validate())
	}
	if c.App.Env == "prod" && !c.Admin.UseInvite {
		errs = append(errs, errors.New("BOOTSTRAP_ADMIN_USE_INVITE: must be true in prod (no static bootstrap password)"))
	}
	return errors.Join(errs...)
}

func (a AdminConfig) Validate() error {
	var errs []error
	if a.Email == "" {
		errs = append(errs, errors.New("BOOTSTRAP_ADMIN_EMAIL: must not be empty"))
	}
	if !a.UseInvite && a.Password == "" {
		errs = append(errs, errors.New("BOOTSTRAP_ADMIN_PASSWORD: must not be empty unless BOOTSTRAP_ADMIN_USE_INVITE=true"))
	}
	if a.FirstName == "" {
		errs = append(errs, errors.New("BOOTSTRAP_ADMIN_FIRST_NAME: must not be empty"))
	}
	if a.LastName == "" {
		errs = append(errs, errors.New("BOOTSTRAP_ADMIN_LAST_NAME: must not be empty"))
	}
	return errors.Join(errs...)
}

func (p PasswordConfig) Validate() error {
	var errs []error
	if p.MinLength < 1 {
		errs = append(errs, errors.New("AUTH_PASSWORD_MIN_LENGTH: must be >= 1"))
	}
	if p.MaxLength < p.MinLength {
		errs = append(errs, errors.New("AUTH_PASSWORD_MAX_LENGTH: must be >= AUTH_PASSWORD_MIN_LENGTH"))
	}
	if p.MaxLength > 1024 {
		errs = append(errs, errors.New("AUTH_PASSWORD_MAX_LENGTH: must be <= 1024"))
	}
	if p.Argon2Memory < 8*1024 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_MEMORY_KIB: must be >= 8192"))
	}
	if p.Argon2Time < 1 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_TIME: must be >= 1"))
	}
	if p.Argon2Threads < 1 {
		errs = append(errs, errors.New("AUTH_PASSWORD_ARGON2_THREADS: must be >= 1"))
	}
	return errors.Join(errs...)
}

func main() {
	var cfg Config
	cfgutil.MustLoad("bootstrap", &cfg)

	log := logger.New(logger.Config{
		Level:     cfg.Logger.Level,
		Format:    cfg.Logger.Format,
		Module:    "bootstrap",
		AddSource: cfg.Logger.AddSource,
	})
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, postgres.Config{
		DSN:             cfg.Postgres.DSN().Raw(),
		MaxConns:        int32(cfg.Postgres.MaxOpenConns),
		MinConns:        int32(cfg.Postgres.MaxIdleConns),
		ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
		ConnectTimeout:  cfg.Postgres.ConnectTimeout,
	})
	if err != nil {
		log.Error("postgres connect", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if err := run(ctx, log, pool, cfg); err != nil {
		log.Error("bootstrap failed", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger, pool *pgxpool.Pool, cfg Config) error {
	accounts := iamrepo.NewAccountRepo(pool)
	credentials := iamrepo.NewCredentialsRepo(pool)
	admins := adminrepo.NewAdministratorRepo(pool)
	hasher := passwordhash.NewArgon2id(passwordhash.Argon2idParams{
		Memory:  cfg.Password.Argon2Memory,
		Time:    cfg.Password.Argon2Time,
		Threads: cfg.Password.Argon2Threads,
		SaltLen: cfg.Password.Argon2SaltLen,
		KeyLen:  cfg.Password.Argon2KeyLen,
	})

	total, err := admins.Count(ctx)
	if err != nil {
		return err
	}
	if total > 0 {
		log.Info("administrators already exist, skipping bootstrap", slog.Int64("total", total))
		return nil
	}

	policy := iam.PasswordPolicy{MinLength: cfg.Password.MinLength, MaxLength: cfg.Password.MaxLength}
	if !cfg.Admin.UseInvite {
		if err := policy.Validate(cfg.Admin.Password); err != nil {
			return err
		}
	}

	acc, err := iam.NewPendingAccountNow(iam.NewAccountID(), cfg.Admin.Email, cfg.Admin.FirstName, cfg.Admin.LastName)
	if err != nil {
		return err
	}
	if !cfg.Admin.UseInvite {
		if err := acc.Activate(); err != nil {
			return err
		}
	}
	if err := accounts.Create(ctx, acc); err != nil {
		if !errors.Is(err, shared.ErrConflict) {
			return err
		}
		log.Info("account already exists, reusing", slog.String("email", cfg.Admin.Email))
		existing, loadErr := accounts.GetByEmail(ctx, acc.Email())
		if loadErr != nil {
			return loadErr
		}
		acc = existing
	}

	if !cfg.Admin.UseInvite {
		hash, err := hasher.Hash(cfg.Admin.Password)
		if err != nil {
			return err
		}
		cred, err := iam.NewCredentials(acc.ID(), hash)
		if err != nil {
			return err
		}
		if err := credentials.Save(ctx, cred); err != nil {
			return err
		}
	}

	a, err := admin.NewAdministratorNow(admin.NewID(), acc)
	if err != nil {
		return err
	}
	if err := admins.Create(ctx, a); err != nil {
		return err
	}

	if cfg.Admin.UseInvite {
		invites, closeFn, err := newInviteStore(ctx, cfg)
		if err != nil {
			return err
		}
		defer closeFn()
		token, err := invites.Issue(ctx, ports.InviteToken{
			Purpose:   ports.InviteTokenSetPassword,
			AccountID: acc.ID().String(),
		}, cfg.Invite.SetPasswordTTL)
		if err != nil {
			return fmt.Errorf("issue invite: %w", err)
		}
		log.Info("bootstrap complete (invite mode)",
			slog.String("admin_id", a.ID().String()),
			slog.String("account_id", acc.ID().String()),
			slog.String("email", cfg.Admin.Email),
			slog.String("set_password_url", strings.ReplaceAll(cfg.Invite.SetPasswordURL, "{token}", token)),
			slog.Duration("ttl", cfg.Invite.SetPasswordTTL),
		)
		return nil
	}

	log.Info("bootstrap complete",
		slog.String("admin_id", a.ID().String()),
		slog.String("account_id", acc.ID().String()),
		slog.String("email", cfg.Admin.Email),
	)
	return nil
}

func newInviteStore(ctx context.Context, cfg Config) (ports.InviteTokenStore, func(), error) {
	rdb, err := redisx.New(ctx, redisx.Config{
		Addr:           cfg.Redis.Addr,
		Password:       cfg.Redis.Password,
		DB:             cfg.Redis.DB,
		ConnectTimeout: cfg.Redis.ConnectTimeout,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("redis connect: %w", err)
	}
	return rediscode.NewStore(rdb, cfg.Invite.KeyPrefix), func() { _ = rdb.Close() }, nil
}
