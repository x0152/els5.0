package api

import (
	"context"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/speech/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/speech"
	authx "github.com/els/backend/internal/utils/auth"
)

const maxAudioBytes = 10 << 20

type Deps struct {
	Authenticator    *authx.Authenticator
	Assess           *usecases.AssessUseCase
	Feedback         *usecases.FeedbackUseCase
	ListPhonemes     *usecases.ListPhonemesUseCase
	GeneratePractice *usecases.GeneratePracticeUseCase
	Synthesize       *usecases.SynthesizeUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "assessSpeech",
		Method:      http.MethodPost,
		Path:        "/api/v1/speech/assess",
		Summary:     "Score the pronunciation of a recorded reading against the reference text",
		Tags:        []string{"speech"},
	}, func(ctx context.Context, actor *iam.Actor, in *AssessInput) (AssessOutput, error) {
		form := in.RawBody.Data()
		if form == nil || !form.Audio.IsSet {
			return AssessOutput{}, huma.Error400BadRequest("audio file is required")
		}
		defer form.Audio.Close()
		audio, err := io.ReadAll(io.LimitReader(form.Audio, maxAudioBytes))
		if err != nil {
			return AssessOutput{}, huma.Error500InternalServerError("failed to read audio")
		}
		cmd := usecases.AssessCommand{
			Audio:      audio,
			Text:       formValue(in.RawBody.Form, "text"),
			Strictness: formFloat(in.RawBody.Form, "strictness", actor.Account().SpeechStrictness()),
		}
		res, err := deps.Assess.Execute(ctx, actor, cmd)
		if err != nil {
			return AssessOutput{}, err
		}
		return toAssessOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "speechFeedback",
		Method:      http.MethodPost,
		Path:        "/api/v1/speech/feedback",
		Summary:     "Get LLM coaching advice for a scored reading",
		Tags:        []string{"speech"},
	}, func(ctx context.Context, actor *iam.Actor, in *FeedbackInput) (FeedbackOutput, error) {
		res, err := deps.Feedback.Execute(ctx, actor, toFeedbackCommand(in))
		if err != nil {
			return FeedbackOutput{}, err
		}
		return toFeedbackOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listSpeechPhonemes",
		Method:      http.MethodGet,
		Path:        "/api/v1/speech/phonemes",
		Summary:     "List the phoneme articulation guide",
		Tags:        []string{"speech"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListPhonemesInput) (PhonemesOutput, error) {
		items, err := deps.ListPhonemes.Execute(ctx, actor)
		if err != nil {
			return PhonemesOutput{}, err
		}
		return toPhonemesOutput(items), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "speechSynthesize",
		Method:      http.MethodPost,
		Path:        "/api/v1/speech/tts",
		Summary:     "Synthesize speech audio for a text",
		Tags:        []string{"speech"},
	}, func(ctx context.Context, actor *iam.Actor, in *SynthesizeInput) (SynthesizeOutput, error) {
		res, err := deps.Synthesize.Execute(ctx, actor, usecases.SynthesizeCommand{Text: in.Body.Text, Voice: in.Body.Voice, Speed: in.Body.Speed})
		if err != nil {
			return SynthesizeOutput{}, err
		}
		return SynthesizeOutput{AudioBase64: base64.StdEncoding.EncodeToString(res.Audio), Voice: res.Voice}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listSpeechVoices",
		Method:      http.MethodGet,
		Path:        "/api/v1/speech/voices",
		Summary:     "List available TTS voices",
		Tags:        []string{"speech"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListVoicesInput) (VoicesOutput, error) {
		return VoicesOutput{Voices: speech.Voices}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "speechGeneratePractice",
		Method:      http.MethodPost,
		Path:        "/api/v1/speech/practice",
		Summary:     "Generate sentences to practice reading aloud",
		Tags:        []string{"speech"},
	}, func(ctx context.Context, actor *iam.Actor, in *GeneratePracticeInput) (SpeechPracticeOutput, error) {
		sentences, err := deps.GeneratePractice.Execute(ctx, actor, usecases.GeneratePracticeCommand{Topic: in.Body.Topic, Sounds: in.Body.Sounds})
		if err != nil {
			return SpeechPracticeOutput{}, err
		}
		return SpeechPracticeOutput{Sentences: sentences}, nil
	})
}

func formValue(form *multipart.Form, key string) string {
	if form == nil {
		return ""
	}
	if vals := form.Value[key]; len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func formFloat(form *multipart.Form, key string, fallback float64) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(formValue(form, key)), 64)
	if err != nil {
		return fallback
	}
	return f
}
