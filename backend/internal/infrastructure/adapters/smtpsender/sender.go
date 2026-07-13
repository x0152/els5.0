package smtpsender

import (
	"bytes"
	"context"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared/ports"
)

var _ ports.MailSender = (*Sender)(nil)

//go:embed templates/*.html
var templatesFS embed.FS

const (
	tmplInvite        = "invite"
	tmplMagicLogin    = "magic_login"
	tmplPasswordReset = "password_reset"

	defaultTimeout = 30 * time.Second
)

type Config struct {
	Host      string
	Port      int
	User      string
	Password  string
	FromEmail string
	FromName  string
	Secure    bool
	Timeout   time.Duration
}

type Sender struct {
	cfg       Config
	templates map[string]*template.Template
}

func New(cfg Config) (*Sender, error) {
	if cfg.Host == "" {
		return nil, errors.New("smtpsender: host must not be empty")
	}
	if cfg.Port <= 0 {
		return nil, errors.New("smtpsender: port must be > 0")
	}
	if cfg.FromEmail == "" {
		return nil, errors.New("smtpsender: from email must not be empty")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}

	templates, err := parseTemplates()
	if err != nil {
		return nil, fmt.Errorf("smtpsender: parse templates: %w", err)
	}
	return &Sender{cfg: cfg, templates: templates}, nil
}

func parseTemplates() (map[string]*template.Template, error) {
	base, err := templatesFS.ReadFile("templates/base.html")
	if err != nil {
		return nil, fmt.Errorf("read base: %w", err)
	}

	out := make(map[string]*template.Template, 3)
	for _, name := range []string{tmplInvite, tmplMagicLogin, tmplPasswordReset} {
		body, rerr := templatesFS.ReadFile("templates/" + name + ".html")
		if rerr != nil {
			return nil, fmt.Errorf("read %s: %w", name, rerr)
		}
		t, perr := template.New(name).Parse(string(base))
		if perr != nil {
			return nil, fmt.Errorf("parse base for %s: %w", name, perr)
		}
		if _, perr = t.Parse(string(body)); perr != nil {
			return nil, fmt.Errorf("parse %s: %w", name, perr)
		}
		out[name] = t
	}
	return out, nil
}

type templateData struct {
	Subject       string
	RecipientName string
	ActionLabel   string
	ActionURL     string
}

func (s *Sender) SendSetPasswordInvite(ctx context.Context, to, recipientName, link string) error {
	return s.send(ctx, to, tmplInvite, templateData{
		Subject:       "Invitation to ELS",
		RecipientName: recipientName,
		ActionLabel:   "SIGN UP",
		ActionURL:     link,
	})
}

func (s *Sender) SendMagicLogin(ctx context.Context, to, recipientName, link string) error {
	return s.send(ctx, to, tmplMagicLogin, templateData{
		Subject:       "Sign in to ELS",
		RecipientName: recipientName,
		ActionLabel:   "SIGN IN",
		ActionURL:     link,
	})
}

func (s *Sender) SendPasswordReset(ctx context.Context, to, recipientName, link string) error {
	return s.send(ctx, to, tmplPasswordReset, templateData{
		Subject:       "ELS password reset",
		RecipientName: recipientName,
		ActionURL:     link,
	})
}

func (s *Sender) send(ctx context.Context, to, tmplName string, data templateData) error {
	if to == "" {
		return errors.New("smtpsender: empty recipient")
	}
	tmpl, ok := s.templates[tmplName]
	if !ok {
		return fmt.Errorf("smtpsender: unknown template %q", tmplName)
	}

	var html bytes.Buffer
	if err := tmpl.Execute(&html, data); err != nil {
		return fmt.Errorf("smtpsender: render %s: %w", tmplName, err)
	}

	msg, err := buildMessage(s.cfg.FromEmail, s.cfg.FromName, to, data.Subject, html.Bytes(), time.Now())
	if err != nil {
		return fmt.Errorf("smtpsender: build message: %w", err)
	}

	timeout := s.cfg.Timeout
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d > 0 && d < timeout {
			timeout = d
		}
	}

	if err := s.deliver(ctx, to, msg, timeout); err != nil {
		return fmt.Errorf("smtpsender: deliver to %s: %w", to, err)
	}
	return nil
}

func (s *Sender) deliver(ctx context.Context, to string, msg []byte, timeout time.Duration) error {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	dialer := &net.Dialer{Timeout: timeout}

	var (
		conn net.Conn
		err  error
	)
	if s.cfg.Secure {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	_ = conn.SetDeadline(time.Now().Add(timeout))

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = c.Quit() }()

	if !s.cfg.Secure {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
				return fmt.Errorf("starttls: %w", err)
			}
		}
	}

	if s.cfg.User != "" {
		if err := c.Auth(smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)); err != nil {
			return fmt.Errorf("auth: %w", err)
		}
	}

	if err := c.Mail(s.cfg.FromEmail); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return fmt.Errorf("write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}
	return nil
}

func buildMessage(fromEmail, fromName, to, subject string, htmlBody []byte, now time.Time) ([]byte, error) {
	if err := assertSafeHeader("from_email", fromEmail); err != nil {
		return nil, err
	}
	if err := assertSafeHeader("from_name", fromName); err != nil {
		return nil, err
	}
	if err := assertSafeHeader("to", to); err != nil {
		return nil, err
	}
	if err := assertSafeHeader("subject", subject); err != nil {
		return nil, err
	}

	from := fromEmail
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", encodeHeader(fromName), fromEmail)
	}

	var buf bytes.Buffer
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + to + "\r\n")
	buf.WriteString("Subject: " + encodeHeader(subject) + "\r\n")
	buf.WriteString("Date: " + now.UTC().Format(time.RFC1123Z) + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buf.WriteString("\r\n")

	w := quotedprintable.NewWriter(&buf)
	if _, err := w.Write(htmlBody); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func assertSafeHeader(field, v string) error {
	if strings.ContainsAny(v, "\r\n") {
		return fmt.Errorf("smtpsender: %s must not contain CR/LF", field)
	}
	return nil
}

func encodeHeader(v string) string {
	for _, r := range v {
		if r > 127 {
			return mimeBEncode(v)
		}
	}
	return v
}

func mimeBEncode(v string) string {
	const prefix, suffix = "=?UTF-8?B?", "?="
	var b strings.Builder
	b.Grow(len(v) + len(prefix) + len(suffix) + 8)
	b.WriteString(prefix)
	b.WriteString(base64.StdEncoding.EncodeToString([]byte(v)))
	b.WriteString(suffix)
	return b.String()
}
