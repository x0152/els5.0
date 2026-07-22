package api

import (
	"github.com/els/backend/internal/domain/onboarding"
)

func toProgressOutput(statuses []onboarding.Status) OnboardingProgressOutput {
	items := make([]OnboardingItemOutput, 0, len(statuses))
	for _, s := range statuses {
		items = append(items, OnboardingItemOutput{
			ID:        s.ID,
			Kind:      string(s.Kind),
			Metric:    s.Metric,
			Threshold: s.Threshold,
			Value:     s.Value,
			Done:      s.Done,
			Acked:     s.Acked,
		})
	}
	return OnboardingProgressOutput{Items: items}
}
