package core

type DictEntry struct {
	Value string
	Label string
	Icon  string
}

var dictionaries = map[string][]DictEntry{
	"skill": {
		{"reading", "Reading", "📖"},
		{"listening", "Listening", "🎧"},
		{"writing", "Writing", "✍️"},
		{"speaking", "Speaking", "🗣️"},
	},
	"action": {
		{"grammar_error", "Grammar error", "⚠️"},
		{"word_error", "Word error", "⚠️"},
		{"read", "Read", "📖"},
		{"heard", "Heard", "🎧"},
		{"used_in_writing", "Used in writing", "✍️"},
		{"used_in_speech", "Used in speech", "🗣️"},
		{"construction_used", "Construction used", "🏗️"},
		{"reviewed_ok", "Reviewed (ok)", "✅"},
		{"reviewed_fail", "Reviewed (fail)", "❌"},
		{"reviewed", "Reviewed", "🔁"},
	},
	"pos": {
		{"noun", "Noun", ""},
		{"verb", "Verb", ""},
		{"adjective", "Adjective", ""},
		{"adverb", "Adverb", ""},
		{"pronoun", "Pronoun", ""},
		{"preposition", "Preposition", ""},
		{"conjunction", "Conjunction", ""},
		{"determiner", "Determiner", ""},
		{"article", "Article", ""},
		{"auxiliary", "Auxiliary", ""},
		{"particle", "Particle", ""},
		{"numeral", "Numeral", ""},
		{"interjection", "Interjection", ""},
	},
	"outcome": {
		{"ok", "OK", "✅"},
		{"fail", "Fail", "❌"},
	},
	"status": {
		{"pending", "Pending", "⏳"},
		{"processed", "Processed", "✅"},
		{"failed", "Failed", "❌"},
	},
}

func Dictionaries() map[string][]DictEntry {
	return dictionaries
}
