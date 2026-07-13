package logsender

import (
	"context"
	"fmt"
	"os"
)

type Sender struct {
	out *os.File
}

func New() *Sender {
	return &Sender{out: os.Stdout}
}

func (s *Sender) SendSetPasswordInvite(_ context.Context, to, recipientName, link string) error {
	_, err := fmt.Fprintf(s.out,
		"\n=== MAIL: set password invite ===\nto: %s\nname: %s\nlink: %s\n=================================\n\n",
		to, recipientName, link,
	)
	return err
}

func (s *Sender) SendMagicLogin(_ context.Context, to, recipientName, link string) error {
	_, err := fmt.Fprintf(s.out,
		"\n=== MAIL: magic login ===\nto: %s\nname: %s\nlink: %s\n=========================\n\n",
		to, recipientName, link,
	)
	return err
}

func (s *Sender) SendPasswordReset(_ context.Context, to, recipientName, link string) error {
	_, err := fmt.Fprintf(s.out,
		"\n=== MAIL: password reset ===\nto: %s\nname: %s\nlink: %s\n============================\n\n",
		to, recipientName, link,
	)
	return err
}
