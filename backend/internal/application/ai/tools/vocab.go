package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/els/backend/internal/domain/agent"
	"github.com/els/backend/internal/domain/vocab"
)

func vocabTools(repo vocab.Repository) []agent.Tool {
	return []agent.Tool{addVocabWord(repo), listVocabWords(repo), deleteVocabWord(repo)}
}

func addVocabWord(repo vocab.Repository) agent.Tool {
	return agent.Tool{
		Name:        "add_vocab_word",
		Description: "Adds a word, phrase, phrasal verb, or idiom to the user's vocabulary collection for memorization. Fill in the translation, definition, and example yourself.",
		Icon:        "book-marked",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text":          map[string]any{"type": "string", "description": "The word/phrase in English (correct spelling)."},
				"kind":          map[string]any{"type": "string", "enum": []string{"word", "phrase", "phrasal_verb", "idiom"}, "description": "Unit type."},
				"transcription": map[string]any{"type": "string", "description": "IPA transcription for a single word, otherwise empty."},
				"translation":   map[string]any{"type": "string", "description": "Short translation into the user's native language (see run context)."},
				"definition":    map[string]any{"type": "string", "description": "Definition in English."},
				"example":       map[string]any{"type": "string", "description": "Usage example in English."},
			},
			"required": []string{"text"},
		},
		Label: func(args string) string {
			var a struct {
				Text string `json:"text"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if a.Text != "" {
				return fmt.Sprintf("Adding \"%s\"", a.Text)
			}
			return "Adding a word"
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			var a struct {
				Text          string `json:"text"`
				Kind          string `json:"kind"`
				Transcription string `json:"transcription"`
				Translation   string `json:"translation"`
				Definition    string `json:"definition"`
				Example       string `json:"example"`
			}
			if err := json.Unmarshal([]byte(args), &a); err != nil {
				return "", fmt.Errorf("invalid arguments: %w", err)
			}
			if strings.TrimSpace(a.Kind) == "" {
				a.Kind = string(vocab.KindWord)
			}
			unit, err := vocab.NewUnit(uuid.NewString(), actor.AccountID().String(), vocab.CheckResult{
				Correct:       true,
				Kind:          a.Kind,
				Text:          a.Text,
				Transcription: a.Transcription,
				Translation:   a.Translation,
				Definition:    a.Definition,
				Example:       a.Example,
			})
			if err != nil {
				return "", err
			}
			if _, err := repo.Create(ctx, unit); err != nil {
				return fmt.Sprintf("Failed to add «%s»: %v", a.Text, err), nil
			}
			return fmt.Sprintf("Added to collection: «%s».", unit.Text), nil
		},
	}
}

func listVocabWords(repo vocab.Repository) agent.Tool {
	return agent.Tool{
		Name:        "list_vocab_words",
		Description: "Returns words from the user's collection (with translation and memorization status).",
		Icon:        "book-marked",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{"type": "string", "enum": []string{"new", "learning", "learned"}, "description": "Filter by status (optional)."},
				"limit":  map[string]any{"type": "integer", "description": "How many words to return (default 50)."},
			},
		},
		Label: func(string) string { return "Reading the word collection" },
		Execute: func(ctx context.Context, args string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			var a struct {
				Status string `json:"status"`
				Limit  int    `json:"limit"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if a.Limit <= 0 || a.Limit > 100 {
				a.Limit = 50
			}
			units, total, err := repo.List(ctx, actor.AccountID().String(), vocab.ListFilter{Status: vocab.Status(a.Status), Limit: a.Limit})
			if err != nil {
				return "", err
			}
			if len(units) == 0 {
				return "Collection is empty.", nil
			}
			items := make([]map[string]any, 0, len(units))
			for _, u := range units {
				items = append(items, map[string]any{
					"text":        u.Text,
					"kind":        string(u.Kind),
					"translation": u.Translation,
					"status":      string(u.Status),
				})
			}
			b, _ := json.MarshalIndent(map[string]any{"total": total, "items": items}, "", "  ")
			return string(b), nil
		},
	}
}

func deleteVocabWord(repo vocab.Repository) agent.Tool {
	return agent.Tool{
		Name:        "delete_vocab_word",
		Description: "Removes a word/phrase from the user's collection by text.",
		Icon:        "trash-2",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"text": map[string]any{"type": "string", "description": "Text of the word/phrase to delete."},
			},
			"required": []string{"text"},
		},
		Label: func(args string) string {
			var a struct {
				Text string `json:"text"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if a.Text != "" {
				return fmt.Sprintf("Removing \"%s\"", a.Text)
			}
			return "Removing a word"
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			var a struct {
				Text string `json:"text"`
			}
			if err := json.Unmarshal([]byte(args), &a); err != nil {
				return "", fmt.Errorf("invalid arguments: %w", err)
			}
			text := strings.TrimSpace(a.Text)
			if text == "" {
				return "Provide the word text to delete.", nil
			}
			accountID := actor.AccountID().String()
			units, _, err := repo.List(ctx, accountID, vocab.ListFilter{Search: text, Limit: 100})
			if err != nil {
				return "", err
			}
			for _, u := range units {
				if strings.EqualFold(u.Text, text) {
					if err := repo.Delete(ctx, accountID, u.ID); err != nil {
						return "", err
					}
					return fmt.Sprintf("Removed from collection: «%s».", u.Text), nil
				}
			}
			return fmt.Sprintf("«%s» is not in the collection.", text), nil
		},
	}
}
