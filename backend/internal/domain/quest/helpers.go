package quest

import (
	"errors"
	"hash/fnv"
	"strconv"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

type VoiceOption struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

const (
	DefaultVoice         = "Jasper"
	DefaultNarratorVoice = "Bella"
)

var kittenVoices = []VoiceOption{
	{Name: "Bella", Gender: "female"},
	{Name: "Jasper", Gender: "male"},
	{Name: "Luna", Gender: "female"},
	{Name: "Bruno", Gender: "male"},
	{Name: "Rosie", Gender: "female"},
	{Name: "Hugo", Gender: "male"},
	{Name: "Kiki", Gender: "female"},
	{Name: "Leo", Gender: "male"},
}

func Voices() []VoiceOption {
	out := make([]VoiceOption, len(kittenVoices))
	copy(out, kittenVoices)
	return out
}

func NormalizeVoice(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	for _, voice := range kittenVoices {
		if strings.EqualFold(voice.Name, value) {
			return voice.Name
		}
	}
	return ""
}

func FallbackVoice(seed string) string {
	return fallbackVoiceFromPool(seed, kittenVoices)
}

func FallbackVoiceByGender(seed string, gender string) string {
	pool := voicesByGender(gender)
	if len(pool) == 0 {
		return FallbackVoice(seed)
	}
	return fallbackVoiceFromPool(seed, pool)
}

func voicesByGender(gender string) []VoiceOption {
	g := strings.ToLower(strings.TrimSpace(gender))
	if g == "" {
		return nil
	}
	var out []VoiceOption
	for _, v := range kittenVoices {
		if strings.ToLower(v.Gender) == g {
			out = append(out, v)
		}
	}
	return out
}

func fallbackVoiceFromPool(seed string, pool []VoiceOption) string {
	if len(pool) == 0 {
		return DefaultVoice
	}
	key := strings.TrimSpace(seed)
	if key == "" {
		return pool[0].Name
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	idx := int(hasher.Sum32()) % len(pool)
	return pool[idx].Name
}

func EnsureVoice(preferred string, fallbackSeed string) string {
	return EnsureVoiceByGender(preferred, fallbackSeed, "")
}

func EnsureVoiceByGender(preferred string, fallbackSeed string, gender string) string {
	if normalized := NormalizeVoice(preferred); normalized != "" {
		if gender == "" || VoiceMatchesGender(normalized, gender) {
			return normalized
		}
	}
	if g := strings.TrimSpace(gender); g != "" {
		return FallbackVoiceByGender(fallbackSeed, g)
	}
	if fallbackVoice := NormalizeVoice(fallbackSeed); fallbackVoice != "" {
		return fallbackVoice
	}
	return FallbackVoice(fallbackSeed)
}

func VoiceMatchesGender(voiceName string, gender string) bool {
	g := strings.ToLower(strings.TrimSpace(gender))
	if g == "" {
		return true
	}
	for _, v := range kittenVoices {
		if strings.EqualFold(v.Name, voiceName) {
			return strings.ToLower(v.Gender) == g
		}
	}
	return true
}

func EnsureNarratorVoice(preferred string, fallbackSeed string) string {
	if normalized := NormalizeVoice(preferred); normalized != "" {
		return normalized
	}
	if fallbackVoice := NormalizeVoice(fallbackSeed); fallbackVoice != "" {
		return fallbackVoice
	}
	if strings.TrimSpace(fallbackSeed) == "" {
		return DefaultNarratorVoice
	}
	return FallbackVoice("narrator:" + fallbackSeed)
}

func AssignCharacterFallback(name string, index int, gender string) string {
	seed := strings.TrimSpace(name)
	if seed == "" {
		seed = "character:" + strconv.Itoa(index)
	}
	if g := strings.TrimSpace(gender); g != "" {
		return FallbackVoiceByGender(seed, g)
	}
	return FallbackVoice(seed)
}

type MissionSanitizer struct{}

func (s MissionSanitizer) Execute(mission *CustomMission) CustomMission {
	clean := *mission
	clean.SecretEnding = ""

	// Purely internal engine fields: history summarization, pacing, legacy XP,
	// image generation timings. The scene archive is slimmed to an outcomes map
	// (stage + summary) for scene transitions in the UI.
	// npcStates, currentScene, and userPrompt remain — they may be useful to the UI.
	if len(clean.Scenes) > 0 {
		scenes := make([]DynamicScene, len(clean.Scenes))
		for i, sc := range clean.Scenes {
			scenes[i] = DynamicScene{Stage: sc.Stage, Summary: sc.Summary, IsFinal: sc.IsFinal}
		}
		clean.Scenes = scenes
	}
	clean.HistorySummary = ""
	clean.SummarizedUpToTurn = 0
	clean.EstimatedScenes = 0
	clean.SkillSignals = nil
	clean.SkillCategories = nil
	clean.SkillsEarned = nil
	clean.TotalXP = 0
	clean.CoverImageGenStartedAt = ""
	clean.SceneImageGenStartedAt = nil
	clean.CharacterAvatarGenStartedAt = nil

	if clean.CurrentScene != nil {
		sceneCopy := *clean.CurrentScene
		sceneCopy.Trigger = ""
		sceneCopy.ScenePurpose = ""
		clean.CurrentScene = &sceneCopy
	}

	if len(clean.Characters) > 0 {
		characters := make([]Character, len(clean.Characters))
		copy(characters, clean.Characters)
		for i := range characters {
			characters[i].Personality = ""
			characters[i].SpeechStyle = ""
			characters[i].Motivation = ""
			characters[i].Arc = ""
			characters[i].InitialTrust = 0
		}
		clean.Characters = characters
	}

	if len(clean.PlotPoints) > 0 {
		sanitized := make([]PlotPoint, len(clean.PlotPoints))
		for i, pp := range clean.PlotPoints {
			sanitized[i] = PlotPoint{
				ID:        pp.ID,
				Required:  pp.Required,
				Delivered: pp.Delivered,
			}
		}
		clean.PlotPoints = sanitized
	}

	if clean.Resolution != nil {
		resCopy := *clean.Resolution
		resCopy.Outcomes = nil
		clean.Resolution = &resCopy
	}

	return clean
}

// Must exceed the in-process image job timeout (8m) so polling does not mark a
// still-running generation as failed while Retry is blocked by the in-memory lock.
const staleImageGeneration = 10 * time.Minute

func CharacterAvatarKey(name string) string {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return ""
	}
	return strings.Join(strings.Fields(key), " ")
}

func RecoverStaleImageStatuses(m *CustomMission) bool {
	if m == nil {
		return false
	}
	now := time.Now()
	changed := false

	if m.CoverImageStatus == "generating" && strings.TrimSpace(m.CoverImage) == "" {
		stale := false
		if t, err := time.Parse(time.RFC3339, m.CoverImageGenStartedAt); err == nil {
			stale = now.Sub(t) > staleImageGeneration
		} else if m.CoverImageGenStartedAt == "" {
			if created, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
				stale = now.Sub(created) > 45*time.Minute
			}
		}
		if stale {
			m.CoverImageStatus = "error"
			if strings.TrimSpace(m.CoverImageError) == "" {
				m.CoverImageError = "Image generation timed out. Check IMAGE_API_* / server logs, then refresh."
			}
			m.CoverImageGenStartedAt = ""
			changed = true
		}
	}

	if m.SceneImageStatus != nil {
		if m.SceneImageGenStartedAt == nil {
			m.SceneImageGenStartedAt = map[string]string{}
		}
		if m.SceneImageErrors == nil {
			m.SceneImageErrors = map[string]string{}
		}
		for k, st := range m.SceneImageStatus {
			if st != "generating" {
				continue
			}
			if m.SceneImages != nil && strings.TrimSpace(m.SceneImages[k]) != "" {
				continue
			}
			started := m.SceneImageGenStartedAt[k]
			stale := false
			if t, err := time.Parse(time.RFC3339, started); err == nil {
				stale = now.Sub(t) > staleImageGeneration
			} else if started == "" {
				if created, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
					stale = now.Sub(created) > 45*time.Minute
				}
			}
			if stale {
				m.SceneImageStatus[k] = "error"
				if strings.TrimSpace(m.SceneImageErrors[k]) == "" {
					m.SceneImageErrors[k] = "Scene image generation timed out."
				}
				delete(m.SceneImageGenStartedAt, k)
				changed = true
			}
		}
	}

	if m.CharacterAvatarStatus != nil {
		if m.CharacterAvatarGenStartedAt == nil {
			m.CharacterAvatarGenStartedAt = map[string]string{}
		}
		if m.CharacterAvatarErrors == nil {
			m.CharacterAvatarErrors = map[string]string{}
		}
		for k, st := range m.CharacterAvatarStatus {
			if st != "generating" {
				continue
			}
			if m.CharacterAvatars != nil && strings.TrimSpace(m.CharacterAvatars[k]) != "" {
				continue
			}
			started := m.CharacterAvatarGenStartedAt[k]
			stale := false
			if t, err := time.Parse(time.RFC3339, started); err == nil {
				stale = now.Sub(t) > staleImageGeneration
			} else if started == "" {
				if created, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
					stale = now.Sub(created) > 45*time.Minute
				}
			}
			if stale {
				m.CharacterAvatarStatus[k] = "error"
				if strings.TrimSpace(m.CharacterAvatarErrors[k]) == "" {
					m.CharacterAvatarErrors[k] = "Character avatar generation timed out."
				}
				delete(m.CharacterAvatarGenStartedAt, k)
				changed = true
			}
		}
	}

	return changed
}
