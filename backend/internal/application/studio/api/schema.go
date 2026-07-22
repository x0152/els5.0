package api

import (
	"time"

	authx "github.com/els/backend/internal/utils/auth"
)

type AreaOutput struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Icon      string    `json:"icon,omitempty"`
	Total     int       `json:"total"`
	Done      int       `json:"done"`
	Due       int       `json:"due"`
	CreatedAt time.Time `json:"created_at"`
}

type ItemOutput struct {
	ID                string    `json:"id"`
	AreaID            string    `json:"area_id"`
	Text              string    `json:"text"`
	Transcription     string    `json:"transcription,omitempty"`
	Translation       string    `json:"translation,omitempty"`
	Explanation       string    `json:"explanation,omitempty"`
	ExplanationNative string    `json:"explanation_native,omitempty"`
	Example           string    `json:"example,omitempty"`
	Task              string     `json:"task,omitempty"`
	Listened          bool       `json:"listened"`
	Spoken            bool       `json:"spoken"`
	Written           bool       `json:"written"`
	Recalled          bool       `json:"recalled"`
	ReviewStage       int        `json:"review_stage"`
	NextReviewAt      *time.Time `json:"next_review_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type ListAreasInput struct {
	authx.BearerInput
}

type AreasOutput struct {
	Items []AreaOutput `json:"items"`
}

type CreateAreaInput struct {
	authx.BearerInput
	Body struct {
		Title string `json:"title" minLength:"1" maxLength:"200"`
		Icon  string `json:"icon,omitempty" maxLength:"32" doc:"Lucide icon name, e.g. coffee"`
	}
}

type DeleteAreaInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type DeleteAreaOutput struct{}

type ListItemsInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type ItemsOutput struct {
	Items []ItemOutput `json:"items"`
}

type AddItemInput struct {
	authx.BearerInput
	ID   string `path:"id" format:"uuid"`
	Body struct {
		Text string `json:"text" minLength:"1" maxLength:"500" doc:"Word or phrase to study"`
	}
}

type DeleteItemInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type DeleteItemOutput struct{}

type MarkSkillInput struct {
	authx.BearerInput
	ID   string `path:"id" format:"uuid"`
	Body struct {
		Skill string `json:"skill" enum:"listened,spoken,written,recalled"`
	}
}

type CaptureItemInput struct {
	authx.BearerInput
	Body struct {
		Text string `json:"text" minLength:"1" maxLength:"500" doc:"Word or phrase to study"`
		Area string `json:"area" minLength:"1" maxLength:"200" doc:"Target area title, created if missing"`
		Icon string `json:"icon,omitempty" maxLength:"32" doc:"Lucide icon name for a newly created area"`
	}
}

type PassReviewInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type RegenExampleInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type RegenTaskInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type CheckReplyInput struct {
	authx.BearerInput
	ID   string `path:"id" format:"uuid"`
	Body struct {
		Reply string `json:"reply" minLength:"1" maxLength:"2000"`
	}
}

type CheckReplyOutput struct {
	OK      bool   `json:"ok"`
	Comment string `json:"comment"`
}
