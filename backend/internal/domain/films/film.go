package films

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared"
)

const (
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusFailed     = "failed"
)

const (
	KindFilm   = "film"
	KindSeries = "series"
)

var Levels = []string{"A1", "A2", "B1", "B2", "C1", "C2"}

func ParseLevel(s string) (string, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	for _, l := range Levels {
		if s == l {
			return l, nil
		}
	}
	return "", fmt.Errorf("%w: film.level: must be one of %s", shared.ErrValidation, strings.Join(Levels, ", "))
}

type Cue struct {
	Index   int    `json:"index"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
	Text    string `json:"text"`
}

type SubtitleTrack struct {
	Lang  string `json:"lang"`
	Label string `json:"label"`
	Cues  []Cue  `json:"cues"`
}

type AudioVariant struct {
	Lang  string `json:"lang"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

type Series struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PosterPath  string `json:"poster_path"`
}

type Film struct {
	ID            string          `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	PosterPath    string          `json:"poster_path"`
	DurationMs    int             `json:"duration_ms"`
	Status        string          `json:"status"`
	Error         string          `json:"error"`
	Kind          string          `json:"kind"`
	Level         string          `json:"level"`
	SeriesTitle   string          `json:"series_title"`
	Season        int             `json:"season"`
	Episode       int             `json:"episode"`
	AudioVariants []AudioVariant  `json:"audio_variants"`
	Subtitles     []SubtitleTrack `json:"subtitles"`
	CreatedAt     time.Time       `json:"created_at"`
}

func (f Film) Validate() error {
	var errs []error
	if strings.TrimSpace(f.Title) == "" {
		errs = append(errs, fmt.Errorf("film.title: must not be empty"))
	}
	if f.Kind == KindSeries && strings.TrimSpace(f.SeriesTitle) == "" {
		errs = append(errs, fmt.Errorf("film.series_title: must not be empty for series"))
	}
	if _, err := ParseLevel(f.Level); err != nil {
		errs = append(errs, fmt.Errorf("film.level: must be one of %s", strings.Join(Levels, ", ")))
	}
	return shared.Validation(errs...)
}

func PickEnglishSubtitle(tracks []SubtitleTrack) (SubtitleTrack, bool) {
	for _, t := range tracks {
		switch strings.ToLower(strings.TrimSpace(t.Lang)) {
		case "en", "eng", "english":
			return t, true
		}
	}
	return SubtitleTrack{}, false
}

func ParseSRT(data []byte) []Cue {
	text := strings.ReplaceAll(strings.TrimPrefix(string(data), "\ufeff"), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	cues := []Cue{}
	for _, block := range strings.Split(strings.TrimSpace(text), "\n\n") {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) < 2 {
			continue
		}
		idx := len(cues) + 1
		timeLine := 0
		if n, err := strconv.Atoi(strings.TrimSpace(lines[0])); err == nil {
			idx, timeLine = n, 1
		}
		if timeLine >= len(lines) {
			continue
		}
		start, end, ok := parseTimeRange(lines[timeLine])
		if !ok {
			continue
		}
		body := strings.TrimSpace(strings.Join(lines[timeLine+1:], "\n"))
		if body == "" {
			continue
		}
		cues = append(cues, Cue{Index: idx, StartMs: start, EndMs: end, Text: body})
	}
	return cues
}

func parseTimeRange(line string) (int, int, bool) {
	parts := strings.SplitN(line, "-->", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	start, ok1 := parseTimestamp(parts[0])
	end, ok2 := parseTimestamp(parts[1])
	return start, end, ok1 && ok2
}

func parseTimestamp(s string) (int, bool) {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	hms, ms := s, "0"
	if dot := strings.LastIndex(s, "."); dot >= 0 {
		hms, ms = s[:dot], s[dot+1:]
	}
	segs := strings.Split(hms, ":")
	if len(segs) != 3 {
		return 0, false
	}
	h, e1 := strconv.Atoi(segs[0])
	m, e2 := strconv.Atoi(segs[1])
	sec, e3 := strconv.Atoi(segs[2])
	milli, e4 := strconv.Atoi(ms)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return 0, false
	}
	return ((h*60+m)*60+sec)*1000 + milli, true
}
