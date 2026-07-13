package lookups

import (
	"context"
	"errors"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

type iamAccountsAdapter struct {
	accounts iam.AccountRepository
}

func NewIAMAccountsSource(accounts iam.AccountRepository) Source {
	return Source{
		ID:      iam.GridSourceAccounts,
		Adapter: &iamAccountsAdapter{accounts: accounts},
	}
}

func (a *iamAccountsAdapter) Hydrate(ctx context.Context, _ *iam.Actor, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	ids := make([]iam.AccountID, 0, len(keys))
	for _, k := range keys {
		id, err := vo.ParseID(k)
		if err != nil {
			continue
		}
		ids = append(ids, iam.AccountID{ID: id})
	}
	if len(ids) == 0 {
		return out, nil
	}
	accounts, err := a.accounts.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, acc := range accounts {
		out[acc.ID().String()] = acc.Email().String()
	}
	return out, nil
}

func (a *iamAccountsAdapter) Resolve(ctx context.Context, _ *iam.Actor, values []string) ([]Resolution, []string, error) {
	resolved := make([]Resolution, 0, len(values))
	unresolved := make([]string, 0)

	ids := make([]iam.AccountID, 0, len(values))
	idByKey := make(map[string]string, len(values))
	emails := make([]string, 0, len(values))
	emailInputs := make(map[string]string, len(values))
	for _, v := range values {
		if id, err := vo.ParseID(v); err == nil {
			acc := iam.AccountID{ID: id}
			ids = append(ids, acc)
			idByKey[acc.String()] = v
			continue
		}
		if email, err := vo.NewEmail(v); err == nil {
			emails = append(emails, email.String())
			emailInputs[email.String()] = v
			continue
		}
		unresolved = append(unresolved, v)
	}

	if len(ids) > 0 {
		accounts, err := a.accounts.GetByIDs(ctx, ids)
		if err != nil {
			return nil, nil, err
		}
		seen := make(map[string]struct{}, len(accounts))
		for _, acc := range accounts {
			key := acc.ID().String()
			seen[key] = struct{}{}
			resolved = append(resolved, Resolution{
				Input:     idByKey[key],
				Key:       key,
				Label:     acc.Email().String(),
				MatchedBy: MatchByKey,
			})
		}
		for k, input := range idByKey {
			if _, ok := seen[k]; !ok {
				unresolved = append(unresolved, input)
			}
		}
	}

	for _, emailStr := range emails {
		email, _ := vo.NewEmail(emailStr)
		acc, err := a.accounts.GetByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, shared.ErrNotFound) {
				unresolved = append(unresolved, emailInputs[emailStr])
				continue
			}
			return nil, nil, err
		}
		resolved = append(resolved, Resolution{
			Input:     emailInputs[emailStr],
			Key:       acc.ID().String(),
			Label:     acc.Email().String(),
			MatchedBy: MatchByLabel,
		})
	}

	return resolved, unresolved, nil
}

func (a *iamAccountsAdapter) Search(ctx context.Context, _ *iam.Actor, q string, limit int32, _ string) (Page, error) {
	if limit <= 0 {
		limit = 20
	}
	accounts, err := a.accounts.SearchByEmail(ctx, q, limit)
	if err != nil {
		return Page{}, err
	}
	items := make([]Item, 0, len(accounts))
	for _, acc := range accounts {
		items = append(items, Item{Key: acc.ID().String(), Label: acc.Email().String()})
	}
	return Page{Items: items}, nil
}
