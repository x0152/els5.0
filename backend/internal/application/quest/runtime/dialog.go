package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/els/backend/internal/domain/iam"
	"github.com/els/backend/internal/domain/quest"
	"github.com/els/backend/internal/domain/shared"
)

type respondJob struct {
	ID        string
	UserID    string
	MissionID string
	Text      string
	Strict    bool
	Status    string
	Step      string
	Grammar   *quest.GrammarCheck
	Partial   *quest.PartialWorld
	Result    *quest.RespondJobResult
	Error     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type dialogJobManager struct {
	mu              sync.RWMutex
	jobs            map[string]*respondJob
	activeMission   map[string]string
	latestMission   map[string]string
	sceneGenerating map[string]bool
}

func newDialogJobManager() *dialogJobManager {
	return &dialogJobManager{
		jobs:            map[string]*respondJob{},
		activeMission:   map[string]string{},
		latestMission:   map[string]string{},
		sceneGenerating: map[string]bool{},
	}
}

// Keys activeMission/latestMission/sceneGenerating include userID: the same
// shared mission is played by different users independently.
func runKey(userID, missionID string) string { return userID + "\x00" + missionID }

func (m *dialogJobManager) SetSceneGenerating(userID, missionID string, generating bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if generating {
		m.sceneGenerating[runKey(userID, missionID)] = true
	} else {
		delete(m.sceneGenerating, runKey(userID, missionID))
	}
}

const finishedJobTTL = time.Hour

func (m *dialogJobManager) Start(userID, missionID, text string, strict bool) (*respondJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := runKey(userID, missionID)
	if _, busy := m.activeMission[key]; busy {
		return nil, fmt.Errorf("%w: response is already generating for this mission", shared.ErrConflict)
	}
	if m.sceneGenerating[key] {
		return nil, fmt.Errorf("%w: next scene is still generating, please wait", shared.ErrConflict)
	}
	m.pruneLocked()

	id := fmt.Sprintf("job_%d", time.Now().UnixMilli())
	now := time.Now()
	job := &respondJob{
		ID:        id,
		UserID:    userID,
		MissionID: missionID,
		Text:      text,
		Strict:    strict,
		Status:    "running",
		Step:      "grammar",
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.jobs[id] = job
	m.activeMission[key] = id
	m.latestMission[key] = id
	return job, nil
}

func (m *dialogJobManager) pruneLocked() {
	cutoff := time.Now().Add(-finishedJobTTL)
	for id, job := range m.jobs {
		if job.Status == "running" || job.UpdatedAt.After(cutoff) {
			continue
		}
		delete(m.jobs, id)
		if m.latestMission[runKey(job.UserID, job.MissionID)] == id {
			delete(m.latestMission, runKey(job.UserID, job.MissionID))
		}
	}
}

func (m *dialogJobManager) SnapshotByMission(userID, missionID string) *quest.RespondJobStatusResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobID, ok := m.latestMission[runKey(userID, missionID)]
	if !ok {
		return nil
	}
	job := m.jobs[jobID]
	if job == nil {
		return nil
	}
	return snapshotLocked(job)
}

func (m *dialogJobManager) Snapshot(jobID string) *quest.RespondJobStatusResponse {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job := m.jobs[jobID]
	if job == nil {
		return nil
	}
	return snapshotLocked(job)
}

func snapshotLocked(job *respondJob) *quest.RespondJobStatusResponse {
	resp := &quest.RespondJobStatusResponse{
		JobID:     job.ID,
		Status:    job.Status,
		Step:      job.Step,
		InputText: job.Text,
		Error:     job.Error,
	}
	if job.Grammar != nil {
		grammarCopy := *job.Grammar
		resp.Grammar = &grammarCopy
	}
	if job.Partial != nil && job.Status == "running" {
		partialCopy := *job.Partial
		resp.Partial = &partialCopy
	}
	if job.Result != nil {
		resultCopy := *job.Result
		resp.Result = &resultCopy
	}
	return resp
}

func (m *dialogJobManager) Get(jobID string) *respondJob {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job := m.jobs[jobID]
	if job == nil {
		return nil
	}
	copyJob := *job
	return &copyJob
}

func (m *dialogJobManager) SetStep(jobID, step string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job := m.jobs[jobID]; job != nil {
		job.Step = step
		job.UpdatedAt = time.Now()
	}
}

func (m *dialogJobManager) SetGrammar(jobID string, grammar *quest.GrammarCheck) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job := m.jobs[jobID]; job != nil {
		job.Grammar = grammar
		job.UpdatedAt = time.Now()
	}
}

