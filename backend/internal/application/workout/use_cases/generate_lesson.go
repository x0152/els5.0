package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/practice"
	"github.com/els/backend/internal/domain/reading"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/domain/speech"
	"github.com/els/backend/internal/domain/vocab"
	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/domain/writing"
	"github.com/els/backend/internal/utils/timex"
)

type LLMClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type AccountSource interface {
	GetByID(ctx context.Context, id iam.AccountID) (*iam.Account, error)
}

type WordSource interface {
	List(ctx context.Context, accountID string, filter vocab.ListFilter) ([]vocab.Unit, int, error)
}

type ErrorSource interface {
	ListRecentErrors(ctx context.Context, accountID string, since time.Time, limit int) ([]workout.GrammarFocus, error)
}

type GenerateLessonUseCase struct {
	repo     workout.Repository
	films    films.Repository
	accounts AccountSource
	words    WordSource
	errs     ErrorSource
	llm      LLMClient
	clock    timex.Clock
}

func NewGenerateLessonUseCase(repo workout.Repository, filmsRepo films.Repository, accounts AccountSource, words WordSource, errs ErrorSource, llm LLMClient, clock timex.Clock) *GenerateLessonUseCase {
	if clock == nil {
		clock = timex.System()
	}
	return &GenerateLessonUseCase{repo: repo, films: filmsRepo, accounts: accounts, words: words, errs: errs, llm: llm, clock: clock}
}

func (uc *GenerateLessonUseCase) Execute(ctx context.Context, accountID string) (workout.Lesson, error) {
	// 1. An active lesson is the current one — generation is idempotent.
	if lesson, err := uc.repo.CurrentLesson(ctx, accountID); err == nil {
		return lesson, nil
	} else if !errors.Is(err, shared.ErrNotFound) {
		return workout.Lesson{}, err
	}
	if !uc.llm.Available() {
		return workout.Lesson{}, fmt.Errorf("llm is not configured: %w", shared.ErrValidation)
	}

	// 2. Load the learner and the lesson history.
	id, err := vo.ParseID(accountID)
	if err != nil {
		return workout.Lesson{}, err
	}
	account, err := uc.accounts.GetByID(ctx, iam.AccountID{ID: id})
	if err != nil {
		return workout.Lesson{}, err
	}
	level := workout.NormalizeLevel(account.EnglishLevel())
	recent, err := uc.repo.ListRecentLessons(ctx, accountID, workout.CycleLength)
	if err != nil {
		return workout.Lesson{}, err
	}
	number := 1
	if len(recent) > 0 {
		number = recent[0].Number + 1
	}

	// 3. The 7th lesson of a cycle reviews the weakest material instead of new content.
	now := uc.clock.Now().In(timex.MSK)
	lesson := workout.Lesson{
		ID:        uuid.NewString(),
		AccountID: accountID,
		Number:    number,
		Status:    workout.LessonStatusActive,
		CreatedAt: now,
	}
	if workout.IsReviewNumber(number) {
		if err := uc.buildReviewSteps(ctx, &lesson, level); err != nil {
			return workout.Lesson{}, err
		}
	}
	// A review with nothing to review (or a regular number) gets a fresh lesson.
	if len(lesson.Steps) == 0 {
		if err := uc.buildRegularSteps(ctx, &lesson, level, recent, now); err != nil {
			return workout.Lesson{}, err
		}
	}
	if len(lesson.Steps) == 0 {
		return workout.Lesson{}, fmt.Errorf("no material for a lesson: %w", shared.ErrValidation)
	}

	// 4. Persist the ready lesson.
	if err := uc.repo.InsertLesson(ctx, lesson); err != nil {
		return workout.Lesson{}, err
	}
	return lesson, nil
}

