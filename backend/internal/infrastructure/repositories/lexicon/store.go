package lexicon

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/els/backend/internal/domain/lexicon"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) SaveSentence(ctx context.Context, mediaID string, a lexicon.Analysis, segments []lexicon.Segment) error {
	return s.save(ctx, mediaID, a, segments, func(u lexicon.Unit, ids map[int]uuid.UUID, _ []int) (uuid.UUID, bool) {
		id, ok := ids[u.SentenceIdx]
		return id, ok
	})
}

func (s *Store) SaveSubtitle(ctx context.Context, mediaID string, cues []lexicon.Cue, a lexicon.Analysis) error {
	segments := make([]lexicon.Segment, 0, len(cues))
	for _, c := range cues {
		segments = append(segments, lexicon.Segment{
			Kind:       lexicon.SegmentKindSubtitle,
			SegmentIdx: c.Index,
			StartPos:   c.StartMs,
			EndPos:     c.EndMs,
			Text:       c.Text,
			Metadata:   map[string]any{},
		})
	}
	return s.save(ctx, mediaID, a, segments, resolveSubtitleSegment)
}

func (s *Store) FindOccurrences(ctx context.Context, lemmas []string) ([]lexicon.LemmaOccurrences, error) {
	if len(lemmas) == 0 {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx,
		`SELECT lower(u.base_form) AS lemma,
		        u.media_id::text,
		        CASE WHEN b.id IS NOT NULL THEN 'book' WHEN f.id IS NOT NULL THEN 'film' ELSE '' END AS media_type,
		        COALESCE(b.title, f.title, '') AS title,
		        COALESCE(f.kind, '') AS kind,
		        COALESCE(f.series_title, '') AS series_title,
		        COALESCE(f.season, 0) AS season,
		        COALESCE(f.episode, 0) AS episode,
		        COALESCE(b.author, '') AS author,
		        COALESCE(bf.is_stop, false) AS is_stop,
		        ms.start_pos AS ref,
		        ms.text AS example
		 FROM units u
		 JOIN media_segments ms ON ms.id = u.segment_id
		 LEFT JOIN reader_books b ON b.id = u.media_id
		 LEFT JOIN films f ON f.id = u.media_id
		 LEFT JOIN base_forms bf ON bf.text = u.base_form
		 WHERE lower(u.base_form) = ANY($1)
		 GROUP BY lemma, u.media_id, b.id, f.id, b.title, f.title, f.kind, f.series_title, f.season, f.episode, b.author, bf.is_stop, ms.id, ms.start_pos, ms.text
		 ORDER BY u.media_id, ms.start_pos`, lemmas)
	if err != nil {
		return nil, fmt.Errorf("find occurrences: %w", err)
	}
	defer rows.Close()

	byLemma := make(map[string]*lexicon.LemmaOccurrences)
	lemmaOrder := make([]string, 0)
	mediaIdx := make(map[string]int)
	for rows.Next() {
		var lemma, mediaID, mediaType, title, kind, seriesTitle, author, example string
		var season, episode, ref int
		var isStop bool
		if err := rows.Scan(&lemma, &mediaID, &mediaType, &title, &kind, &seriesTitle, &season, &episode, &author, &isStop, &ref, &example); err != nil {
			return nil, fmt.Errorf("scan occurrence: %w", err)
		}
		lo, ok := byLemma[lemma]
		if !ok {
			lo = &lexicon.LemmaOccurrences{Lemma: lemma}
			byLemma[lemma] = lo
			lemmaOrder = append(lemmaOrder, lemma)
		}
		lo.IsStop = lo.IsStop || isStop
		lo.Total++
		key := lemma + "\x00" + mediaID
		i, ok := mediaIdx[key]
		if !ok {
			i = len(lo.Media)
			mediaIdx[key] = i
			lo.Media = append(lo.Media, lexicon.MediaOccurrence{
				MediaID:     mediaID,
				MediaType:   mediaType,
				Title:       title,
				Kind:        kind,
				SeriesTitle: seriesTitle,
				Season:      season,
				Episode:     episode,
				Author:      author,
			})
			lo.MediaCount++
		}
		lo.Media[i].Spots = append(lo.Media[i].Spots, lexicon.Spot{Ref: ref, Example: example})
		lo.Media[i].Count = len(lo.Media[i].Spots)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate occurrences: %w", err)
	}

	out := make([]lexicon.LemmaOccurrences, 0, len(lemmaOrder))
	for _, lemma := range lemmaOrder {
		out = append(out, *byLemma[lemma])
	}
	return out, nil
}

