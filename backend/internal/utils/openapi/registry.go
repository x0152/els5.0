package openapi

import (
	"fmt"
	"sort"
	"sync"

	"github.com/danielgtaylor/huma/v2"
)

type Module struct {
	Name     string
	Version  string
	Register func(huma.API)
}

var (
	mu      sync.RWMutex
	modules = map[string]Module{}
)

func Register(m Module) {
	if m.Name == "" {
		panic("openapi: module name is empty")
	}
	if m.Register == nil {
		panic(fmt.Sprintf("openapi: module %q has nil Register", m.Name))
	}
	mu.Lock()
	defer mu.Unlock()
	if _, exists := modules[m.Name]; exists {
		panic(fmt.Sprintf("openapi: duplicate module %q", m.Name))
	}
	modules[m.Name] = m
}

func Get(name string) (Module, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := modules[name]
	return m, ok
}

func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(modules))
	for n := range modules {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
