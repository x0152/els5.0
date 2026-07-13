package api

import (
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/quest"
)

type mediaURLs struct {
	urls   media.PublicURL
	bucket string
}

func (m mediaURLs) url(filename string) string {
	if filename == "" {
		return ""
	}
	path, err := media.NewPath(m.bucket + "/" + filename)
	if err != nil {
		return ""
	}
	return m.urls.Build(path)
}

func toMissionView(m *quest.CustomMission, urls mediaURLs) *quest.CustomMission {
	if m == nil {
		return nil
	}
	clean := quest.MissionSanitizer{}.Execute(m)
	clean.CoverImage = urls.url(clean.CoverImage)
	clean.SceneImages = rewriteURLs(clean.SceneImages, urls)
	clean.CharacterAvatars = rewriteURLs(clean.CharacterAvatars, urls)
	return &clean
}

func rewriteURLs(in map[string]string, urls mediaURLs) map[string]string {
	if len(in) == 0 {
		return in
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = urls.url(v)
	}
	return out
}

func toMissionSummary(item quest.MissionCatalogItem, urls mediaURLs) MissionSummary {
	m := item.Mission
	if !item.Started {
		// In the catalog, someone else's mission is shown as incomplete, without the author's progress.
		m.IsComplete = false
		m.CurrentStage = 0
	}
	return MissionSummary{
		ID:               m.ID,
		Started:          item.Started,
		Title:            m.Title,
		Description:      m.Description,
		Genre:            m.Genre,
		Language:         m.Language,
		CreatedAt:        m.CreatedAt,
		IsComplete:       m.IsComplete,
		CurrentStage:     m.CurrentStage,
		TotalStages:      m.TotalStages,
		GenerationStatus: m.GenerationStatus,
		GenerationStep:   m.GenerationStep,
		GenerationError:  m.GenerationError,
		CoverImage:       urls.url(m.CoverImage),
		CoverImageStatus: m.CoverImageStatus,
	}
}
