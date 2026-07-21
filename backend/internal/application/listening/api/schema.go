package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type GenerateDictationInput struct {
	authx.BearerInput
	Body struct {
		Topic    string `json:"topic,omitempty" maxLength:"200" doc:"Optional topic for the sentences"`
		UseVocab bool   `json:"use_vocab,omitempty" doc:"Use words the user is learning"`
		Level    string `json:"level,omitempty" enum:"easy,medium,hard" doc:"Difficulty; default medium"`
		Count    int    `json:"count,omitempty" minimum:"3" maximum:"10" doc:"Number of sentences; default 5"`
	}
}

type DictationOutput struct {
	Sentences []string `json:"sentences"`
}
