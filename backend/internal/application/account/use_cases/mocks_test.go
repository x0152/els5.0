package usecases_test

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared/ports"
	"github.com/els/backend/internal/domain/shared/vo"
)

var _ ports.ContentSniffer = (*sniffStub)(nil)

type sniffStub struct {
	forceMime string
	forceErr  error
}

func (s *sniffStub) Sniff(r io.Reader) (io.Reader, string, error) {
	if s.forceErr != nil {
		return nil, "", s.forceErr
	}
	br := bufio.NewReaderSize(r, 512)
	head, _ := br.Peek(512)
	if s.forceMime != "" {
		return br, s.forceMime, nil
	}
	mime := http.DetectContentType(head)
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = mime[:i]
	}
	return br, strings.ToLower(strings.TrimSpace(mime)), nil
}

type accountRepoStub struct {
	getByID    func(ctx context.Context, id iam.AccountID) (*iam.Account, error)
	getByIDErr error

	updatePictureCalls []*iam.Account
	updatePicturePrev  string
	updatePictureErr   error
}

func (r *accountRepoStub) Create(_ context.Context, _ *iam.Account) error { return nil }
func (r *accountRepoStub) Update(_ context.Context, _ *iam.Account) error { return nil }
func (r *accountRepoStub) UpdatePicture(_ context.Context, a *iam.Account) (string, error) {
	r.updatePictureCalls = append(r.updatePictureCalls, a)
	return r.updatePicturePrev, r.updatePictureErr
}
func (r *accountRepoStub) Delete(_ context.Context, _ iam.AccountID) error { return nil }
func (r *accountRepoStub) GetByID(ctx context.Context, id iam.AccountID) (*iam.Account, error) {
	if r.getByID != nil {
		return r.getByID(ctx, id)
	}
	return nil, r.getByIDErr
}
func (r *accountRepoStub) GetByIDs(_ context.Context, _ []iam.AccountID) ([]*iam.Account, error) {
	return nil, nil
}
func (r *accountRepoStub) GetByEmail(_ context.Context, _ vo.Email) (*iam.Account, error) {
	return nil, nil
}
func (r *accountRepoStub) SearchByEmail(_ context.Context, _ string, _ int32) ([]*iam.Account, error) {
	return nil, nil
}
func (r *accountRepoStub) ExistsEmail(_ context.Context, _ vo.Email) (bool, error) {
	return false, nil
}

type storageStub struct {
	putCalls    []storagePut
	putErr      error
	deleteCalls []media.Path
	deleteErr   error
}

type storagePut struct {
	Path media.Path
	Opts media.PutOptions
	Body []byte
}

func (s *storageStub) Put(_ context.Context, p media.Path, r io.Reader, opts media.PutOptions) error {
	body, _ := io.ReadAll(r)
	s.putCalls = append(s.putCalls, storagePut{Path: p, Opts: opts, Body: body})
	return s.putErr
}
func (s *storageStub) Get(_ context.Context, _ media.Path) (io.ReadCloser, media.Metadata, error) {
	return nil, media.Metadata{}, errors.New("storageStub.Get not implemented")
}
func (s *storageStub) Delete(_ context.Context, p media.Path) error {
	s.deleteCalls = append(s.deleteCalls, p)
	return s.deleteErr
}
