package usecases_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	usecases "github.com/els/backend/internal/application/account/use_cases"
	iamdom "github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/test"
	iamtest "github.com/els/backend/internal/utils/test/iam"
)

const (
	testBucket   = "avatars"
	testCDNBase  = "https://cdn.example"
	testMaxBytes = 5 * 1024 * 1024
)

var pngBytes = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}

func newUploadAccountUC(repo iamdom.AccountRepository, storage media.Storage) *usecases.UploadAccountPictureUseCase {
	return usecases.NewUploadAccountPictureUseCase(
		repo,
		storage,
		&sniffStub{},
		media.NewPublicURL(testCDNBase),
		usecases.UploadAccountPictureConfig{Bucket: testBucket, MaxSizeBytes: testMaxBytes},
		slog.Default(),
	)
}

func validUploadCmd(target iamdom.AccountID) usecases.UploadAccountPictureCommand {
	return usecases.UploadAccountPictureCommand{
		TargetID:    target,
		Reader:      bytes.NewReader(pngBytes),
		Size:        int64(len(pngBytes)),
		ContentType: "image/png",
		Filename:    "avatar.png",
	}
}

func TestUploadAccountPicture_Unauthorized(t *testing.T) {
	uc := newUploadAccountUC(&accountRepoStub{}, &storageStub{})
	_, err := uc.Execute(context.Background(), nil, validUploadCmd(iamdom.NewAccountID()))
	test.ErrIs(t, err, shared.ErrUnauthorized)
}

func TestUploadAccountPicture_NonAdminCannotChangeOther(t *testing.T) {
	owner := iamtest.NewAccount(t).Build(t)
	repo := &accountRepoStub{
		getByID: func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return owner, nil },
	}
	uc := newUploadAccountUC(repo, &storageStub{})

	expert := iamtest.Expert(t, owner.ID().ID)
	_, err := uc.Execute(context.Background(), expert, validUploadCmd(owner.ID()))

	test.ErrIs(t, err, shared.ErrForbidden)
}

func TestUploadAccountPicture_NotFound(t *testing.T) {
	repo := &accountRepoStub{
		getByID: func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return nil, shared.ErrNotFound },
	}
	uc := newUploadAccountUC(repo, &storageStub{})

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), validUploadCmd(iamdom.NewAccountID()))

	test.ErrIs(t, err, shared.ErrNotFound)
}

func TestUploadAccountPicture_ValidationCases(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*usecases.UploadAccountPictureCommand)
	}{
		{name: "size_zero", mut: func(c *usecases.UploadAccountPictureCommand) { c.Size = 0 }},
		{name: "size_over_limit", mut: func(c *usecases.UploadAccountPictureCommand) { c.Size = testMaxBytes + 1 }},
		{name: "unsupported_content_type", mut: func(c *usecases.UploadAccountPictureCommand) { c.Reader = strings.NewReader("not an image at all") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			existing := iamtest.NewAccount(t).Build(t)
			repo := &accountRepoStub{
				getByID: func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return existing, nil },
			}
			storage := &storageStub{}
			uc := newUploadAccountUC(repo, storage)
			cmd := validUploadCmd(existing.ID())
			tc.mut(&cmd)

			_, err := uc.Execute(context.Background(), iamtest.Admin(t), cmd)

			test.ErrIs(t, err, shared.ErrValidation)
			if len(storage.putCalls) != 0 || len(repo.updatePictureCalls) != 0 {
				t.Errorf("expected no Put/Update on validation error")
			}
		})
	}
}

