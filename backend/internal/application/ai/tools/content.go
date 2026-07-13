package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"

	"github.com/els/backend/internal/domain/agent"
	"github.com/els/backend/internal/domain/book"
	"github.com/els/backend/internal/domain/films"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/reader"
)

type FrameReader interface {
	ReadFrame(ctx context.Context, filmID string, atMs int, question string) (string, error)
}

type ContentDeps struct {
	Books    book.Repository
	Films    films.Repository
	Reader   reader.Repository
	Missions quest.MissionRepository
	Storage  media.Storage
	Vision   FrameReader
}

type ContentPlugin struct{ tools []agent.Tool }

func NewContentPlugin(d ContentDeps) *ContentPlugin {
	tools := []agent.Tool{
		listBookUnits(d.Books),
		readBookUnit(d.Books),
		listFilms(d.Films),
		readFilmSubtitles(d.Films),
		listReaderBooks(d.Reader),
		readBookText(d.Reader, d.Storage),
		listQuests(d.Missions),
		readQuest(d.Missions),
	}
	if d.Vision != nil {
		tools = append(tools, readFilmFrame(d.Vision))
	}
	return &ContentPlugin{tools: tools}
}

func (p *ContentPlugin) Tools(_ agent.RunContext) []agent.Tool { return p.tools }

func bookSlug(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "grammarbook", "grammar", "merfy":
		return "grammarbook"
	case "essentialbook", "essential", "words":
		return "essentialbook"
	case "wordbook", "vocabulary":
		return "wordbook"
	case "phrasebook", "collocations":
		return "phrasebook"
	default:
		return ""
	}
}

func listBookUnits(repo book.Repository) agent.Tool {
	return agent.Tool{
		Name:        "list_book_units",
		Description: "List of textbook units (chapters): grammar (English Grammar in Use), essentialbook (504 Essential Words), wordbook (Vocabulary in Use), or phrasebook (Collocations in Use). Returns numbers and titles.",
		Icon:        "list",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"book": map[string]any{"type": "string", "enum": []string{"grammar", "essentialbook", "wordbook", "phrasebook"}, "description": "Which textbook."},
		}, "required": []string{"book"}},
		Label: func(string) string { return "Listing units" },
		Execute: func(ctx context.Context, args string) (string, error) {
			var a struct {
				Book string `json:"book"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			slug := bookSlug(a.Book)
			if slug == "" {
				return "Provide book: grammar, essentialbook, wordbook, or phrasebook.", nil
			}
			chapters, err := repo.List(ctx, slug)
			if err != nil {
				return "", err
			}
			items := make([]map[string]any, 0, len(chapters))
			for _, c := range chapters {
				items = append(items, map[string]any{"number": c.Number, "title": c.Title, "status": c.Status})
			}
			b, _ := json.MarshalIndent(items, "", "  ")
			return string(b), nil
		},
	}
}

func readBookUnit(repo book.Repository) agent.Tool {
	return agent.Tool{
		Name:        "read_book_unit",
		Description: "Full text of one textbook unit (theory and exercises). book: grammar, essentialbook, wordbook, or phrasebook; number — the unit number.",
		Icon:        "book-open",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"book":   map[string]any{"type": "string", "enum": []string{"grammar", "essentialbook", "wordbook", "phrasebook"}},
			"number": map[string]any{"type": "integer", "description": "Unit number."},
		}, "required": []string{"book", "number"}},
		Label: func(args string) string {
			var a struct {
				Book   string `json:"book"`
				Number int    `json:"number"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			return fmt.Sprintf("Unit %d (%s)", a.Number, a.Book)
		},
		Execute: func(ctx context.Context, args string) (string, error) {
			var a struct {
				Book   string `json:"book"`
				Number int    `json:"number"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			slug := bookSlug(a.Book)
			if slug == "" || a.Number <= 0 {
				return "Provide book and number.", nil
			}
			c, err := repo.GetByNumber(ctx, slug, a.Number)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("# %s — Unit %d: %s\n\n## Theory\n%s\n\n## Exercises\n%s", a.Book, c.Number, c.Title, c.Theory, c.Exercises), nil
		},
	}
}

func listFilms(repo films.Repository) agent.Tool {
	return agent.Tool{
		Name:        "list_films",
		Description: "List of films and series in the catalog (id, title, type, series).",
		Icon:        "film",
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Label:       func(string) string { return "Listing films" },
		Execute: func(ctx context.Context, _ string) (string, error) {
			list, err := repo.List(ctx)
			if err != nil {
				return "", err
			}
			items := make([]map[string]any, 0, len(list))
			for _, f := range list {
				items = append(items, map[string]any{"id": f.ID, "title": f.Title, "kind": f.Kind, "series_title": f.SeriesTitle})
			}
			b, _ := json.MarshalIndent(items, "", "  ")
			return string(b), nil
		},
	}
}

func readFilmSubtitles(repo films.Repository) agent.Tool {
	return agent.Tool{
		Name:        "read_film_subtitles",
		Description: "English film subtitles around the at_ms timecode (milliseconds). If at_ms is omitted — from the start. Use this to understand what is said at a specific moment.",
		Icon:        "captions",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"film_id": map[string]any{"type": "string"},
			"at_ms":   map[string]any{"type": "integer", "description": "Timecode in milliseconds (optional)."},
		}, "required": []string{"film_id"}},
		Label: func(string) string { return "Reading subtitles" },
		Execute: func(ctx context.Context, args string) (string, error) {
			var a struct {
				FilmID string `json:"film_id"`
				AtMs   int    `json:"at_ms"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if strings.TrimSpace(a.FilmID) == "" {
				return "Provide film_id.", nil
			}
			f, err := repo.Get(ctx, a.FilmID)
			if err != nil {
				return "", err
			}
			track, ok := films.PickEnglishSubtitle(f.Subtitles)
			if !ok || len(track.Cues) == 0 {
				return "The film has no English subtitles.", nil
			}
			cues := track.Cues
			center := 0
			for i, c := range cues {
				if c.StartMs <= a.AtMs {
					center = i
				} else {
					break
				}
			}
			from := center - 8
			if from < 0 {
				from = 0
			}
			to := center + 30
			if to > len(cues) {
				to = len(cues)
			}
			var sb strings.Builder
			for _, c := range cues[from:to] {
				fmt.Fprintf(&sb, "[%s] %s\n", msToTime(c.StartMs), c.Text)
			}
			return sb.String(), nil
		},
	}
}

