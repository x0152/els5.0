package api

import (
	usecases "github.com/els/backend/internal/application/writing/use_cases"
	"github.com/els/backend/internal/domain/writing"
)

func toTrainerCheckCommand(in *TrainerCheckInput) (usecases.TrainerCheckCommand, error) {
	level, err := writing.ParseTrainerLevel(in.Body.Level)
	if err != nil {
		return usecases.TrainerCheckCommand{}, err
	}
	return usecases.TrainerCheckCommand{Dialogue: in.Body.Dialogue, Draft: in.Body.Draft, Level: level}, nil
}

func toTrainerCheckOutput(v writing.TrainerVerdict) TrainerCheckOutput {
	issues := make([]TrainerIssueOutput, 0, len(v.Issues))
	for _, i := range v.Issues {
		issues = append(issues, TrainerIssueOutput{Fragment: i.Fragment, Severity: string(i.Severity), Hint: i.Hint})
	}
	return TrainerCheckOutput{Pass: v.Pass, Comment: v.Comment, Issues: issues}
}
