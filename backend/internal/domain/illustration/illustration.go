package illustration

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	StatusPending    = "pending"
	StatusGenerating = "generating"
	StatusReady      = "ready"
	StatusError      = "error"
)

type Status struct {
	ID     string
	Status string
	URL    string
	Error  string
}

func Key(prompt, aspect string) string {
	norm := strings.ToLower(strings.Join(strings.Fields(prompt), " "))
	sum := sha256.Sum256([]byte(norm + "|" + aspect))
	return hex.EncodeToString(sum[:16])
}

func Filename(id string) string { return id + ".png" }

func StyledPrompt(prompt string) string {
	const style = "clean modern flat vector illustration for an English learning textbook, simple friendly shapes, soft colors, white background, no text, no letters, no watermark"
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return style
	}
	return prompt + ". " + style
}
