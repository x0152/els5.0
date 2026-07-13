package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/learn/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/practice"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator *authx.Authenticator

	ListBooks       *usecases.ListBooksUseCase
	ListChapters    *usecases.ListChaptersUseCase
	GetChapter      *usecases.GetChapterUseCase
	CreateChapter   *usecases.CreateChapterUseCase
	UpdateChapter   *usecases.UpdateChapterUseCase
	DeleteChapter   *usecases.DeleteChapterUseCase
	GenerateChapter *usecases.GenerateChapterUseCase

	EnsureIllustration *usecases.EnsureIllustrationUseCase

	ListVariants    *usecases.ListVariantsUseCase
	GenerateVariant *usecases.GenerateVariantUseCase
	DeleteVariant   *usecases.DeleteVariantUseCase
	GetProgress     *usecases.GetProgressUseCase
	SaveProgress    *usecases.SaveProgressUseCase
	ResetProgress   *usecases.ResetProgressUseCase
	CheckFree       *usecases.CheckFreeUseCase
}

func Register(api huma.API, deps Deps) {
	registerChapters(api, deps)
	registerIllustrations(api, deps)
	registerPractice(api, deps)
}

func registerChapters(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listLearnBooks",
		Method:      http.MethodGet,
		Path:        "/api/v1/books",
		Summary:     "List available books with series, level and description",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListBooksInput) (BookListOutput, error) {
		books, err := deps.ListBooks.Execute(ctx, actor)
		if err != nil {
			return BookListOutput{}, err
		}
		return toBookListOutput(books), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listChapters",
		Method:      http.MethodGet,
		Path:        "/api/v1/books/{book}/chapters",
		Summary:     "List chapters of a book",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListChaptersInput) (ChaptersOutput, error) {
		chapters, err := deps.ListChapters.Execute(ctx, actor, in.Book)
		if err != nil {
			return ChaptersOutput{}, err
		}
		return toChaptersOutput(chapters), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getChapter",
		Method:      http.MethodGet,
		Path:        "/api/v1/books/{book}/chapters/{number}",
		Summary:     "Get a single chapter by number",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *GetChapterInput) (ChapterOutput, error) {
		chapter, err := deps.GetChapter.Execute(ctx, actor, in.Book, in.Number)
		if err != nil {
			return ChapterOutput{}, err
		}
		return toChapterOutput(chapter), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "createChapter",
		Method:        http.MethodPost,
		Path:          "/api/v1/books/{book}/chapters",
		Summary:       "Create a chapter (global admin only)",
		Tags:          []string{"learn"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, actor *iam.Actor, in *CreateChapterInput) (ChapterOutput, error) {
		chapter, err := deps.CreateChapter.Execute(ctx, actor, toChapter(in.Book, in.Body))
		if err != nil {
			return ChapterOutput{}, err
		}
		return toChapterOutput(chapter), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "updateChapter",
		Method:      http.MethodPut,
		Path:        "/api/v1/books/{book}/chapters/{number}",
		Summary:     "Update a chapter (global admin only)",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateChapterInput) (ChapterOutput, error) {
		chapter := toChapter(in.Book, in.Body)
		chapter.Number = in.Number
		updated, err := deps.UpdateChapter.Execute(ctx, actor, chapter)
		if err != nil {
			return ChapterOutput{}, err
		}
		return toChapterOutput(updated), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "generateChapter",
		Method:        http.MethodPost,
		Path:          "/api/v1/books/{book}/chapters/generate",
		Summary:       "Generate a chapter (theory + exercises) on a topic with the LLM (global admin only)",
		Tags:          []string{"learn"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, actor *iam.Actor, in *GenerateChapterInput) (ChapterOutput, error) {
		chapter, err := deps.GenerateChapter.Execute(ctx, actor, in.Book, in.Body.Topic)
		if err != nil {
			return ChapterOutput{}, err
		}
		return toChapterOutput(chapter), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "deleteChapter",
		Method:      http.MethodDelete,
		Path:        "/api/v1/books/{book}/chapters/{number}",
		Summary:     "Delete a chapter (global admin only)",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteChapterInput) (DeleteChapterOutput, error) {
		if err := deps.DeleteChapter.Execute(ctx, actor, in.Book, in.Number); err != nil {
			return DeleteChapterOutput{}, err
		}
		return DeleteChapterOutput{OK: true}, nil
	})
}

