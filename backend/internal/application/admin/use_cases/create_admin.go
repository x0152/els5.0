package usecases

import (
	"context"

	authusecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/iam"
)

type AccountInviter interface {
	Execute(ctx context.Context, cmd authusecases.InviteAccountCommand) (*iam.Account, error)
}

type CreateAdminUseCase struct {
	admins  admin.Repository
	invites AccountInviter
}

func NewCreateAdminUseCase(admins admin.Repository, invites AccountInviter) *CreateAdminUseCase {
	return &CreateAdminUseCase{admins: admins, invites: invites}
}

type CreateAdminCommand struct {
	Email     string
	FirstName string
	LastName  string
}

type CreateAdminResult struct {
	Admin *admin.Administrator
}

func (uc *CreateAdminUseCase) Execute(
	ctx context.Context,
	actor *iam.Actor,
	cmd CreateAdminCommand,
) (CreateAdminResult, error) {
	// 1. Only a global admin can create other admins.
	if err := iam.RequireGlobalAdmin(actor); err != nil {
		return CreateAdminResult{}, err
	}

	// 2. Create the account and send an invite to set a password.
	acc, err := uc.invites.Execute(ctx, authusecases.InviteAccountCommand{
		Email:     cmd.Email,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
	})
	if err != nil {
		return CreateAdminResult{}, err
	}

	// 3. Create the domain administrator (no groups — global by default).
	a, err := admin.NewAdministratorNow(admin.NewID(), acc)
	if err != nil {
		return CreateAdminResult{}, err
	}

	// 4. Persist the administrator.
	if err := uc.admins.Create(ctx, a); err != nil {
		return CreateAdminResult{}, err
	}

	return CreateAdminResult{Admin: a}, nil
}
