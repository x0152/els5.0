package quest

import (
	"encoding/json"
	"strings"
)

type PlayerProfile struct {
	Streak            int                   `json:"streak"`
	LastPlayedDate    string                `json:"lastPlayedDate"`
	TotalWordsWritten int                   `json:"totalWordsWritten"`
	Skills            []PlayerSkill         `json:"skills"`
	ErrorPatterns     map[string]int        `json:"errorPatterns"`
	SuccessPatterns   map[string]int        `json:"successPatterns"`
	CompletedMissions []string              `json:"completedMissions"`
	MissionLog        []MissionLogEntry     `json:"missionLog"`
	Codex             map[string]CodexEntry `json:"codex"`
	FirstName         string                `json:"firstName"`
	LastName          string                `json:"lastName"`
	EnglishLevel      string                `json:"englishLevel"`
	AboutMe           string                `json:"aboutMe"`
}

type PlayerSkill struct {
	Name      string `json:"name"`
	Category  string `json:"category"`
	Level     int    `json:"level"`
	TimesUsed int    `json:"timesUsed"`
	LastUsed  string `json:"lastUsed"`
}

type CodexEntry struct {
	Word        string   `json:"word"`
	Translation string   `json:"translation"`
	Example     string   `json:"example"`
	Status      string   `json:"status"`
	UsageCount  int      `json:"usageCount"`
	MissionIDs  []string `json:"missionIds"`
}

type MissionLogEntry struct {
	MissionID  string   `json:"missionId"`
	Title      string   `json:"title"`
	Date       string   `json:"date"`
	SkillsUsed []string `json:"skillsUsed"`
	Summary    string   `json:"summary"`
}

// --- Plot Points & Resolution ---

