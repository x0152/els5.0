package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/utils/llmx"
)

type ChatClient interface {
	Available() bool
	Chat(ctx context.Context, system, user string) (string, error)
}

type LLMGateway struct {
	client  ChatClient
	grammar ChatClient
	timeout time.Duration
	logger  *slog.Logger
}

func NewLLMGateway(client, grammar ChatClient, timeout time.Duration, logger *slog.Logger) *LLMGateway {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	if logger == nil {
		logger = slog.Default()
	}
	if grammar == nil {
		grammar = client
	}
	return &LLMGateway{client: client, grammar: grammar, timeout: timeout, logger: logger}
}

func (g *LLMGateway) IsAvailable() bool {
	return g.client != nil && g.client.Available()
}

type createMissionResult struct {
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	SecretEnding    string            `json:"secretEnding"`
	NarratorVoice   string            `json:"narratorVoice,omitempty"`
	Characters      []quest.Character `json:"characters"`
	TotalStages     int               `json:"totalScenes"`
	EstimatedScenes int               `json:"estimatedScenes"`
	PlotPoints      []quest.PlotPoint `json:"plotPoints"`
	Resolution      *quest.Resolution `json:"resolution"`
}

func (g *LLMGateway) CreateMission(
	ctx context.Context,
	prompt, genre, practiceGoals, language string,
	profile quest.PlayerProfile,
) (*createMissionResult, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("llm not available")
	}
	system, user := quest.BuildCreateMissionPrompts(prompt, genre, practiceGoals, language, profile)
	var result createMissionResult
	if err := g.callJSON(ctx, system, user, "mission", 8, &result); err != nil {
		return nil, err
	}
	if result.Title == "" || len(result.Characters) == 0 {
		return nil, fmt.Errorf("llm returned incomplete mission")
	}
	for i := range result.PlotPoints {
		result.PlotPoints[i].NormalizeFact()
	}
	if result.EstimatedScenes > 0 {
		result.TotalStages = result.EstimatedScenes
	}
	if result.TotalStages < 4 {
		result.TotalStages = 4
	}
	if result.TotalStages > 12 {
		result.TotalStages = 12
	}
	if result.EstimatedScenes <= 0 {
		result.EstimatedScenes = result.TotalStages
	}
	return &result, nil
}

func (g *LLMGateway) GenerateScene(
	ctx context.Context,
	mission *quest.CustomMission,
	profile *quest.PlayerProfile,
	stage int,
	sceneCtx *quest.SceneContext,
) (*quest.DynamicScene, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("llm not available")
	}
	system, user := quest.BuildScenePrompts(mission, profile, stage, sceneCtx)
	var scene quest.DynamicScene
	if err := g.callJSON(ctx, system, user, "scene", 8, &scene); err != nil {
		return nil, err
	}
	g.sanitizeScene(&scene)

	if stage < 0 {
		stage = mission.CurrentStage
	}
	scene.Stage = stage
	scene.IsFinal = stage+1 >= mission.TotalStages || (sceneCtx != nil && sceneCtx.Finale)
	if sceneCtx != nil && strings.TrimSpace(sceneCtx.Flavor) != "" {
		scene.Flavor = sceneCtx.Flavor
	} else {
		scene.Flavor = quest.FlavorPlotBeat
	}
	return &scene, nil
}

func (g *LLMGateway) CheckGrammar(ctx context.Context, playerText, language string, strict bool) (*quest.GrammarCheck, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("llm not available")
	}
	system, user := quest.BuildGrammarPrompts(playerText, language, strict)
	var result quest.GrammarCheck
	if err := g.callJSONWith(ctx, g.grammar, system, user, "grammar", 3, &result); err != nil {
		return nil, err
	}
	result.OK = len(result.Errors) == 0
	return &result, nil
}

func (g *LLMGateway) SuggestNativeReplies(
	ctx context.Context,
	mission *quest.CustomMission,
	playerText string,
	profile *quest.PlayerProfile,
) ([]string, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("llm not available")
	}
	system, user := quest.BuildNativeReplyPrompts(mission, playerText, profile)
	var result struct {
		Variants []string `json:"variants"`
	}
	if err := g.callJSON(ctx, system, user, "native reply variants", 3, &result); err != nil {
		return nil, err
	}

	variants := make([]string, 0, len(result.Variants))
	seen := map[string]struct{}{}
	for _, option := range result.Variants {
		value := strings.TrimSpace(option)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		variants = append(variants, value)
		if len(variants) >= 5 {
			break
		}
	}
	if len(variants) == 0 {
		return nil, fmt.Errorf("empty native reply variants")
	}
	return variants, nil
}

