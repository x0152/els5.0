package api

import (
	"github.com/danielgtaylor/huma/v2"

	authx "github.com/els/backend/internal/utils/auth"
)

type MeInput struct {
	authx.BearerInput
}

type ListAppsInput struct {
	authx.BearerInput
}

type MeOutput struct {
	AccountID            string `json:"account_id"`
	Email                string `json:"email"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name"`
	PictureURL           string `json:"picture_url"`
	EnglishLevel         string `json:"english_level"`
	AboutMe              string `json:"about_me"`
	NativeLanguage       string `json:"native_language"`
	ShowTranslations     bool   `json:"show_translations"`
	Status               string `json:"status"`
	Role                 string `json:"role"`
	EntityID             string `json:"entity_id"`
	IsGlobalAdmin        bool   `json:"is_global_admin"`
	ImpersonationEnabled bool   `json:"impersonation_enabled"`
}

type UpdateProfileBody struct {
	FirstName        string `json:"first_name" minLength:"1" maxLength:"100"`
	LastName         string `json:"last_name" minLength:"1" maxLength:"100"`
	EnglishLevel     string `json:"english_level" maxLength:"100"`
	AboutMe          string `json:"about_me" maxLength:"2000"`
	NativeLanguage   string `json:"native_language" maxLength:"100" doc:"The learner's native language name in English, e.g. Russian, Spanish"`
	ShowTranslations bool   `json:"show_translations" doc:"Show translations into the native language across the platform"`
}

type UpdateProfileInput struct {
	authx.BearerInput
	Body UpdateProfileBody
}

type UploadAccountPictureForm struct {
	File huma.FormFile `form:"file" contentType:"image/png,image/jpeg,image/webp,image/gif" required:"true"`
}

type UploadAccountPictureInput struct {
	authx.BearerInput
	RawBody huma.MultipartFormFiles[UploadAccountPictureForm]
}

type UploadAccountPictureByIDInput struct {
	authx.BearerInput
	AccountID string `path:"account_id" doc:"Target account id"`
	RawBody   huma.MultipartFormFiles[UploadAccountPictureForm]
}

type AccountPictureOutput struct {
	AccountID  string `json:"account_id"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PictureURL string `json:"picture_url"`
	Status     string `json:"status"`
}

type AppOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URI         string `json:"uri"`
	Description string `json:"description,omitempty"`
	Group       string `json:"group,omitempty"`
	Disabled    bool   `json:"disabled"`
}

type AppsOutput struct {
	Items []AppOutput `json:"items"`
	Total int         `json:"total"`
}
