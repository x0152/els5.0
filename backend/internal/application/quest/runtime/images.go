package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared/ports"
)

type imageLLM interface {
	DescribeCoverImage(ctx context.Context, mission *quest.CustomMission) (string, error)
	DescribeSceneImage(ctx context.Context, scene *quest.DynamicScene, mission *quest.CustomMission) (string, error)
	DescribeCharacterAvatar(ctx context.Context, character *quest.Character, scene *quest.DynamicScene, mission *quest.CustomMission) (string, error)
}

type Images struct {
	provider ports.ImageGenerator
	storage  media.Storage
	bucket   string
	llm      imageLLM
	missions quest.MissionRepository
	logger   *slog.Logger
	running  sync.Map
}

func NewImages(provider ports.ImageGenerator, storage media.Storage, bucket string, llm imageLLM, missions quest.MissionRepository, logger *slog.Logger) *Images {
	if logger == nil {
		logger = slog.Default()
	}
	return &Images{provider: provider, storage: storage, bucket: bucket, llm: llm, missions: missions, logger: logger}
}

func (s *Images) IsAvailable() bool {
	return s != nil && s.provider != nil && s.provider.IsAvailable() && s.storage != nil
}

func (s *Images) put(ctx context.Context, filename, contentType string, data []byte) error {
	path, err := media.NewPath(s.bucket + "/" + filename)
	if err != nil {
		return err
	}
	return s.storage.Put(ctx, path, bytes.NewReader(data), media.PutOptions{ContentType: contentType, Size: int64(len(data))})
}

func (s *Images) playerReference(ctx context.Context, mission *quest.CustomMission) []ports.ImageReference {
	raw := strings.TrimSpace(mission.PlayerAvatarImage)
	if raw == "" {
		return nil
	}
	path, err := media.NewPath(raw)
	if err != nil {
		return nil
	}
	rc, meta, err := s.storage.Get(ctx, path)
	if err != nil {
		return nil
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil || len(data) == 0 {
		return nil
	}
	contentType := meta.ContentType
	if contentType == "" {
		contentType = "image/png"
	}
	return []ports.ImageReference{{Data: data, ContentType: contentType}}
}

func (s *Images) GenerateInitialImagesAsync(userID, missionID string) {
	s.generateCover(userID, missionID, true)
}

func (s *Images) GenerateCoverAsync(userID, missionID string) {
	s.generateCover(userID, missionID, false)
}

// chain=true continues the initial pipeline (scene 0, avatars) after the
// cover; manual retries pass false so only the requested image regenerates.
func (s *Images) generateCover(userID, missionID string, chain bool) {
	if !s.IsAvailable() {
		return
	}
	key := userID + ":" + missionID + ":cover"
	if !s.start(key) {
		return
	}

	// Mark generating synchronously so API responses issued right after
	// triggering already report the new status.
	if err := s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
		mission.CoverImageStatus = "generating"
		mission.CoverImageError = ""
		mission.CoverImageGenStartedAt = time.Now().UTC().Format(time.RFC3339)
	}); err != nil {
		s.logger.Warn("quest cover: mark generating failed", slog.String("mission", missionID), slog.String("err", err.Error()))
		s.finish(key)
		return
	}

	go func() {
		defer s.finish(key)
		if chain {
			defer s.chainAfterCover(userID, missionID)
		}
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("quest cover image panic", slog.String("mission", missionID), slog.Any("panic", r))
				s.failCover(userID, missionID, fmt.Sprintf("internal error: %v", r))
			}
		}()

		runCtx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
		defer cancel()

		mission, err := s.missions.GetByID(runCtx, userID, missionID)
		if err != nil {
			s.failCover(userID, missionID, "could not load mission for cover generation")
			return
		}

		var prompt string
		if s.llm != nil {
			descCtx, descCancel := context.WithTimeout(runCtx, 90*time.Second)
			prompt, err = s.llm.DescribeCoverImage(descCtx, mission)
			descCancel()
		}
		if err != nil || strings.TrimSpace(prompt) == "" {
			prompt = buildCoverPrompt(mission)
		}
		prompt = enforceRealisticStoryImageStyle(prompt)

		refs := s.playerReference(runCtx, mission)
		prompt = withPlayerLikeness(prompt, refs)
		data, err := s.provider.GenerateImageBytes(runCtx, prompt, &ports.ImageOptions{Aspect: ports.ImageAspectWide, References: refs})
		if err != nil {
			s.failCover(userID, missionID, err.Error())
			return
		}

		filename := fmt.Sprintf("%s_%s_cover.png", mission.ID, userID)
		if err := s.put(runCtx, filename, "image/png", data); err != nil {
			s.failCover(userID, missionID, "failed to upload cover image")
			return
		}

		if err := s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
			mission.CoverImage = filename
			mission.CoverImageStatus = "ready"
			mission.CoverImageError = ""
			mission.CoverImageGenStartedAt = ""
		}); err != nil {
			s.logger.Warn("quest cover: save ready failed", slog.String("mission", missionID), slog.String("err", err.Error()))
		}
	}()
}

