package vocab

import (
	"regexp"
	"strings"
	"time"
)

type CardMode string

const (
	CardModeChoice CardMode = "choice"
	CardModeInput  CardMode = "input"
)

type CardDirection string

const (
	CardDirectionWord        CardDirection = "word"
	CardDirectionTranslation CardDirection = "translation"
)

const LearnedStreak = 3

type Card struct {
	Unit      Unit
	Mode      CardMode
	Direction CardDirection
	Options   []string
	ImageURL  string
}

func ModeFor(u Unit) CardMode {
	if u.Status == StatusLearned || u.CorrectStreak >= LearnedStreak-1 {
		return CardModeInput
	}
	return CardModeChoice
}

// ApplyAnswer advances streak/status: the streak grows at most once per day,
// so learning->learned requires correct answers on LearnedStreak different days.
func ApplyAnswer(u Unit, correct bool, now time.Time) Unit {
	mode := ModeFor(u)
	answeredToday := u.LastAnsweredAt != nil && sameDay(*u.LastAnsweredAt, now)
	t := now
	u.LastAnsweredAt = &t
	if !correct {
		if u.Status == StatusLearned {
			u.Status = StatusLearning
		}
		if mode == CardModeInput {
			u.CorrectStreak = 1
		} else {
			u.CorrectStreak = 0
		}
		return u
	}
	if u.Status == StatusLearned || answeredToday {
		return u
	}
	u.CorrectStreak++
	if u.Status == StatusNew {
		u.Status = StatusLearning
	}
	if u.CorrectStreak >= LearnedStreak {
		u.Status = StatusLearned
	}
	return u
}

func IsCorrectAnswer(u Unit, answer string) bool {
	n := normalizeAnswer(answer)
	if u.Translation != "" && n == normalizeAnswer(u.Translation) {
		return true
	}
	return n == normalizeAnswer(u.Text)
}

func normalizeAnswer(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

func MaskedDefinition(u Unit) string {
	masked := u.Definition
	for _, w := range strings.Fields(u.Text) {
		if len(w) < 3 {
			continue
		}
		re, err := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(w) + `\w*`)
		if err != nil {
			continue
		}
		masked = re.ReplaceAllString(masked, "___")
	}
	return masked
}

func ImagePrompt(text string) string {
	return `A clear educational illustration of "` + strings.TrimSpace(text) + `": a simple memorable scene that shows the meaning.`
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.In(b.Location()).Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
