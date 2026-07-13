package usecases_test

import (
	"context"
	"fmt"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/domain/shared/vo"
)

type accountsStub struct {
	byEmail map[string]*iam.Account
	byID    map[iam.AccountID]*iam.Account

	createCalls []*iam.Account
	createErr   error

	updateCalls []*iam.Account
	updateErr   error

	getByEmailErr error
	getByIDErr    error
}

func newAccountsStub() *accountsStub {
	return &accountsStub{
		byEmail: map[string]*iam.Account{},
		byID:    map[iam.AccountID]*iam.Account{},
	}
}

func (s *accountsStub) put(a *iam.Account) *accountsStub {
	s.byEmail[a.Email().String()] = a
	s.byID[a.ID()] = a
	return s
}

func (s *accountsStub) Create(_ context.Context, a *iam.Account) error {
	s.createCalls = append(s.createCalls, a)
	if s.createErr != nil {
		return s.createErr
	}
	s.put(a)
	return nil
}
func (s *accountsStub) Update(_ context.Context, a *iam.Account) error {
	s.updateCalls = append(s.updateCalls, a)
	return s.updateErr
}
func (s *accountsStub) UpdatePicture(_ context.Context, _ *iam.Account) (string, error) {
	return "", nil
}
func (s *accountsStub) Delete(_ context.Context, _ iam.AccountID) error { return nil }
func (s *accountsStub) GetByID(_ context.Context, id iam.AccountID) (*iam.Account, error) {
	if s.getByIDErr != nil {
		return nil, s.getByIDErr
	}
	if a, ok := s.byID[id]; ok {
		return a, nil
	}
	return nil, errNotFound
}
func (s *accountsStub) GetByIDs(_ context.Context, _ []iam.AccountID) ([]*iam.Account, error) {
	return nil, nil
}
func (s *accountsStub) GetByEmail(_ context.Context, e vo.Email) (*iam.Account, error) {
	if s.getByEmailErr != nil {
		return nil, s.getByEmailErr
	}
	if a, ok := s.byEmail[e.String()]; ok {
		return a, nil
	}
	return nil, errNotFound
}
func (s *accountsStub) SearchByEmail(_ context.Context, _ string, _ int32) ([]*iam.Account, error) {
	return nil, nil
}
func (s *accountsStub) ExistsEmail(_ context.Context, _ vo.Email) (bool, error) { return false, nil }

type credentialsStub struct {
	byID map[iam.AccountID]*iam.Credentials

	saveCalls []*iam.Credentials
	saveErr   error

	getErr error
}

func newCredentialsStub() *credentialsStub {
	return &credentialsStub{byID: map[iam.AccountID]*iam.Credentials{}}
}

func (s *credentialsStub) Save(_ context.Context, c *iam.Credentials) error {
	s.saveCalls = append(s.saveCalls, c)
	if s.saveErr != nil {
		return s.saveErr
	}
	s.byID[c.AccountID()] = c
	return nil
}
func (s *credentialsStub) GetByAccountID(_ context.Context, id iam.AccountID) (*iam.Credentials, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if c, ok := s.byID[id]; ok {
		return c, nil
	}
	return nil, errNotFound
}

type rolesStub struct {
	byID   map[iam.AccountID]iam.AccountRoleLink
	getErr error
}

func newRolesStub() *rolesStub {
	return &rolesStub{byID: map[iam.AccountID]iam.AccountRoleLink{}}
}

func (s *rolesStub) put(id iam.AccountID, role iam.Role) {
	s.byID[id] = iam.AccountRoleLink{
		AccountID: id,
		Role:      role,
		EntityID:  id.ID,
	}
}

func (s *rolesStub) GetByAccountID(_ context.Context, id iam.AccountID) (iam.AccountRoleLink, error) {
	if s.getErr != nil {
		return iam.AccountRoleLink{}, s.getErr
	}
	if l, ok := s.byID[id]; ok {
		return l, nil
	}
	return iam.AccountRoleLink{}, errNotFound
}

type invitesStub struct {
	issued    []ports.InviteToken
	issueErr  error
	consumed  []string
	consumeOK ports.InviteToken
	consumeErr error
}

