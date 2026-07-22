package api

import (
	usecases "github.com/els/backend/internal/application/diary/use_cases"
	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/utils/timex"
)

func toSubmitEntryCommand(in *SubmitEntryInput) usecases.SubmitEntryCommand {
	return usecases.SubmitEntryCommand{Text: in.Body.Text, Question: in.Body.Question, Draft: in.Body.Draft}
}

func toCheckEntryOutput(res quest.GrammarCheck) CheckEntryOutput {
	out := CheckEntryOutput{OK: res.OK, Errors: make([]GrammarErrorOutput, 0, len(res.Errors))}
	for _, e := range res.Errors {
		out.Errors = append(out.Errors, GrammarErrorOutput{
			Original:    e.Original,
			Correction:  e.Correction,
			Explanation: e.Explanation,
			Type:        e.Type,
		})
	}
	return out
}

func toListEntriesQuery(in *ListEntriesInput) usecases.ListEntriesQuery {
	return usecases.ListEntriesQuery{Limit: in.Limit, Offset: in.Offset}
}

func toCorrections(items []diary.Correction) []CorrectionOutput {
	out := make([]CorrectionOutput, 0, len(items))
	for _, c := range items {
		out = append(out, CorrectionOutput{
			Sentence:    c.Sentence,
			Fragment:    c.Fragment,
			Correction:  c.Correction,
			Description: c.Description,
		})
	}
	return out
}

func toEntryOutput(e diary.Entry) EntryOutput {
	return EntryOutput{
		ID:           e.ID,
		Date:         timex.FormatDate(e.Date),
		Question:     e.Question,
		Draft:        e.Draft,
		Text:         e.Text,
		Reply:        e.Reply,
		NextQuestion: e.NextQuestion,
		NativeSample: e.NativeSample,
		Corrections:  toCorrections(e.Corrections),
		Status:       e.Status,
		CreatedAt:    e.CreatedAt,
	}
}

func toTodayOutput(res usecases.TodayResult) TodayOutput {
	out := TodayOutput{
		Question: res.Question,
		Warmup:   toCorrections(res.Warmup),
		Streak:   res.Streak,
	}
	if res.Entry != nil {
		entry := toEntryOutput(*res.Entry)
		out.Entry = &entry
	}
	return out
}

func toEntriesOutput(res usecases.ListEntriesResult, limit, offset int32) EntriesOutput {
	items := make([]EntryOutput, 0, len(res.Items))
	for _, e := range res.Items {
		items = append(items, toEntryOutput(e))
	}
	return EntriesOutput{Items: items, Total: res.Total, Limit: limit, Offset: offset}
}
