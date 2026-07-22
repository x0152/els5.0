package workout_test

import (
	"testing"
	"time"

	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/workout"
)

func readyFilm(id, kind, series string, season, episode, durationMs int, level string) films.Film {
	return films.Film{ID: id, Title: id, Status: films.StatusReady, Kind: kind, SeriesTitle: series, Season: season, Episode: episode, DurationMs: durationMs, Level: level}
}

func TestPickTitle(t *testing.T) {
	friends1 := readyFilm("f1", films.KindSeries, "Friends", 1, 1, 22*60*1000, "B1")
	friends2 := readyFilm("f2", films.KindSeries, "Friends", 1, 2, 22*60*1000, "B1")
	movie := readyFilm("m1", films.KindFilm, "", 0, 0, 100*60*1000, "B1")
	hard := readyFilm("h1", films.KindFilm, "", 0, 0, 100*60*1000, "C2")

	t.Run("filters by level", func(t *testing.T) {
		// act
		_, _, ok := workout.PickTitle([]films.Film{hard}, nil, "B1")
		// assert
		if ok {
			t.Fatal("C2 film picked for B1 learner")
		}
	})

	t.Run("prefers the least recently used title", func(t *testing.T) {
		// arrange
		positions := []workout.Position{
			{Title: "series:Friends", FilmID: "f1", UsedAt: day("2026-07-18")},
			{Title: "film:m1", FilmID: "m1", UsedAt: day("2026-07-10")},
		}
		// act
		film, _, ok := workout.PickTitle([]films.Film{friends1, friends2, movie}, positions, "B1")
		// assert
		if !ok || film.ID != "m1" {
			t.Fatalf("got %q, want m1", film.ID)
		}
	})

	t.Run("new title starts at the first episode", func(t *testing.T) {
		// act
		film, pos, ok := workout.PickTitle([]films.Film{friends2, friends1}, nil, "B1")
		// assert
		if !ok || film.ID != "f1" || pos.NextSegment != 0 {
			t.Fatalf("got film %q segment %d", film.ID, pos.NextSegment)
		}
	})
}

func segs(durationsMin ...int) []workout.Segment {
	out := []workout.Segment{}
	cursor := 0
	for i, d := range durationsMin {
		out = append(out, workout.Segment{Index: i, StartMs: cursor, EndMs: cursor + d*60*1000})
		cursor += d * 60 * 1000
	}
	return out
}

func TestWatchRange(t *testing.T) {
	t.Run("sitcom episode is watched whole", func(t *testing.T) {
		// arrange
		film := readyFilm("f1", films.KindSeries, "Friends", 1, 1, 22*60*1000, "B1")
		next := readyFilm("f2", films.KindSeries, "Friends", 1, 2, 22*60*1000, "B1")
		plan := workout.FilmPlan{Segments: segs(7, 8, 7)}
		// act
		picked, pos := workout.WatchRange(film, plan, workout.Position{Title: "series:Friends", FilmID: "f1"}, []films.Film{film, next})
		// assert
		if len(picked) != 3 {
			t.Fatalf("got %d segments, want all 3", len(picked))
		}
		if pos.FilmID != "f2" || pos.NextSegment != 0 {
			t.Fatalf("position not advanced to next episode: %+v", pos)
		}
	})

	t.Run("long film goes in chunks", func(t *testing.T) {
		// arrange
		film := readyFilm("m1", films.KindFilm, "", 0, 0, 120*60*1000, "B1")
		plan := workout.FilmPlan{Segments: segs(8, 8, 8, 8)}
		// act
		picked, pos := workout.WatchRange(film, plan, workout.Position{Title: "film:m1", FilmID: "m1"}, []films.Film{film})
		// assert
		if len(picked) != 2 {
			t.Fatalf("got %d segments, want 2 within the 20-minute budget", len(picked))
		}
		if pos.NextSegment != 2 || pos.FilmID != "m1" {
			t.Fatalf("position wrong: %+v", pos)
		}
	})
}

func lessonWith(kinds ...string) workout.Lesson {
	l := workout.Lesson{}
	for _, k := range kinds {
		l.Steps = append(l.Steps, workout.Step{Kind: k})
	}
	return l
}

func TestPickSlots(t *testing.T) {
	// arrange: reading and writing dominated the recent lessons.
	recent := []workout.Lesson{
		lessonWith(workout.StepReading, workout.StepWriting),
		lessonWith(workout.StepReading, workout.StepWriting),
	}
	// act
	slots := workout.PickSlots(recent, 3)
	// assert: the overused kinds go to the back of the queue.
	for _, s := range slots[:2] {
		if s == workout.StepReading || s == workout.StepWriting {
			t.Fatalf("overused slot %q picked first", s)
		}
	}
}

func TestPickWarmupAndReview(t *testing.T) {
	items := []workout.Item{
		{Text: "fresh", LessonNumber: 9, LastScore: 90},
		{Text: "yesterday", LessonNumber: 9, LastScore: 90},
		{Text: "three ago", LessonNumber: 7, LastScore: 90},
		{Text: "weak", LessonNumber: 5, LastScore: 30},
		{Text: "old strong", LessonNumber: 1, LastScore: 95},
	}

	// act: warm-up for lesson 10 → age 1, age 3 and weak items, worst first.
	warmup := workout.PickWarmup(items, 10, 10)
	// assert
	if len(warmup) != 4 || warmup[0].Text != "weak" {
		t.Fatalf("warmup wrong: %+v", warmup)
	}

	// act: review at lesson 14 keeps only the finishing cycle (lessons 8+).
	review := workout.PickReview(items, 14, 10)
	// assert
	for _, it := range review {
		if it.LessonNumber < 8 {
			t.Fatalf("item outside the cycle picked: %+v", it)
		}
	}
}

func TestNormalizeLevel(t *testing.T) {
	cases := map[string]string{"b2": "B2", "Intermediate B1": "B1", "": "B1", "C1 advanced": "C1"}
	for in, want := range cases {
		if got := workout.NormalizeLevel(in); got != want {
			t.Fatalf("NormalizeLevel(%q) = %q, want %q", in, got, want)
		}
	}
	if workout.LevelAtMost("C1", "B1") || !workout.LevelAtMost("A2", "B1") {
		t.Fatal("LevelAtMost wrong")
	}
	_ = time.Now
}
