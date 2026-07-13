package api

import (
	"time"

	authx "github.com/els/backend/internal/utils/auth"
)

type AdminOutput struct {
	ID         string    `json:"id"`
	AccountID  string    `json:"account_id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Status     string    `json:"status"`
	PictureURL string    `json:"picture_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AdminsOutput struct {
	Items  []AdminOutput `json:"items"`
	Total  int64         `json:"total"`
	Limit  int32         `json:"limit"`
	Offset int32         `json:"offset"`
}

type CreateAdminInput struct {
	authx.BearerInput
	Body struct {
		Email     string `json:"email" format:"email"`
		FirstName string `json:"first_name" minLength:"1"`
		LastName  string `json:"last_name"  minLength:"1"`
	}
}

type GetAdminInput struct {
	authx.BearerInput
	ID string `path:"id" format:"uuid"`
}

type ListAdminsInput struct {
	authx.BearerInput
	Limit  int32 `query:"limit"  default:"50" minimum:"1" maximum:"200"`
	Offset int32 `query:"offset" default:"0"  minimum:"0"`
}
