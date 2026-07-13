package settingsflag

import (
	"context"

	"github.com/els/backend/internal/domain/settings"
)

type EventProcessingGate struct {
	repo settings.FlagRepository
}

func NewEventProcessingGate(repo settings.FlagRepository) *EventProcessingGate {
	return &EventProcessingGate{repo: repo}
}

func (g *EventProcessingGate) Enabled(ctx context.Context) (bool, error) {
	return g.repo.GetFlag(ctx, settings.FlagEventProcessing)
}
