package bindings

import (
	"context"
	"fmt"

	authusecases "github.com/els/backend/internal/application/auth/use_cases"
	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
)

func inviteOnNoAuthToPending[E any](
	invites *authusecases.InviteAccountUseCase,
	sideOf func(E) iam.AccountSide,
) func(context.Context, *iam.Actor, grid.Row, grid.Row, E) error {
	return func(ctx context.Context, _ *iam.Actor, before, after grid.Row, entity E) error {
		return sendInviteOnNoAuthToPending(ctx, invites, before, after, sideOf(entity))
	}
}

func sendInviteOnNoAuthToPending(
	ctx context.Context,
	invites *authusecases.InviteAccountUseCase,
	before, after grid.Row,
	side iam.AccountSide,
) error {
	if statusOf(before) != iam.AccountStatusNoAuth.String() ||
		statusOf(after) != iam.AccountStatusPendingPassword.String() {
		return nil
	}
	if invites == nil {
		return fmt.Errorf("%w: invite service is not configured", shared.ErrConflict)
	}
	return invites.ResendFor(ctx, side.Account())
}

func statusOf(row grid.Row) string {
	raw, _ := row.Cells[iam.ColAccountStatus].(string)
	return raw
}
