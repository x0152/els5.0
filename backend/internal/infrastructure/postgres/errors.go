package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
	codeCheckViolation      = "23514"
	codeNotNullViolation    = "23502"
)

func IsUniqueViolation(err error) bool {
	return pgErrCode(err) == codeUniqueViolation
}

func IsForeignKeyViolation(err error) bool {
	return pgErrCode(err) == codeForeignKeyViolation
}

func IsCheckViolation(err error) bool {
	return pgErrCode(err) == codeCheckViolation
}

func IsNotNullViolation(err error) bool {
	return pgErrCode(err) == codeNotNullViolation
}

func pgErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
