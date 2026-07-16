package api

import (
	"github.com/danielgtaylor/huma/v2"

	authx "github.com/els/backend/internal/utils/auth"
)

type AssessForm struct {
	Audio huma.FormFile `form:"audio" required:"true"`
}

type AssessInput struct {
	authx.BearerInput
	RawBody huma.MultipartFormFiles[AssessForm]
}

type PhonemeOutput struct {
	Expected string  `json:"expected"`
	Heard    string  `json:"heard,omitempty"`
	Score    float64 `json:"score"`
	Verdict  string  `json:"verdict" enum:"good,close,wrong,missing"`
}

type WordOutput struct {
	Word     string          `json:"word"`
	IPA      string          `json:"ipa"`
	Score    int             `json:"score"`
	Phonemes []PhonemeOutput `json:"phonemes"`
	Extra    []string        `json:"extra"`
}

type AssessOutput struct {
	Overall int          `json:"overall"`
	Heard   string       `json:"heard"`
	Words   []WordOutput `json:"words"`
}

type FeedbackInput struct {
	authx.BearerInput
	Body struct {
		Text           string   `json:"text" minLength:"1" maxLength:"500" doc:"Reference text that was read aloud"`
		Heard          string   `json:"heard" minLength:"1" maxLength:"2000" doc:"IPA transcription of what was actually said"`
		NativeLanguage string   `json:"native_language" maxLength:"50" doc:"Student's native language for advice wording"`
		Issues         []string `json:"issues,omitempty" maxItems:"100" doc:"Detected per-phoneme issues, e.g. \"think: expected θ, heard s\""`
	}
}

type FeedbackTipOutput struct {
	Sound  string `json:"sound"`
	Advice string `json:"advice"`
}

type FeedbackOutput struct {
	Summary string              `json:"summary"`
	Tips    []FeedbackTipOutput `json:"tips"`
}

type ListPhonemesInput struct {
	authx.BearerInput
}

type PhonemeInfoOutput struct {
	Symbol      string `json:"symbol"`
	Kind        string `json:"kind" enum:"vowel,diphthong,consonant"`
	Examples    string `json:"examples"`
	Description string `json:"description"`
	Pitfall     string `json:"pitfall,omitempty"`
}

type PhonemesOutput struct {
	Items []PhonemeInfoOutput `json:"items"`
}