func (uc *GenerateLessonUseCase) buildRegularSteps(ctx context.Context, lesson *workout.Lesson, level string, recent []workout.Lesson, now time.Time) error {
	// 1. Pick the title and the watch block: sequential within a title, rotating between titles.
	// Without prepared films fall back to a desk lesson (speak / read / write / grammar / vocab).
	available, err := uc.films.List(ctx)
	if err != nil {
		return err
	}
	film, pos, planned, err := uc.pickPlannedTitle(ctx, lesson.AccountID, available, level)
	if err != nil {
		if errors.Is(err, shared.ErrValidation) {
			return uc.buildDeskSteps(ctx, lesson, level, recent, now)
		}
		return err
	}
	segments, nextPos := workout.WatchRange(film, planned, pos, episodesOf(available, film))
	if len(segments) == 0 {
		return fmt.Errorf("no segments for film %s: %w", film.ID, shared.ErrValidation)
	}
	lesson.FilmID = film.ID
	lesson.StartMs = segments[0].StartMs
	lesson.EndMs = segments[len(segments)-1].EndMs

	track, hasTrack := films.PickEnglishSubtitle(film.Subtitles)
	if !hasTrack && len(film.Subtitles) > 0 {
		track = film.Subtitles[0]
	}
	cues := workout.CuesInRange(track, lesson.StartMs, lesson.EndMs)

	// 2. Warm-up from the spiral: material of 1 and 3 lessons ago, weak first.
	items, err := uc.repo.ListItems(ctx, lesson.AccountID, lesson.Number-workout.CycleLength)
	if err != nil {
		return err
	}
	warmup := workout.PickWarmup(items, lesson.Number, 4)
	if len(warmup) > 0 {
		payload := workout.WarmupPayload{Items: warmupItems(warmup)}
		if err := uc.appendStep(lesson, workout.StepWarmup, "Warm-up", payload); err != nil {
			return err
		}
	}

	// 3. The skeleton: watch, comprehension questions, speak the key phrases.
	watch := workout.WatchPayload{
		FilmID:  film.ID,
		Title:   film.Title,
		StartMs: lesson.StartMs,
		EndMs:   lesson.EndMs,
		Recap:   segments[0].Recap,
		Summary: segments[0].Summary,
	}
	if err := uc.appendStep(lesson, workout.StepWatch, "Watch", watch); err != nil {
		return err
	}
	questions, err := uc.generateQuestions(ctx, segments[0].Recap, cues, level)
	if err != nil {
		return err
	}
	if err := uc.appendStep(lesson, workout.StepQuestions, "Did you get it?", workout.QuestionsPayload{Questions: questions}); err != nil {
		return err
	}
	phrases := workout.LevelPhrases(segments, level, 4)
	if len(phrases) > 0 {
		if err := uc.appendStep(lesson, workout.StepSpeak, "Say it like them", speakPayload(phrases, film.ID)); err != nil {
			return err
		}
	}

	// 4. Variable slots balanced over the week; a slot without material yields to the next kind.
	filled := 0
	for _, slot := range workout.PickSlots(recent, lesson.Number) {
		if filled == workout.SlotsPerLesson {
			break
		}
		err := uc.buildSlot(ctx, lesson, slot, segments[0].Summary, cues, film.ID, level, now)
		if err != nil {
			// A flaky optional slot must not kill the whole lesson.
			if !errors.Is(err, errSlotSkipped) && ctx.Err() != nil {
				return err
			}
			continue
		}
		filled++
	}

	// 5. Advance the watching position and count the warm-up as a review pass.
	nextPos.AccountID = lesson.AccountID
	nextPos.UsedAt = now
	if err := uc.repo.SavePosition(ctx, nextPos); err != nil {
		return err
	}
	texts := make([]string, 0, len(warmup))
	for _, it := range warmup {
		texts = append(texts, it.Text)
	}
	return uc.repo.MarkReviewed(ctx, lesson.AccountID, texts, lesson.Number, now)
}

const deskSlots = 4

func (uc *GenerateLessonUseCase) buildDeskSteps(ctx context.Context, lesson *workout.Lesson, level string, recent []workout.Lesson, now time.Time) error {
	// 1. Warm-up from the spiral when there is anything to review.
	items, err := uc.repo.ListItems(ctx, lesson.AccountID, lesson.Number-workout.CycleLength)
	if err != nil {
		return err
	}
	warmup := workout.PickWarmup(items, lesson.Number, 4)
	if len(warmup) > 0 {
		if err := uc.appendStep(lesson, workout.StepWarmup, "Warm-up", workout.WarmupPayload{Items: warmupItems(warmup)}); err != nil {
			return err
		}
	}

	// 2. Pronunciation practice replaces the film speak step.
	system, user := speech.BuildPracticePrompt("everyday conversation", nil)
	sentences, err := generate(ctx, uc.llm, system, user, speech.ParsePractice)
	if err != nil {
		return err
	}
	phrases := make([]workout.SpeakPhrase, 0, len(sentences))
	for _, s := range sentences {
		phrases = append(phrases, workout.SpeakPhrase{Text: s})
	}
	if err := uc.appendStep(lesson, workout.StepSpeak, "Say it clearly", workout.SpeakPayload{Phrases: phrases}); err != nil {
		return err
	}

	// 3. More variable slots than a film lesson — there is no watch / questions skeleton.
	filled := 0
	for _, slot := range workout.PickSlots(recent, lesson.Number) {
		if filled == deskSlots {
			break
		}
		if slot == workout.StepDictation {
			continue
		}
		err := uc.buildSlot(ctx, lesson, slot, "everyday life", nil, "", level, now)
		if err != nil {
			if !errors.Is(err, errSlotSkipped) && ctx.Err() != nil {
				return err
			}
			continue
		}
		filled++
	}
	if len(lesson.Steps) == 0 {
		return fmt.Errorf("no material for a desk lesson: %w", shared.ErrValidation)
	}
	texts := make([]string, 0, len(warmup))
	for _, it := range warmup {
		texts = append(texts, it.Text)
	}
	return uc.repo.MarkReviewed(ctx, lesson.AccountID, texts, lesson.Number, now)
}

