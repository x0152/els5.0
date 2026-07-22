package api

import (
	"github.com/els/backend/internal/domain/studio"
)

func toAreaOutput(a studio.AreaStats) AreaOutput {
	return AreaOutput{
		ID:        a.ID,
		Title:     a.Title,
		Icon:      a.Icon,
		Total:     a.Total,
		Done:      a.Done,
		Due:       a.Due,
		CreatedAt: a.CreatedAt,
	}
}

func toAreasOutput(areas []studio.AreaStats) AreasOutput {
	items := make([]AreaOutput, 0, len(areas))
	for _, a := range areas {
		items = append(items, toAreaOutput(a))
	}
	return AreasOutput{Items: items}
}

func toItemOutput(i studio.Item) ItemOutput {
	return ItemOutput{
		ID:                i.ID,
		AreaID:            i.AreaID,
		Text:              i.Text,
		Transcription:     i.Transcription,
		Translation:       i.Translation,
		Explanation:       i.Explanation,
		ExplanationNative: i.ExplanationNative,
		Example:           i.Example,
		Task:              i.Task,
		Listened:          i.Listened,
		Spoken:            i.Spoken,
		Written:           i.Written,
		Recalled:          i.Recalled,
		ReviewStage:       i.ReviewStage,
		NextReviewAt:      i.NextReviewAt,
		CreatedAt:         i.CreatedAt,
	}
}

func toItemsOutput(items []studio.Item) ItemsOutput {
	out := make([]ItemOutput, 0, len(items))
	for _, i := range items {
		out = append(out, toItemOutput(i))
	}
	return ItemsOutput{Items: out}
}

func toCheckReplyOutput(c studio.ReplyCheck) CheckReplyOutput {
	return CheckReplyOutput{OK: c.OK, Comment: c.Comment}
}
