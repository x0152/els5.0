package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/vocab/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/vocab"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator    *authx.Authenticator
	AddUnit          *usecases.AddUnitUseCase
	Analyze          *usecases.AnalyzeUseCase
	Occurrences      *usecases.OccurrencesUseCase
	ListUnits        *usecases.ListUnitsUseCase
	UpdateStatus     *usecases.UpdateStatusUseCase
	DeleteUnit       *usecases.DeleteUnitUseCase
	GeneratePractice *usecases.GeneratePracticeUseCase
	GetPractice      *usecases.GetPracticeUseCase
	SaveProgress     *usecases.SavePracticeProgressUseCase
	CheckPractice    *usecases.CheckPracticeUseCase
	GenerateCards    *usecases.GenerateCardsUseCase
	AnswerCard       *usecases.AnswerCardUseCase
	DueCards         *usecases.DueCardsUseCase
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "addVocabUnit",
		Method:      http.MethodPost,
		Path:        "/api/v1/vocab/units",
		Summary:     "Add a vocabulary item; the LLM validates it and writes its description",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *AddUnitInput) (AddUnitOutput, error) {
		res, err := deps.AddUnit.Execute(ctx, actor, in.Body.Text)
		if err != nil {
			return AddUnitOutput{}, err
		}
		return toAddUnitOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "analyzeVocab",
		Method:      http.MethodPost,
		Path:        "/api/v1/vocab/analyze",
		Summary:     "Break selected text into candidate vocabulary items via the LLM",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *AnalyzeInput) (AnalyzeOutput, error) {
		items, err := deps.Analyze.Execute(ctx, actor, in.Body.Text, in.Body.Context)
		if err != nil {
			return AnalyzeOutput{}, err
		}
		return toAnalyzeOutput(items), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "vocabOccurrences",
		Method:      http.MethodGet,
		Path:        "/api/v1/vocab/occurrences",
		Summary:     "List media where a word or phrase occurs, from the parsed lexicon",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *OccurrencesInput) (OccurrencesOutput, error) {
		res, err := deps.Occurrences.Execute(ctx, actor, in.Text)
		if err != nil {
			return OccurrencesOutput{}, err
		}
		return toOccurrencesOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listVocabUnits",
		Method:      http.MethodGet,
		Path:        "/api/v1/vocab/units",
		Summary:     "List vocabulary items with search, status filter and pagination",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListUnitsInput) (UnitsOutput, error) {
		res, err := deps.ListUnits.Execute(ctx, actor, toListFilter(in))
		if err != nil {
			return UnitsOutput{}, err
		}
		return toUnitsOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getVocabPractice",
		Method:      http.MethodGet,
		Path:        "/api/v1/vocab/practice",
		Summary:     "Get the learner's latest practice session, with saved answers",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *GetPracticeInput) (PracticeOutput, error) {
		res, err := deps.GetPractice.Execute(ctx, actor)
		if err != nil {
			return PracticeOutput{}, err
		}
		return toPracticeOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "generateVocabPractice",
		Method:      http.MethodPost,
		Path:        "/api/v1/vocab/practice",
		Summary:     "Generate a fresh LLM practice session from the learner's learning words",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *GeneratePracticeInput) (PracticeOutput, error) {
		res, err := deps.GeneratePractice.Execute(ctx, actor)
		if err != nil {
			return PracticeOutput{}, err
		}
		return toPracticeOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "saveVocabPracticeProgress",
		Method:      http.MethodPut,
		Path:        "/api/v1/vocab/practice/progress",
		Summary:     "Save the learner's answers for the current practice session",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *SavePracticeProgressInput) (SavePracticeProgressOutput, error) {
		if err := deps.SaveProgress.Execute(ctx, actor, in.Body.SessionID, toPracticeAnswers(in.Body.Answers), in.Body.Completed); err != nil {
			return SavePracticeProgressOutput{}, err
		}
		return SavePracticeProgressOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "checkVocabPractice",
		Method:      http.MethodPost,
		Path:        "/api/v1/vocab/practice/check",
		Summary:     "Check a free-form practice sentence via the LLM",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *CheckPracticeInput) (CheckPracticeOutput, error) {
		res, err := deps.CheckPractice.Execute(ctx, actor, in.Body.Instruction, in.Body.Answer)
		if err != nil {
			return CheckPracticeOutput{}, err
		}
		return toCheckPracticeOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "generateVocabCards",
		Method:      http.MethodPost,
		Path:        "/api/v1/vocab/cards",
		Summary:     "Build a flashcard deck: guess the word by its image and definition",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *GenerateCardsInput) (CardsOutput, error) {
		cards, err := deps.GenerateCards.Execute(ctx, actor, in.Body.ImagesOnly)
		if err != nil {
			return CardsOutput{}, err
		}
		return toCardsOutput(cards), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "dueVocabCards",
		Method:      http.MethodGet,
		Path:        "/api/v1/vocab/cards/due",
		Summary:     "Count words that can still advance their memorization progress today",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *DueCardsInput) (DueCardsOutput, error) {
		count, err := deps.DueCards.Execute(ctx, actor)
		if err != nil {
			return DueCardsOutput{}, err
		}
		return DueCardsOutput{Count: count}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "answerVocabCard",
		Method:      http.MethodPost,
		Path:        "/api/v1/vocab/cards/answer",
		Summary:     "Check a flashcard answer and advance the word's memorization progress",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *AnswerCardInput) (AnswerCardOutput, error) {
		res, err := deps.AnswerCard.Execute(ctx, actor, in.Body.UnitID, in.Body.Answer)
		if err != nil {
			return AnswerCardOutput{}, err
		}
		return toAnswerCardOutput(res), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "updateVocabUnitStatus",
		Method:      http.MethodPatch,
		Path:        "/api/v1/vocab/units/{id}/status",
		Summary:     "Update a vocabulary item's memorization status",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateStatusInput) (UnitOutput, error) {
		unit, err := deps.UpdateStatus.Execute(ctx, actor, in.ID, vocab.Status(in.Body.Status))
		if err != nil {
			return UnitOutput{}, err
		}
		return toUnitOutput(unit), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "deleteVocabUnit",
		Method:      http.MethodDelete,
		Path:        "/api/v1/vocab/units/{id}",
		Summary:     "Delete a vocabulary item",
		Tags:        []string{"vocab"},
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteUnitInput) (DeleteUnitOutput, error) {
		if err := deps.DeleteUnit.Execute(ctx, actor, in.ID); err != nil {
			return DeleteUnitOutput{}, err
		}
		return DeleteUnitOutput{OK: true}, nil
	})
}
