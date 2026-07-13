package api

import (
	"github.com/els/backend/internal/domain/quest"
	authx "github.com/els/backend/internal/utils/auth"
)

type CreateMissionInput struct {
	authx.BearerInput
	Body struct {
		Prompt        string `json:"prompt,omitempty"`
		Genre         string `json:"genre,omitempty"`
		Language      string `json:"language,omitempty"`
		PracticeGoals string `json:"practiceGoals,omitempty"`
	}
}

type CreateMissionOutput struct {
	MissionID string `json:"missionId"`
}

type ListMissionsInput struct {
	authx.BearerInput
}

type MissionSummary struct {
	ID               string `json:"id"`
	Started          bool   `json:"started"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	Genre            string `json:"genre,omitempty"`
	Language         string `json:"language,omitempty"`
	CreatedAt        string `json:"createdAt"`
	IsComplete       bool   `json:"isComplete"`
	CurrentStage     int    `json:"currentStage"`
	TotalStages      int    `json:"totalStages"`
	GenerationStatus string `json:"generationStatus"`
	GenerationStep   string `json:"generationStep,omitempty"`
	GenerationError  string `json:"generationError,omitempty"`
	CoverImage       string `json:"coverImage,omitempty"`
	CoverImageStatus string `json:"coverImageStatus,omitempty"`
}

type ListMissionsOutput struct {
	Missions []MissionSummary `json:"missions"`
}

type MissionInput struct {
	authx.BearerInput
	ID string `path:"id"`
}

type GetMissionOutput struct {
	Mission     *quest.CustomMission            `json:"mission"`
	ActiveReply *quest.RespondJobStatusResponse `json:"activeReply,omitempty"`
}

type RespondInput struct {
	authx.BearerInput
	ID   string `path:"id"`
	Body struct {
		Text   string `json:"text"`
		Strict *bool  `json:"strict,omitempty"`
	}
}

type RespondOutput struct {
	JobID string `json:"jobId"`
}

type SuggestNativeReplyInput struct {
	authx.BearerInput
	ID   string `path:"id"`
	Body struct {
		Text string `json:"text"`
	}
}

type SuggestNativeReplyOutput struct {
	Variants []string `json:"variants"`
}

type ResetMissionOutput struct {
	Mission *quest.CustomMission `json:"mission"`
}

type RegenerateImagesInput struct {
	authx.BearerInput
	ID   string `path:"id"`
	Body struct {
		Kind string `json:"kind,omitempty" enum:",cover,scene,avatar"`
		Key  string `json:"key,omitempty"`
	}
}

type RegenerateImagesOutput struct {
	Mission *quest.CustomMission `json:"mission"`
}

type DeleteMissionOutput struct {
	OK bool `json:"ok"`
}
