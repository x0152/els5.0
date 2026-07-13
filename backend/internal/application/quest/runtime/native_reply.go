package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

func (s *Dialog) NativeReplyVariants(ctx context.Context, userID, missionID, text string) ([]string, error) {
	draft := strings.TrimSpace(text)
	if draft == "" {
		return nil, fmt.Errorf("%w: text is required", shared.ErrValidation)
	}

	mission, err := s.missions.GetByID(ctx, userID, missionID)
	if err != nil {
		return nil, err
	}

	profile, err := s.profiles.Get(ctx, userID)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}
		profile = quest.NewDefaultProfile()
	}
	applyAccountIdentity(ctx, s.accounts, userID, &profile)

	return s.llm.SuggestNativeReplies(ctx, mission, draft, &profile)
}
