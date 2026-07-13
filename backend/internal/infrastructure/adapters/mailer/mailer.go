package mailer

import (
	"fmt"
	"log/slog"

	"github.com/els/backend/internal/config"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/infrastructure/adapters/logsender"
	"github.com/els/backend/internal/infrastructure/adapters/smtpsender"
)

func New(cfg config.SMTP, log *slog.Logger) ports.MailSender {
	if !cfg.Enabled {
		log.Info("mail: stdout sender (SMTP disabled)")
		return logsender.New()
	}
	sender, err := smtpsender.New(smtpsender.Config{
		Host:      cfg.Host,
		Port:      cfg.Port,
		User:      cfg.User,
		Password:  cfg.Password,
		FromEmail: cfg.FromEmail,
		FromName:  cfg.FromName,
		Secure:    cfg.Secure,
	})
	if err != nil {
		panic(fmt.Errorf("mail: smtp init failed: %w", err))
	}
	log.Info("mail: SMTP sender",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.Bool("secure", cfg.Secure),
		slog.String("from", cfg.FromEmail),
	)
	return sender
}