func (s *Images) chainAfterCover(userID, missionID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	mission, err := s.missions.GetByID(ctx, userID, missionID)
	if err != nil {
		return
	}
	if mission.SceneImages == nil || strings.TrimSpace(mission.SceneImages["0"]) == "" {
		s.GenerateSceneAsync(userID, missionID, 0)
		return
	}
	s.GenerateSceneAvatarsAsync(userID, missionID, 0)
}

func (s *Images) GenerateSceneAsync(userID, missionID string, stage int) {
	if !s.IsAvailable() {
		return
	}
	key := fmt.Sprintf("%s:%s:scene:%d", userID, missionID, stage)
	if !s.start(key) {
		return
	}
	stageKey := fmt.Sprintf("%d", stage)

	if err := s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
		if mission.SceneImageStatus == nil {
			mission.SceneImageStatus = map[string]string{}
		}
		if mission.SceneImageGenStartedAt == nil {
			mission.SceneImageGenStartedAt = map[string]string{}
		}
		if mission.SceneImageErrors == nil {
			mission.SceneImageErrors = map[string]string{}
		}
		mission.SceneImageStatus[stageKey] = "generating"
		delete(mission.SceneImageErrors, stageKey)
		mission.SceneImageGenStartedAt[stageKey] = time.Now().UTC().Format(time.RFC3339)
	}); err != nil {
		s.finish(key)
		return
	}

	go func() {
		defer s.finish(key)
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("quest scene image panic", slog.String("mission", missionID), slog.Int("stage", stage), slog.Any("panic", r))
				s.failScene(userID, missionID, stageKey, fmt.Sprintf("internal error: %v", r))
			}
		}()

		runCtx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
		defer cancel()

		mission, err := s.missions.GetByID(runCtx, userID, missionID)
		if err != nil {
			s.failScene(userID, missionID, stageKey, "could not load mission")
			return
		}

		scene := sceneByStage(mission, stage)
		if scene == nil {
			s.failScene(userID, missionID, stageKey, "scene not found for image generation")
			return
		}

		var prompt string
		if s.llm != nil {
			descCtx, descCancel := context.WithTimeout(runCtx, 90*time.Second)
			prompt, err = s.llm.DescribeSceneImage(descCtx, scene, mission)
			descCancel()
		}
		if err != nil || strings.TrimSpace(prompt) == "" {
			prompt = buildScenePrompt(scene, mission)
		}
		prompt = enforceRealisticStoryImageStyle(prompt)

		refs := s.playerReference(runCtx, mission)
		prompt = withPlayerLikeness(prompt, refs)
		data, err := s.provider.GenerateImageBytes(runCtx, prompt, &ports.ImageOptions{References: refs})
		if err != nil {
			s.failScene(userID, missionID, stageKey, err.Error())
			return
		}

		filename := fmt.Sprintf("%s_%s_%d.png", mission.ID, userID, stage)
		if err := s.put(runCtx, filename, "image/png", data); err != nil {
			s.failScene(userID, missionID, stageKey, "failed to upload scene image")
			return
		}

		if err := s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
			if mission.SceneImages == nil {
				mission.SceneImages = map[string]string{}
			}
			if mission.SceneImageStatus == nil {
				mission.SceneImageStatus = map[string]string{}
			}
			if mission.SceneImageGenStartedAt == nil {
				mission.SceneImageGenStartedAt = map[string]string{}
			}
			if mission.SceneImageErrors == nil {
				mission.SceneImageErrors = map[string]string{}
			}
			mission.SceneImages[stageKey] = filename
			mission.SceneImageStatus[stageKey] = "ready"
			delete(mission.SceneImageGenStartedAt, stageKey)
			delete(mission.SceneImageErrors, stageKey)
		}); err != nil {
			s.logger.Warn("quest scene: save ready failed", slog.String("mission", missionID), slog.String("err", err.Error()))
		}
		s.GenerateSceneAvatarsAsync(userID, missionID, stage)
	}()
}