func (m *dialogJobManager) SetPartial(jobID string, partial *quest.PartialWorld) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if job := m.jobs[jobID]; job != nil && job.Status == "running" {
		job.Partial = partial
		job.UpdatedAt = time.Now()
	}
}

func (m *dialogJobManager) Complete(jobID string, result *quest.RespondJobResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job := m.jobs[jobID]
	if job == nil {
		return
	}
	job.Result = result
	job.Status = "done"
	job.Step = "done"
	job.UpdatedAt = time.Now()
	delete(m.activeMission, runKey(job.UserID, job.MissionID))
}

func (m *dialogJobManager) Fail(jobID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job := m.jobs[jobID]
	if job == nil {
		return
	}
	job.Status = "error"
	job.Step = "error"
	job.Error = err.Error()
	job.UpdatedAt = time.Now()
	delete(m.activeMission, runKey(job.UserID, job.MissionID))
}

type Dialog struct {
	llm      *LLMGateway
	missions quest.MissionRepository
	profiles quest.ProfileRepository
	accounts iam.AccountRepository
	images   *Images
	jobs     *dialogJobManager
	logger   *slog.Logger
}

func NewDialog(llm *LLMGateway, missions quest.MissionRepository, profiles quest.ProfileRepository, accounts iam.AccountRepository, images *Images, logger *slog.Logger) *Dialog {
	if logger == nil {
		logger = slog.Default()
	}
	return &Dialog{llm: llm, missions: missions, profiles: profiles, accounts: accounts, images: images, jobs: newDialogJobManager(), logger: logger}
}

func (s *Dialog) Start(ctx context.Context, userID, missionID, text string, strict bool) (*quest.StartRespondJobResponse, error) {
	mission, err := s.missions.GetByID(ctx, userID, missionID)
	if err != nil {
		return nil, err
	}
	if mission.IsComplete {
		return nil, fmt.Errorf("%w: mission already complete", shared.ErrConflict)
	}

	job, err := s.jobs.Start(userID, missionID, text, strict)
	if err != nil {
		return nil, err
	}

	go s.runRespondJob(context.WithoutCancel(ctx), job.ID)
	return &quest.StartRespondJobResponse{JobID: job.ID}, nil
}

func (s *Dialog) SnapshotByMission(userID, missionID string) *quest.RespondJobStatusResponse {
	return s.jobs.SnapshotByMission(userID, missionID)
}