func readFilmFrame(vision FrameReader) agent.Tool {
	return agent.Tool{
		Name:        "read_film_frame",
		Description: "Extracts a film frame at at_ms (milliseconds) and describes what is shown. Use when you need to understand the visual scene, not only the subtitles.",
		Icon:        "image",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"film_id":  map[string]any{"type": "string"},
			"at_ms":    map[string]any{"type": "integer", "description": "Frame timecode in milliseconds."},
			"question": map[string]any{"type": "string", "description": "What specifically to ask about the frame (optional)."},
		}, "required": []string{"film_id", "at_ms"}},
		Label: func(string) string { return "Viewing a frame" },
		Execute: func(ctx context.Context, args string) (string, error) {
			var a struct {
				FilmID   string `json:"film_id"`
				AtMs     int    `json:"at_ms"`
				Question string `json:"question"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if strings.TrimSpace(a.FilmID) == "" {
				return "Provide film_id.", nil
			}
			return vision.ReadFrame(ctx, a.FilmID, a.AtMs, a.Question)
		},
	}
}

func listReaderBooks(repo reader.Repository) agent.Tool {
	return agent.Tool{
		Name:        "list_books",
		Description: "List of the user's books in the reader (id, title, author, reading position).",
		Icon:        "library",
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Label:       func(string) string { return "Listing books" },
		Execute: func(ctx context.Context, _ string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			list, err := repo.List(ctx, actor.AccountID().String())
			if err != nil {
				return "", err
			}
			items := make([]map[string]any, 0, len(list))
			for _, b := range list {
				items = append(items, map[string]any{"id": b.ID, "title": b.Title, "author": b.Author, "position": b.Position, "text_length": b.TextLength})
			}
			out, _ := json.MarshalIndent(items, "", "  ")
			return string(out), nil
		},
	}
}

func readBookText(repo reader.Repository, storage media.Storage) agent.Tool {
	return agent.Tool{
		Name:        "read_book_text",
		Description: "A fragment of book text from the reader around a position (character offset). If position is omitted — the user's current reading position is used.",
		Icon:        "book-open",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"book_id":  map[string]any{"type": "string"},
			"position": map[string]any{"type": "integer", "description": "Character offset (optional)."},
		}, "required": []string{"book_id"}},
		Label: func(string) string { return "Reading the book" },
		Execute: func(ctx context.Context, args string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			if storage == nil {
				return "Book reading is unavailable.", nil
			}
			var a struct {
				BookID   string `json:"book_id"`
				Position int    `json:"position"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if strings.TrimSpace(a.BookID) == "" {
				return "Provide book_id.", nil
			}
			b, err := repo.Get(ctx, actor.AccountID().String(), a.BookID)
			if err != nil {
				return "", err
			}
			path, err := media.NewPath(b.ContentPath)
			if err != nil {
				return "Book text is unavailable.", nil
			}
			rc, _, err := storage.Get(ctx, path)
			if err != nil {
				return "", err
			}
			defer rc.Close()
			data, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}
			runes := []rune(stripText(string(data)))
			pos := a.Position
			if pos <= 0 {
				pos = b.Position
			}
			start := pos - 1000
			if start < 0 {
				start = 0
			}
			end := pos + 2000
			if end > len(runes) {
				end = len(runes)
			}
			if start >= end {
				return "Empty fragment.", nil
			}
			return string(runes[start:end]), nil
		},
	}
}