func (g *LLMGateway) GenerateWorldResponse(
	ctx context.Context,
	mission *quest.CustomMission,
	playerText string,
	profile *quest.PlayerProfile,
) (*quest.WorldResult, error) {
	return g.GenerateWorldResponseStream(ctx, mission, playerText, profile, nil)
}

// GenerateWorldResponseStream generates the world reply, streaming a draft via
// onPartial as the LLM streams. If the client cannot stream or
// onPartial == nil — a normal blocking call.
func (g *LLMGateway) GenerateWorldResponseStream(
	ctx context.Context,
	mission *quest.CustomMission,
	playerText string,
	profile *quest.PlayerProfile,
	onPartial func(*quest.PartialWorld),
) (*quest.WorldResult, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("llm not available")
	}
	system, user := quest.BuildWorldPrompts(mission, playerText, profile)

	var result quest.WorldResult
	if err := g.callJSONStream(ctx, system, user, "world response", 5, &result, onPartial); err != nil {
		return nil, err
	}
	g.sanitizeWorldResult(&result)
	return &result, nil
}

type chatStreamer interface {
	ChatStream(ctx context.Context, system, user string, onDelta func(string)) error
}

func (g *LLMGateway) callJSONStream(
	ctx context.Context,
	system, user, label string,
	timeoutMul int,
	out any,
	onPartial func(*quest.PartialWorld),
) error {
	streamer, ok := g.client.(chatStreamer)
	if !ok || onPartial == nil {
		return g.callJSON(ctx, system, user, label, timeoutMul, out)
	}

	mul := time.Duration(timeoutMul)
	if mul < 1 {
		mul = 1
	}
	llmCtx, cancel := context.WithTimeout(ctx, g.timeout*mul)
	defer cancel()

	var buf strings.Builder
	err := streamer.ChatStream(llmCtx, system, user, func(delta string) {
		buf.WriteString(delta)
		if partial := quest.ParsePartialWorld(buf.String()); partial != nil {
			onPartial(partial)
		}
	})
	if err != nil {
		g.logger.Warn("quest: llm stream failed, falling back to blocking call", slog.String("label", label), slog.String("err", err.Error()))
		return g.callJSON(ctx, system, user, label, timeoutMul, out)
	}

	raw := llmx.CleanLLMResponse(buf.String())
	if strings.TrimSpace(raw) == "" {
		return fmt.Errorf("empty llm response")
	}
	if jsonErr := json.Unmarshal([]byte(raw), out); jsonErr != nil {
		return g.callJSON(ctx, system, user, label, timeoutMul, out)
	}
	return nil
}

func (g *LLMGateway) EvaluateStoryState(
	ctx context.Context,
	mission *quest.CustomMission,
	worldResponse *quest.WorldResult,
	playerText string,
) (*quest.EvaluationResult, error) {
	if !g.IsAvailable() {
		return nil, fmt.Errorf("llm not available")
	}
	system, user := quest.BuildEvaluatorPrompts(mission, worldResponse, playerText)
	var result quest.EvaluationResult
	if err := g.callJSON(ctx, system, user, "story evaluator", 2, &result); err != nil {
		return nil, err
	}
	if result.SceneState == "" {
		result.SceneState = quest.SceneActive
	}
	if result.PlayerIntent == "" {
		result.PlayerIntent = quest.IntentExploring
	}
	if result.NarrativeMomentum == "" {
		result.NarrativeMomentum = "medium"
	}
	return &result, nil
}

func (g *LLMGateway) GenerateEpilogue(ctx context.Context, mission *quest.CustomMission, outcome string, lastWorld *quest.WorldResult, profile *quest.PlayerProfile) (string, error) {
	if !g.IsAvailable() {
		return "", fmt.Errorf("llm not available")
	}
	system, user := quest.BuildEpiloguePrompts(mission, outcome, lastWorld, profile)
	var result struct {
		Epilogue string `json:"epilogue"`
	}
	if err := g.callJSON(ctx, system, user, "epilogue", 2, &result); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Epilogue), nil
}

func (g *LLMGateway) SummarizeHistory(ctx context.Context, mission *quest.CustomMission, turns []quest.DialogueTurn) (string, error) {
	if !g.IsAvailable() {
		return "", fmt.Errorf("llm not available")
	}
	system, user := quest.BuildSummarizePrompts(mission, turns)
	var result struct {
		Summary string `json:"summary"`
	}
	if err := g.callJSON(ctx, system, user, "summarize history", 2, &result); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Summary), nil
}

func (g *LLMGateway) DescribeCoverImage(ctx context.Context, mission *quest.CustomMission) (string, error) {
	if !g.IsAvailable() {
		return "", fmt.Errorf("llm not available")
	}
	system, user := quest.BuildCoverImageDescriptionPrompts(mission)
	return g.describe(ctx, system, user, "cover image")
}