func (s *Dialog) runRespondJob(ctx context.Context, jobID string) {
	job := s.jobs.Get(jobID)
	if job == nil {
		return
	}

	userID := job.UserID
	mission, err := s.missions.GetByID(ctx, userID, job.MissionID)
	if err != nil {
		s.jobs.Fail(jobID, err)
		return
	}
	ensureMissionVoices(mission)
	mission.EnsureNPCStates()
	if mission.IsComplete {
		s.jobs.Fail(jobID, fmt.Errorf("mission already complete"))
		return
	}
	historyBase := len(mission.History)

	profile, err := s.profiles.Get(ctx, userID)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			s.jobs.Fail(jobID, err)
			return
		}
		profile = quest.NewDefaultProfile()
	}
	applyAccountIdentity(ctx, s.accounts, userID, &profile)
	promptProfile := profile

	s.jobs.SetStep(jobID, "grammar")

	// Grammar check and world reply start in parallel: the world does not depend on
	// the check result, and on a grammar error its call is cancelled.
	// The world draft is streamed to the player, but only after grammar
	// returns OK: a turn with errors is rejected and nobody sees its draft.
	worldCtx, cancelWorld := context.WithCancel(ctx)
	defer cancelWorld()
	var grammarPassed atomic.Bool
	var pendingPartial atomic.Pointer[quest.PartialWorld]
	publishPartial := func(partial *quest.PartialWorld) {
		if grammarPassed.Load() {
			s.jobs.SetPartial(jobID, partial)
		} else {
			pendingPartial.Store(partial)
		}
	}
	type worldOut struct {
		world *quest.WorldResult
		err   error
	}
	worldCh := make(chan worldOut, 1)
	go func() {
		world, err := s.llm.GenerateWorldResponseStream(worldCtx, mission, job.Text, &promptProfile, publishPartial)
		worldCh <- worldOut{world: world, err: err}
	}()

	grammar, llmErr := s.llm.CheckGrammar(ctx, job.Text, mission.Language, job.Strict)
	if llmErr != nil {
		s.jobs.Fail(jobID, llmErr)
		return
	}
	s.jobs.SetGrammar(jobID, grammar)
	if !grammar.OK {
		s.jobs.Complete(jobID, &quest.RespondJobResult{
			GrammarOK:    false,
			Errors:       grammar.Errors,
			CurrentStage: mission.CurrentStage,
			TotalStages:  mission.TotalStages,
		})
		return
	}

	// --- Phase 1: World Response ---
	s.jobs.SetStep(jobID, "world")
	grammarPassed.Store(true)
	if partial := pendingPartial.Load(); partial != nil {
		s.jobs.SetPartial(jobID, partial)
	}
	wo := <-worldCh
	if wo.err != nil {
		s.jobs.Fail(jobID, wo.err)
		return
	}
	world := wo.world
	applyWorldVoices(mission, world)

	wordCount := len(strings.Fields(job.Text))
	mission.History = append(mission.History, quest.DialogueTurn{
		Scene:   mission.CurrentStage,
		Speaker: "player",
		Text:    job.Text,
	})
	if strings.TrimSpace(world.Narration) != "" {
		mission.History = append(mission.History, quest.DialogueTurn{
			Scene:   mission.CurrentStage,
			Speaker: "narrator",
			Voice:   world.NarrationVoice,
			Text:    world.Narration,
		})
	}
	for _, line := range world.Responses {
		if strings.TrimSpace(line.Name) == "" || strings.TrimSpace(line.Text) == "" {
			continue
		}
		mission.History = append(mission.History, quest.DialogueTurn{
			Scene:   mission.CurrentStage,
			Speaker: line.Name,
			Voice:   line.Voice,
			Text:    line.Text,
		})
	}

	profile.TotalWordsWritten += wordCount

	// --- Phase 2: Story Evaluator ---
	s.jobs.SetStep(jobID, "evaluating")

	var eval *quest.EvaluationResult
	if mission.HasPlotPoints() {
		eval, err = s.llm.EvaluateStoryState(ctx, mission, world, job.Text)
		if err != nil {
			s.logger.Warn("quest: story evaluator failed, continuing without", slog.String("err", err.Error()))
			eval = &quest.EvaluationResult{
				SceneState:        quest.SceneActive,
				PlayerIntent:      quest.IntentExploring,
				NarrativeMomentum: "medium",
			}
		}
	} else {
		eval = s.legacyEvaluation(world)
	}

	for _, ppID := range eval.DeliveredPlotPoints {
		mission.DeliverPlotPoint(ppID, mission.CurrentStage)
	}

	mission.EnsureNPCStates()
	for name, delta := range eval.TrustChanges {
		if delta == 0 {
			continue
		}
		mission.ApplyTrustChange(name, delta)
	}
	for _, f := range eval.PlayerLearnedFromNPCs {
		mission.RecordPlayerLearnedFromNPC(f.NPC, f.Fact)
	}
	for _, f := range eval.NPCsLearnedAboutPlayer {
		mission.RecordNPCLearnedAboutPlayer(f.NPC, f.Fact)
	}

	result := &quest.RespondJobResult{
		GrammarOK:      true,
		Narration:      world.Narration,
		NarrationVoice: world.NarrationVoice,
		Responses:      world.Responses,
		SceneState:     eval.SceneState,
		PlayerIntent:   eval.PlayerIntent,
		CurrentStage:   mission.CurrentStage,
		TotalStages:    mission.TotalStages,
	}

	shouldAdvance := s.shouldAdvanceScene(mission, eval, world)

	if shouldAdvance {
		nextStage := mission.CurrentStage + 1
		resolutionAttempt := mission.HasPlotPoints() && eval.PlayerIntent == quest.IntentAttemptingResolution && mission.CanAttemptResolution()
		inFinalScene := mission.CurrentScene != nil && mission.CurrentScene.IsFinal
		endTriggered := mission.IsReadyForFinale() || nextStage >= mission.MaxScenes() || resolutionAttempt
		// When the story is ready for the finale but the player has not played the resolution yet,
		// set a climactic final scene instead of an instant epilogue.
		playFinaleScene := endTriggered && !resolutionAttempt && mission.HasPlotPoints() && !inFinalScene && nextStage < mission.MaxScenes()

		if endTriggered && !playFinaleScene {
			mission.IsComplete = true
			mission.Outcome = mission.DetermineOutcome()
			if eval.PlayerIntent == quest.IntentAttemptingResolution && !mission.AllRequiredDelivered() {
				if mission.RequiredPlotPointsRemaining() > mission.DeliveredCount() {
					mission.Outcome = quest.OutcomeFailed
				} else {
					mission.Outcome = quest.OutcomePartial
				}
			}

			epilogue, epilogueErr := s.llm.GenerateEpilogue(ctx, mission, mission.Outcome, world, &promptProfile)
			if epilogueErr != nil {
				s.logger.Warn("quest: epilogue generation failed", slog.String("mission", mission.ID), slog.String("err", epilogueErr.Error()))
				epilogue = ""
			}

			mission.CurrentStage = nextStage
			result.IsComplete = true
			result.SceneAdvanced = true
			result.Outcome = mission.Outcome
			result.CurrentStage = mission.CurrentStage

			if epilogue != "" {
				result.Epilogue = epilogue
				mission.History = append(mission.History, quest.DialogueTurn{
					Scene:   mission.CurrentStage,
					Speaker: "narrator",
					Voice:   mission.NarratorVoice,
					Text:    epilogue,
				})
			}
			mission.History = append(mission.History, quest.DialogueTurn{
				Scene:   mission.CurrentStage,
				Speaker: "system",
				Text:    "- Mission complete -",
			})
		} else {
			mission.History = append(mission.History, quest.DialogueTurn{
				Scene:   mission.CurrentStage,
				Speaker: "system",
				Text:    "- Scene completed -",
			})

			nextSceneCtx := s.decideSceneContext(mission, eval, world, job.Text)
			if playFinaleScene {
				nextSceneCtx.Finale = true
				nextSceneCtx.Flavor = quest.FlavorPlotBeat
			}

			if err := s.profiles.Save(ctx, userID, profile); err != nil {
				s.jobs.Fail(jobID, err)
				return
			}
			if err := s.persistRespondResult(ctx, userID, mission, historyBase, eval); err != nil {
				s.jobs.Fail(jobID, err)
				return
			}

			result.SceneAdvanced = true
			// Tell the client the target scene: while the mission has not caught up,
			// the UI shows "scene is generating".
			result.CurrentStage = nextStage
			s.jobs.Complete(jobID, result)

			s.maybeSummarizeAsync(userID, job.MissionID)

			s.jobs.SetSceneGenerating(userID, job.MissionID, true)
			promptProfileCopy := promptProfile
			go s.generateNextSceneAsync(context.WithoutCancel(ctx), userID, job.MissionID, nextStage, &promptProfileCopy, jobID, nextSceneCtx)
			return
		}
	}

	if err := s.profiles.Save(ctx, userID, profile); err != nil {
		s.jobs.Fail(jobID, err)
		return
	}
	if err := s.persistRespondResult(ctx, userID, mission, historyBase, eval); err != nil {
		s.jobs.Fail(jobID, err)
		return
	}

	s.jobs.Complete(jobID, result)

	s.maybeSummarizeAsync(userID, mission.ID)
}

