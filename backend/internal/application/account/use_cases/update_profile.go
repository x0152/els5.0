package usecases

import (
	"context"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

type UpdateProfileUseCase struct {
	accounts iam.AccountRepository
}

func NewUpdateProfileUseCase(accounts iam.AccountRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{accounts: accounts}
}

type UpdateProfileCommand struct {
	FirstName        string
	LastName         string
	EnglishLevel     string
	AboutMe          string
	NativeLanguage   string
	ShowTranslations bool
	SpeechStrictness float64
}

type UpdateProfileResult struct {
	Account *iam.Account
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, actor *iam.Actor, cmd UpdateProfileCommand) (UpdateProfileResult, error) {
	account, err := uc.accounts.GetByID(ctx, actor.Account().ID())
	if err != nil {
		return UpdateProfileResult{}, err
	}
	name, err := vo.NewPersonName(cmd.FirstName, cmd.LastName)
	if err != nil {
		return UpdateProfileResult{}, shared.Validation(err)
	}
	strictness := cmd.SpeechStrictness
	if strictness == 0 {
		strictness = account.SpeechStrictness()
	}
	if err := account.UpdateProfile(name, cmd.EnglishLevel, cmd.AboutMe, cmd.NativeLanguage, cmd.ShowTranslations, strictness); err != nil {
		return UpdateProfileResult{}, err
	}
	if err := uc.accounts.Update(ctx, account); err != nil {
		return UpdateProfileResult{}, err
	}
	return UpdateProfileResult{Account: account}, nil
}