func (uc *GenerateLessonUseCase) buildReviewSteps(ctx context.Context, lesson *workout.Lesson, level string) error {
	// 1. Gather the weakest material of the finishing cycle.
	items, err := uc.repo.ListItems(ctx, lesson.AccountID, lesson.Number-workout.CycleLength)
	if err != nil {
		return err
	}
	review := workout.PickReview(items, lesson.Number, 12)
	phrases, words := splitItems(review)

	// 2. Speak and re-type the phrases that went badly.
	if len(phrases) > 0 {
		if err := uc.appendStep(lesson, workout.StepSpeak, "Say it again", workout.SpeakPayload{Phrases: phrases}); err != nil {
			return err
		}
		if err := uc.appendStep(lesson, workout.StepDictation, "Type what you hear", workout.DictationPayload{Sentences: phrases}); err != nil {
			return err
		}
	}
	if len(words) > 0 {
		if err := uc.appendStep(lesson, workout.StepVocab, "Words of the week", workout.VocabPayload{Words: words}); err != nil {
			return err
		}
	}

	// 3. A grammar drill on the week's recurring mistakes.
	now := uc.clock.Now().In(timex.MSK)
	if err := uc.buildSlot(ctx, lesson, workout.StepGrammar, "", nil, "", level, now); err != nil && !errors.Is(err, errSlotSkipped) {
		return err
	}
	texts := make([]string, 0, len(review))
	for _, it := range review {
		texts = append(texts, it.Text)
	}
	return uc.repo.MarkReviewed(ctx, lesson.AccountID, texts, lesson.Number, now)
}

var errSlotSkipped = errors.New("slot skipped")

// generate calls the LLM and parses the answer, retrying once — models occasionally emit broken JSON.
func generate[T any](ctx context.Context, llm LLMClient, system, user string, parse func(string) (T, error)) (T, error) {
	var zero T
	raw, err := llm.Chat(ctx, system, user)
	if err == nil {
		v, perr := parse(raw)
		if perr == nil {
			return v, nil
		}
		err = perr
	}
	if ctx.Err() != nil {
		return zero, err
	}
	raw, err = llm.Chat(ctx, system, user)
	if err != nil {
		return zero, err
	}
	return parse(raw)
}

func (uc *GenerateLessonUseCase) buildSlot(ctx context.Context, lesson *workout.Lesson, slot, topic string, cues []films.Cue, filmID, level string, now time.Time) error {
	switch slot {
	case workout.StepDictation:
		lines := workout.DictationLines(cues, filmID, 5)
		if len(lines) == 0 {
			return errSlotSkipped
		}
		return uc.appendStep(lesson, workout.StepDictation, "Type what you hear", workout.DictationPayload{Sentences: lines})

	case workout.StepReading:
		system, user := reading.BuildTextPrompt(topic, uc.learningWords(ctx, lesson.AccountID), readingLevel(level), reading.LengthShort)
		text, err := generate(ctx, uc.llm, system, user, reading.ParseText)
		if err != nil {
			return err
		}
		return uc.appendStep(lesson, workout.StepReading, "Read", workout.ReadingPayload{Title: text.Title, Body: text.Body, Words: text.Words})

	case workout.StepWriting:
		system, user := writing.BuildSituationPrompt(topic)
		situation, err := generate(ctx, uc.llm, system, user, writing.ParseSituation)
		if err != nil {
			return err
		}
		return uc.appendStep(lesson, workout.StepWriting, "Write back", workout.WritingPayload{Scenario: situation.Scenario, Dialogue: situation.Dialogue})

	case workout.StepGrammar:
		focuses, err := uc.errs.ListRecentErrors(ctx, lesson.AccountID, now.AddDate(0, 0, -14), 12)
		if err != nil {
			return errSlotSkipped
		}
		system, user := workout.BuildGrammarPrompt(focuses, topic, level, practice.BlockCatalog)
		payload, err := generate(ctx, uc.llm, system, user, workout.ParseGrammar)
		if err != nil {
			return err
		}
		return uc.appendStep(lesson, workout.StepGrammar, "Grammar drill", payload)

	case workout.StepVocab:
		units, _, err := uc.words.List(ctx, lesson.AccountID, vocab.ListFilter{Status: vocab.StatusLearning, Limit: 8})
		if err != nil || len(units) == 0 {
			return errSlotSkipped
		}
		payload := workout.VocabPayload{Words: make([]workout.VocabWord, 0, len(units))}
		for _, u := range units {
			payload.Words = append(payload.Words, workout.VocabWord{Text: u.Text, Translation: u.Translation, Definition: u.Definition, Example: u.Example})
		}
		return uc.appendStep(lesson, workout.StepVocab, "Your words", payload)
	}
	return errSlotSkipped
}

