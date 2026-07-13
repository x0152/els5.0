package usecases

import (
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

func iamID(raw string) (iam.AccountID, error) {
	id, err := vo.ParseID(raw)
	if err != nil {
		return iam.AccountID{}, shared.ErrUnauthorized
	}
	return iam.AccountID{ID: id}, nil
}

const tokenPlaceholder = "{token}"

func renderLink(template, token string) string {
	return strings.ReplaceAll(template, tokenPlaceholder, token)
}

func maskEmail(email string) string {
	at := strings.IndexByte(email, '@')
	if at < 0 {
		return "***"
	}
	local := email[:at]
	domain := email[at:]
	switch {
	case len(local) <= 1:
		return "*" + domain
	case len(local) <= 3:
		return local[:1] + "***" + domain
	default:
		return local[:2] + "***" + domain
	}
}
