package api

import (
	"fmt"

	usecases "github.com/els/backend/internal/application/account/use_cases"
	"github.com/els/backend/internal/domain/apps"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

func parseAccountID(s string) (iam.AccountID, error) {
	id, err := vo.ParseID(s)
	if err != nil {
		return iam.AccountID{}, fmt.Errorf("%w: account_id: %v", shared.ErrValidation, err)
	}
	return iam.AccountID{ID: id}, nil
}

func toAccountPictureOutput(a *iam.Account) AccountPictureOutput {
	return AccountPictureOutput{
		AccountID:  a.ID().String(),
		Email:      a.Email().String(),
		FirstName:  a.Name().First(),
		LastName:   a.Name().Last(),
		PictureURL: a.PictureURL(),
		Status:     a.Status().String(),
	}
}

func accountToMeOutput(a *iam.Account, actor *iam.Actor) MeOutput {
	return MeOutput{
		AccountID:        a.ID().String(),
		Email:            a.Email().String(),
		FirstName:        a.Name().First(),
		LastName:         a.Name().Last(),
		PictureURL:       a.PictureURL(),
		EnglishLevel:     a.EnglishLevel(),
		AboutMe:          a.AboutMe(),
		NativeLanguage:   a.NativeLanguage(),
		ShowTranslations: a.ShowTranslations(),
		AutoWordImages:   a.AutoWordImages(),
		Status:           a.Status().String(),
		Role:             actor.Role().String(),
		EntityID:         actor.EntityID().String(),
		IsGlobalAdmin:    actor.IsGlobalAdmin(),
	}
}

func toMeOutput(r usecases.MeResult, impersonationEnabled bool) MeOutput {
	acc := r.Actor.Account()
	return MeOutput{
		AccountID:            acc.ID().String(),
		Email:                acc.Email().String(),
		FirstName:            acc.Name().First(),
		LastName:             acc.Name().Last(),
		PictureURL:           acc.PictureURL(),
		EnglishLevel:         acc.EnglishLevel(),
		AboutMe:              acc.AboutMe(),
		NativeLanguage:       acc.NativeLanguage(),
		ShowTranslations:     acc.ShowTranslations(),
		AutoWordImages:       acc.AutoWordImages(),
		Status:               acc.Status().String(),
		Role:                 r.Actor.Role().String(),
		EntityID:             r.Actor.EntityID().String(),
		IsGlobalAdmin:        r.Actor.IsGlobalAdmin(),
		ImpersonationEnabled: impersonationEnabled && r.Actor.IsGlobalAdmin(),
	}
}

func toAppsOutput(r usecases.ListAppsResult) AppsOutput {
	items := make([]AppOutput, 0, len(r.Apps))
	for _, a := range r.Apps {
		items = append(items, toAppOutput(a))
	}
	return AppsOutput{Items: items, Total: len(items)}
}

func toAppOutput(a apps.App) AppOutput {
	return AppOutput{
		ID:          a.ID,
		Name:        a.Name,
		URI:         a.URI,
		Description: a.Description,
		Group:       a.Group,
		Disabled:    a.Disabled,
	}
}
