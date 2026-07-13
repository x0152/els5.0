package api

import (
	"fmt"

	usecases "github.com/els/backend/internal/application/admin/use_cases"
	"github.com/els/backend/internal/domain/admin"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

func toCreateAdminCommand(in *CreateAdminInput) usecases.CreateAdminCommand {
	return usecases.CreateAdminCommand{
		Email:     in.Body.Email,
		FirstName: in.Body.FirstName,
		LastName:  in.Body.LastName,
	}
}

func toListAdminsQuery(in *ListAdminsInput) usecases.ListAdminsQuery {
	return usecases.ListAdminsQuery{Limit: in.Limit, Offset: in.Offset}
}

func toGetAdminQuery(in *GetAdminInput) (usecases.GetAdminQuery, error) {
	id, err := parseAdminID(in.ID)
	if err != nil {
		return usecases.GetAdminQuery{}, err
	}
	return usecases.GetAdminQuery{ID: id}, nil
}

func parseAdminID(s string) (admin.ID, error) {
	id, err := vo.ParseID(s)
	if err != nil {
		return admin.ID{}, fmt.Errorf("%w: id: %v", shared.ErrValidation, err)
	}
	return admin.ID{ID: id}, nil
}

func toAdminOutput(a *admin.Administrator) AdminOutput {
	return AdminOutput{
		ID:         a.ID().String(),
		AccountID:  a.AccountID().String(),
		Email:      a.Email().String(),
		FirstName:  a.FirstName(),
		LastName:   a.LastName(),
		Status:     string(a.Status()),
		PictureURL: a.Account().PictureURL(),
		CreatedAt:  a.CreatedAt(),
		UpdatedAt:  a.UpdatedAt(),
	}
}

func toAdminsOutput(r usecases.ListAdminsResult) AdminsOutput {
	items := make([]AdminOutput, 0, len(r.Admins))
	for _, a := range r.Admins {
		items = append(items, toAdminOutput(a))
	}
	return AdminsOutput{
		Items:  items,
		Total:  r.Total,
		Limit:  r.Limit,
		Offset: r.Offset,
	}
}
