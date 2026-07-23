package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/illustration"
	"github.com/els/backend/internal/domain/settings"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type ImageEnsurer interface {
	Ensure(ctx context.Context, prompt, aspect string, trigger bool) illustration.Status
}

type AddUnitUseCase struct {
	units  vocab.Repository
	llm    LLMClient
	flags  settings.FlagRepository
	images ImageEnsurer
}

func NewAddUnitUseCase(units vocab.Repository, llm LLMClient, flags settings.FlagRepository, images ImageEnsurer) *AddUnitUseCase {
	return &AddUnitUseCase{units: units, llm: llm, flags: flags, images: images}
}

type AddUnitResult struct {
	Correct     bool
	Correction  string
	Explanation string
	Unit        *vocab.Unit
}

// AddUnitMeta carries word details already produced by the analyze step,
// letting the unit be stored instantly while the LLM enriches it in the background.
type AddUnitMeta struct {
	Kind        string
	Translation string
	Description string
	Frequency   int
	Cefr        string
}

func (uc *AddUnitUseCase) Execute(ctx context.Context, actor *iam.Actor, input string, meta *AddUnitMeta) (AddUnitResult, error) {
	if actor == nil {
		return AddUnitResult{}, shared.ErrUnauthorized
	}
	accountID := actor.AccountID().String()

	input = strings.TrimSpace(input)
	if input == "" {
		return AddUnitResult{}, shared.Validation(fmt.Errorf("text: must not be empty"))
	}
	if !uc.llm.Available() {
		return AddUnitResult{}, shared.ErrUnavailable
	}

	// Finish processing even if the client disconnects (page reload, navigation).
	ctx = context.WithoutCancel(ctx)

	if meta != nil {
		return uc.addFromMeta(ctx, actor, input, *meta)
	}

	// Ask the LLM to validate and describe the word.
	system, user := vocab.BuildCheckPrompt(input, actor.Account().NativeLanguage())
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return AddUnitResult{}, err
	}
	check, err := vocab.ParseCheckResult(raw)
	if err != nil {
		return AddUnitResult{}, err
	}

	// If the input is invalid — return a correction and save nothing.
	if !check.Correct {
		return AddUnitResult{Correct: false, Correction: check.Correction, Explanation: check.Explanation}, nil
	}

	return uc.persist(ctx, accountID, check)
}

// addFromMeta stores the unit immediately from analyze data and enriches it in the background.
func (uc *AddUnitUseCase) addFromMeta(ctx context.Context, actor *iam.Actor, input string, meta AddUnitMeta) (AddUnitResult, error) {
	check := vocab.CheckResult{
		Correct:     true,
		Kind:        meta.Kind,
		Text:        input,
		Translation: meta.Translation,
		Definition:  meta.Description,
		Frequency:   meta.Frequency,
		Cefr:        meta.Cefr,
	}
	res, err := uc.persist(ctx, actor.AccountID().String(), check)
	if err != nil {
		return AddUnitResult{}, err
	}
	go uc.enrich(*res.Unit, actor.Account().NativeLanguage())
	return res, nil
}

func (uc *AddUnitUseCase) persist(ctx context.Context, accountID string, check vocab.CheckResult) (AddUnitResult, error) {
	exists, err := uc.units.ExistsText(ctx, accountID, strings.TrimSpace(check.Text), check.Kind)
	if err != nil {
		return AddUnitResult{}, err
	}
	if exists {
		return AddUnitResult{}, shared.ErrConflict
	}

	unit, err := vocab.NewUnit(uuid.NewString(), accountID, check)
	if err != nil {
		return AddUnitResult{}, err
	}
	stored, err := uc.units.Create(ctx, unit)
	if err != nil {
		return AddUnitResult{}, err
	}

	if uc.images != nil && uc.flags != nil {
		if on, err := uc.flags.GetFlag(ctx, settings.FlagAutoWordImages); err == nil && on {
			uc.images.Ensure(ctx, vocab.ImagePrompt(stored.Text), "square", true)
		}
	}
	return AddUnitResult{Correct: true, Unit: &stored}, nil
}

// enrich fills in transcription, definition and example via the LLM after the unit is already stored.
func (uc *AddUnitUseCase) enrich(unit vocab.Unit, nativeLanguage string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	system, user := vocab.BuildCheckPrompt(unit.Text, nativeLanguage)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return
	}
	check, err := vocab.ParseCheckResult(raw)
	if err != nil || !check.Correct {
		return
	}
	if v := strings.TrimSpace(check.Transcription); v != "" {
		unit.Transcription = v
	}
	if v := strings.TrimSpace(check.Translation); v != "" {
		unit.Translation = v
	}
	if v := strings.TrimSpace(check.Definition); v != "" {
		unit.Definition = v
	}
	if v := strings.TrimSpace(check.Example); v != "" {
		unit.Example = v
	}
	_ = uc.units.UpdateDetails(ctx, unit)
}
