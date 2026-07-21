package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/utils/timex"
)

type PlanFilmsUseCase struct {
	repo   workout.Repository
	films  films.Repository
	llm    LLMClient
	clock  timex.Clock
	logger *slog.Logger
}

func NewPlanFilmsUseCase(repo workout.Repository, filmsRepo films.Repository, llm LLMClient, clock timex.Clock, logger *slog.Logger) *PlanFilmsUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &PlanFilmsUseCase{repo: repo, films: filmsRepo, llm: llm, clock: clock, logger: logger}
}

func (uc *PlanFilmsUseCase) Execute(ctx context.Context) error {
	// 1. Find one ready film that has no segmentation plan yet.
	if !uc.llm.Available() {
		return nil
	}
	all, err := uc.films.List(ctx)
	if err != nil {
		return err
	}
	plannedIDs, err := uc.repo.ListPlannedFilmIDs(ctx, "")
	if err != nil {
		return err
	}
	// Failed plans older than a day get one more chance: an LLM hiccup must not block a film forever.
	stale, err := uc.repo.ListStaleFailedPlanFilmIDs(ctx, uc.clock.Now().Add(-24*time.Hour))
	if err != nil {
		return err
	}
	planned := make(map[string]bool, len(plannedIDs))
	for _, id := range plannedIDs {
		planned[id] = true
	}
	for _, id := range stale {
		planned[id] = false
	}
	var target *films.Film
	for i, f := range all {
		if f.Status == films.StatusReady && !planned[f.ID] {
			target = &all[i]
			break
		}
	}
	if target == nil {
		return nil
	}

	// 2. Segment its English subtitles into watch blocks with recaps and key phrases.
	track, ok := films.PickEnglishSubtitle(target.Subtitles)
	if !ok && len(target.Subtitles) > 0 {
		track = target.Subtitles[0]
	}
	now := uc.clock.Now().In(timex.MSK)
	plan := workout.FilmPlan{FilmID: target.ID, Status: workout.PlanStatusReady, CreatedAt: now}
	if len(track.Cues) == 0 {
		plan.Status = workout.PlanStatusFailed
		plan.Error = "no subtitles"
		return uc.repo.SavePlan(ctx, plan)
	}
	system, user := workout.BuildSegmentationPrompt(target.Title, track.Cues)
	raw, err := uc.llm.Chat(ctx, system, user)
	if err == nil {
		plan.Segments, err = workout.ParseSegments(raw, track.Cues)
	}
	if err != nil {
		uc.logger.Error("workout film planning failed", slog.String("film", target.ID), slog.String("err", err.Error()))
		plan.Status = workout.PlanStatusFailed
		plan.Error = err.Error()
	}

	// 3. Persist the plan; failed plans are kept so the worker does not retry forever.
	return uc.repo.SavePlan(ctx, plan)
}
