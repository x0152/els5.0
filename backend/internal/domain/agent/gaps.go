package agent

import (
	"regexp"
	"strings"
)

var gapRe = regexp.MustCompile(`\{\{[^}]*\}\}`)

// FillGap writes the user's answer into the ordinal-th {{gap}} of the content
// as {{spec||answer}}, replacing a previous fill if present.
func FillGap(content string, ordinal int, answer string) (string, bool) {
	answer = strings.NewReplacer("{", "", "}", "", "\n", " ").Replace(strings.TrimSpace(answer))
	n := 0
	found := false
	out := gapRe.ReplaceAllStringFunc(content, func(m string) string {
		if n != ordinal {
			n++
			return m
		}
		n++
		found = true
		spec := m[2 : len(m)-2]
		if i := strings.Index(spec, "||"); i != -1 {
			spec = spec[:i]
		}
		return "{{" + spec + "||" + answer + "}}"
	})
	return out, found
}
