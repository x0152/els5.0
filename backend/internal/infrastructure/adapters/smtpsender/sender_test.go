package smtpsender

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestParseTemplates(t *testing.T) {
	t.Parallel()
	templates, err := parseTemplates()
	if err != nil {
		t.Fatalf("parseTemplates: %v", err)
	}
	for _, name := range []string{tmplInvite, tmplMagicLogin, tmplPasswordReset} {
		if _, ok := templates[name]; !ok {
			t.Errorf("missing template %q", name)
		}
	}
}

func TestRenderInvite(t *testing.T) {
	t.Parallel()
	tmpls, err := parseTemplates()
	if err != nil {
		t.Fatalf("parseTemplates: %v", err)
	}
	s := &Sender{cfg: Config{FromEmail: "noreply@els.local", FromName: "Els"}, templates: tmpls}

	var buf strings.Builder
	if err := s.templates[tmplInvite].Execute(&buf, templateData{
		Subject:       "Invitation to ELS",
		RecipientName: "Ivan",
		ActionLabel:   "SIGN UP",
		ActionURL:     "https://platform.els.example.com/set-password?token=abc",
	}); err != nil {
		t.Fatalf("render: %v", err)
	}
	body := buf.String()
	for _, fragment := range []string{
		"ELS",
		"Ivan",
		"https://platform.els.example.com/set-password?token=abc",
		"#059669",
		"Set password",
		"SIGN UP",
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("rendered body missing %q", fragment)
		}
	}
}

func TestRenderMagicLoginActionLabel(t *testing.T) {
	t.Parallel()
	tmpls, err := parseTemplates()
	if err != nil {
		t.Fatalf("parseTemplates: %v", err)
	}

	var buf strings.Builder
	if err := tmpls[tmplMagicLogin].Execute(&buf, templateData{
		Subject:       "Sign in to ELS",
		RecipientName: "Ivan Petrov",
		ActionLabel:   "SIGN IN",
		ActionURL:     "https://platform.els.example.com/login/confirm?token=abc",
	}); err != nil {
		t.Fatalf("render: %v", err)
	}
	body := buf.String()
	for _, fragment := range []string{"Ivan Petrov", "SIGN IN", "Sign in to cabinet"} {
		if !strings.Contains(body, fragment) {
			t.Errorf("rendered body missing %q", fragment)
		}
	}
}

func TestBuildMessageEncoding(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	msg, err := buildMessage("noreply@els.local", "Els Expert", "user@example.com",
		"Hello, Ivan", []byte("<p>body</p>"), now)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	out := string(msg)
	if !strings.Contains(out, "Subject: =?UTF-8?B?") {
		t.Errorf("subject must be RFC 2047 encoded, got: %s", firstLine(out, "Subject"))
	}
	if !strings.Contains(out, "From: =?UTF-8?B?") {
		t.Errorf("from name with cyrillic must be encoded, got: %s", firstLine(out, "From"))
	}
	if !strings.Contains(out, "Content-Type: text/html; charset=UTF-8") {
		t.Errorf("missing html content-type")
	}
	if !strings.Contains(out, "Content-Transfer-Encoding: quoted-printable") {
		t.Errorf("missing quoted-printable encoding")
	}
	if !strings.Contains(out, "Date: Mon, 27 Apr 2026 12:00:00 +0000") {
		t.Errorf("missing or malformed Date header, got: %s", firstLine(out, "Date"))
	}
}

func TestBuildMessageAsciiSubjectPassthrough(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	msg, err := buildMessage("noreply@els.local", "Els", "user@example.com",
		"Reset password", []byte("<p>body</p>"), now)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	out := string(msg)
	if !strings.Contains(out, "Subject: Reset password\r\n") {
		t.Errorf("ASCII subject must pass through verbatim, got: %s", firstLine(out, "Subject"))
	}
	if !strings.Contains(out, "From: Els <noreply@els.local>\r\n") {
		t.Errorf("ASCII from must pass through verbatim, got: %s", firstLine(out, "From"))
	}
}

func TestNewValidatesConfig(t *testing.T) {
	t.Parallel()
	_, err := New(Config{})
	if err == nil {
		t.Errorf("expected error on empty host")
	}
}

func TestSendUnknownTemplate(t *testing.T) {
	t.Parallel()
	tmpls, err := parseTemplates()
	if err != nil {
		t.Fatalf("parseTemplates: %v", err)
	}
	s := &Sender{cfg: Config{FromEmail: "noreply@els.local"}, templates: tmpls}
	err = s.send(context.Background(), "user@example.com", "nope", templateData{})
	if err == nil || !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("expected unknown template error, got: %v", err)
	}
}

func TestBuildMessage_RejectsHeaderInjection(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name     string
		fromName string
		to       string
		subject  string
	}{
		{name: "to_with_crlf", to: "user@example.com\r\nBcc: leak@example.com", subject: "ok"},
		{name: "subject_with_lf", to: "user@example.com", subject: "ok\nBcc: leak@example.com"},
		{name: "from_with_cr", fromName: "Els\rEvil", to: "user@example.com", subject: "ok"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fromName := tc.fromName
			if fromName == "" {
				fromName = "Els"
			}
			_, err := buildMessage("noreply@els.local", fromName, tc.to, tc.subject, []byte("body"), now)
			if err == nil {
				t.Fatal("expected header injection to be rejected")
			}
			if !strings.Contains(err.Error(), "must not contain CR/LF") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func firstLine(out, header string) string {
	for _, line := range strings.Split(out, "\r\n") {
		if strings.HasPrefix(line, header+":") {
			return line
		}
	}
	return ""
}
