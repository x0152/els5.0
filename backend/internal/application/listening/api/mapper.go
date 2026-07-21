package api

import (
	usecases "github.com/els/backend/internal/application/listening/use_cases"
	"github.com/els/backend/internal/domain/listening"
)

func toGenerateDictationCommand(in *GenerateDictationInput) (usecases.GenerateDictationCommand, error) {
	level, err := listening.ParseLevel(in.Body.Level)
	if err != nil {
		return usecases.GenerateDictationCommand{}, err
	}
	count, err := listening.ParseSentenceCount(in.Body.Count)
	if err != nil {
		return usecases.GenerateDictationCommand{}, err
	}
	return usecases.GenerateDictationCommand{Topic: in.Body.Topic, UseVocab: in.Body.UseVocab, Level: level, Count: count}, nil
}
