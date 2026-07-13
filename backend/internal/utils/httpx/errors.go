package httpx

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/els/backend/internal/domain/shared"
)

type ErrorCode string

const (
	CodeValidation       ErrorCode = "VALIDATION_ERROR"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeConflict         ErrorCode = "CONFLICT"
	CodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	CodeForbidden        ErrorCode = "FORBIDDEN"
	CodeUnavailable      ErrorCode = "SERVICE_UNAVAILABLE"
	CodeInternal         ErrorCode = "INTERNAL_ERROR"
	CodeBadRequest       ErrorCode = "BAD_REQUEST"
	CodeMethodNotAllowed ErrorCode = "METHOD_NOT_ALLOWED"
	CodeTooManyRequests  ErrorCode = "TOO_MANY_REQUESTS"
	CodePayloadTooLarge  ErrorCode = "PAYLOAD_TOO_LARGE"
)

type ErrorBody struct {
	Code    ErrorCode     `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type Error struct {
	Status  int
	Code    ErrorCode
	Message string
	Details []ErrorDetail
	Err     error
	Stack   []byte
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func NewError(status int, code ErrorCode, message string) *Error {
	e := &Error{Status: status, Code: code, Message: message}
	if status >= http.StatusInternalServerError {
		e.Stack = debug.Stack()
	}
	return e
}

func Wrap(err error, status int, code ErrorCode, message string) *Error {
	e := &Error{Status: status, Code: code, Message: message, Err: err}
	if status >= http.StatusInternalServerError {
		e.Stack = stackFromErr(err)
		if e.Stack == nil {
			e.Stack = debug.Stack()
		}
	}
	return e
}

func stackFromErr(err error) []byte {
	var he *Error
	if errors.As(err, &he) && len(he.Stack) > 0 {
		return he.Stack
	}
	return nil
}

func WithDetails(err *Error, details ...ErrorDetail) *Error {
	if err == nil {
		return nil
	}
	clone := *err
	clone.Details = make([]ErrorDetail, 0, len(err.Details)+len(details))
	clone.Details = append(clone.Details, err.Details...)
	clone.Details = append(clone.Details, details...)
	return &clone
}

func MapError(_ context.Context, err error) (int, ErrorBody) {
	if err == nil {
		return http.StatusOK, ErrorBody{}
	}

	var he *Error
	if errors.As(err, &he) {
		return he.Status, ErrorBody{
			Code:    he.Code,
			Message: he.Message,
			Details: he.Details,
		}
	}

	switch {
	case errors.Is(err, shared.ErrNotFound):
		return http.StatusNotFound, ErrorBody{Code: CodeNotFound, Message: err.Error()}
	case errors.Is(err, shared.ErrConflict):
		return http.StatusConflict, ErrorBody{Code: CodeConflict, Message: err.Error()}
	case errors.Is(err, shared.ErrValidation):
		return http.StatusUnprocessableEntity, ErrorBody{Code: CodeValidation, Message: err.Error()}
	case errors.Is(err, shared.ErrUnauthorized):
		return http.StatusUnauthorized, ErrorBody{Code: CodeUnauthorized, Message: err.Error()}
	case errors.Is(err, shared.ErrForbidden):
		return http.StatusForbidden, ErrorBody{Code: CodeForbidden, Message: err.Error()}
	case errors.Is(err, shared.ErrUnavailable):
		return http.StatusServiceUnavailable, ErrorBody{Code: CodeUnavailable, Message: err.Error()}
	}

	return http.StatusInternalServerError, ErrorBody{
		Code:    CodeInternal,
		Message: "internal server error",
	}
}
