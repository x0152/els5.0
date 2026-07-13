package lexicon

import (
	"sort"
	"strings"
	"unicode"
)

func MapAnalysis(raw Analysis) Analysis {
	out := Analysis{
		SentenceCount: raw.SentenceCount,
		Language:      langOrDefault(raw.Language),
		Units:         make([]Unit, 0, len(raw.Units)),
		BaseForms:     make(map[string]BaseForm, len(raw.BaseForms)),
	}
	for key, bf := range raw.BaseForms {
		cleanKey := cleanBaseForm(key)
		if cleanKey == "" {
			continue
		}
		text := cleanBaseForm(bf.Text)
		if text == "" {
			text = cleanKey
		}
		out.BaseForms[cleanKey] = BaseForm{
			Text:     text,
			POS:      strings.TrimSpace(bf.POS),
			IsStop:   bf.IsStop,
			Language: langOrDefault(bf.Language),
		}
	}
	for _, u := range raw.Units {
		bf := cleanBaseForm(u.BaseForm)
		if bf == "" {
			continue
		}
		meta := u.Metadata
		if meta == nil {
			meta = map[string]any{}
		}
		out.Units = append(out.Units, Unit{
			UnitType:    u.UnitType,
			BaseForm:    bf,
			POS:         strings.TrimSpace(u.POS),
			SentenceIdx: u.SentenceIdx,
			Metadata:    meta,
			Language:    langOrDefault(u.Language),
			Spans:       u.Spans,
		})
	}
	return out
}

func cleanBaseForm(s string) string {
	return strings.TrimSpace(strings.Trim(s, "-–—"))
}

func langOrDefault(s string) string {
	if v := strings.TrimSpace(strings.ToLower(s)); v != "" {
		return v
	}
	return "en"
}

func BuildSentenceSegments(units []Unit) []Segment {
	type builder struct {
		start int
		end   int
		spans []Span
	}
	builders := map[int]*builder{}
	for _, u := range units {
		b := builders[u.SentenceIdx]
		if b == nil {
			b = &builder{start: -1, end: -1}
			builders[u.SentenceIdx] = b
		}
		for _, s := range u.Spans {
			if strings.TrimSpace(s.Text) == "" {
				continue
			}
			b.spans = append(b.spans, s)
			if s.End > s.Start {
				if b.start < 0 || s.Start < b.start {
					b.start = s.Start
				}
				if s.End > b.end {
					b.end = s.End
				}
			}
		}
	}
	idxs := make([]int, 0, len(builders))
	for i := range builders {
		idxs = append(idxs, i)
	}
	sort.Ints(idxs)
	segs := make([]Segment, 0, len(idxs))
	for _, i := range idxs {
		b := builders[i]
		start := maxInt(b.start, 0)
		segs = append(segs, Segment{
			Kind:       SegmentKindSentence,
			SegmentIdx: i,
			StartPos:   start,
			EndPos:     maxInt(b.end, start),
			Text:       buildSegmentText(b.spans),
			Metadata:   map[string]any{},
		})
	}
	return segs
}

func buildSegmentText(spans []Span) string {
	filtered := make([]Span, 0, len(spans))
	for _, s := range spans {
		switch strings.ToLower(strings.TrimSpace(s.SpanType)) {
		case "word", "punct", "token":
			filtered = append(filtered, s)
		}
	}
	if len(filtered) == 0 {
		filtered = spans
	}
	best := map[int]Span{}
	for _, s := range filtered {
		cur, ok := best[s.Start]
		if !ok {
			best[s.Start] = s
			continue
		}
		curLen := cur.End - cur.Start
		sLen := s.End - s.Start
		if sLen < curLen || (sLen == curLen && len(s.Text) < len(cur.Text)) {
			best[s.Start] = s
		}
	}
	starts := make([]int, 0, len(best))
	for st := range best {
		starts = append(starts, st)
	}
	sort.Ints(starts)
	var b strings.Builder
	for i, st := range starts {
		tok := strings.TrimSpace(best[st].Text)
		if tok == "" {
			continue
		}
		if i > 0 && !attachToPrev(tok) {
			b.WriteByte(' ')
		}
		b.WriteString(tok)
	}
	return strings.TrimSpace(b.String())
}

func attachToPrev(tok string) bool {
	if strings.HasPrefix(tok, "'") || strings.HasPrefix(tok, "’") {
		return true
	}
	for _, r := range tok {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

type SubtitleSegment struct {
	CueIndex    int
	StartMs     int
	EndMs       int
	StartOffset int
	EndOffset   int
	Text        string
}

func BuildSubtitleInput(cues []Cue) (string, []SubtitleSegment) {
	var b strings.Builder
	segs := make([]SubtitleSegment, 0, len(cues))
	offset := 0
	for _, c := range cues {
		text := strings.TrimSpace(c.Text)
		if text == "" {
			continue
		}
		if len(segs) > 0 {
			b.WriteString("\n\n")
			offset += 2
		}
		start := offset
		b.WriteString(text)
		offset += len([]rune(text))
		segs = append(segs, SubtitleSegment{
			CueIndex:    c.Index,
			StartMs:     c.StartMs,
			EndMs:       c.EndMs,
			StartOffset: start,
			EndOffset:   offset,
			Text:        text,
		})
	}
	return b.String(), segs
}

func AttachSubtitleSegments(a Analysis, segs []SubtitleSegment) Analysis {
	if len(a.Units) == 0 || len(segs) == 0 {
		return a
	}
	sort.Slice(segs, func(i, j int) bool { return segs[i].StartOffset < segs[j].StartOffset })
	for i := range a.Units {
		seg, ok := locateSubtitleSegment(a.Units[i], segs)
		if !ok {
			continue
		}
		a.Units[i].SentenceIdx = seg.CueIndex
		if a.Units[i].Metadata == nil {
			a.Units[i].Metadata = map[string]any{}
		}
		a.Units[i].Metadata["cue_index"] = seg.CueIndex
		a.Units[i].Metadata["start_ms"] = seg.StartMs
		a.Units[i].Metadata["end_ms"] = seg.EndMs
	}
	return a
}

func locateSubtitleSegment(u Unit, segs []SubtitleSegment) (SubtitleSegment, bool) {
	anchor := -1
	for _, s := range u.Spans {
		if strings.TrimSpace(s.Text) == "" {
			continue
		}
		if anchor < 0 || s.Start < anchor {
			anchor = s.Start
		}
	}
	if anchor < 0 {
		return SubtitleSegment{}, false
	}
	idx := sort.Search(len(segs), func(i int) bool { return segs[i].EndOffset > anchor })
	if idx >= len(segs) {
		return SubtitleSegment{}, false
	}
	seg := segs[idx]
	if anchor < seg.StartOffset || anchor >= seg.EndOffset {
		return SubtitleSegment{}, false
	}
	if len(u.Spans) > 0 {
		word := strings.ToLower(strings.TrimSpace(u.Spans[0].Text))
		if word != "" && !strings.Contains(strings.ToLower(seg.Text), word) {
			if idx > 0 && strings.Contains(strings.ToLower(segs[idx-1].Text), word) {
				return segs[idx-1], true
			}
			if idx < len(segs)-1 && strings.Contains(strings.ToLower(segs[idx+1].Text), word) {
				return segs[idx+1], true
			}
		}
	}
	return seg, true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