// persistRespondResult applies turn changes onto a fresh mission version inside
// one locking transaction: parallel writes (images, summarization)
// are not lost by saving a stale snapshot.
func (s *Dialog) persistRespondResult(ctx context.Context, userID string, snapshot *quest.CustomMission, historyBase int, eval *quest.EvaluationResult) error {
	appended := snapshot.History[historyBase:]
	return s.missions.Update(ctx, userID, snapshot.ID, func(fresh *quest.CustomMission) error {
		ensureMissionVoices(fresh)
		fresh.EnsureNPCStates()
		fresh.History = append(fresh.History, appended...)
		for _, ppID := range eval.DeliveredPlotPoints {
			fresh.DeliverPlotPoint(ppID, fresh.CurrentStage)
		}
		for name, delta := range eval.TrustChanges {
			if delta != 0 {
				fresh.ApplyTrustChange(name, delta)
			}
		}
		for _, f := range eval.PlayerLearnedFromNPCs {
			fresh.RecordPlayerLearnedFromNPC(f.NPC, f.Fact)
		}
		for _, f := range eval.NPCsLearnedAboutPlayer {
			fresh.RecordNPCLearnedAboutPlayer(f.NPC, f.Fact)
		}
		if summary := strings.TrimSpace(eval.SceneSummary); summary != "" && fresh.CurrentScene != nil {
			fresh.CurrentScene.Summary = summary
			for i := range fresh.Scenes {
				if fresh.Scenes[i].Stage == fresh.CurrentScene.Stage {
					fresh.Scenes[i].Summary = summary
				}
			}
		}
		fresh.IsComplete = snapshot.IsComplete
		fresh.Outcome = snapshot.Outcome
		fresh.CurrentStage = snapshot.CurrentStage
		return nil
	})
}

