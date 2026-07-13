package vo

import (
	"errors"
	"fmt"
	"time"

	"github.com/els/backend/internal/utils/timex"
)

type Timestamps struct {
	createdAt time.Time
	updatedAt time.Time
}

func NewTimestamps(createdAt, updatedAt time.Time) (Timestamps, error) {
	var errs []error
	if createdAt.IsZero() {
		errs = append(errs, fmt.Errorf("created_at must not be zero"))
	}
	if updatedAt.IsZero() {
		errs = append(errs, fmt.Errorf("updated_at must not be zero"))
	}
	if !createdAt.IsZero() && !updatedAt.IsZero() && updatedAt.Before(createdAt) {
		errs = append(errs, fmt.Errorf("updated_at must be >= created_at"))
	}
	if err := errors.Join(errs...); err != nil {
		return Timestamps{}, err
	}
	return Timestamps{createdAt: createdAt, updatedAt: updatedAt}, nil
}

func NewCurrentTimestamps() (Timestamps, error) {
	now := timex.Now()
	return NewTimestamps(now, now)
}

func (t Timestamps) CreatedAt() time.Time { return t.createdAt }
func (t Timestamps) UpdatedAt() time.Time { return t.updatedAt }

func (t Timestamps) Touch(updatedAt time.Time) (Timestamps, error) {
	return NewTimestamps(t.createdAt, updatedAt)
}
