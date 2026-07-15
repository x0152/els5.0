package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/lexicon"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/vocab"
)

const commonMediaThreshold = 8

type LexiconReader interface {
	FindOccurrences(ctx context.Context, lemmas []string) ([]lexicon.LemmaOccurrences, error)
}

type UnitsReader interface {
	ExistsText(ctx context.Context, accountID, text string) (bool, error)
}

type AnalyzeItem struct {
	Text        string
	Kind        string
	Description string
	Translation string
	Frequency   int
	Cefr        string
	Common      bool
	Existing    bool
	Total       int
	MediaCount  int
	Media       []lexicon.MediaOccurrence
}

type AnalyzeUseCase struct {
	llm     LLMClient
	lexicon LexiconReader
	units   UnitsReader
}

func NewAnalyzeUseCase(llm LLMClient, lexicon LexiconReader, units UnitsReader) *AnalyzeUseCase {
	return &AnalyzeUseCase{llm: llm, lexicon: lexicon, units: units}
}

func (uc *AnalyzeUseCase) exists(ctx context.Context, accountID, text string) bool {
	if uc.units == nil || strings.TrimSpace(text) == "" {
		return false
	}
	ok, err := uc.units.ExistsText(ctx, accountID, strings.TrimSpace(text))
	return err == nil && ok
}

func (uc *AnalyzeUseCase) Execute(ctx context.Context, actor *iam.Actor, selection, surrounding string) ([]AnalyzeItem, error) {
	if actor == nil {
		return nil, shared.ErrUnauthorized
	}
	selection = strings.TrimSpace(selection)
	if selection == "" {
		return nil, shared.Validation(fmt.Errorf("text: must not be empty"))
	}
	if !uc.llm.Available() {
		return nil, shared.ErrUnavailable
	}

	// 1. Ask the LLM to parse the selection into units with lemmas.
	system, user := vocab.BuildAnalyzePrompt(selection, strings.TrimSpace(surrounding), actor.Account().EnglishLevel(), actor.Account().NativeLanguage())
	raw, err := uc.llm.Chat(ctx, system, user)
	if err != nil {
		return nil, err
	}
	items, err := vocab.ParseAnalyzeResult(raw)
	if err != nil {
		return nil, err
	}

	// 2. Look up lemma occurrences in the media registry (best-effort).
	accountID := actor.AccountID().String()
	occurrences := uc.lookupOccurrences(ctx, items)

	// 3. Merge the parse with occurrences.
	out := buildAnalyzeItems(items, occurrences)
	for i := range out {
		out[i].Existing = uc.exists(ctx, accountID, out[i].Text)
	}
	return out, nil
}

func (uc *AnalyzeUseCase) Stream(ctx context.Context, actor *iam.Actor, selection, surrounding string, emit func(AnalyzeItem)) error {
	if actor == nil {
		return shared.ErrUnauthorized
	}
	selection = strings.TrimSpace(selection)
	if selection == "" {
		return shared.Validation(fmt.Errorf("text: must not be empty"))
	}
	if !uc.llm.Available() {
		return shared.ErrUnavailable
	}

	streamer, ok := uc.llm.(interface {
		ChatStream(context.Context, string, string, func(string)) error
	})
	if !ok {
		items, err := uc.Execute(ctx, actor, selection, surrounding)
		if err != nil {
			return err
		}
		for _, it := range items {
			emit(it)
		}
		return nil
	}

	system, user := vocab.BuildAnalyzePrompt(selection, strings.TrimSpace(surrounding), actor.Account().EnglishLevel(), actor.Account().NativeLanguage())
	accountID := actor.AccountID().String()
	flush := func(line string) {
		if it, ok := vocab.ParseAnalyzeLine(line); ok {
			emit(uc.enrich(ctx, accountID, it))
		}
	}
	var buf strings.Builder
	err := streamer.ChatStream(ctx, system, user, func(delta string) {
		buf.WriteString(delta)
		s := buf.String()
		for {
			i := strings.IndexByte(s, '\n')
			if i < 0 {
				break
			}
			flush(s[:i])
			s = s[i+1:]
		}
		buf.Reset()
		buf.WriteString(s)
	})
	if err != nil {
		return err
	}
	flush(buf.String())
	return nil
}

