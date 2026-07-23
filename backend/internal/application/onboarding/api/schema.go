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

type GetToursInput struct {
	authx.BearerInput
}

type ToursOutput struct {
	IDs []string `json:"ids"`
}

type MarkTourInput struct {
	authx.BearerInput
	Body struct {
		ID string `json:"id" minLength:"1" maxLength:"64" doc:"Tour id to mark as completed"`
	}
}

type MarkTourOutput struct{}

type ResetToursInput struct {
	authx.BearerInput
}

type ResetToursOutput struct{}