func (s *Dialog) shouldAdvanceScene(mission *quest.CustomMission, eval *quest.EvaluationResult, world *quest.WorldResult) bool {
	if !mission.HasPlotPoints() {
		return world.TriggerMet
	}

	if world.Transition != nil && strings.TrimSpace(world.Transition.Type) != "" {
		return true
	}

	switch eval.SceneState {
	case quest.SceneResolved, quest.SceneTransitioning:
		return true
	case quest.SceneWindingDown:
		turnsInScene := countPlayerTurnsInScene(mission.History, mission.CurrentStage)
		return turnsInScene >= 3 && eval.NarrativeMomentum == "low"
	}

	if eval.PlayerIntent == quest.IntentDeparting {
		return true
	}

	if eval.PlayerIntent == quest.IntentAttemptingResolution {
		// In the final scene the confrontation must have room to unfold: the first
		// accusation gets a world reply; resolution comes on the second turn
		// or when the evaluator considers the scene resolved.
		if mission.CurrentScene != nil && mission.CurrentScene.IsFinal {
			turnsInScene := countPlayerTurnsInScene(mission.History, mission.CurrentStage)
			if turnsInScene < 2 {
				return false
			}
		}
		return mission.CanAttemptResolution()
	}

	turnsInScene := countPlayerTurnsInScene(mission.History, mission.CurrentStage)
	if turnsInScene >= 5 && eval.NarrativeMomentum != "high" {
		return true
	}
	if turnsInScene >= 4 && eval.NarrativeMomentum == "low" {
		return true
	}

	return false
}

func (s *Dialog) decideSceneContext(mission *quest.CustomMission, eval *quest.EvaluationResult, world *quest.WorldResult, playerText string) *quest.SceneContext {
	ctx := &quest.SceneContext{
		Flavor:         quest.FlavorPlotBeat,
		LastPlayerText: strings.TrimSpace(playerText),
	}
	if eval != nil {
		ctx.LastPlayerIntent = eval.PlayerIntent
	}
	if world != nil && world.Transition != nil {
		ctx.TransitionType = world.Transition.Type
		ctx.TransitionDetail = world.Transition.Detail
	}

	recentNonPlot := 0
	for i := len(mission.Scenes) - 1; i >= 0 && i >= len(mission.Scenes)-2; i-- {
		flav := mission.Scenes[i].Flavor
		if flav == quest.FlavorBreather || flav == quest.FlavorDetour {
			recentNonPlot++
			continue
		}
		break
	}

	intent := ""
	if eval != nil {
		intent = eval.PlayerIntent
	}
	transitionType := ""
	if world != nil && world.Transition != nil {
		transitionType = strings.ToLower(strings.TrimSpace(world.Transition.Type))
	}

	if mission.IsReadyForFinale() {
		ctx.Flavor = quest.FlavorPlotBeat
		return ctx
	}

	switch {
	case intent == quest.IntentDeparting || transitionType == "player_leaves":
		ctx.Flavor = quest.FlavorDetour
	case intent == quest.IntentOffTopic:
		ctx.Flavor = quest.FlavorBreather
	case recentNonPlot >= 2:
		ctx.Flavor = quest.FlavorChanceEncounter
	default:
		ctx.Flavor = quest.FlavorPlotBeat
	}

	return ctx
}

func countPlayerTurnsInScene(history []quest.DialogueTurn, stage int) int {
	count := 0
	for _, turn := range history {
		if turn.Scene == stage && turn.Speaker == "player" {
			count++
		}
	}
	return count
}

func (s *Dialog) legacyEvaluation(world *quest.WorldResult) *quest.EvaluationResult {
	state := quest.SceneActive
	if world.TriggerMet {
		state = quest.SceneResolved
	}
	return &quest.EvaluationResult{
		SceneState:        state,
		PlayerIntent:      quest.IntentExploring,
		NarrativeMomentum: "medium",
	}
}

