package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type GetProgressInput struct {
	authx.BearerInput
}

type OnboardingItemOutput struct {
	ID        string `json:"id"`
	Kind      string `json:"kind" enum:"checklist,achievement"`
	Metric    string `json:"metric"`
	Threshold int    `json:"threshold"`
	Value     int    `json:"value"`
	Done      bool   `json:"done"`
	Acked     bool   `json:"acked"`
}

type OnboardingProgressOutput struct {
	Items []OnboardingItemOutput `json:"items"`
}

type AckItemsInput struct {
	authx.BearerInput
	Body struct {
		IDs []string `json:"ids" minItems:"1" maxItems:"100" doc:"Achievement item ids to mark as seen"`
	}
}

type AckItemsOutput struct{}
