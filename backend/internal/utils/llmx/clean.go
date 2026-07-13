package llmx

import (
	"encoding/json"
	"strings"
)

func CleanLLMResponse(s string) string {
	s = strings.TrimSpace(s)

	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(s, "</think>")
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + s[end+len("</think>"):]
	}
	s = strings.ReplaceAll(s, "<think>", "")
	s = strings.ReplaceAll(s, "</think>", "")
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	if extracted := extractJSON(s); extracted != "" {
		return extracted
	}

	return s
}

func extractJSON(s string) string {
	last := strings.LastIndex(s, "}")
	if last == -1 {
		return ""
	}

	depth := 0
	start := -1
	for i := last; i >= 0; i-- {
		switch s[i] {
		case '}':
			depth++
		case '{':
			depth--
			if depth == 0 {
				start = i
			}
		}
		if start != -1 {
			break
		}
	}

	if start == -1 {
		return ""
	}

	candidate := strings.TrimSpace(s[start : last+1])
	if json.Valid([]byte(candidate)) {
		return candidate
	}
	return ""
}
