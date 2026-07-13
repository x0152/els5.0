package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed all:_template
var templateFS embed.FS

func main() {
	var (
		module = flag.String("module", "", "short module name (lowercase letters/digits, e.g. billing)")
		root   = flag.String("root", ".", "repository root (directory containing go.mod)")
	)
	flag.Parse()

	if *module == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/new-app -module=<name> [-root=.]")
		fmt.Fprintln(os.Stderr, "example: make new-app MODULE=billing")
		os.Exit(2)
	}

	mod := strings.ToLower(strings.TrimSpace(*module))
	if !regexp.MustCompile(`^[a-z][a-z0-9]{0,62}$`).MatchString(mod) {
		fmt.Fprintf(os.Stderr, "invalid -module %q: use ^[a-z][a-z0-9]{0,62}$\n", *module)
		os.Exit(2)
	}
	envPrefix := strings.ToUpper(mod)
	titleMod := strings.ToUpper(mod[:1]) + mod[1:]

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err := os.Stat(filepath.Join(absRoot, "go.mod")); err != nil {
		fmt.Fprintf(os.Stderr, "go.mod not found under %s (wrong -root?)\n", absRoot)
		os.Exit(1)
	}

	appDir := filepath.Join(absRoot, "internal", "application", mod)
	if _, err := os.Stat(appDir); err == nil {
		fmt.Fprintf(os.Stderr, "refuse to overwrite existing %s\n", appDir)
		os.Exit(1)
	}
	domainDir := filepath.Join(absRoot, "internal", "domain", mod)
	if _, err := os.Stat(domainDir); err == nil {
		fmt.Fprintf(os.Stderr, "refuse to overwrite existing %s\n", domainDir)
		os.Exit(1)
	}
	cmdDir := filepath.Join(absRoot, "cmd", mod)
	if _, err := os.Stat(cmdDir); err == nil {
		fmt.Fprintf(os.Stderr, "refuse to overwrite existing %s\n", cmdDir)
		os.Exit(1)
	}

	err = fs.WalkDir(templateFS, "_template", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, ok := strings.CutPrefix(path, "_template/")
		if !ok {
			return nil
		}
		rel = strings.ReplaceAll(rel, "templateapp", mod)
		if d.IsDir() {
			dst := filepath.Join(absRoot, rel)
			return os.MkdirAll(dst, 0o755) // #nosec G301 -- scaffolding code is intentionally readable/executable for local dev users
		}
		b, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(b)
		text = strings.ReplaceAll(text, "TEMPLATEAPP", envPrefix)
		text = strings.ReplaceAll(text, "Templateapp", titleMod)
		text = strings.ReplaceAll(text, "templateapp", mod)
		dst := filepath.Join(absRoot, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { // #nosec G301 -- generated project directories are shared in dev environments
			return err
		}
		return os.WriteFile(dst, []byte(text), 0o644) // #nosec G306 -- generated source/templates should be readable by developer tooling
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("created application %q under:\n  %s\n  %s\n  %s\n\n", mod, cmdDir, appDir, domainDir)
	fmt.Println("next steps (manual):")
	fmt.Printf("  1. cmd/main.go: import _ %q and add %s.Name to mounts map\n",
		"github.com/els/backend/internal/application/"+mod, mod)
	fmt.Printf("  2. tools/openapi-dump/main.go: add blank import _ %q\n",
		"github.com/els/backend/internal/application/"+mod)
	fmt.Printf("  3. env: set %s_HTTP_* and %s_SESSION_* (see internal/application/%s/config.go)\n", envPrefix, envPrefix, mod)
	fmt.Println("  4. make openapi MODULE=<name>   # after wiring openapi init()")
}
