package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

// --- books ---

type BookSchema struct {
	Slug        string `json:"slug"`
	Series      string `json:"series"`
	Level       string `json:"level,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type BookListOutput struct {
	Items []BookSchema `json:"items"`
}

type ListBooksInput struct {
	authx.BearerInput
}

// --- chapters ---

type ChapterOutput struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Page      int      `json:"page"`
	Words     []string `json:"words,omitempty"`
	Footer    string   `json:"footer,omitempty"`
	Theory    string   `json:"theory"`
	Exercises string   `json:"exercises"`
	Status    string   `json:"status" enum:"generating,ready,error"`
	Error     string   `json:"error,omitempty"`
}

type ChaptersOutput struct {
	Items []ChapterOutput `json:"items"`
}

type ChapterBody struct {
	Number    int      `json:"number" minimum:"1"`
	Title     string   `json:"title" minLength:"1" maxLength:"200"`
	Page      int      `json:"page" minimum:"0"`
	Words     []string `json:"words"`
	Footer    string   `json:"footer" maxLength:"500"`
	Theory    string   `json:"theory"`
	Exercises string   `json:"exercises"`
}

type ListChaptersInput struct {
	authx.BearerInput
	Book string `path:"book" minLength:"1"`
}

type GetChapterInput struct {
	authx.BearerInput
	Book   string `path:"book" minLength:"1"`
	Number int    `path:"number" minimum:"1"`
}

type CreateChapterInput struct {
	authx.BearerInput
	Book string `path:"book" minLength:"1"`
	Body ChapterBody
}

type UpdateChapterInput struct {
	authx.BearerInput
	Book   string `path:"book" minLength:"1"`
	Number int    `path:"number" minimum:"1"`
	Body   ChapterBody
}

type GenerateChapterBody struct {
	Topic string `json:"topic" minLength:"1" maxLength:"500"`
}

type GenerateChapterInput struct {
	authx.BearerInput
	Book string `path:"book" minLength:"1"`
	Body GenerateChapterBody
}

type DeleteChapterInput struct {
	authx.BearerInput
	Book   string `path:"book" minLength:"1"`
	Number int    `path:"number" minimum:"1"`
}

type DeleteChapterOutput struct {
	OK bool `json:"ok"`
}

// --- illustrations ---

type EnsureBody struct {
	Prompt  string `json:"prompt" minLength:"1" maxLength:"2000"`
	Aspect  string `json:"aspect,omitempty" enum:"square,landscape,portrait" default:"square"`
	Trigger bool   `json:"trigger"`
}

type EnsureInput struct {
	authx.BearerInput
	Body EnsureBody
}

type IllustrationOutput struct {
	ID     string `json:"id"`
	Status string `json:"status" enum:"pending,generating,ready,error"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

// --- practice ---

type VariantSchema struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Exercises string `json:"exercises"`
	Status    string `json:"status" enum:"generating,ready,error"`
	Error     string `json:"error,omitempty"`
}

type VariantsOutput struct {
	Items []VariantSchema `json:"items"`
}

type AnswerSchema struct {
	Answer      string `json:"answer"`
	Correct     bool   `json:"correct"`
	Correction  string `json:"correction,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type ProgressOutput struct {
	Answers   map[string]AnswerSchema `json:"answers"`
	Completed bool                    `json:"completed"`
}

type CheckOutput struct {
	Correct     bool   `json:"correct"`
	Correction  string `json:"correction,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type PracticeOKOutput struct {
	OK bool `json:"ok"`
}

type ListVariantsInput struct {
	authx.BearerInput
	Kind   string `path:"kind" minLength:"1"`
	Number int    `path:"number" minimum:"1"`
}

type GenerateVariantInput struct {
	authx.BearerInput
	Kind   string `path:"kind" minLength:"1"`
	Number int    `path:"number" minimum:"1"`
}

type DeleteVariantInput struct {
	authx.BearerInput
	ID string `path:"id"`
}

type GetProgressInput struct {
	authx.BearerInput
	Kind    string `path:"kind" minLength:"1"`
	Number  int    `path:"number" minimum:"1"`
	Variant string `query:"variant"`
}

type SaveProgressBody struct {
	Variant   string                  `json:"variant"`
	Answers   map[string]AnswerSchema `json:"answers"`
	Completed bool                    `json:"completed"`
}

type SaveProgressInput struct {
	authx.BearerInput
	Kind   string `path:"kind" minLength:"1"`
	Number int    `path:"number" minimum:"1"`
	Body   SaveProgressBody
}

type ResetProgressInput struct {
	authx.BearerInput
	Kind    string `path:"kind" minLength:"1"`
	Number  int    `path:"number" minimum:"1"`
	Variant string `query:"variant"`
}

type CheckBody struct {
	Kind        string `json:"kind" minLength:"1"`
	Number      int    `json:"number" minimum:"1"`
	Instruction string `json:"instruction" maxLength:"2000"`
	Answer      string `json:"answer" minLength:"1" maxLength:"4000"`
}

type CheckInput struct {
	authx.BearerInput
	Body CheckBody
}
