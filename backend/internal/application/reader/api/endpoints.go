package api

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/danielgtaylor/huma/v2"

	usecases "github.com/els/backend/internal/application/reader/use_cases"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	authx "github.com/els/backend/internal/utils/auth"
)

type Deps struct {
	Authenticator    *authx.Authenticator
	ListBooks        *usecases.ListBooksUseCase
	GetBook          *usecases.GetBookUseCase
	UploadBook       *usecases.UploadBookUseCase
	ImportArticle    *usecases.ImportArticleUseCase
	UpdateBook       *usecases.UpdateBookUseCase
	DeleteBook       *usecases.DeleteBookUseCase
	SaveProgress     *usecases.SaveProgressUseCase
	ListCollections  *usecases.ListCollectionsUseCase
	UpdateCollection *usecases.UpdateCollectionUseCase
	MediaURLs        media.PublicURL
	TempDir          string
}

func readAsset(f huma.FormFile, limit int64) *usecases.UploadAsset {
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil || len(data) == 0 {
		return nil
	}
	return &usecases.UploadAsset{Data: data, ContentType: f.ContentType, Filename: f.Filename}
}

func Register(api huma.API, deps Deps) {
	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listBooks",
		Method:      http.MethodGet,
		Path:        "/api/v1/reader/books",
		Summary:     "List the reader's books",
		Tags:        []string{"reader"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListBooksInput) (BooksOutput, error) {
		list, err := deps.ListBooks.Execute(ctx, actor)
		if err != nil {
			return BooksOutput{}, err
		}
		return toBooksOutput(list, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "getBook",
		Method:      http.MethodGet,
		Path:        "/api/v1/reader/books/{id}",
		Summary:     "Get a book with its content URL and reading position",
		Tags:        []string{"reader"},
	}, func(ctx context.Context, actor *iam.Actor, in *GetBookInput) (BookOutput, error) {
		book, err := deps.GetBook.Execute(ctx, actor, in.ID)
		if err != nil {
			return BookOutput{}, err
		}
		return toBookOutput(book, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "uploadBook",
		Method:        http.MethodPost,
		Path:          "/api/v1/reader/books",
		Summary:       "Upload a book (FB2/EPUB/HTML); it is converted to HTML in the background",
		Tags:          []string{"reader"},
		DefaultStatus: http.StatusAccepted,
	}, func(ctx context.Context, actor *iam.Actor, in *UploadBookInput) (BookOutput, error) {
		if deps.UploadBook == nil {
			return BookOutput{}, huma.Error503ServiceUnavailable("book upload is not configured")
		}
		form := in.RawBody.Data()
		if form == nil || !form.File.IsSet {
			return BookOutput{}, huma.Error400BadRequest("book file is required")
		}
		defer form.File.Close()

		tempPath, err := saveTemp(deps.TempDir, form.File, form.File.Filename)
		if err != nil {
			return BookOutput{}, huma.Error500InternalServerError("failed to buffer upload")
		}

		cmd := usecases.UploadBookCommand{
			Title:       formValue(in.RawBody.Form, "title"),
			Author:      formValue(in.RawBody.Form, "author"),
			Description: formValue(in.RawBody.Form, "description"),
			Filename:    form.File.Filename,
			TempPath:    tempPath,
			Kind:        formValue(in.RawBody.Form, "kind"),
			GroupTitle:  formValue(in.RawBody.Form, "group_title"),
		}
		if form.Cover.IsSet {
			cmd.Cover = readAsset(form.Cover, 20<<20)
		}

		book, err := deps.UploadBook.Execute(ctx, actor, cmd)
		if err != nil {
			os.Remove(tempPath)
			return BookOutput{}, err
		}
		return toBookOutput(book, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID:   "importArticle",
		Method:        http.MethodPost,
		Path:          "/api/v1/reader/articles",
		Summary:       "Import an article by URL; readable text and images are extracted and converted in the background",
		Tags:          []string{"reader"},
		DefaultStatus: http.StatusAccepted,
	}, func(ctx context.Context, actor *iam.Actor, in *ImportArticleInput) (BookOutput, error) {
		if deps.ImportArticle == nil {
			return BookOutput{}, huma.Error503ServiceUnavailable("article import is not configured")
		}
		book, err := deps.ImportArticle.Execute(ctx, actor, usecases.ImportArticleCommand{
			URL:        in.Body.URL,
			GroupTitle: in.Body.GroupTitle,
		})
		if err != nil {
			return BookOutput{}, err
		}
		return toBookOutput(book, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "updateBook",
		Method:      http.MethodPatch,
		Path:        "/api/v1/reader/books/{id}",
		Summary:     "Update a book's title, author, description and cover",
		Tags:        []string{"reader"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateBookInput) (BookOutput, error) {
		if deps.UpdateBook == nil {
			return BookOutput{}, huma.Error503ServiceUnavailable("book editing is not configured")
		}
		form := in.RawBody.Data()
		cmd := usecases.UpdateBookCommand{
			Title:       formValue(in.RawBody.Form, "title"),
			Author:      formValue(in.RawBody.Form, "author"),
			Description: formValue(in.RawBody.Form, "description"),
		}
		if form != nil && form.Cover.IsSet {
			cmd.Cover = readAsset(form.Cover, 20<<20)
		}
		book, err := deps.UpdateBook.Execute(ctx, actor, in.ID, cmd)
		if err != nil {
			return BookOutput{}, err
		}
		return toBookOutput(book, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "saveBookProgress",
		Method:      http.MethodPut,
		Path:        "/api/v1/reader/books/{id}/progress",
		Summary:     "Save the reading position for a book",
		Tags:        []string{"reader"},
	}, func(ctx context.Context, actor *iam.Actor, in *SaveBookProgressInput) (SaveBookProgressOutput, error) {
		if err := deps.SaveProgress.Execute(ctx, actor, in.ID, in.Body.Position); err != nil {
			return SaveBookProgressOutput{}, err
		}
		return SaveBookProgressOutput{OK: true}, nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "listCollections",
		Method:      http.MethodGet,
		Path:        "/api/v1/reader/collections",
		Summary:     "List article collections (cover and description)",
		Tags:        []string{"reader"},
	}, func(ctx context.Context, actor *iam.Actor, _ *ListCollectionsInput) (CollectionsOutput, error) {
		list, err := deps.ListCollections.Execute(ctx, actor)
		if err != nil {
			return CollectionsOutput{}, err
		}
		return toCollectionsOutput(list, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "updateCollection",
		Method:      http.MethodPatch,
		Path:        "/api/v1/reader/collections",
		Summary:     "Update an article collection's title, description and cover",
		Tags:        []string{"reader"},
	}, func(ctx context.Context, actor *iam.Actor, in *UpdateCollectionInput) (CollectionOutput, error) {
		if deps.UpdateCollection == nil {
			return CollectionOutput{}, huma.Error503ServiceUnavailable("collection editing is not configured")
		}
		form := in.RawBody.Data()
		cmd := usecases.UpdateCollectionCommand{
			Title:       formValue(in.RawBody.Form, "title"),
			NewTitle:    formValue(in.RawBody.Form, "new_title"),
			Description: formValue(in.RawBody.Form, "description"),
		}
		if form != nil && form.Cover.IsSet {
			cmd.Cover = readAsset(form.Cover, 20<<20)
		}
		collection, err := deps.UpdateCollection.Execute(ctx, actor, cmd)
		if err != nil {
			return CollectionOutput{}, err
		}
		return toCollectionOutput(collection, deps.MediaURLs), nil
	})

	authx.Authed(api, deps.Authenticator, huma.Operation{
		OperationID: "deleteBook",
		Method:      http.MethodDelete,
		Path:        "/api/v1/reader/books/{id}",
		Summary:     "Delete a book",
		Tags:        []string{"reader"},
	}, func(ctx context.Context, actor *iam.Actor, in *DeleteBookInput) (DeleteBookOutput, error) {
		if err := deps.DeleteBook.Execute(ctx, actor, in.ID); err != nil {
			return DeleteBookOutput{}, err
		}
		return DeleteBookOutput{OK: true}, nil
	})
}

func saveTemp(dir string, r io.Reader, filename string) (string, error) {
	f, err := os.CreateTemp(dir, "book-upload-*"+filepath.Ext(filename))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
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