func listQuests(repo quest.MissionRepository) agent.Tool {
	return agent.Tool{
		Name:        "list_quests",
		Description: "Short list of the user's quests (missions): id, title, genre, whether completed.",
		Icon:        "swords",
		Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		Label:       func(string) string { return "Listing quests" },
		Execute: func(ctx context.Context, _ string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			list, err := repo.List(ctx, actor.AccountID().String())
			if err != nil {
				return "", err
			}
			items := make([]map[string]any, 0, len(list))
			for _, item := range list {
				m := item.Mission
				items = append(items, map[string]any{"id": m.ID, "title": m.Title, "genre": m.Genre, "is_complete": m.IsComplete, "started": item.Started})
			}
			out, _ := json.MarshalIndent(items, "", "  ")
			return string(out), nil
		},
	}
}

func readQuest(repo quest.MissionRepository) agent.Tool {
	return agent.Tool{
		Name:        "read_quest",
		Description: "Data for one quest. part=info — description, characters, goals, and plot points; part=dialogue — dialogue history. Pick the part you need to avoid reading extras.",
		Icon:        "swords",
		Parameters: map[string]any{"type": "object", "properties": map[string]any{
			"id":   map[string]any{"type": "string"},
			"part": map[string]any{"type": "string", "enum": []string{"info", "dialogue"}, "description": "What to return (default info)."},
		}, "required": []string{"id"}},
		Label: func(string) string { return "Reading the quest" },
		Execute: func(ctx context.Context, args string) (string, error) {
			actor := agent.ActorFrom(ctx)
			if actor == nil {
				return "", fmt.Errorf("no authenticated user in context")
			}
			var a struct {
				ID   string `json:"id"`
				Part string `json:"part"`
			}
			_ = json.Unmarshal([]byte(args), &a)
			if strings.TrimSpace(a.ID) == "" {
				return "Provide the quest id.", nil
			}
			m, err := repo.GetByID(ctx, actor.AccountID().String(), a.ID)
			if err != nil {
				m, _, err = repo.GetOrigin(ctx, a.ID)
				if err != nil {
					return "", err
				}
			}
			if a.Part == "dialogue" {
				var sb strings.Builder
				for _, t := range m.History {
					fmt.Fprintf(&sb, "%s: %s\n", t.Speaker, t.Text)
				}
				if sb.Len() == 0 {
					return "This quest has no dialogue yet.", nil
				}
				return sb.String(), nil
			}
			characters := make([]map[string]any, 0, len(m.Characters))
			for _, c := range m.Characters {
				characters = append(characters, map[string]any{"name": c.Name, "role": c.Role, "personality": c.Personality})
			}
			plot := make([]map[string]any, 0, len(m.PlotPoints))
			for _, pp := range m.PlotPoints {
				plot = append(plot, map[string]any{"fact": pp.Fact, "required": pp.Required, "delivered": pp.Delivered})
			}
			info := map[string]any{
				"title":          m.Title,
				"description":    m.Description,
				"genre":          m.Genre,
				"practice_goals": m.PracticeGoals,
				"characters":     characters,
				"plot_points":    plot,
				"current_stage":  m.CurrentStage,
				"total_stages":   m.TotalStages,
				"is_complete":    m.IsComplete,
				"outcome":        m.Outcome,
			}
			out, _ := json.MarshalIndent(info, "", "  ")
			return string(out), nil
		},
	}
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripText(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	return strings.Join(strings.Fields(s), " ")
}

func msToTime(ms int) string {
	total := ms / 1000
	h := total / 3600
	m := (total % 3600) / 60
	sec := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, sec)
	}
	return fmt.Sprintf("%d:%02d", m, sec)
}