func TestUploadAccountPicture_OK_AdminUpdatesOther(t *testing.T) {
	existing := iamtest.NewAccount(t).Build(t)
	repo := &accountRepoStub{
		getByID: func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return existing, nil },
	}
	storage := &storageStub{}
	uc := newUploadAccountUC(repo, storage)

	res, err := uc.Execute(context.Background(), iamtest.Admin(t), validUploadCmd(existing.ID()))

	test.NoErr(t, err)
	if len(storage.putCalls) != 1 {
		t.Fatalf("expected one Put call, got %d", len(storage.putCalls))
	}
	put := storage.putCalls[0]
	if put.Path.Bucket() != testBucket {
		t.Errorf("expected bucket=%s, got %s", testBucket, put.Path.Bucket())
	}
	if !strings.HasPrefix(put.Path.Key(), "accounts/"+existing.ID().String()+"/") {
		t.Errorf("expected key starts with accounts/<id>/, got %s", put.Path.Key())
	}
	if !strings.HasSuffix(res.Account.PictureURL(), ".png") {
		t.Errorf("expected .png url, got %s", res.Account.PictureURL())
	}
}

func TestUploadAccountPicture_OK_DeletesPreviousPicture(t *testing.T) {
	urls := media.NewPublicURL(testCDNBase)
	prevPath, err := media.NewPath(testBucket + "/accounts/old/old.png")
	test.Must(t, err)
	prevURL := urls.Build(prevPath)

	existing := iamtest.NewAccount(t).Build(t)
	repo := &accountRepoStub{
		getByID:           func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return existing, nil },
		updatePicturePrev: prevURL,
	}
	storage := &storageStub{}
	uc := newUploadAccountUC(repo, storage)

	_, err = uc.Execute(context.Background(), iamtest.Admin(t), validUploadCmd(existing.ID()))

	test.NoErr(t, err)
	if len(storage.deleteCalls) != 1 || storage.deleteCalls[0].String() != prevPath.String() {
		t.Errorf("expected delete of previous path, got %v", storage.deleteCalls)
	}
}

func TestUploadAccountPicture_StorageFails(t *testing.T) {
	boom := errors.New("s3 down")
	existing := iamtest.NewAccount(t).Build(t)
	repo := &accountRepoStub{
		getByID: func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return existing, nil },
	}
	storage := &storageStub{putErr: boom}
	uc := newUploadAccountUC(repo, storage)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), validUploadCmd(existing.ID()))

	if !errors.Is(err, boom) {
		t.Errorf("expected put error to propagate, got %v", err)
	}
	if len(repo.updatePictureCalls) != 0 {
		t.Errorf("expected no UpdatePicture when Put failed")
	}
}

func TestUploadAccountPicture_RejectsLyingContentType(t *testing.T) {
	existing := iamtest.NewAccount(t).Build(t)
	repo := &accountRepoStub{
		getByID: func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return existing, nil },
	}
	storage := &storageStub{}
	uc := newUploadAccountUC(repo, storage)
	cmd := validUploadCmd(existing.ID())
	cmd.Reader = strings.NewReader("<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>")
	cmd.ContentType = "image/png"
	cmd.Filename = "evil.png"

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), cmd)

	test.ErrIs(t, err, shared.ErrValidation)
	if len(storage.putCalls) != 0 || len(repo.updatePictureCalls) != 0 {
		t.Errorf("expected no Put/UpdatePicture when Content-Type is forged")
	}
}

func TestUploadAccountPicture_RepoFails_RollsBackPut(t *testing.T) {
	boom := errors.New("update failed")
	existing := iamtest.NewAccount(t).Build(t)
	repo := &accountRepoStub{
		getByID:          func(_ context.Context, _ iamdom.AccountID) (*iamdom.Account, error) { return existing, nil },
		updatePictureErr: boom,
	}
	storage := &storageStub{}
	uc := newUploadAccountUC(repo, storage)

	_, err := uc.Execute(context.Background(), iamtest.Admin(t), validUploadCmd(existing.ID()))

	if !errors.Is(err, boom) {
		t.Errorf("expected update error to propagate, got %v", err)
	}
	if len(storage.deleteCalls) != 1 || storage.deleteCalls[0].String() != storage.putCalls[0].Path.String() {
		t.Errorf("expected rollback to delete the just-put path")
	}
}
