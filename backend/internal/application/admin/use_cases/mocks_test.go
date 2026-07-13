package usecases_test

import (
	"context"

	authusecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
)

type adminRepoStub struct {
	createCalls []*admin.Administrator
	createErr   error

	getByID    func(ctx context.Context, id admin.ID) (*admin.Administrator, error)
	getByIDErr error

	listFilter admin.Filter
	listLimit  int32
	listOffset int32
	listReply  []*admin.Administrator
	listTotal  int64
	listErr    error
	listCalls  int
}

func (s *adminRepoStub) Create(_ context.Context, a *admin.Administrator) error {
	s.createCalls = append(s.createCalls, a)
	return s.createErr
}

func (s *adminRepoStub) Update(_ context.Context, _ *admin.Administrator) error { return nil }

func (s *adminRepoStub) Delete(_ context.Context, _ admin.ID) error { return nil }

func (s *adminRepoStub) GetByID(ctx context.Context, id admin.ID) (*admin.Administrator, error) {
	if s.getByID != nil {
		return s.getByID(ctx, id)
	}
	return nil, s.getByIDErr
}

func (s *adminRepoStub) GetByAccountID(_ context.Context, _ iam.AccountID) (*admin.Administrator, error) {
	return nil, nil
}

func (s *adminRepoStub) List(_ context.Context, f admin.Filter, limit, offset int32) ([]*admin.Administrator, int64, error) {
	s.listCalls++
	s.listFilter = f
	s.listLimit = limit
	s.listOffset = offset
	if s.listErr != nil {
		return nil, 0, s.listErr
	}
	return s.listReply, s.listTotal, nil
}

func (s *adminRepoStub) Count(_ context.Context) (int64, error) { return 0, nil }

type accountInviterStub struct {
	calls []authusecases.InviteAccountCommand
	reply *iam.Account
	err   error
}

func (s *accountInviterStub) Execute(_ context.Context, cmd authusecases.InviteAccountCommand) (*iam.Account, error) {
	s.calls = append(s.calls, cmd)
	return s.reply, s.err
}
