package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/els/backend/internal/utils/reqctx"
)

type Response[T any] struct {
	Body SuccessBody[T] `json:"body"`
}

type SuccessBody[T any] struct {
	OK   bool  `json:"ok"`
	Data T     `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

func Success[T any](ctx context.Context, data T) *Response[T] {
	return &Response[T]{
		Body: SuccessBody[T]{
			OK:   true,
			Data: data,
			Meta: metaFromCtx(ctx, nil),
		},
	}
}

func SuccessList[T any](ctx context.Context, items T, p *Pagination) *Response[T] {
	return &Response[T]{
		Body: SuccessBody[T]{
			OK:   true,
			Data: items,
			Meta: metaFromCtx(ctx, p),
		},
	}
}

func metaFromCtx(ctx context.Context, p *Pagination) *Meta {
	rid := reqctx.RequestID(ctx)
	if rid == "" && p == nil {
		return nil
	}
	return &Meta{RequestID: rid, Pagination: p}
}

type ErrorResponse struct {
	OK   bool       `json:"ok"`
	Err  *ErrorBody `json:"error"`
	Meta *Meta      `json:"meta,omitempty"`

	status int
}

func (e *ErrorResponse) Error() string {
	if e.Err != nil {
		return e.Err.Message
	}
	return "error"
}

func (e *ErrorResponse) GetStatus() int { return e.status }

func ErrorFrom(ctx context.Context, err error) *ErrorResponse {
	status, body := MapError(ctx, err)
	LogResponseError(ctx, err, status, body.Code)
	return &ErrorResponse{
		OK:     false,
		Err:    &body,
		Meta:   metaFromCtx(ctx, nil),
		status: status,
	}
}

func Return[T any](ctx context.Context, data T, err error) (*Response[T], error) {
	if err != nil {
		return nil, ErrorFrom(ctx, err)
	}
	return Success(ctx, data), nil
}

func InstallHumaErrorHandler() {
	huma.NewError = func(status int, msg string, errs ...error) huma.StatusError {
		return buildHumaError(context.Background(), status, msg, errs...)
	}
	huma.NewErrorWithContext = func(ctx huma.Context, status int, msg string, errs ...error) huma.StatusError {
		var reqCtx context.Context
		if ctx != nil {
			reqCtx = ctx.Context()
		}
		return buildHumaError(reqCtx, status, msg, errs...)
	}
}

func buildHumaError(ctx context.Context, status int, msg string, errs ...error) *ErrorResponse {
	details := make([]ErrorDetail, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		if d, ok := err.(huma.ErrorDetailer); ok {
			ed := d.ErrorDetail()
			details = append(details, ErrorDetail{
				Field:   ed.Location,
				Message: ed.Message,
			})
			continue
		}
		details = append(details, ErrorDetail{Message: err.Error()})
	}

	code := codeFromStatus(status)
	logHumaError(ctx, status, code, msg, details)

	return &ErrorResponse{
		OK: false,
		Err: &ErrorBody{
			Code:    code,
			Message: msg,
			Details: details,
		},
		Meta:   metaFromCtx(ctx, nil),
		status: status,
	}
}

func logHumaError(ctx context.Context, status int, code ErrorCode, msg string, details []ErrorDetail) {
	if status < 400 {
		return
	}
	if reqctx.IsSilent(ctx) {
		return
	}
	level := slog.LevelWarn
	if status >= 500 {
		level = slog.LevelError
	}
	attrs := []slog.Attr{
		slog.String("err", msg),
		slog.Int("status", status),
		slog.String("code", string(code)),
		slog.String("err_type", "huma"),
	}
	if len(details) > 0 {
		detailAttrs := make([]any, 0, len(details))
		for _, d := range details {
			detailAttrs = append(detailAttrs, slog.GroupValue(
				slog.String("field", d.Field),
				slog.String("message", d.Message),
			))
		}
		attrs = append(attrs, slog.Any("details", detailAttrs))
	}
	if level == slog.LevelError {
		attrs = append(attrs, slog.String("stack", string(debug.Stack())))
	}
	slog.LogAttrs(ctx, level, "request error", attrs...)
}

func codeFromStatus(s int) ErrorCode {
	switch s {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusMethodNotAllowed:
		return CodeMethodNotAllowed
	case http.StatusConflict:
		return CodeConflict
	case http.StatusUnprocessableEntity:
		return CodeValidation
	case http.StatusRequestEntityTooLarge:
		return CodePayloadTooLarge
	case http.StatusTooManyRequests:
		return CodeTooManyRequests
	case http.StatusServiceUnavailable:
		return CodeUnavailable
	}
	if s >= 500 {
		return CodeInternal
	}
	return CodeBadRequest
}

type APIOption func(*huma.Config)

func WithDocsDisabled() APIOption {
	return func(c *huma.Config) {
		c.OpenAPIPath = ""
		c.DocsPath = ""
		c.SchemasPath = ""
	}
}

func NewAPI(mux *http.ServeMux, title, version string, opts ...APIOption) huma.API {
	cfg := huma.DefaultConfig(title, version)
	for _, hook := range cfg.CreateHooks {
		cfg = hook(cfg)
	}
	cfg.CreateHooks = nil
	cfg.Transformers = dropByFuncName(cfg.Transformers, "SchemaLinkTransformer")
	cfg.OnAddOperation = dropByFuncName(cfg.OnAddOperation, "SchemaLinkTransformer")
	for _, opt := range opts {
		opt(&cfg)
	}
	return humago.New(mux, cfg)
}

func dropByFuncName[T any](fns []T, needle string) []T {
	if len(fns) == 0 {
		return fns
	}
	out := make([]T, 0, len(fns))
	for _, f := range fns {
		name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		if strings.Contains(name, needle) {
			continue
		}
		out = append(out, f)
	}
	return out
}