func (uc *GenerateLessonUseCase) pickPlannedTitle(ctx context.Context, accountID string, available []films.Film, level string) (films.Film, workout.Position, workout.FilmPlan, error) {
	ready, err := uc.repo.ListPlannedFilmIDs(ctx, workout.PlanStatusReady)
	if err != nil {
		return films.Film{}, workout.Position{}, workout.FilmPlan{}, err
	}
	readySet := make(map[string]bool, len(ready))
	for _, id := range ready {
		readySet[id] = true
	}
	planned := make([]films.Film, 0, len(available))
	for _, f := range available {
		if readySet[f.ID] {
			planned = append(planned, f)
		}
	}
	positions, err := uc.repo.ListPositions(ctx, accountID)
	if err != nil {
		return films.Film{}, workout.Position{}, workout.FilmPlan{}, err
	}
	film, pos, ok := workout.PickTitle(planned, positions, level)
	if !ok {
		return films.Film{}, workout.Position{}, workout.FilmPlan{}, fmt.Errorf("no prepared films for level %s: %w", level, shared.ErrValidation)
	}
	plan, err := uc.repo.GetPlan(ctx, film.ID)
	if err != nil {
		return films.Film{}, workout.Position{}, workout.FilmPlan{}, err
	}
	return film, pos, plan, nil
}

func (uc *GenerateLessonUseCase) generateQuestions(ctx context.Context, recap string, cues []films.Cue, level string) ([]workout.Question, error) {
	system, user := workout.BuildQuestionsPrompt(recap, cues, level, 5)
	return generate(ctx, uc.llm, system, user, workout.ParseQuestions)
}

func (uc *GenerateLessonUseCase) learningWords(ctx context.Context, accountID string) []string {
	units, _, err := uc.words.List(ctx, accountID, vocab.ListFilter{Status: vocab.StatusLearning, Limit: 8})
	if err != nil {
		return nil
	}
	words := make([]string, 0, len(units))
	for _, u := range units {
		words = append(words, u.Text)
	}
	return words
}

func (uc *GenerateLessonUseCase) appendStep(lesson *workout.Lesson, kind, title string, payload any) error {
	step, err := workout.NewStep(fmt.Sprintf("s%d", len(lesson.Steps)+1), kind, title, payload)
	if err != nil {
		return err
	}
	lesson.Steps = append(lesson.Steps, step)
	return nil
}

func warmupItems(items []workout.Item) []workout.WarmupItem {
	out := make([]workout.WarmupItem, 0, len(items))
	for i, it := range items {
		mode := "speak"
		if i%2 == 1 {
			mode = "dictation"
		}
		out = append(out, workout.WarmupItem{Mode: mode, Text: it.Text, FilmID: it.FilmID, StartMs: it.StartMs, EndMs: it.EndMs})
	}
	return out
}

func speakPayload(phrases []workout.KeyPhrase, filmID string) workout.SpeakPayload {
	out := workout.SpeakPayload{Phrases: make([]workout.SpeakPhrase, 0, len(phrases))}
	for _, p := range phrases {
		out.Phrases = append(out.Phrases, workout.SpeakPhrase{Text: p.Text, FilmID: filmID, StartMs: p.StartMs, EndMs: p.EndMs})
	}
	return out
}

func splitItems(items []workout.Item) (phrases []workout.SpeakPhrase, words []workout.VocabWord) {
	for _, it := range items {
		if it.Kind == workout.ItemWord {
			words = append(words, workout.VocabWord{Text: it.Text})
			continue
		}
		phrases = append(phrases, workout.SpeakPhrase{Text: it.Text, FilmID: it.FilmID, StartMs: it.StartMs, EndMs: it.EndMs})
	}
	return phrases, words
}

func episodesOf(all []films.Film, film films.Film) []films.Film {
	if film.Kind != films.KindSeries {
		return []films.Film{film}
	}
	out := []films.Film{}
	for _, f := range all {
		if f.Kind == films.KindSeries && f.SeriesTitle == film.SeriesTitle && f.Status == films.StatusReady {
			out = append(out, f)
		}
	}
	return out
}

func readingLevel(level string) reading.Level {
	switch strings.ToUpper(level) {
	case "A1", "A2":
		return reading.LevelEasy
	case "B2", "C1", "C2":
		return reading.LevelHard
	}
	return reading.LevelMedium
}