func (s *Store) DeleteByMedia(ctx context.Context, mediaID string) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM units WHERE media_id = $1`, mediaID); err != nil {
		return fmt.Errorf("delete units: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `DELETE FROM media_segments WHERE media_id = $1`, mediaID); err != nil {
		return fmt.Errorf("delete media segments: %w", err)
	}
	return nil
}

type segmentResolver func(u lexicon.Unit, ids map[int]uuid.UUID, ordered []int) (uuid.UUID, bool)

func (s *Store) save(ctx context.Context, mediaID string, a lexicon.Analysis, segments []lexicon.Segment, resolve segmentResolver) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM units WHERE media_id = $1`, mediaID); err != nil {
		return fmt.Errorf("delete units: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM media_segments WHERE media_id = $1`, mediaID); err != nil {
		return fmt.Errorf("delete media segments: %w", err)
	}
	if err := upsertBaseForms(ctx, tx, a.BaseForms); err != nil {
		return err
	}
	segmentIDs, err := insertSegments(ctx, tx, mediaID, segments)
	if err != nil {
		return err
	}
	ordered := orderedKeys(segmentIDs)
	if err := insertUnits(ctx, tx, mediaID, a.Units, segmentIDs, ordered, resolve); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func upsertBaseForms(ctx context.Context, tx pgx.Tx, baseForms map[string]lexicon.BaseForm) error {
	for _, bf := range baseForms {
		if _, err := tx.Exec(ctx,
			`INSERT INTO base_forms (text, pos, is_stop, language) VALUES ($1,$2,$3,$4)
			 ON CONFLICT (text) DO NOTHING`,
			bf.Text, nullStr(bf.POS), bf.IsStop, langOrDefault(bf.Language)); err != nil {
			return fmt.Errorf("upsert base form %q: %w", bf.Text, err)
		}
	}
	return nil
}

func insertSegments(ctx context.Context, tx pgx.Tx, mediaID string, segments []lexicon.Segment) (map[int]uuid.UUID, error) {
	sort.Slice(segments, func(i, j int) bool { return segments[i].SegmentIdx < segments[j].SegmentIdx })
	ids := make(map[int]uuid.UUID, len(segments))
	for _, seg := range segments {
		meta := seg.Metadata
		if meta == nil {
			meta = map[string]any{}
		}
		metaJSON, err := json.Marshal(meta)
		if err != nil {
			return nil, fmt.Errorf("marshal segment metadata: %w", err)
		}
		id := uuid.New()
		if _, err := tx.Exec(ctx,
			`INSERT INTO media_segments (id, media_id, kind, segment_idx, start_pos, end_pos, text, metadata)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			id, mediaID, string(seg.Kind), seg.SegmentIdx, seg.StartPos, seg.EndPos, seg.Text, metaJSON); err != nil {
			return nil, fmt.Errorf("insert segment %d: %w", seg.SegmentIdx, err)
		}
		ids[seg.SegmentIdx] = id
	}
	return ids, nil
}

func insertUnits(ctx context.Context, tx pgx.Tx, mediaID string, units []lexicon.Unit, segmentIDs map[int]uuid.UUID, ordered []int, resolve segmentResolver) error {
	for _, u := range units {
		metaJSON, err := json.Marshal(u.Metadata)
		if err != nil {
			return fmt.Errorf("marshal unit metadata: %w", err)
		}
		unitID := uuid.New()
		var segmentID *uuid.UUID
		if id, ok := resolve(u, segmentIDs, ordered); ok {
			segmentID = &id
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO units (id, media_id, segment_id, unit_type, base_form, pos, sentence_idx, unit_metadata, language)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			unitID, mediaID, segmentID, string(u.UnitType), u.BaseForm, nullStr(u.POS), u.SentenceIdx, metaJSON, langOrDefault(u.Language)); err != nil {
			return fmt.Errorf("insert unit %q: %w", u.BaseForm, err)
		}
		for _, span := range u.Spans {
			if _, err := tx.Exec(ctx,
				`INSERT INTO unit_spans (id, unit_id, position, span_type, start, "end", text)
				 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				uuid.New(), unitID, span.Position, span.SpanType, span.Start, span.End, span.Text); err != nil {
				return fmt.Errorf("insert span for unit %q: %w", u.BaseForm, err)
			}
		}
	}
	return nil
}

func resolveSubtitleSegment(u lexicon.Unit, ids map[int]uuid.UUID, ordered []int) (uuid.UUID, bool) {
	if cue, ok := metadataInt(u.Metadata, "cue_index"); ok {
		if id, found := ids[cue]; found {
			return id, true
		}
	}
	if u.SentenceIdx >= 0 {
		if id, found := ids[u.SentenceIdx]; found {
			return id, true
		}
		if id, found := ids[u.SentenceIdx+1]; found {
			return id, true
		}
		if u.SentenceIdx < len(ordered) {
			return ids[ordered[u.SentenceIdx]], true
		}
	}
	if len(ordered) > 0 {
		return ids[ordered[0]], true
	}
	return uuid.UUID{}, false
}

func orderedKeys(ids map[int]uuid.UUID) []int {
	keys := make([]int, 0, len(ids))
	for k := range ids {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func metadataInt(metadata map[string]any, key string) (int, bool) {
	if metadata == nil {
		return 0, false
	}
	switch v := metadata[key].(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	}
	return 0, false
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func langOrDefault(s string) string {
	if s == "" {
		return "en"
	}
	return s
}