func (s *Images) GenerateSceneAvatarsAsync(userID, missionID string, stage int) {
	if !s.IsAvailable() {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		mission, err := s.missions.GetByID(ctx, userID, missionID)
		if err != nil {
			return
		}
		scene := sceneByStage(mission, stage)
		if scene == nil {
			return
		}

		for _, present := range scene.Present {
			name := strings.TrimSpace(present.Name)
			if name == "" {
				continue
			}
			key := quest.CharacterAvatarKey(name)
			if key == "" {
				continue
			}
			if mission.CharacterAvatars != nil && strings.TrimSpace(mission.CharacterAvatars[key]) != "" {
				continue
			}
			s.GenerateCharacterAvatarAsync(userID, missionID, name, stage)
		}
	}()
}

func (s *Images) GenerateCharacterAvatarAsync(userID, missionID, characterName string, stage int) {
	if !s.IsAvailable() {
		return
	}
	avatarKey := quest.CharacterAvatarKey(characterName)
	if avatarKey == "" {
		return
	}
	key := fmt.Sprintf("%s:%s:avatar:%s", userID, missionID, avatarKey)
	if !s.start(key) {
		return
	}

	if err := s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
		if mission.CharacterAvatarStatus == nil {
			mission.CharacterAvatarStatus = map[string]string{}
		}
		if mission.CharacterAvatarGenStartedAt == nil {
			mission.CharacterAvatarGenStartedAt = map[string]string{}
		}
		if mission.CharacterAvatarErrors == nil {
			mission.CharacterAvatarErrors = map[string]string{}
		}
		mission.CharacterAvatarStatus[avatarKey] = "generating"
		delete(mission.CharacterAvatarErrors, avatarKey)
		mission.CharacterAvatarGenStartedAt[avatarKey] = time.Now().UTC().Format(time.RFC3339)
	}); err != nil {
		s.finish(key)
		return
	}

	go func() {
		defer s.finish(key)
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("quest avatar panic", slog.String("mission", missionID), slog.String("avatar", avatarKey), slog.Any("panic", r))
				s.failAvatar(userID, missionID, avatarKey, fmt.Sprintf("internal error: %v", r))
			}
		}()

		runCtx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
		defer cancel()

		mission, err := s.missions.GetByID(runCtx, userID, missionID)
		if err != nil {
			s.failAvatar(userID, missionID, avatarKey, "could not load mission")
			return
		}

		character := characterByName(mission, characterName)
		if character == nil {
			s.failAvatar(userID, missionID, avatarKey, "character not found for avatar generation")
			return
		}

		scene := sceneByStage(mission, stage)
		if scene == nil {
			scene = mission.CurrentScene
		}

		var prompt string
		if s.llm != nil {
			descCtx, descCancel := context.WithTimeout(runCtx, 90*time.Second)
			prompt, err = s.llm.DescribeCharacterAvatar(descCtx, character, scene, mission)
			descCancel()
		}
		if err != nil || strings.TrimSpace(prompt) == "" {
			prompt = buildAvatarPrompt(character, scene, mission)
		}
		prompt = enforceRealisticAvatarStyle(prompt)

		data, err := s.provider.GenerateImageBytes(runCtx, prompt, &ports.ImageOptions{Aspect: ports.ImageAspectSquare, Size: "128x128"})
		if err != nil {
			s.failAvatar(userID, missionID, avatarKey, err.Error())
			return
		}

		filename := fmt.Sprintf("%s_%s_avatar_%s.png", mission.ID, userID, avatarFileToken(character.Name, avatarKey))
		if err := s.put(runCtx, filename, "image/png", data); err != nil {
			s.failAvatar(userID, missionID, avatarKey, "failed to upload character avatar")
			return
		}

		if err := s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
			if mission.CharacterAvatars == nil {
				mission.CharacterAvatars = map[string]string{}
			}
			if mission.CharacterAvatarStatus == nil {
				mission.CharacterAvatarStatus = map[string]string{}
			}
			if mission.CharacterAvatarErrors == nil {
				mission.CharacterAvatarErrors = map[string]string{}
			}
			if mission.CharacterAvatarGenStartedAt == nil {
				mission.CharacterAvatarGenStartedAt = map[string]string{}
			}
			mission.CharacterAvatars[avatarKey] = filename
			mission.CharacterAvatarStatus[avatarKey] = "ready"
			delete(mission.CharacterAvatarErrors, avatarKey)
			delete(mission.CharacterAvatarGenStartedAt, avatarKey)
		}); err != nil {
			s.logger.Warn("quest avatar: save ready failed", slog.String("mission", missionID), slog.String("err", err.Error()))
		}
	}()
}

