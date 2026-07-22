package workout

import (
	"sort"
	"time"

	"github.com/els/backend/internal/domain/films"
)

// Position tracks how far the learner has advanced through one title (a series or a standalone film).
type Position struct {
	AccountID   string
	Title       string
	FilmID      string
	NextSegment int
	UsedAt      time.Time
}

// sitcomMaxMs: episodes up to this length are watched whole in one lesson.
const sitcomMaxMs = 26 * 60 * 1000

// watchBudgetMs: for longer films a lesson covers consecutive segments up to this runtime.
const watchBudgetMs = 20 * 60 * 1000

var slotPool = []string{StepReading, StepWriting, StepGrammar, StepVocab, StepDictation}

// SlotsPerLesson: every lesson carries all skill slots; PickSlots only decides their order.
const SlotsPerLesson = 5

// TitleKey groups episodes of one series under a single watching position.
func TitleKey(f films.Film) string {
	if f.Kind == films.KindSeries {
		return "series:" + f.SeriesTitle
	}
	return "film:" + f.ID
}

// PickTitle chooses which title the lesson anchors on: among suitable films prefer the
// title the learner is already inside but hasn't seen for the longest time.
func PickTitle(available []films.Film, positions []Position, userLevel string) (films.Film, Position, bool) {
	byKey := map[string][]films.Film{}
	order := []string{}
	for _, f := range available {
		if f.Status != films.StatusReady || !LevelAtMost(f.Level, userLevel) {
			continue
		}
		key := TitleKey(f)
		if _, seen := byKey[key]; !seen {
			order = append(order, key)
		}
		byKey[key] = append(byKey[key], f)
	}
	if len(order) == 0 {
		return films.Film{}, Position{}, false
	}
	for key := range byKey {
		sort.SliceStable(byKey[key], func(i, j int) bool {
			a, b := byKey[key][i], byKey[key][j]
			if a.Season != b.Season {
				return a.Season < b.Season
			}
			return a.Episode < b.Episode
		})
	}
	posByKey := map[string]Position{}
	for _, p := range positions {
		posByKey[p.Title] = p
	}
	sort.SliceStable(order, func(i, j int) bool {
		pi, iOk := posByKey[order[i]]
		pj, jOk := posByKey[order[j]]
		if iOk != jOk {
			return iOk
		}
		return pi.UsedAt.Before(pj.UsedAt)
	})
	for _, key := range order {
		pos, ok := posByKey[key]
		if !ok {
			return byKey[key][0], Position{Title: key, FilmID: byKey[key][0].ID}, true
		}
		for i, f := range byKey[key] {
			if f.ID != pos.FilmID {
				continue
			}
			return byKey[key][i], pos, true
		}
	}
	first := byKey[order[0]][0]
	return first, Position{Title: TitleKey(first), FilmID: first.ID}, true
}

// WatchRange picks the consecutive segments the lesson covers and the advanced position.
// A short episode is watched whole; a long film goes in ~9-minute chunks. When the title
// runs out of segments the position moves to the next episode (or wraps to the start).
func WatchRange(film films.Film, plan FilmPlan, pos Position, episodes []films.Film) ([]Segment, Position) {
	start := pos.NextSegment
	if start >= len(plan.Segments) {
		start = 0
	}
	var picked []Segment
	if film.DurationMs > 0 && film.DurationMs <= sitcomMaxMs {
		picked = plan.Segments
		start = 0
	} else {
		budget := 0
		for i := start; i < len(plan.Segments); i++ {
			s := plan.Segments[i]
			if len(picked) > 0 && budget+(s.EndMs-s.StartMs) > watchBudgetMs {
				break
			}
			picked = append(picked, s)
			budget += s.EndMs - s.StartMs
		}
	}
	next := Position{AccountID: pos.AccountID, Title: pos.Title, FilmID: film.ID, NextSegment: start + len(picked)}
	if next.NextSegment >= len(plan.Segments) {
		next.FilmID = nextEpisodeID(film, episodes)
		next.NextSegment = 0
	}
	return picked, next
}

func nextEpisodeID(current films.Film, episodes []films.Film) string {
	sort.SliceStable(episodes, func(i, j int) bool {
		if episodes[i].Season != episodes[j].Season {
			return episodes[i].Season < episodes[j].Season
		}
		return episodes[i].Episode < episodes[j].Episode
	})
	for i, f := range episodes {
		if f.ID == current.ID && i+1 < len(episodes) {
			return episodes[i+1].ID
		}
	}
	return current.ID
}

// PickSlots orders the variable skills for this lesson: the least practised slot kinds
// over the recent lessons come first, ties rotate with the lesson number so the mix keeps
// changing. The caller takes the first SlotsPerLesson kinds it can actually fill.
func PickSlots(recent []Lesson, lessonNumber int) []string {
	usage := map[string]int{}
	for _, l := range recent {
		for _, s := range l.Steps {
			usage[s.Kind]++
		}
	}
	pool := make([]string, len(slotPool))
	copy(pool, slotPool)
	sort.SliceStable(pool, func(i, j int) bool {
		if usage[pool[i]] != usage[pool[j]] {
			return usage[pool[i]] < usage[pool[j]]
		}
		return (indexOf(pool[i])+lessonNumber)%len(slotPool) < (indexOf(pool[j])+lessonNumber)%len(slotPool)
	})
	return pool
}

func indexOf(kind string) int {
	for i, k := range slotPool {
		if k == kind {
			return i
		}
	}
	return 0
}

// LevelPhrases filters segment key phrases to the learner's level, hardest suitable first.
func LevelPhrases(segments []Segment, userLevel string, limit int) []KeyPhrase {
	out := []KeyPhrase{}
	for _, s := range segments {
		for _, p := range s.Phrases {
			if LevelAtMost(p.Level, userLevel) {
				out = append(out, p)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return levelRank[out[i].Level] > levelRank[out[j].Level]
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
