package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/practice"
)

type Variants struct {
	variants practice.VariantRepository
	sources  *Sources
	service  *Service
	logger   *slog.Logger
}

func NewVariants(variants practice.VariantRepository, sources *Sources, service *Service, logger *slog.Logger) *Variants {
	if logger == nil {
		logger = slog.Default()
	}
	return &Variants{variants: variants, sources: sources, service: service, logger: logger}
}

func (g *Variants) Enqueue(accountID, variantID string, kind practice.Kind, number int) {
	go g.run(accountID, variantID, kind, number)
}

func (g *Variants) run(accountID, variantID string, kind practice.Kind, number int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			g.logger.Error("practice: variant generation panic", slog.String("variant", variantID), slog.Any("panic", r))
			g.fail(accountID, variantID, kind, number, fmt.Errorf("internal error during generation"))
		}
	}()

	src, err := g.sources.Source(ctx, kind, number)
	if err != nil {
		g.fail(accountID, variantID, kind, number, fmt.Errorf("load source: %w", err))
		return
	}
	title, items, err := g.service.Plan(ctx, src)
	if err != nil {
		g.fail(accountID, variantID, kind, number, err)
		return
	}

	// Generate exercises one by one and append them to the variant — the frontend sees them via polling.
	var exercises strings.Builder
	count := 0
	for _, item := range items {
		block, err := g.service.GenerateExercise(ctx, src, item, count+1)
		if err != nil {
			g.logger.Warn("practice: exercise generation failed, retrying", slog.String("variant", variantID), slog.String("err", err.Error()))
			block, err = g.service.GenerateExercise(ctx, src, item, count+1)
		}
		if err != nil {
			g.logger.Error("practice: exercise generation failed", slog.String("variant", variantID), slog.String("err", err.Error()))
			continue
		}
		count++
		if exercises.Len() > 0 {
			exercises.WriteString("\n\n")
		}
		exercises.WriteString(block)
		g.write(accountID, variantID, kind, number, title, exercises.String(), practice.StatusGenerating)
	}
	if count == 0 {
		g.fail(accountID, variantID, kind, number, fmt.Errorf("no exercises generated"))
		return
	}
	g.write(accountID, variantID, kind, number, title, exercises.String(), practice.StatusReady)
}

func (g *Variants) write(accountID, variantID string, kind practice.Kind, number int, title, exercises, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := g.variants.Update(ctx, accountID, practice.Variant{
		ID:        variantID,
		Kind:      kind,
		Number:    number,
		Title:     title,
		Exercises: exercises,
		Status:    status,
	}); err != nil {
		g.logger.Error("practice: save variant failed", slog.String("variant", variantID), slog.String("err", err.Error()))
	}
}

func (g *Variants) fail(accountID, variantID string, kind practice.Kind, number int, cause error) {
	g.logger.Warn("practice: variant generation failed", slog.String("variant", variantID), slog.String("err", cause.Error()))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = g.variants.Update(ctx, accountID, practice.Variant{
		ID:     variantID,
		Kind:   kind,
		Number: number,
		Status: practice.StatusError,
		Error:  cause.Error(),
	})
}
