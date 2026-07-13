package api

import (
	"github.com/danielgtaylor/huma/v2"

	authx "github.com/els/backend/internal/utils/auth"
)

type BookSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
	CoverURL    string `json:"cover_url,omitempty"`
	Status      string `json:"status"`
	Kind        string `json:"kind"`
	GroupTitle  string `json:"group_title,omitempty"`
	TextLength  int    `json:"text_length"`
	Position    int    `json:"position"`
	Percent     int    `json:"percent"`
	CreatedAt   string `json:"created_at"`
}

type BookOutput struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
	CoverURL    string `json:"cover_url,omitempty"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	Kind        string `json:"kind"`
	GroupTitle  string `json:"group_title,omitempty"`
	TextLength  int    `json:"text_length"`
	Position    int    `json:"position"`
	Percent     int    `json:"percent"`
	ContentURL  string `json:"content_url,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type BooksOutput struct {
	Items []BookSummary `json:"items"`
}

type CollectionOutput struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	CoverURL    string `json:"cover_url,omitempty"`
}

type CollectionsOutput struct {
	Items []CollectionOutput `json:"items"`
}

type ListCollectionsInput struct {
	authx.BearerInput
}

type UpdateCollectionForm struct {
	Cover huma.FormFile `form:"cover" contentType:"image/png,image/jpeg,image/webp" required:"false"`
}

type UpdateCollectionInput struct {
	authx.BearerInput
	RawBody huma.MultipartFormFiles[UpdateCollectionForm]
}

type ListBooksInput struct {
	authx.BearerInput
}

type GetBookInput struct {
	authx.BearerInput
	ID string `path:"id"`
}

type UploadBookForm struct {
	File  huma.FormFile `form:"file" required:"true"`
	Cover huma.FormFile `form:"cover" contentType:"image/png,image/jpeg,image/webp" required:"false"`
}

type UploadBookInput struct {
	authx.BearerInput
	RawBody huma.MultipartFormFiles[UploadBookForm]
}

type ImportArticleInput struct {
	authx.BearerInput
	Body struct {
		URL        string `json:"url" format:"uri" minLength:"1" doc:"Public article URL"`
		GroupTitle string `json:"group_title,omitempty" doc:"Optional collection to group the article under"`
	}
}

type UpdateBookForm struct {
	Cover huma.FormFile `form:"cover" contentType:"image/png,image/jpeg,image/webp" required:"false"`
}

type UpdateBookInput struct {
	authx.BearerInput
	ID      string `path:"id"`
	RawBody huma.MultipartFormFiles[UpdateBookForm]
}

type DeleteBookInput struct {
	authx.BearerInput
	ID string `path:"id"`
}

type DeleteBookOutput struct {
	OK bool `json:"ok"`
}

type SaveBookProgressInput struct {
	authx.BearerInput
	ID   string `path:"id"`
	Body struct {
		Position int `json:"position" minimum:"0"`
	}
}

type SaveBookProgressOutput struct {
	OK bool `json:"ok"`
}
