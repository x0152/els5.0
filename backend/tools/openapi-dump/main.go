package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	_ "github.com/els/backend/internal/application/account"
	_ "github.com/els/backend/internal/application/admin"
	_ "github.com/els/backend/internal/application/ai"
	_ "github.com/els/backend/internal/application/auth"
	_ "github.com/els/backend/internal/application/core"
	_ "github.com/els/backend/internal/application/diary"
	_ "github.com/els/backend/internal/application/films"
	_ "github.com/els/backend/internal/application/learn"
	_ "github.com/els/backend/internal/application/listening"
	_ "github.com/els/backend/internal/application/quest"
	_ "github.com/els/backend/internal/application/reader"
	_ "github.com/els/backend/internal/application/reading"
	_ "github.com/els/backend/internal/application/settings"
	_ "github.com/els/backend/internal/application/speech"
	_ "github.com/els/backend/internal/application/vocab"
	_ "github.com/els/backend/internal/application/workout"
	_ "github.com/els/backend/internal/application/writing"
	"github.com/els/backend/internal/utils/httpx"
	"github.com/els/backend/internal/utils/openapi"
	"github.com/els/backend/internal/utils/probes"
)

func main() {
	var (
		moduleFlag = flag.String("module", "", "module to dump (empty → list available)")
		format     = flag.String("format", "yaml", "output format: yaml | json")
		out        = flag.String("out", "-", "output file path or '-' for stdout")
		list       = flag.Bool("list", false, "list known modules and exit")
	)
	flag.Parse()

	if *list {
		for _, name := range openapi.Names() {
			fmt.Println(name)
		}
		return
	}
	if *moduleFlag == "" {
		fmt.Fprintln(os.Stderr, "usage: openapi-dump -module=<name> [-format=yaml|json] [-out=path]")
		fmt.Fprintf(os.Stderr, "available modules: %s\n", strings.Join(openapi.Names(), ", "))
		os.Exit(2)
	}

	mod, ok := openapi.Get(*moduleFlag)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown module %q; available: %s\n", *moduleFlag, strings.Join(openapi.Names(), ", "))
		os.Exit(2)
	}

	httpx.InstallHumaErrorHandler()
	mux := http.NewServeMux()
	api := httpx.NewAPI(mux, mod.Name, mod.Version)

	probes.Register(api, probes.Deps{Module: mod.Name, Version: mod.Version})
	mod.Register(api)

	data, err := render(api, *format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render openapi: %v\n", err)
		os.Exit(1)
	}

	if *out == "-" || *out == "" {
		_, _ = os.Stdout.Write(data)
		return
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil { // #nosec G306 -- OpenAPI artifacts are intentionally shareable/readable
		fmt.Fprintf(os.Stderr, "write %s: %v\n", *out, err)
		os.Exit(1)
	}
}

func render(api huma.API, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		return api.OpenAPI().MarshalJSON()
	default:
		return api.OpenAPI().YAML()
	}
}
