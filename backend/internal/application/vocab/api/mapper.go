package api

import (
	"time"

	usecases "github.com/els/backend/internal/application/vocab/use_cases"
	"github.com/els/backend/internal/domain/lexicon"
	"github.com/els/backend/internal/domain/vocab"
)

func toUnitOutput(u vocab.Unit) UnitOutput {
	return UnitOutput{
		ID:            u.ID,
		Text:          u.Text,
		Kind:          string(u.Kind),
		Transcription: u.Transcription,
		Translation:   u.Translation,
		Definition:    u.Definition,
		Example:       u.Example,
		Frequency:     u.Frequency,
		CEFR:          u.CEFR,
		Status:        string(u.Status),
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}
}

func toAddUnitOutput(res usecases.AddUnitResult) AddUnitOutput {
	out := AddUnitOutput{
		Correct:     res.Correct,
		Correction:  res.Correction,
		Explanation: res.Explanation,
	}
	if res.Unit != nil {
		unit := toUnitOutput(*res.Unit)
		out.Unit = &unit
	}
	return out
}

func toOccurrenceOutputs(media []lexicon.MediaOccurrence) []OccurrenceOutput {
	out := make([]OccurrenceOutput, 0, len(media))
	for _, m := range media {
		spots := make([]SpotOutput, 0, len(m.Spots))
		for _, s := range m.Spots {
			spots = append(spots, SpotOutput{Ref: s.Ref, Example: s.Example})
		}
		out = append(out, OccurrenceOutput{
			MediaID:     m.MediaID,
			MediaType:   m.MediaType,
			Title:       m.Title,
			Kind:        m.Kind,
			SeriesTitle: m.SeriesTitle,
			Season:      m.Season,
			Episode:     m.Episode,
			Author:      m.Author,
			Count:       m.Count,
			Spots:       spots,
		})
	}
	return out
}

func toAnalyzeItemOutput(it usecases.AnalyzeItem) AnalyzeItemOutput {
	return AnalyzeItemOutput{
		Text:        it.Text,
		Kind:        it.Kind,
		Description: it.Description,
		Frequency:   it.Frequency,
		CEFR:        it.Cefr,
		Common:      it.Common,
		Existing:    it.Existing,
		Total:       it.Total,
		MediaCount:  it.MediaCount,
		Media:       toOccurrenceOutputs(it.Media),
	}
}

func toAnalyzeOutput(items []usecases.AnalyzeItem) AnalyzeOutput {
	out := make([]AnalyzeItemOutput, 0, len(items))
	for _, it := range items {
		out = append(out, toAnalyzeItemOutput(it))
	}
	return AnalyzeOutput{Items: out}
}

func toOccurrencesOutput(res usecases.OccurrenceResult) OccurrencesOutput {
	return OccurrencesOutput{
		Common:     res.Common,
		Total:      res.Total,
		MediaCount: res.MediaCount,
		Media:      toOccurrenceOutputs(res.Media),
	}
}

func toUnitsOutput(res usecases.ListUnitsResult) UnitsOutput {
	items := make([]UnitOutput, 0, len(res.Items))
	for _, u := range res.Items {
		items = append(items, toUnitOutput(u))
	}
	return UnitsOutput{Items: items, Total: res.Total, Limit: res.Limit, Offset: res.Offset}
}

func toPracticeOutput(s vocab.PracticeSession) PracticeOutput {
	words := make([]UnitOutput, 0, len(s.Words))
	for _, u := range s.Words {
		words = append(words, toUnitOutput(u))
	}
	answers := make(map[string]PracticeAnswerDTO, len(s.Answers))
	for k, a := range s.Answers {
		answers[k] = PracticeAnswerDTO(a)
	}
	return PracticeOutput{
		ID:        s.ID,
		Status:    s.Status,
		Error:     s.Error,
		Exercises: s.Exercises,
		Words:     words,
		Answers:   answers,
		Completed: s.Completed,
	}
}

func toPracticeAnswers(in map[string]PracticeAnswerDTO) map[string]vocab.PracticeAnswer {
	out := make(map[string]vocab.PracticeAnswer, len(in))
	for k, a := range in {
		out[k] = vocab.PracticeAnswer(a)
	}
	return out
}

func toCheckPracticeOutput(res vocab.PracticeCheckResult) CheckPracticeOutput {
	return CheckPracticeOutput{Correct: res.Correct, Correction: res.Correction, Explanation: res.Explanation}
}

func toListFilter(in *ListUnitsInput) vocab.ListFilter {
	return vocab.ListFilter{
		Search: in.Search,
		Status: vocab.Status(in.Status),
		Limit:  in.Limit,
		Offset: in.Offset,
	}
}
