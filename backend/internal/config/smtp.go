package config

import (
	"errors"
	"fmt"
	"strings"
)

type SMTP struct {
	Enabled   bool   `env:"ENABLED" envDefault:"false"`
	Host      string `env:"HOST" envDefault:""`
	Port      int    `env:"PORT" envDefault:"465"`
	User      string `env:"USER" envDefault:"" secret:"true"`
	Password  string `env:"PASSWORD" envDefault:"" secret:"true"`
	FromEmail string `env:"FROM_EMAIL" envDefault:""`
	FromName  string `env:"FROM_NAME" envDefault:"ELS"`
	Secure    bool   `env:"SECURE" envDefault:"true"`
}

func (s SMTP) Validate() error {
	if !s.Enabled {
		return nil
	}
	var errs []error
	if strings.TrimSpace(s.Host) == "" {
		errs = append(errs, errors.New("SMTP_HOST: must not be empty when SMTP_ENABLED=true"))
	}
	if s.Port <= 0 || s.Port > 65535 {
		errs = append(errs, fmt.Errorf("SMTP_PORT: must be 1-65535, got %d", s.Port))
	}
	if strings.TrimSpace(s.FromEmail) == "" {
		errs = append(errs, errors.New("SMTP_FROM_EMAIL: must not be empty when SMTP_ENABLED=true"))
	}
	if strings.TrimSpace(s.User) == "" {
		errs = append(errs, errors.New("SMTP_USER: must not be empty when SMTP_ENABLED=true"))
	}
	if s.Password == "" {
		errs = append(errs, errors.New("SMTP_PASSWORD: must not be empty when SMTP_ENABLED=true"))
	}
	return errors.Join(errs...)
}