func registerIllustrations(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "ensureIllustration",
		Method:      http.MethodPost,
		Path:        "/api/v1/illustrations",
		Summary:     "Trigger or poll generation of an illustration from a prompt",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, _ *iam.Actor, in *EnsureInput) (IllustrationOutput, error) {
		return toIllustrationOutput(deps.EnsureIllustration.Execute(ctx, toEnsureCommand(in.Body))), nil
	})
}

func registerPractice(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listPracticeVariants",
		Method:      http.MethodGet,
		Path:        "/api/v1/practice/{kind}/{number}/variants",
		Summary:     "List generated practice variants for a chapter",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *ListVariantsInput) (VariantsOutput, error) {
		variants, err := deps.ListVariants.Execute(ctx, actor, practice.Kind(in.Kind), in.Number)
		if err != nil {
			return VariantsOutput{}, err
		}
		return toVariantsOutput(variants), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "generatePracticeVariant",
		Method:        http.MethodPost,
		Path:          "/api/v1/practice/{kind}/{number}/variants",
		Summary:       "Generate a new practice variant with the LLM",
		Tags:          []string{"learn"},
		DefaultStatus: http.StatusCreated,
	}, func(ctx context.Context, actor *iam.Actor, in *GenerateVariantInput) (VariantSchema, error) {
		variant, err := deps.GenerateVariant.Execute(ctx, actor, practice.Kind(in.Kind), in.Number)
		if err != nil {
			return VariantSchema{}, err
		}
		return toVariantSchema(variant), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "deletePracticeVariant",
		Method:      http.MethodDelete,
		Path:        "/api/v1/practice/variants/{id}",
		Summary:     "Delete a generated practice variant",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteVariantInput) (PracticeOKOutput, error) {
		if err := deps.DeleteVariant.Execute(ctx, actor, in.ID); err != nil {
			return PracticeOKOutput{}, err
		}
		return PracticeOKOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getPracticeProgress",
		Method:      http.MethodGet,
		Path:        "/api/v1/practice/{kind}/{number}/progress",
		Summary:     "Get saved answers and completion for a chapter variant",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *GetProgressInput) (ProgressOutput, error) {
		p, err := deps.GetProgress.Execute(ctx, actor, practice.Kind(in.Kind), in.Number, variantKey(in.Variant))
		if err != nil {
			return ProgressOutput{}, err
		}
		return toProgressOutput(p), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "savePracticeProgress",
		Method:      http.MethodPut,
		Path:        "/api/v1/practice/{kind}/{number}/progress",
		Summary:     "Save answers and completion for a chapter variant",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *SaveProgressInput) (PracticeOKOutput, error) {
		if err := deps.SaveProgress.Execute(ctx, actor, practice.Kind(in.Kind), in.Number, variantKey(in.Body.Variant), toProgress(in.Body)); err != nil {
			return PracticeOKOutput{}, err
		}
		return PracticeOKOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "resetPracticeProgress",
		Method:      http.MethodDelete,
		Path:        "/api/v1/practice/{kind}/{number}/progress",
		Summary:     "Reset saved answers for a chapter variant",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *ResetProgressInput) (PracticeOKOutput, error) {
		if err := deps.ResetProgress.Execute(ctx, actor, practice.Kind(in.Kind), in.Number, variantKey(in.Variant)); err != nil {
			return PracticeOKOutput{}, err
		}
		return PracticeOKOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "checkPracticeAnswer",
		Method:      http.MethodPost,
		Path:        "/api/v1/practice/check",
		Summary:     "Validate a free-form answer with the LLM",
		Tags:        []string{"learn"},
	}, func(ctx context.Context, actor *iam.Actor, in *CheckInput) (CheckOutput, error) {
		res, err := deps.CheckFree.Execute(ctx, actor, usecases.CheckFreeCommand{
			Kind:        practice.Kind(in.Body.Kind),
			Number:      in.Body.Number,
			Instruction: in.Body.Instruction,
			Answer:      in.Body.Answer,
		})
		if err != nil {
			return CheckOutput{}, err
		}
		return toCheckOutput(res), nil
	})
}
