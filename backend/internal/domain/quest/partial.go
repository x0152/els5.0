package quest

import (
	"strconv"
	"strings"
)

// ParsePartialWorld extracts from an incomplete JSON stream of the world reply
// narration and line text for draft display to the player. Values may
// cut off mid-word — Done flags show whether the value is complete.
func ParsePartialWorld(raw string) *PartialWorld {
	if i := strings.IndexByte(raw, '{'); i > 0 {
		raw = raw[i:]
	} else if i < 0 {
		return nil
	}

	partial := &PartialWorld{}
	if v, _, ok, closed := extractStringValue(raw, `"narration"`); ok {
		partial.Narration = v
		partial.NarrationDone = closed
	}

	if idx := strings.Index(raw, `"responses"`); idx >= 0 {
		rest := raw[idx+len(`"responses"`):]
		for {
			name, next, ok, nameClosed := extractStringValue(rest, `"name"`)
			if !ok || !nameClosed || strings.TrimSpace(name) == "" {
				break
			}
			text, next2, ok, textClosed := extractStringValue(next, `"text"`)
			if !ok {
				partial.Responses = append(partial.Responses, PartialLine{Name: name})
				break
			}
			partial.Responses = append(partial.Responses, PartialLine{Name: name, Text: text, Done: textClosed})
			if !textClosed {
				break
			}
			rest = next2
		}
	}

	if partial.Narration == "" && len(partial.Responses) == 0 {
		return nil
	}
	return partial
}

// extractStringValue finds key and returns its string value
// (possibly truncated), the remainder after the value, and whether
// the value is complete (closing quote was seen).
func extractStringValue(s, key string) (value, rest string, ok, closed bool) {
	idx := strings.Index(s, key)
	if idx < 0 {
		return "", "", false, false
	}
	after := s[idx+len(key):]

	quote := -1
	for i := 0; i < len(after); i++ {
		c := after[i]
		if c == ':' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		if c == '"' {
			quote = i
		}
		break
	}
	if quote < 0 {
		return "", "", false, false
	}
	after = after[quote+1:]

	end := len(after)
	for i := 0; i < len(after); i++ {
		if after[i] == '\\' {
			i++
			continue
		}
		if after[i] == '"' {
			end = i
			closed = true
			break
		}
	}

	value = unescapeJSONString(after[:end])
	if closed {
		return value, after[end+1:], true, true
	}
	return value, "", true, false
}

func unescapeJSONString(s string) string {
	// Trim a truncated escape at the end of the stream so Unquote does not fail.
	trailing := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\\'; i-- {
		trailing++
	}
	if trailing%2 == 1 {
		s = s[:len(s)-1]
	}
	if unquoted, err := strconv.Unquote(`"` + s + `"`); err == nil {
		return unquoted
	}
	replacer := strings.NewReplacer(`\"`, `"`, `\n`, "\n", `\t`, "\t", `\\`, `\`)
	return replacer.Replace(s)
}
