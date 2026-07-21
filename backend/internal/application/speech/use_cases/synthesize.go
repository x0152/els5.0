package usecases

import (
	"context"
	"math/rand"
	"strings"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/speech"
)

type SynthesizeCommand struct {
	Text  string
	Voice string
	Speed float64
}

type SynthesizeResult struct {
	Audio []byte
	Voice string
}

type SynthesizeUseCase struct {
	tts speech.Synthesizer
}

func NewSynthesizeUseCase(tts speech.Synthesizer) *SynthesizeUseCase {
	return &SynthesizeUseCase{tts: tts}
}

func (uc *SynthesizeUseCase) Execute(ctx context.Context, _ *iam.Actor, cmd SynthesizeCommand) (SynthesizeResult, error) {
	voice := normalizeVoice(cmd.Voice)
	speed := cmd.Speed
	if speed < 0.5 || speed > 2.0 {
		speed = 1.0
	}
	audio, err := uc.tts.Synthesize(ctx, cmd.Text, voice, speed)
	if err != nil {
		return SynthesizeResult{}, err
	}
	return SynthesizeResult{Audio: audio, Voice: voice}, nil
}

func normalizeVoice(raw string) string {
	for _, v := range speech.Voices {
		if strings.EqualFold(v, strings.TrimSpace(raw)) {
			return v
		}
	}
	return speech.Voices[rand.Intn(len(speech.Voices))]
}
