package workout

import "strings"

var levelRank = map[string]int{"A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6}

// NormalizeLevel extracts a CEFR level from a free-form string ("B1", "b2 intermediate"); unknown input falls back to B1.
func NormalizeLevel(s string) string {
	upper := strings.ToUpper(s)
	for _, l := range []string{"C2", "C1", "B2", "B1", "A2", "A1"} {
		if strings.Contains(upper, l) {
			return l
		}
	}
	return "B1"
}

func LevelAtMost(level, max string) bool {
	r, ok := levelRank[NormalizeLevel(level)]
	if !ok {
		return true
	}
	return r <= levelRank[NormalizeLevel(max)]
}
