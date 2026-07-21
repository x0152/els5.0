package runtime

import (
	"strings"

	"github.com/els/backend/internal/domain/quest"
)

func commonFallbackVoice() string { return quest.DefaultVoice }

func ensureMissionVoices(mission *quest.CustomMission) {
	if mission == nil {
		return
	}
	usedVoices := make(map[string]bool)
	for i := range mission.Characters {
		ch := &mission.Characters[i]
		// Prefer the voice the LLM picked for the character; fall back only when it is invalid or already taken.
		voice := quest.EnsureVoiceByGender(ch.Voice, ch.Name, ch.Gender)
		for attempts := 0; attempts < len(quest.Voices()) && usedVoices[voice]; attempts++ {
			voice = quest.AssignCharacterFallback(ch.Name+":"+strings.Repeat("x", attempts+1), i+attempts, ch.Gender)
		}
		ch.Voice = voice
		usedVoices[voice] = true
	}
	mission.NarratorVoice = quest.EnsureNarratorVoice(mission.NarratorVoice, commonFallbackVoice())
}

func missionCharacterVoice(mission *quest.CustomMission, name string) (string, bool) {
	if mission == nil {
		return "", false
	}
	target := strings.TrimSpace(name)
	if target == "" {
		return "", false
	}
	targetLower := strings.ToLower(target)
	for _, ch := range mission.Characters {
		if strings.EqualFold(strings.TrimSpace(ch.Name), target) {
			return quest.EnsureVoice(ch.Voice, commonFallbackVoice()), true
		}
	}
	for _, ch := range mission.Characters {
		charName := strings.TrimSpace(ch.Name)
		if charName == "" {
			continue
		}
		charLower := strings.ToLower(charName)
		firstName := strings.Split(charLower, " ")[0]
		if firstName == "" {
			continue
		}
		if strings.HasPrefix(targetLower, firstName) || strings.HasPrefix(charLower, targetLower) {
			return quest.EnsureVoice(ch.Voice, commonFallbackVoice()), true
		}
	}
	return "", false
}

func applySceneVoices(mission *quest.CustomMission, scene *quest.DynamicScene) {
	if scene == nil {
		return
	}
	commonVoice := commonFallbackVoice()
	narratorVoice := ""
	if mission != nil && mission.NarratorVoice != "" {
		narratorVoice = mission.NarratorVoice
	}
	scene.NarrationVoice = quest.EnsureNarratorVoice(narratorVoice, commonVoice)

	for i := range scene.Present {
		if voice, ok := missionCharacterVoice(mission, scene.Present[i].Name); ok {
			scene.Present[i].Voice = voice
			continue
		}
		scene.Present[i].Voice = quest.EnsureVoice("", commonVoice)
	}
}

func applyWorldVoices(mission *quest.CustomMission, world *quest.WorldResult) {
	if world == nil {
		return
	}
	commonVoice := commonFallbackVoice()
	narratorVoice := ""
	if mission != nil && mission.NarratorVoice != "" {
		narratorVoice = mission.NarratorVoice
	}
	world.NarrationVoice = quest.EnsureNarratorVoice(narratorVoice, commonVoice)

	for i := range world.Responses {
		if voice, ok := missionCharacterVoice(mission, world.Responses[i].Name); ok {
			world.Responses[i].Voice = voice
			continue
		}
		world.Responses[i].Voice = quest.EnsureVoice("", commonVoice)
	}
}
