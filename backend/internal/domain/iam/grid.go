package iam

import (
	"fmt"

	"github.com/els/backend/internal/domain/grid"
	"github.com/els/backend/internal/domain/shared"
)

const (
	GridSourceAccounts grid.SourceID = "iam.accounts"
)

const (
	ColAccountEmail      grid.ColumnID = "email"
	ColAccountFirstName  grid.ColumnID = "first_name"
	ColAccountLastName   grid.ColumnID = "last_name"
	ColAccountStatus     grid.ColumnID = "status"
	ColAccountPictureURL grid.ColumnID = "picture_url"
	ColAccountID         grid.ColumnID = "account_id"
)

func AccountColumns() []grid.Column {
	minLen, maxLen := 1, 200
	return []grid.Column{
		{
			ID:       ColAccountEmail,
			Title:    "Email",
			Type:     grid.TypeEmail,
			Required: true,
			Constraints: &grid.Constraints{
				MinLength: &minLen,
				MaxLength: &maxLen,
				Unique:    true,
			},
		},
		{
			ID:       ColAccountFirstName,
			Title:    "First name",
			Type:     grid.TypeText,
			Required: true,
			Constraints: &grid.Constraints{
				MinLength: &minLen,
				MaxLength: &maxLen,
			},
		},
		{
			ID:       ColAccountLastName,
			Title:    "Last name",
			Type:     grid.TypeText,
			Required: true,
			Constraints: &grid.Constraints{
				MinLength: &minLen,
				MaxLength: &maxLen,
			},
		},
		{
			ID:       ColAccountStatus,
			Title:    "Status",
			Type:     grid.TypeEnum,
			Required: true,
			Enum: []grid.EnumOption{
				{Value: string(AccountStatusActive), Label: "Active"},
				{Value: string(AccountStatusBlocked), Label: "Blocked"},
				{Value: string(AccountStatusPendingPassword), Label: "Pending password"},
				{Value: string(AccountStatusNoAuth), Label: "No authentication"},
			},
		},
		{
			ID:       ColAccountPictureURL,
			Title:    "Avatar",
			Type:     grid.TypeText,
			Readonly: true,
		},
		{
			ID:       ColAccountID,
			Title:    "Account ID",
			Type:     grid.TypeText,
			Readonly: true,
		},
	}
}

func AccountCells(side AccountSide) map[grid.ColumnID]any {
	cells := map[grid.ColumnID]any{
		ColAccountEmail:     side.Email().String(),
		ColAccountFirstName: side.FirstName(),
		ColAccountLastName:  side.LastName(),
		ColAccountStatus:    string(side.Status()),
		ColAccountID:        side.AccountID().String(),
	}
	if url := side.Account().PictureURL(); url != "" {
		cells[ColAccountPictureURL] = url
	} else {
		cells[ColAccountPictureURL] = nil
	}
	return cells
}

func ApplyAccountPatch(side AccountSide, data map[grid.ColumnID]any) (handled map[grid.ColumnID]struct{}, err error) {
	handled = map[grid.ColumnID]struct{}{}

	first := side.FirstName()
	last := side.LastName()
	renameDirty := false

	for id, v := range data {
		switch id {
		case ColAccountEmail:
			s, e := asString(v)
			if e != nil {
				return nil, shared.Validation(fmt.Errorf("column %q: %w", id, e))
			}
			if e := side.ChangeEmail(s); e != nil {
				return nil, e
			}
			handled[id] = struct{}{}
		case ColAccountFirstName:
			s, e := asString(v)
			if e != nil {
				return nil, shared.Validation(fmt.Errorf("column %q: %w", id, e))
			}
			first = s
			renameDirty = true
			handled[id] = struct{}{}
		case ColAccountLastName:
			s, e := asString(v)
			if e != nil {
				return nil, shared.Validation(fmt.Errorf("column %q: %w", id, e))
			}
			last = s
			renameDirty = true
			handled[id] = struct{}{}
		case ColAccountStatus:
			s, e := asString(v)
			if e != nil {
				return nil, shared.Validation(fmt.Errorf("column %q: %w", id, e))
			}
			status, e := ParseAccountStatus(s)
			if e != nil {
				return nil, shared.Validation(fmt.Errorf("column %q: %w", id, e))
			}
			if e := side.SetStatus(status); e != nil {
				return nil, e
			}
			handled[id] = struct{}{}
		}
	}

	if renameDirty {
		if err := side.Rename(first, last); err != nil {
			return nil, err
		}
	}
	return handled, nil
}

func asString(v any) (string, error) {
	if v == nil {
		return "", fmt.Errorf("must not be null")
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("must be a string")
	}
	return s, nil
}