func (s *Images) RegenerateFailed(userID string, mission *quest.CustomMission) {
	if !s.IsAvailable() || mission == nil {
		return
	}
	if mission.CoverImageStatus == "error" {
		s.GenerateCoverAsync(userID, mission.ID)
	}
	for stageKey, status := range mission.SceneImageStatus {
		if status != "error" {
			continue
		}
		stage, err := strconv.Atoi(stageKey)
		if err != nil {
			continue
		}
		s.GenerateSceneAsync(userID, mission.ID, stage)
	}
	for avatarKey, status := range mission.CharacterAvatarStatus {
		if status != "error" {
			continue
		}
		for _, character := range mission.Characters {
			if quest.CharacterAvatarKey(character.Name) == avatarKey {
				s.GenerateCharacterAvatarAsync(userID, mission.ID, character.Name, mission.CurrentStage)
				break
			}
		}
	}
}

func (s *Images) RegenerateOne(userID string, mission *quest.CustomMission, kind, key string) {
	if !s.IsAvailable() || mission == nil {
		return
	}
	switch kind {
	case "cover":
		s.GenerateCoverAsync(userID, mission.ID)
	case "scene":
		if stage, err := strconv.Atoi(key); err == nil {
			s.GenerateSceneAsync(userID, mission.ID, stage)
		}
	case "avatar":
		s.GenerateCharacterAvatarAsync(userID, mission.ID, key, mission.CurrentStage)
	}
}

func (s *Images) DeleteMissionImages(ctx context.Context, mission *quest.CustomMission) {
	if s == nil || s.storage == nil || mission == nil {
		return
	}
	for _, filename := range collectImageFiles(mission) {
		path, err := media.NewPath(s.bucket + "/" + filename)
		if err != nil {
			continue
		}
		if err := s.storage.Delete(ctx, path); err != nil {
			s.logger.Warn("quest: delete image failed", slog.String("file", filename), slog.String("err", err.Error()))
		}
	}
}

