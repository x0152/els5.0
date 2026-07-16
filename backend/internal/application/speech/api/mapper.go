package api

import (
	usecases "github.com/els/backend/internal/application/speech/use_cases"
	"github.com/els/backend/internal/domain/speech"
)

func toAssessOutput(a speech.Assessment) AssessOutput {
	words := make([]WordOutput, 0, len(a.Words))
	for _, w := range a.Words {
		phonemes := make([]PhonemeOutput, 0, len(w.Phonemes))
		for _, p := range w.Phonemes {
			phonemes = append(phonemes, PhonemeOutput{
				Expected: p.Expected,
				Heard:    p.Heard,
				Score:    p.Score,
				Verdict:  string(p.Verdict),
			})
		}
		words = append(words, WordOutput{
			Word:     w.Word,
			IPA:      w.IPA,
			Score:    w.Score,
			Phonemes: phonemes,
			Extra:    w.Extra,
		})
	}
	return AssessOutput{Overall: a.Overall, Heard: a.Heard, Words: words}
}

func toFeedbackCommand(in *FeedbackInput) usecases.FeedbackCommand {
	native := in.Body.NativeLanguage
	if native == "" {
		native = "Russian"
	}
	return usecases.FeedbackCommand{
		Text:           in.Body.Text,
		Heard:          in.Body.Heard,
		NativeLanguage: native,
		Issues:         in.Body.Issues,
	}
}

func toFeedbackOutput(fb speech.Feedback) FeedbackOutput {
	tips := make([]FeedbackTipOutput, 0, len(fb.Tips))
	for _, t := range fb.Tips {
		tips = append(tips, FeedbackTipOutput{Sound: t.Sound, Advice: t.Advice})
	}
	return FeedbackOutput{Summary: fb.Summary, Tips: tips}
}

func toPhonemesOutput(items []speech.PhonemeInfo) PhonemesOutput {
	out := make([]PhonemeInfoOutput, 0, len(items))
	for _, p := range items {
		out = append(out, PhonemeInfoOutput{
			Symbol:      p.Symbol,
			Kind:        p.Kind,
			Examples:    p.Examples,
			Description: p.Description,
			Pitfall:     p.Pitfall,
		})
	}
	return PhonemesOutput{Items: out}
}
