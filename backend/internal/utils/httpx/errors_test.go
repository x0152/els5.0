package httpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/els/backend/internal/domain/shared"
)

func TestMapError(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   ErrorCode
	}{
		{"nil → 200", nil, http.StatusOK, ""},
		{"ErrNotFound → 404", shared.ErrNotFound, http.StatusNotFound, CodeNotFound},
		{"ErrConflict → 409", shared.ErrConflict, http.StatusConflict, CodeConflict},
		{"ErrValidation → 422", shared.ErrValidation, http.StatusUnprocessableEntity, CodeValidation},
		{"ErrUnauthorized → 401", shared.ErrUnauthorized, http.StatusUnauthorized, CodeUnauthorized},
		{"ErrForbidden → 403", shared.ErrForbidden, http.StatusForbidden, CodeForbidden},
		{"ErrUnavailable → 503", shared.ErrUnavailable, http.StatusServiceUnavailable, CodeUnavailable},
		{"wrapped ErrNotFound → 404", fmt.Errorf("user: %w", shared.ErrNotFound), http.StatusNotFound, CodeNotFound},
		{"unknown → 500", errors.New("boom"), http.StatusInternalServerError, CodeInternal},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotStatus, gotBody := MapError(context.Background(), tc.err)
			if gotStatus != tc.wantStatus {
				t.Fatalf("status: want %d, got %d", tc.wantStatus, gotStatus)
			}
			if gotBody.Code != tc.wantCode {
				t.Fatalf("code: want %q, got %q", tc.wantCode, gotBody.Code)
			}
		})
	}
}

func TestMapError_PreservesExplicitError(t *testing.T) {
	explicit := NewError(http.StatusTeapot, "CUSTOM", "tea time")
	explicit = WithDetails(explicit, ErrorDetail{Field: "body.tea", Message: "required"})

	status, body := MapError(context.Background(), explicit)

	if status != http.StatusTeapot {
		t.Fatalf("status: want 418, got %d", status)
	}
	if body.Code != "CUSTOM" {
		t.Fatalf("code: want CUSTOM, got %q", body.Code)
	}
	if len(body.Details) != 1 || body.Details[0].Field != "body.tea" {
		t.Fatalf("details lost: %+v", body.Details)
	}
}

func TestMapError_PreservesExplicitErrorThroughWrap(t *testing.T) {
	wrapped := fmt.Errorf("context: %w", NewError(http.StatusConflict, CodeConflict, "duplicate email"))

	status, body := MapError(context.Background(), wrapped)

	if status != http.StatusConflict {
		t.Fatalf("status: want 409, got %d", status)
	}
	if body.Code != CodeConflict {
		t.Fatalf("code: want %q, got %q", CodeConflict, body.Code)
	}
}

func TestMapError_DoesNotKnowDomainValidationDetails(t *testing.T) {
	err := fmt.Errorf(
		"items[0]: %w",
		shared.Validation(fmt.Errorf("timesheet.hours: must be in [0..24]")),
	)

	status, body := MapError(context.Background(), err)

	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status: want 422, got %d", status)
	}
	const want = "items[0]: validation failed: timesheet.hours: must be in [0..24]"
	if body.Message != want {
		t.Fatalf("message: want %q, got %q", want, body.Message)
	}
}
