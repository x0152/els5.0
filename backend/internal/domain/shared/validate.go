package shared

import (
	"errors"
	"fmt"
)

func Validation(errs ...error) error {
	joined := errors.Join(errs...)
	if joined == nil {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrValidation, joined)
}