func (uc *AnalyzeUseCase) enrich(ctx context.Context, accountID string, it vocab.AnalyzeItem) AnalyzeItem {
	occ := uc.lookupOccurrences(ctx, []vocab.AnalyzeItem{it})
	out := buildAnalyzeItems([]vocab.AnalyzeItem{it}, occ)
	item := AnalyzeItem{Text: it.Text, Kind: it.Kind, Description: it.Description, Translation: it.Translation, Frequency: vocab.ClampFrequency(it.Frequency), Cefr: vocab.NormalizeCEFR(it.Cefr)}
	if len(out) > 0 {
		item = out[0]
	}
	item.Existing = uc.exists(ctx, accountID, item.Text)
	return item
}

func (uc *AnalyzeUseCase) lookupOccurrences(ctx context.Context, items []vocab.AnalyzeItem) map[string]lexicon.LemmaOccurrences {
	if uc.lexicon == nil {
		return nil
	}
	lemmas := collectLemmas(items)
	if len(lemmas) == 0 {
		return nil
	}
	found, err := uc.lexicon.FindOccurrences(ctx, lemmas)
	if err != nil {
		return nil
	}
	byLemma := make(map[string]lexicon.LemmaOccurrences, len(found))
	for _, lo := range found {
		byLemma[lo.Lemma] = lo
	}
	return byLemma
}

func collectLemmas(items []vocab.AnalyzeItem) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, it := range items {
		for _, l := range it.Lemmas {
			l = strings.ToLower(strings.TrimSpace(l))
			if l == "" {
				continue
			}
			if _, ok := seen[l]; ok {
				continue
			}
			seen[l] = struct{}{}
			out = append(out, l)
		}
	}
	return out
}

func buildAnalyzeItems(items []vocab.AnalyzeItem, occurrences map[string]lexicon.LemmaOccurrences) []AnalyzeItem {
	out := make([]AnalyzeItem, 0, len(items))
	for _, it := range items {
		matched := make([]lexicon.LemmaOccurrences, 0, len(it.Lemmas))
		for _, l := range it.Lemmas {
			if lo, ok := occurrences[strings.ToLower(strings.TrimSpace(l))]; ok {
				matched = append(matched, lo)
			}
		}
		media, total, mediaCount, stop := combineMedia(matched)
		item := AnalyzeItem{
			Text:        it.Text,
			Kind:        it.Kind,
			Description: it.Description,
			Translation: it.Translation,
			Frequency:   vocab.ClampFrequency(it.Frequency),
			Cefr:        vocab.NormalizeCEFR(it.Cefr),
			Total:       total,
			MediaCount:  mediaCount,
			Media:       media,
			Common:      stop || mediaCount > commonMediaThreshold,
		}
		if item.Common {
			item.Media = nil
		}
		out = append(out, item)
	}
	return out
}

func combineMedia(occs []lexicon.LemmaOccurrences) (media []lexicon.MediaOccurrence, total, mediaCount int, stop bool) {
	byMedia := make(map[string]*lexicon.MediaOccurrence)
	order := make([]string, 0)
	for _, lo := range occs {
		stop = stop || lo.IsStop
		for _, m := range lo.Media {
			existing, ok := byMedia[m.MediaID]
			if !ok {
				cp := m
				cp.Spots = append([]lexicon.Spot(nil), m.Spots...)
				byMedia[m.MediaID] = &cp
				order = append(order, m.MediaID)
				continue
			}
			seen := make(map[int]bool, len(existing.Spots))
			for _, s := range existing.Spots {
				seen[s.Ref] = true
			}
			for _, s := range m.Spots {
				if !seen[s.Ref] {
					existing.Spots = append(existing.Spots, s)
					seen[s.Ref] = true
				}
			}
		}
	}
	media = make([]lexicon.MediaOccurrence, 0, len(order))
	for _, id := range order {
		m := byMedia[id]
		m.Count = len(m.Spots)
		total += m.Count
		media = append(media, *m)
	}
	return media, total, len(media), stop
}
