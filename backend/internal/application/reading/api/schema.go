package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type GenerateTextInput struct {
	authx.BearerInput
	Body struct {
		Topic    string `json:"topic,omitempty" maxLength:"200" doc:"Optional topic for the text"`
		UseVocab bool   `json:"use_vocab,omitempty" doc:"Weave in words the user is learning"`
		Level    string `json:"level,omitempty" enum:"easy,medium,hard" doc:"Difficulty; default medium"`
		Length   string `json:"length,omitempty" enum:"short,medium,long" doc:"Text length; default medium"`
	}
}

type TextOutput struct {
	Title string   `json:"title"`
	Body  string   `json:"body" doc:"Markdown paragraphs; standalone [image: ...] lines mark optional illustrations"`
	Words []string `json:"words" doc:"Learner words actually used in the text"`
}
