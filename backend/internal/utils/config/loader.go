package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var durationType = reflect.TypeOf(time.Duration(0))

type Validator interface {
	Validate() error
}

func LoadDotEnv() {
	_ = godotenv.Load()
}

func Parse(cfg any) error {
	return env.Parse(cfg)
}

func LoadAndValidate(cfg Validator) error {
	LoadDotEnv()
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate:\n%w", err)
	}
	return nil
}

func MustLoad(name string, cfg Validator) {
	if err := LoadAndValidate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "[config] %s: %v\n", name, err)
		os.Exit(1)
	}
}

func Dump(log *slog.Logger, cfg any) {
	log.LogAttrs(context.Background(), slog.LevelDebug, "config loaded",
		slog.Any("values", snapshot(cfg)),
	)
}

func snapshot(cfg any) map[string]string {
	out := map[string]string{}
	walk(reflect.ValueOf(cfg), "", out)
	return out
}

func walk(v reflect.Value, prefix string, out map[string]string) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		val := v.Field(i)
		name := prefix + f.Name

		if val.Kind() == reflect.Struct && val.Type() != durationType {
			walk(val, name+".", out)
			continue
		}

		display := fmt.Sprintf("%v", val.Interface())
		if f.Tag.Get("secret") == "true" {
			display = mask(display)
		}
		out[name] = display
	}
}

func mask(v string) string {
	if v == "" {
		return ""
	}
	if len(v) <= 4 {
		return strings.Repeat("*", len(v))
	}
	return v[:2] + strings.Repeat("*", len(v)-4) + v[len(v)-2:]
}
