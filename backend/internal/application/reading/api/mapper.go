package api

import (
	usecases "github.com/els/backend/internal/application/reading/use_cases"
	"github.com/els/backend/internal/domain/reading"
)

func toGenerateTextCommand(in *GenerateTextInput) (usecases.GenerateTextCommand, error) {
	level, err := reading.ParseLevel(in.Body.Level)
	if err != nil {
		return usecases.GenerateTextCommand{}, err
	}
	length, err := reading.ParseLength(in.Body.Length)
	if err != nil {
		return usecases.GenerateTextCommand{}, err
	}
	return usecases.GenerateTextCommand{Topic: in.Body.Topic, UseVocab: in.Body.UseVocab, Level: level, Length: length}, nil
}
