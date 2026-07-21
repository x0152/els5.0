package workout

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/films"
)

const (
	PlanStatusPending = "pending"
	PlanStatusReady   = "ready"
	PlanStatusFailed  = "failed"
)

type KeyPhrase struct {
	Text    string `json:"text"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
	Level   string `json:"level"`
}

type Segment struct {
	Index   int         `json:"index"`
	StartMs int         `json:"start_ms"`
	EndMs   int         `json:"end_ms"`
	Recap   string      `json:"recap"`
	Summary string      `json:"summary"`
	Phrases []KeyPhrase `json:"phrases"`
}

type FilmPlan struct {
	FilmID    string
	Status    string
	Error     string
	Segments  []Segment
	CreatedAt time.Time
}

const segmentationSystem = `You prepare a film or an episode for English lessons using its subtitles.
Each subtitle cue is given as "[index] text".

Split the whole runtime into sequential watch blocks of roughly 4-8 minutes. Rules:
- Blocks must cover the film in order, without gaps: each block starts right after the previous one ends.
- Cut on scene boundaries. A block must END with dialogue-rich material; merge low-dialogue stretches (action, montage) into the neighbouring block instead of making them a block of their own.
- "recap" = 1-2 English sentences reminding what happened BEFORE this block (empty string for the first block). No spoilers of the block itself.
- "summary" = one English sentence about the block, no spoilers of later events.
- "phrases" = 3-8 catchy natural spoken phrases from THIS block worth practising aloud: idioms, phrasal verbs, connected speech. Copy the phrase text verbatim from a single cue and give that cue index. Estimate each phrase's CEFR level (A2..C2).

Return ONLY a JSON object:
{"segments": [{"from_cue": 1, "to_cue": 120, "recap": "", "summary": "...", "phrases": [{"cue": 42, "text": "...", "level": "B1"}]}]}`

func BuildSegmentationPrompt(title string, cues []films.Cue) (system, user string) {
	var b strings.Builder
	fmt.Fprintf(&b, "Title: %s\n\nSubtitles:\n", title)
	for _, c := range cues {
		fmt.Fprintf(&b, "[%d] %s\n", c.Index, strings.ReplaceAll(c.Text, "\n", " "))
	}
	return segmentationSystem, b.String()
}

func ParseSegments(raw string, cues []films.Cue) ([]Segment, error) {
	var out struct {
		Segments []struct {
			FromCue int    `json:"from_cue"`
			ToCue   int    `json:"to_cue"`
			Recap   string `json:"recap"`
			Summary string `json:"summary"`
			Phrases []struct {
				Cue   int    `json:"cue"`
				Text  string `json:"text"`
				Level string `json:"level"`
			} `json:"phrases"`
		} `json:"segments"`
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("parse segments: %w", err)
	}
	byIndex := make(map[int]films.Cue, len(cues))
	for _, c := range cues {
		byIndex[c.Index] = c
	}
	segments := make([]Segment, 0, len(out.Segments))
	for _, s := range out.Segments {
		from, okFrom := byIndex[s.FromCue]
		to, okTo := byIndex[s.ToCue]
		if !okFrom || !okTo || to.EndMs <= from.StartMs {
			continue
		}
		seg := Segment{
			Index:   len(segments),
			StartMs: from.StartMs,
			EndMs:   to.EndMs,
			Recap:   strings.TrimSpace(s.Recap),
			Summary: strings.TrimSpace(s.Summary),
			Phrases: []KeyPhrase{},
		}
		for _, p := range s.Phrases {
			cue, ok := byIndex[p.Cue]
			if !ok || strings.TrimSpace(p.Text) == "" {
				continue
			}
			seg.Phrases = append(seg.Phrases, KeyPhrase{
				Text:    strings.TrimSpace(p.Text),
				StartMs: cue.StartMs,
				EndMs:   cue.EndMs,
				Level:   NormalizeLevel(p.Level),
			})
		}
		segments = append(segments, seg)
	}
	if len(segments) == 0 {
		return nil, fmt.Errorf("parse segments: no valid segments")
	}
	return segments, nil
}

// CuesInRange returns the cues of a subtitle track that fall inside a watch block.
func CuesInRange(track films.SubtitleTrack, startMs, endMs int) []films.Cue {
	out := []films.Cue{}
	for _, c := range track.Cues {
		if c.StartMs >= startMs && c.EndMs <= endMs {
			out = append(out, c)
		}
	}
	return out
}
