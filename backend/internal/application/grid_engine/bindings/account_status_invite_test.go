package bindings

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	authusecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

func TestSendInviteOnNoAuthToPendingPassword(t *testing.T) {
	acc := iamtest.NewAccount(t).
		WithEmail("seed@example.com").
		WithStatus(iam.AccountStatusPendingPassword).
		Build(t)
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)
	invites := &bindingInviteStore{}
	mail := &bindingMailSender{}
	uc := authusecases.NewInviteAccountUseCase(authusecases.InviteAccountDeps{
		Invites:  invites,
		Mail:     mail,
		TTL:      time.Hour,
		LinkTmpl: "https://app/set-password?token={token}",
	})

	err = sendInviteOnNoAuthToPending(context.Background(), uc, statusRow(iam.AccountStatusNoAuth), statusRow(iam.AccountStatusPendingPassword), side)

	test.NoErr(t, err)
	if len(invites.issued) != 1 || invites.issued[0].Purpose != ports.InviteTokenSetPassword || invites.issued[0].AccountID != acc.ID().String() {
		t.Fatalf("expected one set-password invite for account, got %+v", invites.issued)
	}
	if len(mail.invites) != 1 || mail.invites[0].to != "seed@example.com" || !strings.Contains(mail.invites[0].link, "issued-token") {
		t.Fatalf("expected one invite mail, got %+v", mail.invites)
	}
}

func TestSendInviteOnNoAuthToPendingPassword_NoopsForOtherTransitions(t *testing.T) {
	acc := iamtest.NewAccount(t).WithStatus(iam.AccountStatusActive).Build(t)
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)
	invites := &bindingInviteStore{}
	mail := &bindingMailSender{}
	uc := authusecases.NewInviteAccountUseCase(authusecases.InviteAccountDeps{Invites: invites, Mail: mail})

	err = sendInviteOnNoAuthToPending(context.Background(), uc, statusRow(iam.AccountStatusActive), statusRow(iam.AccountStatusBlocked), side)

	test.NoErr(t, err)
	if len(invites.issued) != 0 || len(mail.invites) != 0 {
		t.Fatalf("expected no invite side effects, got invites=%+v mail=%+v", invites.issued, mail.invites)
	}
}

func TestSendInviteOnNoAuthToPendingPassword_PropagatesInviteFailure(t *testing.T) {
	acc := iamtest.NewAccount(t).
		WithStatus(iam.AccountStatusPendingPassword).
		Build(t)
	side, err := iam.NewAccountSide(acc)
	test.NoErr(t, err)
	boom := errors.New("token store")
	uc := authusecases.NewInviteAccountUseCase(authusecases.InviteAccountDeps{
		Invites: &bindingInviteStore{issueErr: boom},
		Mail:    &bindingMailSender{},
	})

	err = sendInviteOnNoAuthToPending(context.Background(), uc, statusRow(iam.AccountStatusNoAuth), statusRow(iam.AccountStatusPendingPassword), side)

	if !errors.Is(err, boom) {
		t.Fatalf("expected invite failure, got %v", err)
	}
}

func statusRow(status iam.AccountStatus) grid.Row {
	return grid.Row{
		Cells: map[grid.ColumnID]any{
			iam.ColAccountStatus: status.String(),
		},
	}
}

type bindingInviteStore struct {
	issued   []ports.InviteToken
	issueErr error
}

func (s *bindingInviteStore) Issue(_ context.Context, tok ports.InviteToken, _ time.Duration) (string, error) {
	if s.issueErr != nil {
		return "", s.issueErr
	}
	s.issued = append(s.issued, tok)
	return "issued-token", nil
}

func (s *bindingInviteStore) Consume(_ context.Context, _ string) (ports.InviteToken, error) {
	return ports.InviteToken{}, nil
}

type bindingMailSender struct {
	invites []struct {
		to   string
		link string
	}
}

func (s *bindingMailSender) SendSetPasswordInvite(_ context.Context, to, _ string, link string) error {
	s.invites = append(s.invites, struct {
		to   string
		link string
	}{to: to, link: link})
	return nil
}

func (s *bindingMailSender) SendMagicLogin(context.Context, string, string, string) error {
	return nil
}

func (s *bindingMailSender) SendPasswordReset(context.Context, string, string, string) error {
	return nil
}
