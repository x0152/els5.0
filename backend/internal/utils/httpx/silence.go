package httpx

import (
	"net/http"

	"github.com/els/backend/internal/utils/reqctx"
)

var probePaths = map[string]struct{}{
	"/health": {},
	"/ready":  {},
}

func SilenceProbes(next http.Handler) http.Handler {
	return silenceByPaths(next, probePaths)
}

func SilencePaths(paths ...string) func(http.Handler) http.Handler {
	set := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		set[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return silenceByPaths(next, set)
	}
}

func silenceByPaths(next http.Handler, paths map[string]struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paths[r.URL.Path]; ok {
			r = r.WithContext(reqctx.WithSilent(r.Context()))
		}
		next.ServeHTTP(w, r)
	})
}