func (g *LLMGateway) DescribeSceneImage(ctx context.Context, scene *quest.DynamicScene, mission *quest.CustomMission) (string, error) {
	if !g.IsAvailable() {
		return "", fmt.Errorf("llm not available")
	}
	system, user := quest.BuildSceneImageDescriptionPrompts(scene, mission)
	return g.describe(ctx, system, user, "scene image")
}

func (g *LLMGateway) DescribeCharacterAvatar(ctx context.Context, character *quest.Character, scene *quest.DynamicScene, mission *quest.CustomMission) (string, error) {
	if !g.IsAvailable() {
		return "", fmt.Errorf("llm not available")
	}
	if character == nil {
		return "", fmt.Errorf("character is required")
	}
	system, user := quest.BuildCharacterAvatarDescriptionPrompts(character, scene, mission)
	return g.describe(ctx, system, user, "character avatar")
}

func (g *LLMGateway) describe(ctx context.Context, system, user, label string) (string, error) {
	var result struct {
		Description string `json:"description"`
	}
	if err := g.callJSON(ctx, system, user, label, 2, &result); err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Description), nil
}

func (g *LLMGateway) callJSON(ctx context.Context, system, user, label string, timeoutMul int, out any) error {
	return g.callJSONWith(ctx, g.client, system, user, label, timeoutMul, out)
}

func (g *LLMGateway) callJSONWith(ctx context.Context, client ChatClient, system, user, label string, timeoutMul int, out any) error {
	raw, err := g.callWith(ctx, client, system, user, timeoutMul)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(raw), out); err == nil {
		return nil
	}

	retryUser := user + "\n\nIMPORTANT: Return exactly one valid JSON object and nothing else."
	rawRetry, retryErr := g.callWith(ctx, client, system, retryUser, timeoutMul)
	if retryErr != nil {
		return fmt.Errorf("failed to parse %s (retry failed: %w)", label, retryErr)
	}
	if err := json.Unmarshal([]byte(rawRetry), out); err != nil {
		return fmt.Errorf("failed to parse %s: %w", label, err)
	}
	return nil
}

func (g *LLMGateway) call(ctx context.Context, system, user string, timeoutMul int) (string, error) {
	return g.callWith(ctx, g.client, system, user, timeoutMul)
}

func (g *LLMGateway) callWith(ctx context.Context, client ChatClient, system, user string, timeoutMul int) (string, error) {
	if client == nil || !client.Available() {
		return "", fmt.Errorf("llm not available")
	}
	mul := time.Duration(timeoutMul)
	if mul < 1 {
		mul = 1
	}
	llmCtx, cancel := context.WithTimeout(ctx, g.timeout*mul)
	defer cancel()

	raw, err := client.Chat(llmCtx, system, user)
	if err != nil {
		return "", fmt.Errorf("llm call failed: %w", err)
	}
	raw = llmx.CleanLLMResponse(raw)
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("empty llm response")
	}
	return raw, nil
}

func (g *LLMGateway) sanitizeScene(scene *quest.DynamicScene) {
	if scene == nil {
		return
	}
	scene.Narration = strings.TrimSpace(scene.Narration)
	scene.NarrationVoice = quest.NormalizeVoice(scene.NarrationVoice)

	if len(scene.Present) > 0 {
		filtered := make([]quest.SceneCharacter, 0, len(scene.Present))
		for _, ch := range scene.Present {
			name := strings.TrimSpace(ch.Name)
			voice := quest.NormalizeVoice(ch.Voice)
			dialogue := strings.TrimSpace(ch.Dialogue)
			if name == "" || dialogue == "" {
				continue
			}
			filtered = append(filtered, quest.SceneCharacter{Name: name, Voice: voice, Dialogue: dialogue})
		}
		scene.Present = filtered
	}

	if len(scene.Objects) > 0 {
		filtered := make([]string, 0, len(scene.Objects))
		for _, obj := range scene.Objects {
			value := strings.TrimSpace(obj)
			if value == "" {
				continue
			}
			filtered = append(filtered, value)
		}
		scene.Objects = filtered
	}
}

func (g *LLMGateway) sanitizeWorldResult(result *quest.WorldResult) {
	if result == nil {
		return
	}
	result.Narration = strings.TrimSpace(result.Narration)
	result.NarrationVoice = quest.NormalizeVoice(result.NarrationVoice)
	if len(result.Responses) == 0 {
		return
	}
	filtered := make([]quest.CharacterLine, 0, len(result.Responses))
	for _, line := range result.Responses {
		name := strings.TrimSpace(line.Name)
		voice := quest.NormalizeVoice(line.Voice)
		text := strings.TrimSpace(line.Text)
		if name == "" || text == "" {
			continue
		}
		filtered = append(filtered, quest.CharacterLine{Name: name, Voice: voice, Text: text})
	}
	result.Responses = filtered
}