func collectImageFiles(mission *quest.CustomMission) []string {
	var files []string
	if strings.TrimSpace(mission.CoverImage) != "" {
		files = append(files, mission.CoverImage)
	}
	for _, f := range mission.SceneImages {
		if strings.TrimSpace(f) != "" {
			files = append(files, f)
		}
	}
	for _, f := range mission.CharacterAvatars {
		if strings.TrimSpace(f) != "" {
			files = append(files, f)
		}
	}
	return files
}

func (s *Images) start(key string) bool {
	_, loaded := s.running.LoadOrStore(key, true)
	return !loaded
}

func (s *Images) finish(key string) { s.running.Delete(key) }

func truncateImageErr(msg string) string {
	msg = strings.TrimSpace(msg)
	msg = strings.ReplaceAll(msg, "\n", " ")
	if len(msg) > 300 {
		return msg[:300] + "…"
	}
	return msg
}

func (s *Images) failCover(userID, missionID, msg string) {
	msg = truncateImageErr(msg)
	_ = s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
		mission.CoverImageStatus = "error"
		mission.CoverImageError = msg
		mission.CoverImageGenStartedAt = ""
	})
}

func (s *Images) failScene(userID, missionID, stageKey, msg string) {
	msg = truncateImageErr(msg)
	_ = s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
		if mission.SceneImageStatus == nil {
			mission.SceneImageStatus = map[string]string{}
		}
		if mission.SceneImageErrors == nil {
			mission.SceneImageErrors = map[string]string{}
		}
		if mission.SceneImageGenStartedAt == nil {
			mission.SceneImageGenStartedAt = map[string]string{}
		}
		mission.SceneImageStatus[stageKey] = "error"
		mission.SceneImageErrors[stageKey] = msg
		delete(mission.SceneImageGenStartedAt, stageKey)
	})
}

func (s *Images) failAvatar(userID, missionID, avatarKey, msg string) {
	msg = truncateImageErr(msg)
	_ = s.updateMission(userID, missionID, func(mission *quest.CustomMission) {
		if mission.CharacterAvatarStatus == nil {
			mission.CharacterAvatarStatus = map[string]string{}
		}
		if mission.CharacterAvatarErrors == nil {
			mission.CharacterAvatarErrors = map[string]string{}
		}
		if mission.CharacterAvatarGenStartedAt == nil {
			mission.CharacterAvatarGenStartedAt = map[string]string{}
		}
		mission.CharacterAvatarStatus[avatarKey] = "error"
		mission.CharacterAvatarErrors[avatarKey] = msg
		delete(mission.CharacterAvatarGenStartedAt, avatarKey)
	})
}

func (s *Images) updateMission(userID, missionID string, mutate func(*quest.CustomMission)) error {
	dbCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return s.missions.Update(dbCtx, userID, missionID, func(mission *quest.CustomMission) error {
		mutate(mission)
		return nil
	})
}

func sceneByStage(mission *quest.CustomMission, stage int) *quest.DynamicScene {
	if mission == nil {
		return nil
	}
	for i := range mission.Scenes {
		if mission.Scenes[i].Stage == stage {
			scene := mission.Scenes[i]
			return &scene
		}
	}
	if mission.CurrentScene != nil && mission.CurrentScene.Stage == stage {
		return mission.CurrentScene
	}
	return nil
}

func buildCoverPrompt(mission *quest.CustomMission) string {
	var b strings.Builder
	b.WriteString("cinematic movie poster, photorealistic realism, realistic lighting, ")
	if mission.Genre != "" {
		b.WriteString(mission.Genre)
		b.WriteString(" genre, ")
	}
	b.WriteString(mission.Title)
	b.WriteString(". ")
	b.WriteString(mission.Description)
	b.WriteString(", dramatic lighting, detailed, no text, no letters")
	for _, character := range mission.Characters {
		if strings.TrimSpace(character.Appearance) == "" {
			continue
		}
		b.WriteString(fmt.Sprintf(". NPC %s (NOT the player, different person)%s: %s", character.Name, ageSuffix(character.Age), character.Appearance))
	}
	return b.String()
}

