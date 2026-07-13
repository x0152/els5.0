package runtime

import (
	"context"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared/vo"
)

// applyAccountIdentity overlays the account's name and English level onto the
// quest profile so both mission generation and dialogs use the real player data.
func applyAccountIdentity(ctx context.Context, accounts iam.AccountRepository, userID string, profile *quest.PlayerProfile) {
	if accounts == nil || profile == nil {
		return
	}
	id, err := vo.ParseID(userID)
	if err != nil {
		return
	}
	account, err := accounts.GetByID(ctx, iam.AccountID{ID: id})
	if err != nil {
		return
	}
	if name := account.Name(); !name.IsZero() {
		profile.FirstName = name.First()
		profile.LastName = name.Last()
	}
	if level := strings.TrimSpace(account.EnglishLevel()); level != "" {
		profile.EnglishLevel = level
	}
	if about := strings.TrimSpace(account.AboutMe()); about != "" {
		profile.AboutMe = about
	}
}
