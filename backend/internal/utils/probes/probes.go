package probes

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/timex"
)

const defaultCheckTimeout = 3 * time.Second

type Check func(ctx context.Context) error

type NamedCheck struct {
	Name  string
	Check Check
}

type Deps struct {
	Module       string
	Version      string
	Ready        []NamedCheck
	CheckTimeout time.Duration
	Clock        timex.Clock
}

type HealthOutput struct {
	Status string    `json:"status" example:"ok"`
	Module string    `json:"module,omitempty" example:"auth"`
	Time   time.Time `json:"time"`
}

type CheckResult struct {
	Name  string `json:"name" example:"postgres"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type ReadyOutput struct {
	Ready  bool          `json:"ready"`
	Module string        `json:"module,omitempty" example:"auth"`
	Checks []CheckResult `json:"checks,omitempty"`
	Time   time.Time     `json:"time"`
}

func Register(api huma.API, deps Deps) {
	if deps.CheckTimeout <= 0 {
		deps.CheckTimeout = defaultCheckTimeout
	}
	if deps.Clock == nil {
		deps.Clock = timex.System()
	}
	registerHealth(api, deps)
	registerReady(api, deps)
}

func registerHealth(api huma.API, deps Deps) {
	huma.Register(api, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Liveness probe",
		Description: "Returns 200 if the process is alive. Does not depend on external services.",
		Tags:        []string{"probes"},
	}, func(ctx context.Context, _ *struct{}) (*httpx.Response[HealthOutput], error) {
		return httpx.Success(ctx, HealthOutput{
			Status: "ok",
			Module: deps.Module,
			Time:   deps.Clock.Now(),
		}), nil
	})
}

func registerReady(api huma.API, deps Deps) {
	huma.Register(api, huma.Operation{
		OperationID: "ready",
		Method:      http.MethodGet,
		Path:        "/ready",
		Summary:     "Readiness probe",
		Description: "Returns 200 if the module is ready to accept traffic (all dependencies alive). 503 if any check failed.",
		Tags:        []string{"probes"},
	}, func(ctx context.Context, _ *struct{}) (*httpx.Response[ReadyOutput], error) {
		results, firstErr := runChecks(ctx, deps.Ready, deps.CheckTimeout)

		if firstErr != nil {
			details := make([]httpx.ErrorDetail, 0, len(results))
			for _, r := range results {
				if r.OK {
					continue
				}
				details = append(details, httpx.ErrorDetail{
					Field:   r.Name,
					Message: r.Error,
				})
			}
			he := httpx.Wrap(shared.ErrUnavailable, http.StatusServiceUnavailable, httpx.CodeUnavailable, "not ready")
			he = httpx.WithDetails(he, details...)
			return nil, httpx.ErrorFrom(ctx, he)
		}

		return httpx.Success(ctx, ReadyOutput{
			Ready:  true,
			Module: deps.Module,
			Checks: results,
			Time:   deps.Clock.Now(),
		}), nil
	})
}

func runChecks(ctx context.Context, checks []NamedCheck, timeout time.Duration) ([]CheckResult, error) {
	if len(checks) == 0 {
		return nil, nil
	}

	results := make([]CheckResult, len(checks))
	errs := make([]error, len(checks))

	var wg sync.WaitGroup
	wg.Add(len(checks))
	for i, c := range checks {
		i, c := i, c
		go func() {
			defer wg.Done()
			r := CheckResult{Name: c.Name}
			if c.Check == nil {
				r.OK = true
				results[i] = r
				return
			}
			cctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			if err := c.Check(cctx); err != nil {
				r.OK = false
				r.Error = err.Error()
				errs[i] = err
			} else {
				r.OK = true
			}
			results[i] = r
		}()
	}
	wg.Wait()

	var firstErr error
	for i, err := range errs {
		if err != nil {
			firstErr = fmt.Errorf("%s: %w", checks[i].Name, err)
			break
		}
	}
	return results, firstErr
}