type PlotPoint struct {
	ID          string `json:"id"`
	Fact        string `json:"fact"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Delivered   bool   `json:"delivered"`
	DeliveredAt int    `json:"deliveredAt,omitempty"`
}

func (pp *PlotPoint) NormalizeFact() {
	if pp.Fact == "" && pp.Description != "" {
		pp.Fact = pp.Description
	}
	if pp.Description == "" && pp.Fact != "" {
		pp.Description = pp.Fact
	}
}

const (
	ResolutionMystery       = "mystery"
	ResolutionChoice        = "choice"
	ResolutionConfrontation = "confrontation"
	ResolutionRelationship  = "relationship"
	ResolutionEscape        = "escape"
	ResolutionNegotiation   = "negotiation"
	ResolutionDiscovery     = "discovery"
)

type ResolutionOutcome struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

type Resolution struct {
	Type     string              `json:"type"`
	Goal     string              `json:"goal"`
	Outcomes []ResolutionOutcome `json:"outcomes"`
}

// --- Scene States & Player Intents ---

const (
	SceneActive        = "active"
	SceneBuilding      = "building"
	SceneWindingDown   = "winding_down"
	SceneTransitioning = "transitioning"
	SceneResolved      = "resolved"
)

const (
	IntentInvestigating        = "investigating"
	IntentSocializing          = "socializing"
	IntentExploring            = "exploring"
	IntentConfronting          = "confronting"
	IntentAttemptingResolution = "attempting_resolution"
	IntentDeparting            = "departing"
	IntentOffTopic             = "off_topic"
)

// --- Mission Outcomes ---

const (
	OutcomePerfect    = "perfect"
	OutcomeGood       = "good"
	OutcomePartial    = "partial"
	OutcomeUnexpected = "unexpected"
	OutcomeFailed     = "failed"
	OutcomeAbandoned  = "abandoned"
)

const (
	GenerationStatusGenerating = "generating"
	GenerationStatusReady      = "ready"
	GenerationStatusError      = "error"
)

const (
	GenerationStepCreating   = "creating"
	GenerationStepFirstScene = "first_scene"
)

// --- Core Domain ---

type CustomMission struct {
	ID                          string            `json:"id"`
	Title                       string            `json:"title"`
	Description                 string            `json:"description"`
	UserPrompt                  string            `json:"userPrompt"`
	Genre                       string            `json:"genre"`
	Language                    string            `json:"language"`
	PracticeGoals               string            `json:"practiceGoals"`
	SecretEnding                string            `json:"secretEnding"`
	NarratorVoice               string            `json:"narratorVoice,omitempty"`
	Characters                  []Character       `json:"characters"`
	PlotPoints                  []PlotPoint       `json:"plotPoints,omitempty"`
	Resolution                  *Resolution       `json:"resolution,omitempty"`
	CurrentStage                int               `json:"currentStage"`
	TotalStages                 int               `json:"totalStages"`
	EstimatedScenes             int               `json:"estimatedScenes,omitempty"`
	IsComplete                  bool              `json:"isComplete"`
	Outcome                     string            `json:"outcome,omitempty"`
	CurrentScene                *DynamicScene     `json:"currentScene"`
	Scenes                      []DynamicScene    `json:"scenes"`
	History                     []DialogueTurn    `json:"history"`
	SkillsEarned                []SkillReward     `json:"skillsEarned"`
	SkillSignals                map[string]int    `json:"skillSignals,omitempty"`
	SkillCategories             map[string]string `json:"skillCategories,omitempty"`
	CreatedAt                   string            `json:"createdAt"`
	PlayerAvatarImage           string            `json:"playerAvatarImage,omitempty"`
	CoverImage                  string            `json:"coverImage,omitempty"`
	SceneImages                 map[string]string `json:"sceneImages,omitempty"`
	CharacterAvatars            map[string]string `json:"characterAvatars,omitempty"`
	CoverImageStatus            string            `json:"coverImageStatus,omitempty"`
	SceneImageStatus            map[string]string `json:"sceneImageStatus,omitempty"`
	CharacterAvatarStatus       map[string]string `json:"characterAvatarStatus,omitempty"`
	CoverImageError             string            `json:"coverImageError,omitempty"`
	SceneImageErrors            map[string]string `json:"sceneImageErrors,omitempty"`
	CharacterAvatarErrors       map[string]string `json:"characterAvatarErrors,omitempty"`
	CoverImageGenStartedAt      string            `json:"coverImageGenStartedAt,omitempty"`
	SceneImageGenStartedAt      map[string]string `json:"sceneImageGenStartedAt,omitempty"`
	CharacterAvatarGenStartedAt map[string]string `json:"characterAvatarGenStartedAt,omitempty"`

	GenerationStatus string `json:"generationStatus,omitempty"`
	GenerationStep   string `json:"generationStep,omitempty"`
	GenerationError  string `json:"generationError,omitempty"`

	HistorySummary     string `json:"historySummary,omitempty"`
	SummarizedUpToTurn int    `json:"summarizedUpToTurn,omitempty"`

	NPCStates map[string]*NPCState `json:"npcStates,omitempty"`

	TotalXP int `json:"totalXp,omitempty"`

	// Epoch increments on every reset; in-flight async scene generation from a
	// previous playthrough must not write into the new one.
	Epoch int `json:"epoch,omitempty"`
}

type Character struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Gender       string `json:"gender,omitempty"`
	Age          string `json:"age,omitempty"`
	Voice        string `json:"voice,omitempty"`
	Personality  string `json:"personality"`
	SpeechStyle  string `json:"speechStyle"`
	Appearance   string `json:"appearance"`
	Motivation   string `json:"motivation"`
	Arc          string `json:"arc"`
	InitialTrust int    `json:"initialTrust,omitempty"`
}

const (
	TrustMin = -3
	TrustMax = 3
)

type NPCState struct {
	Trust            int      `json:"trust"`
	KnowsAboutPlayer []string `json:"knowsAboutPlayer,omitempty"`
	PlayerKnowsAbout []string `json:"playerKnowsAbout,omitempty"`
}

type NPCFactUpdate struct {
	NPC  string `json:"npc"`
	Fact string `json:"fact"`
}

type DynamicScene struct {
	Narration      string           `json:"narration"`
	NarrationVoice string           `json:"narrationVoice,omitempty"`
	Present        []SceneCharacter `json:"present"`
	Objects        []string         `json:"objects"`
	ScenePurpose   string           `json:"scenePurpose,omitempty"`
	Tips           []LanguageTip    `json:"tips"`
	Stage          int              `json:"stage"`
	IsFinal        bool             `json:"isFinal"`
	Flavor         string           `json:"flavor,omitempty"`
	Summary        string           `json:"summary,omitempty"`

	Trigger string `json:"trigger,omitempty"`
}

const (
	FlavorPlotBeat        = "plot_beat"
	FlavorBreather        = "breather"
	FlavorDetour          = "detour"
	FlavorChanceEncounter = "chance_encounter"
)

type SceneContext struct {
	Flavor           string
	LastPlayerText   string
	LastPlayerIntent string
	TransitionType   string
	TransitionDetail string
	Finale           bool
}

type SceneCharacter struct {
	Name     string `json:"name"`
	Voice    string `json:"voice,omitempty"`
	Dialogue string `json:"dialogue"`
}

type LanguageTip struct {
	Construction string `json:"construction"`
	Tip          string `json:"tip"`
	Example      string `json:"example"`
	Explanation  string `json:"explanation"`
	Category     string `json:"category"`
}

type DialogueTurn struct {
	Scene   int    `json:"scene"`
	Speaker string `json:"speaker"`
	Voice   string `json:"voice,omitempty"`
	Text    string `json:"text"`
}

// --- World Response (pure roleplay, no game logic) ---

type WorldResult struct {
	Narration      string          `json:"narration"`
	NarrationVoice string          `json:"narrationVoice,omitempty"`
	Responses      []CharacterLine `json:"responses"`
	Transition     *Transition     `json:"transition,omitempty"`

	TriggerMet bool        `json:"triggerMet,omitempty"`
	XPEarned   int         `json:"xp,omitempty"`
	Skills     FlexStrings `json:"skills,omitempty"`
}

type Transition struct {
	Type   string `json:"type"`
	Detail string `json:"detail,omitempty"`
}

// PartialWorld is a draft world reply built from an incomplete LLM stream.
// Shown to the player until generation finishes; the source of truth
// remains the final WorldResult. Done flags distinguish completed values
// from ones cut mid-stream — the UI uses them for generation state.
type PartialWorld struct {
	Narration     string        `json:"narration,omitempty"`
	NarrationDone bool          `json:"narrationDone,omitempty"`
	Responses     []PartialLine `json:"responses,omitempty"`
}

type PartialLine struct {
	Name string `json:"name"`
	Text string `json:"text,omitempty"`
	Done bool   `json:"done,omitempty"`
}

// --- Story Evaluator Result ---

type EvaluationResult struct {
	DeliveredPlotPoints    []string        `json:"deliveredPlotPoints"`
	SceneState             string          `json:"sceneState"`
	PlayerIntent           string          `json:"playerIntent"`
	NarrativeMomentum      string          `json:"narrativeMomentum"`
	SceneSummary           string          `json:"sceneSummary,omitempty"`
	TransitionReason       string          `json:"transitionReason,omitempty"`
	TrustChanges           map[string]int  `json:"trustChanges,omitempty"`
	PlayerLearnedFromNPCs  []NPCFactUpdate `json:"playerLearnedFromNPCs,omitempty"`
	NPCsLearnedAboutPlayer []NPCFactUpdate `json:"npcsLearnedAboutPlayer,omitempty"`
}

// --- API Request / Response ---

type CreateMissionRequest struct {
	Prompt        string `json:"prompt"`
	Genre         string `json:"genre"`
	Language      string `json:"language"`
	PracticeGoals string `json:"practiceGoals"`
}

type RespondRequest struct {
	Text   string `json:"text"`
	Strict *bool  `json:"strict,omitempty"`
}

type StartRespondJobResponse struct {
	JobID string `json:"jobId"`
}

type RespondJobStatusResponse struct {
	JobID     string            `json:"jobId"`
	Status    string            `json:"status"`
	Step      string            `json:"step"`
	InputText string            `json:"inputText,omitempty"`
	Grammar   *GrammarCheck     `json:"grammar,omitempty"`
	Partial   *PartialWorld     `json:"partial,omitempty"`
	Result    *RespondJobResult `json:"result,omitempty"`
	Error     string            `json:"error,omitempty"`
}

type GrammarCheck struct {
	OK     bool           `json:"ok"`
	Errors []GrammarError `json:"errors"`
}

type RespondJobResult struct {
	GrammarOK      bool            `json:"grammarOk"`
	Errors         []GrammarError  `json:"errors"`
	Narration      string          `json:"narration"`
	NarrationVoice string          `json:"narrationVoice,omitempty"`
	Responses      []CharacterLine `json:"responses"`
	SceneAdvanced  bool            `json:"sceneAdvanced"`
	NextScene      *DynamicScene   `json:"nextScene,omitempty"`
	IsComplete     bool            `json:"isComplete"`
	Outcome        string          `json:"outcome,omitempty"`
	Epilogue       string          `json:"epilogue,omitempty"`
	SceneState     string          `json:"sceneState,omitempty"`
	PlayerIntent   string          `json:"playerIntent,omitempty"`
	CurrentStage   int             `json:"currentStage"`
	TotalStages    int             `json:"totalStages"`
}

type CharacterLine struct {
	Name  string `json:"name"`
	Voice string `json:"voice,omitempty"`
	Text  string `json:"text"`
}

type GrammarError struct {
	Original    string `json:"original"`
	Correction  string `json:"correction"`
	Explanation string `json:"explanation"`
	Type        string `json:"type"`
}

type FlexStrings []string

func (f *FlexStrings) UnmarshalJSON(data []byte) error {
	var strs []string
	if err := json.Unmarshal(data, &strs); err == nil {
		*f = strs
		return nil
	}

	var objs []map[string]any
	if err := json.Unmarshal(data, &objs); err == nil {
		result := make([]string, 0, len(objs))
		for _, obj := range objs {
			for _, key := range []string{"construction", "name", "skill"} {
				if v, ok := obj[key]; ok {
					if s, ok := v.(string); ok && s != "" {
						result = append(result, s)
						break
					}
				}
			}
		}
		*f = result
		return nil
	}

	*f = nil
	return nil
}

type SkillReward struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	IsNew    bool   `json:"isNew"`
}

type TranslateRequest struct {
	Text string `json:"text"`
}

type TranslateResponse struct {
	Translation string `json:"translation"`
}

type SuggestionResponse struct {
	Title       string   `json:"title"`
	Templates   []string `json:"templates,omitempty"`
	Chunks      []string `json:"chunks,omitempty"`
	Words       []string `json:"words,omitempty"`
	Tip         string   `json:"tip"`
	Explanation string   `json:"explanation,omitempty"`
}

func NewDefaultProfile() PlayerProfile {
	return PlayerProfile{
		Skills:            []PlayerSkill{},
		ErrorPatterns:     map[string]int{},
		SuccessPatterns:   map[string]int{},
		CompletedMissions: []string{},
		MissionLog:        []MissionLogEntry{},
		Codex:             map[string]CodexEntry{},
	}
}

// --- Plot Point helpers ---

func (m *CustomMission) HasPlotPoints() bool {
	return len(m.PlotPoints) > 0
}

func (m *CustomMission) DeliverPlotPoint(ppID string, stage int) bool {
	for i := range m.PlotPoints {
		if m.PlotPoints[i].ID == ppID && !m.PlotPoints[i].Delivered {
			m.PlotPoints[i].Delivered = true
			m.PlotPoints[i].DeliveredAt = stage
			return true
		}
	}
	return false
}

func (m *CustomMission) RequiredPlotPointsRemaining() int {
	count := 0
	for _, pp := range m.PlotPoints {
		if pp.Required && !pp.Delivered {
			count++
		}
	}
	return count
}

func (m *CustomMission) AllRequiredDelivered() bool {
	return m.RequiredPlotPointsRemaining() == 0
}

func (m *CustomMission) DeliveredCount() int {
	count := 0
	for _, pp := range m.PlotPoints {
		if pp.Delivered {
			count++
		}
	}
	return count
}

func (m *CustomMission) PlotPointSummary() (total, delivered, requiredTotal, requiredDelivered int) {
	for _, pp := range m.PlotPoints {
		total++
		if pp.Delivered {
			delivered++
		}
		if pp.Required {
			requiredTotal++
			if pp.Delivered {
				requiredDelivered++
			}
		}
	}
	return
}

func (m *CustomMission) DetermineOutcome() string {
	if !m.HasPlotPoints() {
		return OutcomeGood
	}

	total, delivered, reqTotal, reqDelivered := m.PlotPointSummary()

	if reqTotal == 0 {
		if delivered == total {
			return OutcomePerfect
		}
		return OutcomeGood
	}

	ratio := float64(reqDelivered) / float64(reqTotal)
	switch {
	case ratio >= 1.0 && delivered == total:
		return OutcomePerfect
	case ratio >= 1.0:
		return OutcomeGood
	case ratio >= 0.5:
		return OutcomePartial
	default:
		return OutcomeFailed
	}
}

func (m *CustomMission) IsReadyForFinale() bool {
	if !m.HasPlotPoints() {
		return m.CurrentStage+1 >= m.EstimatedScenes || m.CurrentStage+1 >= m.TotalStages
	}
	return m.AllRequiredDelivered()
}

func (m *CustomMission) CanAttemptResolution() bool {
	if !m.HasPlotPoints() {
		return true
	}
	_, _, reqTotal, reqDelivered := m.PlotPointSummary()
	if reqTotal == 0 {
		return m.CurrentStage >= 1
	}
	if reqDelivered >= reqTotal {
		return true
	}
	minNeeded := (reqTotal + 1) / 2
	if reqDelivered < minNeeded {
		return false
	}
	return m.CurrentStage >= 1
}

func (m *CustomMission) MaxScenes() int {
	est := m.EstimatedScenes
	if est <= 0 {
		est = m.TotalStages
	}
	if est <= 0 {
		est = 8
	}
	return est + 3
}

const npcMemoryCap = 8

func (m *CustomMission) EnsureNPCStates() {
	if m.NPCStates == nil {
		m.NPCStates = map[string]*NPCState{}
	}
	for _, c := range m.Characters {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}
		if _, ok := m.NPCStates[name]; !ok {
			trust := c.InitialTrust
			if trust < TrustMin {
				trust = TrustMin
			}
			if trust > TrustMax {
				trust = TrustMax
			}
			m.NPCStates[name] = &NPCState{Trust: trust}
		}
	}
}

func (m *CustomMission) ApplyTrustChange(npcName string, delta int) (before, after int, ok bool) {
	name := strings.TrimSpace(npcName)
	if name == "" || delta == 0 {
		return 0, 0, false
	}
	m.EnsureNPCStates()
	state, exists := m.NPCStates[name]
	if !exists {
		return 0, 0, false
	}
	before = state.Trust
	state.Trust = clampInt(state.Trust+delta, TrustMin, TrustMax)
	return before, state.Trust, state.Trust != before
}

func (m *CustomMission) RecordNPCLearnedAboutPlayer(npcName, fact string) bool {
	state := m.npcStateIfKnown(npcName)
	if state == nil {
		return false
	}
	fact = strings.TrimSpace(fact)
	if fact == "" {
		return false
	}
	state.KnowsAboutPlayer = appendCappedUnique(state.KnowsAboutPlayer, fact, npcMemoryCap)
	return true
}

func (m *CustomMission) RecordPlayerLearnedFromNPC(npcName, fact string) bool {
	state := m.npcStateIfKnown(npcName)
	if state == nil {
		return false
	}
	fact = strings.TrimSpace(fact)
	if fact == "" {
		return false
	}
	state.PlayerKnowsAbout = appendCappedUnique(state.PlayerKnowsAbout, fact, npcMemoryCap)
	return true
}

func (m *CustomMission) npcStateIfKnown(npcName string) *NPCState {
	name := strings.TrimSpace(npcName)
	if name == "" || m.NPCStates == nil {
		return nil
	}
	return m.NPCStates[name]
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func appendCappedUnique(list []string, fact string, capacity int) []string {
	for _, existing := range list {
		if strings.EqualFold(existing, fact) {
			return list
		}
	}
	list = append(list, fact)
	if len(list) > capacity {
		list = list[len(list)-capacity:]
	}
	return list
}

func TrustTier(trust int) string {
	switch {
	case trust <= -3:
		return "hostile"
	case trust <= -1:
		return "wary"
	case trust == 0:
		return "neutral"
	case trust == 1:
		return "warming"
	case trust == 2:
		return "trusts you"
	default:
		return "close"
	}
}

func TrustSharingRule(trust int) string {
	switch {
	case trust <= -1:
		return "Deflects personal questions, gives only objective/public facts, may be curt or sarcastic."
	case trust == 0:
		return "Cordial but guarded. Shares small talk; deflects anything personal or sensitive."
	case trust == 1:
		return "Shares small personal things when asked; guards bigger secrets."
	case trust == 2:
		return "Opens up about deeper matters when asked. May volunteer small personal details."
	default:
		return "May volunteer important things unprompted when the moment feels right."
	}
}

// ForkForPlayer builds a personal copy of a shared mission: the template (plot,
// characters, cover, avatars) is kept; the author's runtime is cleared.
func (m *CustomMission) ForkForPlayer() *CustomMission {
	fork := *m
	fork.PlotPoints = append([]PlotPoint(nil), m.PlotPoints...)
	fork.Reset()
	fork.TotalXP = 0
	fork.PlayerAvatarImage = ""
	fork.SceneImages = map[string]string{}
	fork.SceneImageStatus = map[string]string{}
	fork.SceneImageErrors = map[string]string{}
	fork.SceneImageGenStartedAt = map[string]string{}
	if img := m.SceneImages["0"]; img != "" {
		fork.SceneImages["0"] = img
		fork.SceneImageStatus["0"] = m.SceneImageStatus["0"]
	}
	return &fork
}

func (m *CustomMission) Reset() {
	var firstScene *DynamicScene
	if len(m.Scenes) > 0 {
		sceneCopy := m.Scenes[0]
		firstScene = &sceneCopy
	}

	m.CurrentStage = 0
	m.Epoch++
	m.IsComplete = false
	m.Outcome = ""
	m.HistorySummary = ""
	m.SummarizedUpToTurn = 0
	m.SkillsEarned = nil
	m.SkillSignals = map[string]int{}
	m.NPCStates = nil
	m.EnsureNPCStates()

	for i := range m.PlotPoints {
		m.PlotPoints[i].Delivered = false
		m.PlotPoints[i].DeliveredAt = 0
	}

	m.CurrentScene = firstScene
	m.Scenes = []DynamicScene{}
	if firstScene != nil {
		m.Scenes = append(m.Scenes, *firstScene)
	}

	m.History = []DialogueTurn{}
	if firstScene != nil {
		m.History = append(m.History,
			DialogueTurn{Scene: 0, Speaker: "system", Text: "- Scene 1 -"},
			DialogueTurn{Scene: 0, Speaker: "narrator", Voice: firstScene.NarrationVoice, Text: firstScene.Narration},
		)
		for _, ch := range firstScene.Present {
			name := strings.TrimSpace(ch.Name)
			text := strings.TrimSpace(ch.Dialogue)
			if name == "" || text == "" {
				continue
			}
			m.History = append(m.History, DialogueTurn{
				Scene:   0,
				Speaker: name,
				Voice:   ch.Voice,
				Text:    text,
			})
		}
	}
}