func ageSuffix(age string) string {
	if strings.TrimSpace(age) == "" {
		return ""
	}
	return ", age " + strings.TrimSpace(age)
}

func buildScenePrompt(scene *quest.DynamicScene, mission *quest.CustomMission) string {
	var b strings.Builder
	b.WriteString("cinematic dramatic scene, photorealistic realism, realistic lighting, ")
	b.WriteString(scene.Narration)
	if strings.TrimSpace(mission.CoverImage) != "" {
		b.WriteString(". visual style consistent with mission cover art")
	}
	for _, present := range scene.Present {
		for _, character := range mission.Characters {
			if !strings.EqualFold(character.Name, present.Name) || strings.TrimSpace(character.Appearance) == "" {
				continue
			}
			b.WriteString(fmt.Sprintf(". NPC %s (NOT the player, different person)%s: %s", character.Name, ageSuffix(character.Age), character.Appearance))
		}
	}
	return b.String()
}

func buildAvatarPrompt(character *quest.Character, scene *quest.DynamicScene, mission *quest.CustomMission) string {
	_ = mission
	var b strings.Builder
	b.WriteString("single character portrait photo, head and shoulders, centered composition, square framing, realistic face details, natural skin texture, realistic proportions, natural lighting, clean background, no text, no letters, no logos, european caucasian person. ")
	b.WriteString(character.Name)
	b.WriteString(", ")
	if strings.TrimSpace(character.Role) != "" {
		b.WriteString(character.Role)
		b.WriteString(", ")
	}
	if strings.TrimSpace(character.Age) != "" {
		b.WriteString("age ")
		b.WriteString(character.Age)
		b.WriteString(", ")
	}
	if strings.TrimSpace(character.Appearance) != "" {
		b.WriteString(character.Appearance)
		b.WriteString(". ")
	}
	if strings.TrimSpace(character.Personality) != "" {
		b.WriteString("mood: ")
		b.WriteString(character.Personality)
		b.WriteString(". ")
	}
	if scene != nil && strings.TrimSpace(scene.Narration) != "" {
		b.WriteString("scene context: ")
		b.WriteString(scene.Narration)
		b.WriteString(". ")
	}
	b.WriteString("photorealistic realism only. no anime, no cartoon, no illustration, no painting, no 3d render, no stylized art.")
	return b.String()
}

func enforceRealisticAvatarStyle(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	styleLock := "photorealistic realism only, realistic skin texture, realistic proportions, natural lighting, no anime, no cartoon, no comic style, no illustration, no painting, no watercolor, no 3d render, no stylized art, no text, no watermark"
	if trimmed == "" {
		return styleLock
	}
	return trimmed + ". " + styleLock
}

func withPlayerLikeness(prompt string, refs []ports.ImageReference) string {
	if len(refs) == 0 {
		return prompt
	}
	return strings.TrimSpace(prompt) + " The protagonist is the real person shown in the attached reference photo: depict this exact person as the main hero, prominently featured, keeping their face and likeness."
}

func enforceRealisticStoryImageStyle(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	styleLock := "photorealistic realism only, realistic proportions, natural skin texture, cinematic natural lighting, no anime, no cartoon, no comic style, no illustration, no painting, no watercolor, no 3d render, no stylized art, no text, no watermark"
	if trimmed == "" {
		return styleLock
	}
	return trimmed + ". " + styleLock
}

func characterByName(mission *quest.CustomMission, name string) *quest.Character {
	if mission == nil {
		return nil
	}
	key := quest.CharacterAvatarKey(name)
	if key == "" {
		return nil
	}
	for i := range mission.Characters {
		if quest.CharacterAvatarKey(mission.Characters[i].Name) == key {
			character := mission.Characters[i]
			return &character
		}
	}
	return nil
}

func avatarFileToken(name, fallback string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = fallback
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToLower(trimmed) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	token := strings.Trim(b.String(), "_")
	if token == "" {
		return "character"
	}
	return token
}
