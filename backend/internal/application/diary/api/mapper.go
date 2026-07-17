package api

import (
	usecases "github.com/els/backend/internal/application/diary/use_cases"
	"github.com/els/backend/internal/domain/diary"
	"github.com/els/backend/internal/utils/timex"
)

func toSubmitEntryCommand(in *SubmitEntryInput) usecases.SubmitEntryCommand {
	return usecases.SubmitEntryCommand{Text: in.Body.Text, Question: in.Body.Question}
}

func toListEntriesQuery(in *ListEntriesInput) usecases.ListEntriesQuery {
	return usecases.ListEntriesQuery{Limit: in.Limit, Offset: in.Offset}
}

func toTrainerCheckCommand(in *TrainerCheckInput) (usecases.TrainerCheckCommand, error) {
	level, err := diary.ParseTrainerLevel(in.Body.Level)
	if err != nil {
		return usecases.TrainerCheckCommand{}, err
	}
	return usecases.TrainerCheckCommand{Dialogue: in.Body.Dialogue, Draft: in.Body.Draft, Level: level}, nil
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
		Text:         e.Text,
		Reply:        e.Reply,
		NextQuestion: e.NextQuestion,
		NativeSample: e.NativeSample,
		Corrections:  toCorrections(e.Corrections),
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

func toTrainerCheckOutput(v diary.TrainerVerdict) TrainerCheckOutput {
	issues := make([]TrainerIssueOutput, 0, len(v.Issues))
	for _, i := range v.Issues {
		issues = append(issues, TrainerIssueOutput{Fragment: i.Fragment, Severity: string(i.Severity), Hint: i.Hint})
	}
	return TrainerCheckOutput{Pass: v.Pass, Comment: v.Comment, Issues: issues}
}
