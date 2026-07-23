package api

import (
	authx "github.com/els/backend/internal/utils/auth"
)

type UnitOutput struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	Kind          string `json:"kind"`
	Transcription string `json:"transcription,omitempty"`
	Translation   string `json:"translation,omitempty"`
	Definition    string `json:"definition,omitempty"`
	Example       string `json:"example,omitempty"`
	Frequency     int    `json:"frequency"`
	CEFR          string `json:"cefr"`
	Status        string `json:"status"`
	CorrectStreak int    `json:"correct_streak"`
	CreatedAt     string `json:"created_at"`
}

type AddUnitInput struct {
	authx.BearerInput
	Body struct {
		Text        string `json:"text" minLength:"1" maxLength:"200" doc:"Word, phrase, phrasal verb or idiom to add"`
		Kind        string `json:"kind,omitempty" maxLength:"20" doc:"Known kind from analyze; when set the unit is stored instantly and enriched in the background"`
		Translation string `json:"translation,omitempty" maxLength:"500" doc:"Known translation from analyze"`
		Description string `json:"description,omitempty" maxLength:"1000" doc:"Known short definition from analyze"`
		Frequency   int    `json:"frequency,omitempty" minimum:"0" maximum:"5" doc:"Known frequency from analyze"`
		Cefr        string `json:"cefr,omitempty" maxLength:"2" doc:"Known CEFR level from analyze"`
	}
}

type AddUnitOutput struct {
	Correct     bool        `json:"correct"`
	Correction  string      `json:"correction,omitempty"`
	Explanation string      `json:"explanation,omitempty"`
	Unit        *UnitOutput `json:"unit,omitempty"`
}

type AnalyzeInput struct {
	authx.BearerInput
	Body struct {
		Text    string `json:"text" minLength:"1" maxLength:"500" doc:"Selected text to break into vocabulary items"`
		Context string `json:"context,omitempty" maxLength:"2000" doc:"Optional surrounding text for disambiguation"`
	}
}

type SpotOutput struct {
	Ref     int    `json:"ref"`
	Example string `json:"example,omitempty"`
}

type OccurrenceOutput struct {
	MediaID     string       `json:"media_id"`
	MediaType   string       `json:"media_type"`
	Title       string       `json:"title"`
	Kind        string       `json:"kind,omitempty"`
	SeriesTitle string       `json:"series_title,omitempty"`
	Season      int          `json:"season,omitempty"`
	Episode     int          `json:"episode,omitempty"`
	Author      string       `json:"author,omitempty"`
	Count       int          `json:"count"`
	Spots       []SpotOutput `json:"spots"`
}

type AnalyzeItemOutput struct {
	Text        string             `json:"text"`
	Kind        string             `json:"kind"`
	Description string             `json:"description"`
	Translation string             `json:"translation,omitempty"`
	Frequency   int                `json:"frequency"`
	CEFR        string             `json:"cefr"`
	Common      bool               `json:"common"`
	Existing    bool               `json:"existing"`
	Total       int                `json:"total"`
	MediaCount  int                `json:"media_count"`
	Media       []OccurrenceOutput `json:"media"`
}

type AnalyzeOutput struct {
	Items []AnalyzeItemOutput `json:"items"`
}

type OccurrencesInput struct {
	authx.BearerInput
	Text string `query:"text" required:"true" minLength:"1" maxLength:"200" doc:"Word or phrase to look up in the parsed lexicon"`
}

type OccurrencesOutput struct {
	Common     bool               `json:"common"`
	Total      int                `json:"total"`
	MediaCount int                `json:"media_count"`
	Media      []OccurrenceOutput `json:"media"`
}

type ListUnitsInput struct {
	authx.BearerInput
	Search string `query:"q" maxLength:"200"`
	Status string `query:"status" enum:",new,learning,learned"`
	Limit  int    `query:"limit" minimum:"1" maximum:"100" default:"50"`
	Offset int    `query:"offset" minimum:"0" default:"0"`
}

type UnitsOutput struct {
	Items  []UnitOutput `json:"items"`
	Total  int          `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

type UpdateStatusInput struct {
	authx.BearerInput
	ID   string `path:"id"`
	Body struct {
		Status string `json:"status" enum:"new,learning,learned"`
	}
}

type GeneratePracticeInput struct {
	authx.BearerInput
}

type GetPracticeInput struct {
	authx.BearerInput
}

type PracticeAnswerDTO struct {
	Answer      string `json:"answer"`
	Correct     bool   `json:"correct"`
	Correction  string `json:"correction,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type PracticeOutput struct {
	ID        string                       `json:"id"`
	Status    string                       `json:"status"`
	Error     string                       `json:"error,omitempty"`
	Exercises string                       `json:"exercises"`
	Words     []UnitOutput                 `json:"words"`
	Answers   map[string]PracticeAnswerDTO `json:"answers"`
	Completed bool                         `json:"completed"`
}

type SavePracticeProgressInput struct {
	authx.BearerInput
	Body struct {
		SessionID string                       `json:"session_id"`
		Answers   map[string]PracticeAnswerDTO `json:"answers"`
		Completed bool                         `json:"completed"`
	}
}

type SavePracticeProgressOutput struct {
	OK bool `json:"ok"`
}

type GenerateCardsInput struct {
	authx.BearerInput
	Body struct {
		ImagesOnly bool `json:"images_only,omitempty" doc:"Build the deck only from words with a generated illustration"`
	}
}

type CardOutput struct {
	UnitID        string   `json:"unit_id"`
	Mode          string   `json:"mode" enum:"choice,input"`
	Direction     string   `json:"direction" enum:"word,translation"`
	Kind          string   `json:"kind"`
	Status        string   `json:"status"`
	Definition    string   `json:"definition,omitempty"`
	Word          string   `json:"word,omitempty"`
	Transcription string   `json:"transcription,omitempty"`
	ImageURL      string   `json:"image_url,omitempty"`
	Options       []string `json:"options,omitempty"`
}

type CardsOutput struct {
	Cards []CardOutput `json:"cards"`
}

type DueCardsInput struct {
	authx.BearerInput
}

type DueCardsOutput struct {
	Count int `json:"count"`
}

type AnswerCardInput struct {
	authx.BearerInput
	Body struct {
		UnitID string `json:"unit_id" minLength:"1"`
		Answer string `json:"answer" maxLength:"200"`
	}
}

type AnswerCardOutput struct {
	Correct bool       `json:"correct"`
	Unit    UnitOutput `json:"unit"`
}

type CheckPracticeInput struct {
	authx.BearerInput
	Body struct {
		Instruction string `json:"instruction" maxLength:"500" doc:"The exercise instruction the answer responds to"`
		Answer      string `json:"answer" minLength:"1" maxLength:"1000" doc:"The learner's free-form sentence"`
	}
}

type CheckPracticeOutput struct {
	Correct     bool   `json:"correct"`
	Correction  string `json:"correction,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type DeleteUnitInput struct {
	authx.BearerInput
	ID string `path:"id"`
}

type DeleteUnitOutput struct {
	OK bool `json:"ok"`
}
