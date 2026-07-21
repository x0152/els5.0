package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type TrainerCheckInput struct {
	authx.BearerInput
	Body struct {
		Dialogue string `json:"dialogue,omitempty" maxLength:"3000" doc:"Dialogue context the draft replies to"`
		Draft    string `json:"draft" minLength:"1" maxLength:"2000" doc:"Draft reply to check"`
		Level    int    `json:"level" minimum:"1" maximum:"3" doc:"Strictness: 1 grammar, 2 natural, 3 native"`
	}
}

type TrainerIssueOutput struct {
	Fragment string `json:"fragment"`
	Severity string `json:"severity" enum:"grammar,style,native"`
	Hint     string `json:"hint"`
}

type TrainerCheckOutput struct {
	Pass    bool                 `json:"pass"`
	Comment string               `json:"comment"`
	Issues  []TrainerIssueOutput `json:"issues"`
}

type GenerateSituationInput struct {
	authx.BearerInput
	Body struct {
		Topic string `json:"topic,omitempty" maxLength:"200" doc:"Optional topic for the situation"`
	}
}

type SituationOutput struct {
	Scenario string `json:"scenario"`
	Dialogue string `json:"dialogue"`
}
