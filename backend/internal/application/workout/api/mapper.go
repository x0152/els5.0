package api

import (
	"encoding/json"
	"time"

	usecases "github.com/els/backend/internal/application/workout/use_cases"
	"github.com/els/backend/internal/domain/workout"
	"github.com/els/backend/internal/utils/timex"
)

func toLessonOutput(l workout.Lesson) LessonOutput {
	steps := make([]StepOutput, 0, len(l.Steps))
	for _, s := range l.Steps {
		steps = append(steps, toStepOutput(s))
	}
	return LessonOutput{
		ID:         l.ID,
		Number:     l.Number,
		CycleIndex: l.CycleIndex(),
		Review:     workout.IsReviewNumber(l.Number),
		FilmID:     l.FilmID,
		StartMs:    l.StartMs,
		EndMs:      l.EndMs,
		Status:     l.Status,
		Steps:      steps,
		CreatedAt:  l.CreatedAt.Format(time.RFC3339),
	}
}

func toStepOutput(s workout.Step) StepOutput {
	out := StepOutput{ID: s.ID, Kind: s.Kind, Title: s.Title, Done: s.Done, Score: s.Score}
	switch s.Kind {
	case workout.StepWarmup:
		var p workout.WarmupPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			for _, it := range p.Items {
				out.Warmup = append(out.Warmup, WarmupItemOutput(it))
			}
		}
	case workout.StepWatch:
		var p workout.WatchPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			out.Watch = &WatchOutput{FilmID: p.FilmID, Title: p.Title, StartMs: p.StartMs, EndMs: p.EndMs, Recap: p.Recap, Summary: p.Summary}
		}
	case workout.StepQuestions:
		var p workout.QuestionsPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			for _, q := range p.Questions {
				out.Questions = append(out.Questions, QuestionOutput{Text: q.Text, Options: q.Options, Answer: q.Answer})
			}
		}
	case workout.StepSpeak:
		var p workout.SpeakPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			out.Phrases = toPhrases(p.Phrases)
		}
	case workout.StepDictation:
		var p workout.DictationPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			out.Phrases = toPhrases(p.Sentences)
		}
	case workout.StepReading:
		var p workout.ReadingPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			out.Reading = &ReadingOutput{Title: p.Title, Body: p.Body, Words: p.Words}
		}
	case workout.StepWriting:
		var p workout.WritingPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			out.Writing = &WritingOutput{Scenario: p.Scenario, Dialogue: p.Dialogue}
		}
	case workout.StepGrammar:
		var p workout.GrammarPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			out.Grammar = &GrammarOutput{Topic: p.Topic, Exercises: p.Exercises}
		}
	case workout.StepVocab:
		var p workout.VocabPayload
		if json.Unmarshal(s.Payload, &p) == nil {
			for _, w := range p.Words {
				out.Vocab = append(out.Vocab, VocabWordOutput(w))
			}
		}
	}
	return out
}

func toPhrases(phrases []workout.SpeakPhrase) []PhraseOutput {
	out := make([]PhraseOutput, 0, len(phrases))
	for _, p := range phrases {
		out = append(out, PhraseOutput(p))
	}
	return out
}

func toTodayOutput(res usecases.TodayResult) WorkoutTodayOutput {
	days := make([]string, 0, len(res.Days))
	for _, d := range res.Days {
		days = append(days, d.In(timex.MSK).Format("2006-01-02"))
	}
	out := WorkoutTodayOutput{Streak: res.Streak, Days: days, Completed: res.Completed}
	if res.Lesson != nil {
		lesson := toLessonOutput(*res.Lesson)
		out.Lesson = &lesson
	}
	return out
}

func toSubmitStepCommand(in *SubmitStepInput) usecases.SubmitStepCommand {
	results := make([]workout.ItemResult, 0, len(in.Body.Results))
	for _, r := range in.Body.Results {
		results = append(results, workout.ItemResult{Kind: r.Kind, Text: r.Text, FilmID: r.FilmID, StartMs: r.StartMs, EndMs: r.EndMs, Score: r.Score})
	}
	return usecases.SubmitStepCommand{LessonID: in.ID, StepID: in.StepID, Score: in.Body.Score, Results: results}
}
