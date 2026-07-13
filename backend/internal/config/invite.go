package config

import (
	"errors"
	"strings"
	"time"
)

type Invite struct {
	KeyPrefix            string        `env:"KEY_PREFIX" envDefault:"invite:"`
	SetPasswordTTL       time.Duration `env:"SET_PASSWORD_TTL" envDefault:"24h"`
	MagicLoginTTL        time.Duration `env:"MAGIC_LOGIN_TTL" envDefault:"15m"`
	ResetPasswordTTL     time.Duration `env:"RESET_PASSWORD_TTL" envDefault:"1h"`
	MagicLoginPersistent bool          `env:"MAGIC_LOGIN_PERSISTENT" envDefault:"false"`
	SetPasswordURL       string        `env:"SET_PASSWORD_URL" envDefault:"http://localhost:5173/set-password?token={token}"`
	MagicLoginURL        string        `env:"MAGIC_LOGIN_URL" envDefault:"http://localhost:5173/auth/confirm?token={token}"`
	ResetPasswordURL     string        `env:"RESET_PASSWORD_URL" envDefault:"http://localhost:5173/reset-password?token={token}"`
}

func (i Invite) Validate() error {
	var errs []error
	if i.SetPasswordTTL <= 0 {
		errs = append(errs, errors.New("INVITE_SET_PASSWORD_TTL: must be > 0"))
	}
	if !i.MagicLoginPersistent && i.MagicLoginTTL <= 0 {
		errs = append(errs, errors.New("INVITE_MAGIC_LOGIN_TTL: must be > 0"))
	}
	if i.ResetPasswordTTL <= 0 {
		errs = append(errs, errors.New("INVITE_RESET_PASSWORD_TTL: must be > 0"))
	}
	if !strings.Contains(i.SetPasswordURL, "{token}") {
		errs = append(errs, errors.New("INVITE_SET_PASSWORD_URL: must contain {token} placeholder"))
	}
	if !strings.Contains(i.MagicLoginURL, "{token}") {
		errs = append(errs, errors.New("INVITE_MAGIC_LOGIN_URL: must contain {token} placeholder"))
	}
	if !strings.Contains(i.ResetPasswordURL, "{token}") {
		errs = append(errs, errors.New("INVITE_RESET_PASSWORD_URL: must contain {token} placeholder"))
	}
	return errors.Join(errs...)
}
