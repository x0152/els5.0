package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UUIDFromGoogle(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func UUIDFromGoogleOpt(id uuid.UUID, valid bool) pgtype.UUID {
	if !valid {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func UUIDToGoogle(v pgtype.UUID) uuid.UUID {
	return uuid.UUID(v.Bytes)
}

func TimestamptzFromTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func TimestamptzToTime(v pgtype.Timestamptz) time.Time {
	return v.Time
}

func TimestamptzFromPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func TimestamptzToPtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