func (s *invitesStub) Issue(_ context.Context, tok ports.InviteToken, _ time.Duration) (string, error) {
	s.issued = append(s.issued, tok)
	if s.issueErr != nil {
		return "", s.issueErr
	}
	return "issued-token-" + string(tok.Purpose), nil
}
func (s *invitesStub) Consume(_ context.Context, token string) (ports.InviteToken, error) {
	s.consumed = append(s.consumed, token)
	if s.consumeErr != nil {
		return ports.InviteToken{}, s.consumeErr
	}
	return s.consumeOK, nil
}

type sessionsStub struct {
	createCalls []sessionCreate
	createReply string
	createErr   error

	revokedTokens   []string
	revokeErr       error
	revokedAccounts []string
	revokeAcctErr   error
}

type sessionCreate struct {
	Subject ports.SessionSubject
	TTL     time.Duration
}

func (s *sessionsStub) Create(_ context.Context, subj ports.SessionSubject, ttl time.Duration) (string, error) {
	s.createCalls = append(s.createCalls, sessionCreate{Subject: subj, TTL: ttl})
	if s.createErr != nil {
		return "", s.createErr
	}
	if s.createReply == "" {
		return "session-token", nil
	}
	return s.createReply, nil
}
func (s *sessionsStub) Lookup(_ context.Context, _ string) (ports.SessionSubject, error) {
	return ports.SessionSubject{}, nil
}
func (s *sessionsStub) Revoke(_ context.Context, token string) error {
	s.revokedTokens = append(s.revokedTokens, token)
	return s.revokeErr
}
func (s *sessionsStub) RevokeByAccountID(_ context.Context, accountID string) error {
	s.revokedAccounts = append(s.revokedAccounts, accountID)
	return s.revokeAcctErr
}

type hasherStub struct {
	hashReply  vo.PasswordHash
	hashErr    error
	verifyErr  error
}

func (h *hasherStub) Hash(_ string) (vo.PasswordHash, error) {
	if h.hashErr != nil {
		return vo.PasswordHash{}, h.hashErr
	}
	if h.hashReply.IsZero() {
		ph, _ := vo.NewPasswordHash("hashed-password")
		return ph, nil
	}
	return h.hashReply, nil
}
func (h *hasherStub) Verify(_ vo.PasswordHash, _ string) error {
	return h.verifyErr
}

type mailStub struct {
	invites []mailMsg
	logins  []mailMsg
	resets  []mailMsg

	inviteErr error
	loginErr  error
	resetErr  error
}

type mailMsg struct {
	To, Name, Link string
}

func (m *mailStub) SendSetPasswordInvite(_ context.Context, to, name, link string) error {
	m.invites = append(m.invites, mailMsg{To: to, Name: name, Link: link})
	return m.inviteErr
}
func (m *mailStub) SendMagicLogin(_ context.Context, to, name, link string) error {
	m.logins = append(m.logins, mailMsg{To: to, Name: name, Link: link})
	return m.loginErr
}
func (m *mailStub) SendPasswordReset(_ context.Context, to, name, link string) error {
	m.resets = append(m.resets, mailMsg{To: to, Name: name, Link: link})
	return m.resetErr
}

var errNotFound = fmt.Errorf("stub: %w", shared.ErrNotFound)

type loginAttemptsStub struct {
	locked    map[string]bool
	failCount map[string]int
	failCalls int
	resetCalls int
	failErr   error
	lockErr   error
}

func newLoginAttemptsStub() *loginAttemptsStub {
	return &loginAttemptsStub{
		locked:    map[string]bool{},
		failCount: map[string]int{},
	}
}

func (s *loginAttemptsStub) IsLocked(_ context.Context, accountID string) (bool, error) {
	if s.lockErr != nil {
		return false, s.lockErr
	}
	return s.locked[accountID], nil
}

func (s *loginAttemptsStub) Fail(_ context.Context, accountID string, attempts int, _ time.Duration) error {
	s.failCalls++
	if s.failErr != nil {
		return s.failErr
	}
	s.failCount[accountID]++
	if s.failCount[accountID] >= attempts {
		s.locked[accountID] = true
	}
	return nil
}

func (s *loginAttemptsStub) Reset(_ context.Context, accountID string) error {
	s.resetCalls++
	delete(s.failCount, accountID)
	delete(s.locked, accountID)
	return nil
}
