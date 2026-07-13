package usecases

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/shared/ports"
)

type UploadAccountPictureUseCase struct {
	accounts iam.AccountRepository
	storage  media.Storage
	sniffer  ports.ContentSniffer
	urls     media.PublicURL
	policy   media.UploadPolicy
	bucket   string
	logger   *slog.Logger
}

type UploadAccountPictureConfig struct {
	Bucket       string
	MaxSizeBytes int64
}

func NewUploadAccountPictureUseCase(
	accounts iam.AccountRepository,
	storage media.Storage,
	sniffer ports.ContentSniffer,
	urls media.PublicURL,
	cfg UploadAccountPictureConfig,
	logger *slog.Logger,
) *UploadAccountPictureUseCase {
	if cfg.Bucket == "" {
		cfg.Bucket = "avatars"
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &UploadAccountPictureUseCase{
		accounts: accounts,
		storage:  storage,
		sniffer:  sniffer,
		urls:     urls,
		policy:   media.ImagePolicy(cfg.MaxSizeBytes),
		bucket:   cfg.Bucket,
		logger:   logger,
	}
}

type UploadAccountPictureCommand struct {
	TargetID    iam.AccountID
	Reader      io.Reader
	Size        int64
	ContentType string
	Filename    string
}

type UploadAccountPictureResult struct {
	Account *iam.Account
}

func (uc *UploadAccountPictureUseCase) Execute(
	ctx context.Context,
	actor *iam.Actor,
	cmd UploadAccountPictureCommand,
) (UploadAccountPictureResult, error) {
	target := cmd.TargetID
	if target.IsZero() && actor != nil {
		target = actor.Account().ID()
	}
	if err := iam.RequireSelfOrGlobalAdmin(actor, target); err != nil {
		return UploadAccountPictureResult{}, err
	}

	existing, err := uc.accounts.GetByID(ctx, target)
	if err != nil {
		return UploadAccountPictureResult{}, err
	}

	body, mime, err := uc.sniffer.Sniff(cmd.Reader)
	if err != nil {
		return UploadAccountPictureResult{}, err
	}
	ext, err := uc.policy.Validate(cmd.Size, mime, cmd.Filename)
	if err != nil {
		return UploadAccountPictureResult{}, err
	}

	newPath, err := media.NewPath(fmt.Sprintf("%s/accounts/%s/%s%s", uc.bucket, existing.ID().String(), uuid.NewString(), ext))
	if err != nil {
		return UploadAccountPictureResult{}, err
	}
	if err := uc.storage.Put(ctx, newPath, body, media.PutOptions{
		ContentType: mime,
		Size:        cmd.Size,
	}); err != nil {
		return UploadAccountPictureResult{}, err
	}

	publicURL := uc.urls.Build(newPath)
	if err := existing.ChangePictureURL(publicURL); err != nil {
		_ = uc.storage.Delete(ctx, newPath)
		return UploadAccountPictureResult{}, err
	}
	previousURL, err := uc.accounts.UpdatePicture(ctx, existing)
	if err != nil {
		_ = uc.storage.Delete(ctx, newPath)
		return UploadAccountPictureResult{}, err
	}

	if previousURL != "" && previousURL != publicURL {
		if oldPath, ok := uc.urls.ParsePath(previousURL); ok {
			if derr := uc.storage.Delete(ctx, oldPath); derr != nil {
				uc.logger.Warn("delete previous account picture failed",
					slog.String("account_id", existing.ID().String()),
					slog.String("path", oldPath.String()),
					slog.String("err", derr.Error()),
				)
			}
		}
	}

	return UploadAccountPictureResult{Account: existing}, nil
}
