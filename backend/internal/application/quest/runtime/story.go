package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/media"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
)

type Generator struct {
	llm      *LLMGateway
	missions quest.MissionRepository
	profiles quest.ProfileRepository
	images   *Images
	accounts iam.AccountRepository
	urls     media.PublicURL
	logger   *slog.Logger
}

func NewGenerator(llm *LLMGateway, missions quest.MissionRepository, profiles quest.ProfileRepository, images *Images, accounts iam.AccountRepository, urls media.PublicURL, logger *slog.Logger) *Generator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Generator{llm: llm, missions: missions, profiles: profiles, images: images, accounts: accounts, urls: urls, logger: logger}
}

func (g *Generator) resolvePlayerAvatar(ctx context.Context, userID string) string {
	if g.accounts == nil {
		return ""
	}
	id, err := vo.ParseID(userID)
	if err != nil {
		return ""
	}
	account, err := g.accounts.GetByID(ctx, iam.AccountID{ID: id})
	if err != nil {
		return ""
	}
	path, ok := g.urls.ParsePath(account.PictureURL())
	if !ok {
		return ""
	}
	return path.String()
}

func (g *Generator) Enqueue(userID, missionID string, req quest.CreateMissionRequest) {
	go g.run(userID, missionID, req)
}

func (g *Generator) run(userID, missionID string, req quest.CreateMissionRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			g.logger.Error("quest: story generation panic", slog.String("mission", missionID), slog.Any("panic", r))
			g.fail(userID, missionID, fmt.Errorf("internal error during generation"))
		}
	}()

	profile, err := g.profiles.Get(ctx, userID)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			g.fail(userID, missionID, fmt.Errorf("failed to load profile: %w", err))
			return
		}
		profile = quest.NewDefaultProfile()
		if saveErr := g.profiles.Save(ctx, userID, profile); saveErr != nil {
			g.fail(userID, missionID, fmt.Errorf("failed to create default profile: %w", saveErr))
			return
		}
	}
	applyAccountIdentity(ctx, g.accounts, userID, &profile)

	language := strings.TrimSpace(req.Language)
	if language == "" {
		language = "English"
	}

	meta, err := g.llm.CreateMission(
		ctx,
		strings.TrimSpace(req.Prompt),
		strings.TrimSpace(req.Genre),
		strings.TrimSpace(req.PracticeGoals),
		language,
		profile,
	)
	if err != nil {
		g.fail(userID, missionID, fmt.Errorf("failed to create mission: %w", err))
		return
	}

	mission := &quest.CustomMission{
		ID:                          missionID,
		Title:                       meta.Title,
		Description:                 meta.Description,
		UserPrompt:                  req.Prompt,
		Genre:                       req.Genre,
		Language:                    language,
		PracticeGoals:               req.PracticeGoals,
		SecretEnding:                meta.SecretEnding,
		NarratorVoice:               meta.NarratorVoice,
		Characters:                  meta.Characters,
		PlotPoints:                  meta.PlotPoints,
		Resolution:                  meta.Resolution,
		CurrentStage:                0,
		TotalStages:                 meta.TotalStages,
		EstimatedScenes:             meta.EstimatedScenes,
		IsComplete:                  false,
		Scenes:                      []quest.DynamicScene{},
		History:                     []quest.DialogueTurn{},
		SkillSignals:                map[string]int{},
		SkillCategories:             map[string]string{},
		SceneImages:                 map[string]string{},
		SceneImageStatus:            map[string]string{},
		CharacterAvatars:            map[string]string{},
		CharacterAvatarStatus:       map[string]string{},
		CharacterAvatarErrors:       map[string]string{},
		CharacterAvatarGenStartedAt: map[string]string{},
		CreatedAt:                   time.Now().Format(time.RFC3339),
		GenerationStatus:            quest.GenerationStatusGenerating,
		GenerationStep:              quest.GenerationStepFirstScene,
	}
	mission.PlayerAvatarImage = g.resolvePlayerAvatar(ctx, userID)
	ensureMissionVoices(mission)
	mission.EnsureNPCStates()

	if err := g.missions.Save(ctx, userID, mission); err != nil {
		g.fail(userID, missionID, fmt.Errorf("failed to save mission: %w", err))
		return
	}

	openingCtx := &quest.SceneContext{Flavor: quest.FlavorPlotBeat}
	scene, err := g.llm.GenerateScene(ctx, mission, &profile, 0, openingCtx)
	if err != nil {
		g.fail(userID, missionID, fmt.Errorf("failed to create first scene: %w", err))
		return
	}
	applySceneVoices(mission, scene)

	mission.CurrentScene = scene
	mission.Scenes = append(mission.Scenes, *scene)
	mission.History = append(mission.History,
		quest.DialogueTurn{Scene: 0, Speaker: "system", Text: "- Scene 1 -"},
		quest.DialogueTurn{Scene: 0, Speaker: "narrator", Voice: scene.NarrationVoice, Text: scene.Narration},
	)
	for _, character := range scene.Present {
		if strings.TrimSpace(character.Name) == "" || strings.TrimSpace(character.Dialogue) == "" {
			continue
		}
		mission.History = append(mission.History, quest.DialogueTurn{
			Scene:   0,
			Speaker: character.Name,
			Voice:   character.Voice,
			Text:    character.Dialogue,
		})
	}

	if g.images != nil && g.images.IsAvailable() {
		mission.CoverImageStatus = "generating"
		mission.SceneImageStatus["0"] = "generating"
	}

	mission.GenerationStatus = quest.GenerationStatusReady
	mission.GenerationStep = ""
	mission.GenerationError = ""

	saveCtx, saveCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer saveCancel()
	if err := g.missions.Save(saveCtx, userID, mission); err != nil {
		g.fail(userID, missionID, fmt.Errorf("failed to save mission: %w", err))
		return
	}

	if g.images != nil && g.images.IsAvailable() {
		g.images.GenerateInitialImagesAsync(userID, missionID)
	}
}

func (g *Generator) fail(userID, missionID string, cause error) {
	g.logger.Warn("quest: story generation failed", slog.String("mission", missionID), slog.String("err", cause.Error()))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := g.missions.Update(ctx, userID, missionID, func(mission *quest.CustomMission) error {
		mission.GenerationStatus = quest.GenerationStatusError
		mission.GenerationStep = ""
		mission.GenerationError = cause.Error()
		return nil
	})
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		g.logger.Error("quest: mark mission failed", slog.String("mission", missionID), slog.String("err", err.Error()))
	}
}
