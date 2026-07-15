package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/els/backend/internal/application/auth/api"
	usecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/passwordhash"
	"github.com/els/backend/internal/infrastructure/adapters/rediscode"
	"github.com/els/backend/internal/infrastructure/adapters/redislockout"
	"github.com/els/backend/internal/infrastructure/adapters/redisratelimit"
	"github.com/els/backend/internal/infrastructure/adapters/redissession"
	iamrepo "github.com/els/backend/internal/infrastructure/repositories/iam"
	authx "github.com/els/backend/internal/utils/auth"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/openapi"
	"github.com/els/backend/internal/utils/probes"
	"github.com/els/backend/internal/utils/timex"
)

const (
	Name    = "auth"
	Version = "0.1.0"
)

func init() {
	openapi.Register(openapi.Module{
		Name:    Name,
		Version: Version,
		Register: func(a huma.API) {
			api.Register(a, api.Deps{})
		},
	})
}

func Mount(humaAPI huma.API, cfg Config, pool *pgxpool.Pool, rdb *redis.Client, mail ports.MailSender) {
	accounts := iamrepo.NewAccountRepo(pool)
	credentials := iamrepo.NewCredentialsRepo(pool)
	roles := iamrepo.NewAccountRoleRepo(pool)
	hasher := passwordhash.NewArgon2id(passwordhash.Argon2idParams{
		Memory:  cfg.Password.Argon2Memory,
		Time:    cfg.Password.Argon2Time,
		Threads: cfg.Password.Argon2Threads,
		SaltLen: cfg.Password.Argon2SaltLen,
		KeyLen:  cfg.Password.Argon2KeyLen,
	})
	sessions := redissession.NewStore(rdb, cfg.Session.KeyPrefix)
	invites := rediscode.NewStore(rdb, cfg.Invite.KeyPrefix)
	attempts := redislockout.NewStore(rdb, "")
	policy := iam.PasswordPolicy{MinLength: cfg.Password.MinLength, MaxLength: cfg.Password.MaxLength}
	clock := timex.System()

	authn := authx.New(sessions, accounts)
	loginStartUC := usecases.NewLoginStartUseCase(usecases.LoginStartDeps{
		Accounts:        accounts,
		Credentials:     credentials,
		Roles:           roles,
		Hasher:          hasher,
		Sessions:        sessions,
		Attempts:        attempts,
		LockoutAttempts: cfg.Security.AuthLockoutAttempts,
		LockoutWindow:   cfg.Security.AuthLockoutWindow,
		SessionTTL:      cfg.Session.TTL,
		Clock:           clock,
	})
	loginConfirmUC := usecases.NewLoginConfirmUseCase(accounts, roles, invites, sessions, cfg.Session.TTL, clock)
	setPasswordUC := usecases.NewSetPasswordUseCase(usecases.SetPasswordDeps{
		Accounts:    accounts,
		Credentials: credentials,
		Hasher:      hasher,
		Invites:     invites,
		Sessions:    sessions,
		Policy:      policy,
	})
	resendInviteUC := usecases.NewResendInviteUseCase(usecases.ResendInviteDeps{
		Accounts: accounts,
		Invites:  invites,
		Mail:     mail,
		TTL:      cfg.Invite.SetPasswordTTL,
		LinkTmpl: cfg.Invite.SetPasswordURL,
	})
	forgotPasswordUC := usecases.NewForgotPasswordUseCase(usecases.ForgotPasswordDeps{
		Accounts: accounts,
		Invites:  invites,
		Mail:     mail,
		TTL:      cfg.Invite.ResetPasswordTTL,
		LinkTmpl: cfg.Invite.ResetPasswordURL,
	})
	resetPasswordUC := usecases.NewResetPasswordUseCase(usecases.ResetPasswordDeps{
		Accounts:    accounts,
		Credentials: credentials,
		Hasher:      hasher,
		Invites:     invites,
		Sessions:    sessions,
		Policy:      policy,
	})
	logoutUC := usecases.NewLogoutUseCase(sessions)
	impersonateUC := usecases.NewImpersonateUseCase(cfg.Impersonation.Enabled, accounts, roles, sessions, cfg.Session.TTL, clock)

	api.Register(humaAPI, api.Deps{
		Authenticator:  authn,
		LoginStart:     loginStartUC,
		LoginConfirm:   loginConfirmUC,
		SetPassword:    setPasswordUC,
		ResendInvite:   resendInviteUC,
		ForgotPassword: forgotPasswordUC,
		ResetPassword:  resetPasswordUC,
		Logout:         logoutUC,
		Impersonate:    impersonateUC,
	})
}

type App struct {
	cfg    Config
	logger *slog.Logger
	srv    *http.Server
}

func New(cfg Config, pool *pgxpool.Pool, rdb *redis.Client, logger *slog.Logger, mail ports.MailSender) *App {
	mux := http.NewServeMux()
	humaAPI := httpx.NewAPI(mux, Name, Version, httpx.APIOptions(cfg.Security)...)

	probes.Register(humaAPI, probes.Deps{
		Module:  Name,
		Version: Version,
		Ready: []probes.NamedCheck{
			{Name: "postgres", Check: pool.Ping},
			{Name: "redis", Check: func(ctx context.Context) error { return rdb.Ping(ctx).Err() }},
		},
	})

	Mount(humaAPI, cfg, pool, rdb, mail)

	handler := httpx.DefaultChain(mux, cfg.Security, httpx.ChainOptions{
		Logger:    logger,
		Limiter:   redisratelimit.New(rdb, "ratelimit:auth:"),
		AuthPaths: []string{"/api/v1/auth/"},
	})

	return &App{
		cfg:    cfg,
		logger: logger,
		srv: &http.Server{
			Addr:         cfg.HTTP.Addr,
			Handler:      handler,
			ReadTimeout:  cfg.HTTP.ReadTimeout,
			WriteTimeout: cfg.HTTP.WriteTimeout,
			IdleTimeout:  cfg.HTTP.IdleTimeout,
		},
	}
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("http listening", slog.String("addr", a.cfg.HTTP.Addr))
		if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("http shutting down",
			slog.Duration("timeout", a.cfg.HTTP.ShutdownTimeout),
		)
		shCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()
		return a.srv.Shutdown(shCtx)
	case err := <-errCh:
		return err
	}
}