func (s *Dialog) maybeSummarizeAsync(userID, missionID string) {
	go func() {
		ctx := context.Background()
		mission, err := s.missions.GetByID(ctx, userID, missionID)
		if err != nil {
			return
		}
		if !quest.NeedsSummarization(mission) {
			return
		}
		turns := quest.TurnsToSummarize(mission)
		if len(turns) == 0 {
			return
		}

		summary, err := s.llm.SummarizeHistory(ctx, mission, turns)
		if err != nil {
			s.logger.Warn("quest: history summarization failed", slog.String("mission", missionID), slog.String("err", err.Error()))
			return
		}

		cutoff := mission.SummarizedUpToTurn + len(turns)
		err = s.missions.Update(ctx, userID, missionID, func(fresh *quest.CustomMission) error {
			// History may have changed while summarization ran — trim only the summarized prefix.
			if len(fresh.History) < cutoff {
				return fmt.Errorf("history changed during summarization")
			}
			fresh.HistorySummary = summary
			fresh.History = fresh.History[cutoff:]
			fresh.SummarizedUpToTurn = 0
			return nil
		})
		if err != nil {
			s.logger.Warn("quest: save after summarization failed", slog.String("mission", missionID), slog.String("err", err.Error()))
		}
	}()
}

func (s *Dialog) generateNextSceneAsync(ctx context.Context, userID, missionID string, nextStage int, profile *quest.PlayerProfile, parentJobID string, sceneCtx *quest.SceneContext) {
	defer s.jobs.SetSceneGenerating(userID, missionID, false)

	mission, err := s.missions.GetByID(ctx, userID, missionID)
	if err != nil {
		s.logger.Warn("quest: async scene load failed", slog.String("job", parentJobID), slog.String("err", err.Error()))
		return
	}
	ensureMissionVoices(mission)
	mission.EnsureNPCStates()

	nextScene, err := s.llm.GenerateScene(ctx, mission, profile, nextStage, sceneCtx)
	if err != nil {
		s.logger.Warn("quest: async next scene generation failed, retrying", slog.String("job", parentJobID), slog.Int("stage", nextStage), slog.String("err", err.Error()))
		nextScene, err = s.llm.GenerateScene(ctx, mission, profile, nextStage, sceneCtx)
	}
	if err != nil {
		s.logger.Error("quest: async next scene generation failed", slog.String("job", parentJobID), slog.Int("stage", nextStage), slog.String("err", err.Error()))
		s.recordSceneFailure(ctx, userID, missionID)
		return
	}
	applySceneVoices(mission, nextScene)

	err = s.missions.Update(ctx, userID, missionID, func(fresh *quest.CustomMission) error {
		if fresh.CurrentStage >= nextStage || fresh.IsComplete {
			return fmt.Errorf("mission advanced elsewhere")
		}
		fresh.CurrentStage = nextStage
		fresh.CurrentScene = nextScene
		fresh.Scenes = append(fresh.Scenes, *nextScene)
		fresh.History = append(
			fresh.History,
			quest.DialogueTurn{Scene: nextStage, Speaker: "system", Text: fmt.Sprintf("- Scene %d -", nextStage+1)},
			quest.DialogueTurn{Scene: nextStage, Speaker: "narrator", Voice: nextScene.NarrationVoice, Text: nextScene.Narration},
		)
		for _, character := range nextScene.Present {
			if strings.TrimSpace(character.Name) == "" || strings.TrimSpace(character.Dialogue) == "" {
				continue
			}
			fresh.History = append(fresh.History, quest.DialogueTurn{
				Scene:   nextStage,
				Speaker: character.Name,
				Voice:   character.Voice,
				Text:    character.Dialogue,
			})
		}
		if s.images != nil && s.images.IsAvailable() {
			if fresh.SceneImageStatus == nil {
				fresh.SceneImageStatus = map[string]string{}
			}
			fresh.SceneImageStatus[fmt.Sprintf("%d", nextStage)] = "generating"
		}
		return nil
	})
	if err != nil {
		s.logger.Warn("quest: async next scene save failed", slog.String("job", parentJobID), slog.String("err", err.Error()))
		return
	}

	if s.images != nil && s.images.IsAvailable() {
		s.images.GenerateSceneAsync(userID, missionID, nextStage)
	}
}

// recordSceneFailure leaves a visible mark in history so the player understands
// that the scene failed to generate; the next player reply will restart generation.
func (s *Dialog) recordSceneFailure(ctx context.Context, userID, missionID string) {
	err := s.missions.Update(ctx, userID, missionID, func(fresh *quest.CustomMission) error {
		fresh.History = append(fresh.History, quest.DialogueTurn{
			Scene:   fresh.CurrentStage,
			Speaker: "system",
			Text:    "- Scene generation failed, reply to continue -",
		})
		return nil
	})
	if err != nil {
		s.logger.Error("quest: record scene failure failed", slog.String("mission", missionID), slog.String("err", err.Error()))
	}
}
