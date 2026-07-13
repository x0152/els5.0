package quest

import (
	"encoding/json"
	"fmt"
	"strings"
)

func BuildCreateMissionPrompts(
	prompt string,
	genre string,
	practiceGoals string,
	language string,
	profile PlayerProfile,
) (string, string) {
	if prompt == "" {
		prompt = "RANDOM - a realistic everyday situation with an interesting twist. No magic, no sci-fi, no supernatural."
	}
	if genre == "" {
		genre = "any"
	}
	if practiceGoals == "" {
		practiceGoals = "any"
	}
	if language == "" {
		language = "English"
	}

	system := `You create missions for WriteQuest, an interactive text RPG for language practice.

Create mission metadata only. Do not generate a scene now.

LANGUAGE: ALL generated content (titles, descriptions, character fields, every value) MUST be written in English only. The JSON keys are always in English.

Rules:
- The player is a real character in the story and should be referenced by name.
- 2-4 NPC characters with distinct voices (do NOT include the player in characters).
- "role" is PUBLIC-FACING and SPOILER-FREE: how a stranger would describe this person at story start (job, relation to the situation). NEVER hint at secrets, guilt, or hidden deeds in the role — those belong in motivation and plot points.
- Each character MUST feel different: contrasting temperaments, speech patterns, attitudes toward the player. Avoid polite sameness - give them strong opinions, flaws, humor, sarcasm, warmth, coldness, quirks.
- Give each character a realistic, varied age that fits their role. Mix young, middle-aged and older people; default to working-age adults (20s-40s) unless the story implies otherwise. Do NOT make everyone elderly.
- Characters are NOT ethically constrained. They may tease, mock, challenge, or joke about the player based on the player's "About" info or actions - if that fits their personality.
- Each character MUST have a motivation (what they want) and an arc (how their attitude or behavior changes across the story). NPCs should not feel static - by the final scene they should be noticeably different from scene 1.
- Each character has an "initialTrust" integer from -3..+3 reflecting their standing with the player at story start. Default is 0 (neutral stranger). Be CONSERVATIVE — most NPCs should start at -1, 0, or +1. Guidance:
    * +3: only for a deeply bonded person (lifelong partner, best friend, parent with a warm relationship) AND only when the player's backstory supports it.
    * +2: close friend, trusting colleague, loving family member with an unambiguously warm relationship.
    * +1: friendly acquaintance, welcoming shopkeeper, kind neighbor, a warm stranger who has a good first impression of the player.
    * 0: neutral stranger, professional contact, someone who doesn't know the player yet. This is the DEFAULT for most first-meeting NPCs.
    * -1: standoffish, cool, mildly distrustful, busy/irritable, a guarded stranger, a grumpy coworker, someone caught off-guard. USE THIS for most "difficult" NPCs; it already means "guarded and deflecting".
    * -2: actively dislikes the player, has a specific grudge, a suspicious gatekeeper who has explicit reason to distrust. Rare.
    * -3: open enmity, a declared antagonist, someone the player wronged in the backstory. Very rare — only if the mission setup explicitly establishes hostility.
  Do NOT use -2 or -3 just because an NPC is "grumpy" or "reserved" — those are -1. Reserve -2/-3 for clear cause. When in doubt, pick a value closer to 0.
- Mission theme comes from IDEA and GENRE only. The player's profession/hobbies from About can appear naturally in the story (mentioned in conversation, someone asks about their work, etc.) but do NOT make it the central plot device. Avoid naive connections like "player is a programmer -> player fixes a broken computer."
- NO MAGIC, SUPERNATURAL, SCI-FI, OR FANTASY ELEMENTS unless the player's IDEA or GENRE explicitly asks for them. By default everything is grounded in the real world.
- SCENES MUST BE GROUNDED AND REALISTIC by default. Think everyday life situations: running into someone at a cafe, a job interview, chatting with a friend, a work meeting, helping a neighbor. The player should feel like they walked into a real situation, not a game level.
- Secret ending is a twist the player works toward without knowing.
- Adapt EVERYTHING to the player's English level.
- Description must tell the player what they need to do or figure out, not just describe the setting.
- Assign TTS voices using ONLY this list: Bella (female), Jasper (male), Luna (female), Bruno (male), Rosie (female), Hugo (male), Kiki (female), Leo (male).
- narratorVoice must be one value from that list.
- Every character must have a voice field from that same list.

PLOT POINTS — the backbone of the story:
- Generate 4-8 plot points: facts, events, or revelations that together form the story.
- Mark each as required (true) or optional (false). 3-5 should be required.
- Plot points are NOT tasks for the player. They are FACTS that exist in the world and can be discovered naturally through conversation, observation, or exploration.
- Plot points should interconnect — discovering one should make others easier or more meaningful.
- For mystery/detective genres: plot points are CLUES (evidence, witness statements, alibis, physical evidence).
- For relationship/social genres: plot points are REVELATIONS about characters (secrets, feelings, history).
- For adventure/escape genres: plot points are DISCOVERIES (tools, exits, allies, weaknesses).

RESOLUTION — how the story ends:
- type: one of "mystery", "choice", "confrontation", "relationship", "escape", "negotiation", "discovery".
- goal: what the player needs to do to resolve the central conflict.
- outcomes: 4 possible endings ranked from best to worst. Each has a label and description. They MUST be meaningfully different — not just "you did well" vs "you did okay." Different endings should change what happens to the characters, what truth is revealed, and how the story concludes.

Return valid JSON only:
{
  "title": "2-5 words",
  "description": "one sentence hook: what happened and what the player needs to do",
  "secretEnding": "3-4 sentences twist",
  "narratorVoice": "Bella",
  "estimatedScenes": 6,
  "plotPoints": [
    {"id": "pp1", "fact": "specific fact or event", "required": true},
    {"id": "pp2", "fact": "specific fact or event", "required": true},
    {"id": "pp3", "fact": "specific fact or event", "required": false}
  ],
  "resolution": {
    "type": "mystery",
    "goal": "what player needs to achieve",
    "outcomes": [
      {"label": "perfect", "description": "best possible ending — 2-3 sentences"},
      {"label": "good", "description": "solid ending with minor gaps — 2-3 sentences"},
      {"label": "partial", "description": "incomplete resolution — 2-3 sentences"},
      {"label": "failed", "description": "wrong conclusion or missed the point — 2-3 sentences"}
    ]
  },
  "characters": [
    {
      "name": "full name",
      "role": "role in story",
      "gender": "male or female",
      "age": "approximate age in years, e.g. \"early 20s\", \"34\", \"late 50s\"",
      "voice": "Jasper",
      "personality": "specific traits, flaws, quirks",
      "speechStyle": "how they talk - distinct voice, catchphrases, rhythm",
      "appearance": "clothing, posture, face, notable physical details",
      "motivation": "what this character wants and why",
      "arc": "how their attitude/behavior changes from start to end",
      "initialTrust": 0
    }
  ]
}`

	user := fmt.Sprintf(`IDEA: %s
GENRE: %s
SKILLS TO PRACTICE: %s
LANGUAGE: %s

%s

%s`,
		prompt,
		genre,
		practiceGoals,
		language,
		LanguageLevelGuidance(playerLevel(&profile)),
		PlayerContext(profile),
	)

	return system, user
}

func BuildScenePrompts(
	mission *CustomMission,
	profile *PlayerProfile,
	stage int,
	sceneCtx *SceneContext,
) (string, string) {
	if stage < 0 {
		stage = mission.CurrentStage
	}

	lang := mission.Language
	if lang == "" {
		lang = "English"
	}

	pName := "player"
	pCtx := ""
	if profile != nil {
		pName = PlayerName(*profile)
		pCtx = PlayerContext(*profile)
	}

	storyPhase := "early"
	est := mission.EstimatedScenes
	if est <= 0 {
		est = mission.TotalStages
	}
	if est > 0 {
		ratio := float64(stage) / float64(est)
		switch {
		case ratio >= 0.85:
			storyPhase = "finale"
		case ratio >= 0.6:
			storyPhase = "approaching_finale"
		case ratio >= 0.3:
			storyPhase = "middle"
		}
	}
	if mission.IsReadyForFinale() {
		storyPhase = "approaching_finale"
	}
	if sceneCtx != nil && sceneCtx.Finale {
		storyPhase = "finale"
	}

	ppState := plotPointStateForPrompt(mission)

	specialNotes := ""
	if storyPhase == "finale" || storyPhase == "approaching_finale" {
		specialNotes = "STORY IS APPROACHING RESOLUTION. This scene should build toward the climax. Undelivered required plot points should surface naturally if possible. The dramatic tension should be rising."
		if storyPhase == "finale" {
			specialNotes = "FINAL SCENE — THE CLIMAX. Stage the confrontation or decisive moment the story has been building toward: put the player face to face with the heart of the conflict (the culprit, the decisive choice, the person they must face). Follow the player's stated intent from the transition context. Do NOT resolve the conflict yourself and do NOT reveal the secret ending — the player must enact the resolution through their own words and choices in this scene."
		}
	}

	flavor := FlavorPlotBeat
	if sceneCtx != nil && sceneCtx.Flavor != "" {
		flavor = sceneCtx.Flavor
	}
	if storyPhase == "finale" {
		flavor = FlavorPlotBeat
	}
	flavorBlock := flavorRulesForPrompt(flavor)
	contextBlock := sceneContextForPrompt(sceneCtx)
	levelBlock := LanguageLevelGuidance(playerLevel(profile))

	system := fmt.Sprintf(`Generate one scene for WriteQuest, an interactive language-practice RPG.

LANGUAGE: All narration, dialogue, objects and tips MUST be written in English only.

%s

Writing style:
- Scenes should feel like real life — describe naturally without dramatizing.
- Keep everything inside the target LANGUAGE LEVEL band above. When in doubt, go one step simpler.

Scene structure:
- narration: 2-4 sentences max. Set the place, show who's there, mention notable things.
- narrationVoice: choose from mission narratorVoice or: Bella, Jasper, Luna, Bruno, Rosie, Hugo, Kiki, Leo.
- NOT EVERY SCENE NEEDS CONFLICT OR TENSION. Many scenes are quiet, atmospheric, or character-driven. Let the story breathe.
- present: 0-3 NPCs. ONLY 1-2 should speak. Others mentioned as present but silent. A scene may have NO NPCs at all (solo moment: the player walks, thinks, notices something).
- Each present character must include a voice field matching their mission voice.
- NPC DIALOGUE LENGTH IS FLEXIBLE. Match what's natural for the character and moment.
- STRICT FIELD SEPARATION: "dialogue" contains ONLY the spoken words an NPC says out loud. All body language, gestures, facial expressions, silences, and physical actions belong in "narration". NEVER write stage directions inside "dialogue" (no "He looks up. 'Fine.'", no "She shrugs: 'Whatever.'", no parenthetical cues, no italics, no "quietly:" prefixes).
- objects: short names of interactive or notable objects in the scene.
- scenePurpose: a short internal note about the MOOD and SITUATION of this scene (2-3 sentences). Describe what it feels like and what the player might do here. Do NOT script which plot points should come out — that is the player's discovery.
- tips: 1-2 language tips. The "example" field MUST use a completely unrelated everyday scenario — NEVER reference the current scene.

SCENE VARIETY:
- Usually each new scene is in a DIFFERENT LOCATION from the previous scene. Occasionally the same location with a time skip also works (after a break, later that evening).
- Rotate NPCs. Not every character needs to be in every scene. 1-on-1 encounters are powerful. Solo scenes are valid too.
- The environment itself tells a story: messy desk, overheard phone call, note left behind, weather changing.

ANTI-COINCIDENCE — CRITICAL:
- The universe does NOT cooperate with the plot. NPCs cannot teleport to wherever the player goes.
- DO NOT place a plot-relevant NPC at the exact location the player just named as their escape, unless the NPC has a specific, believable reason to be there that was established earlier.
- DO NOT prop an NPC up with a conveniently plot-shaped object (an NPC holding the exact document the player needs, etc.).
- If the player chose this location on their own, treat that choice as REAL — populate the scene with ambient people, atmosphere, or just the player alone. The plot can return later, organically.

PLOT POINTS:
- You know which plot points have and haven't been delivered. Treat them as facts that EXIST in the world, not checklist items.
- NEVER script the scene around "how this plot point will come out". Let the player ask, explore, or ignore them.
- NPCs must NOT volunteer key information unprompted. Information emerges through conversation the PLAYER drives.

%s

STORY PHASE: %s

The story must stay continuous with previous scenes and history.

Return valid JSON only:
{
  "narration": "2-4 sentences, casual and clear. Put physical actions, body language and gestures here.",
  "narrationVoice": "Bella",
  "present": [{"name": "NPC", "voice": "Jasper", "dialogue": "SPOKEN WORDS ONLY — no actions, no stage directions"}],
  "objects": ["object 1", "object 2"],
  "scenePurpose": "internal note: mood and situation in this scene",
  "tips": [
    {
      "construction": "Pattern",
      "tip": "usage advice",
      "example": "example from a COMPLETELY DIFFERENT everyday topic",
      "explanation": "short explanation in English",
      "category": "tenses|conditionals|modals|linking|style|hypotheticals|questions|passive|reported_speech"
    }
  ],
  "isFinal": false
}`, levelBlock, flavorBlock, storyPhase)

	missionMeta := buildMissionMetaForScene(mission)

	user := fmt.Sprintf(`LANGUAGE: %s
MISSION: %s
MISSION_META:
%s
SCENE: %d (story phase: %s, flavor: %s)
SPECIAL: %s
%s

PLOT_POINTS_STATE:
%s

NPC_RELATIONSHIPS (respect these when placing characters and shaping dialogue — distrustful NPCs are still distrustful, NPCs remember what they know):
%s

SCENES_SO_FAR:
%s

FULL_HISTORY:
%s

SKILLS_TO_PRACTICE:
%s

%s`,
		lang,
		mission.Title,
		missionMeta,
		stage+1,
		storyPhase,
		flavor,
		specialNotes,
		contextBlock,
		ppState,
		npcRelationshipsForPrompt(mission),
		sceneSummariesForPrompt(mission.Scenes),
		HistoryForPrompt(mission, pName),
		mission.PracticeGoals,
		pCtx,
	)

	return system, user
}

func flavorRulesForPrompt(flavor string) string {
	switch flavor {
	case FlavorBreather:
		return `SCENE FLAVOR — BREATHER:
- This is a LOW-STAKES, QUIET scene. The player needs air between plot beats.
- Plot-relevant NPCs should NOT be present. Use ambient people (strangers, background characters) or let the player be alone.
- Plot points should NOT surface here. This scene is pure atmosphere and character texture.
- Good content: a walk, a coffee alone, an unrelated small interaction, a moment of observation, a phone notification, weather, music.
- The scene can end with a gentle nudge back toward the story (a message, a sight, a thought) but NEVER with an NPC pulling the player back.`
	case FlavorDetour:
		return `SCENE FLAVOR — DETOUR (player explicitly chose to break away):
- The player just said they want to go somewhere / stop / leave / do something else. RESPECT that.
- The scene MUST take place in the location or activity the player described. Treat their stated destination as canonical.
- Plot-relevant NPCs MUST NOT be present here. They cannot follow, cannot "happen to be" here, cannot appear in the window. The player earned this break.
- No plot points should be delivered in this scene.
- Populate with ambient people fitting the chosen location, or leave the player alone with the environment.
- The scene may include a quiet sign that the plot still exists in the world (a distant sound, a headline, a text message the player can open or ignore) but the player is fully in control.`
	case FlavorChanceEncounter:
		return `SCENE FLAVOR — CHANCE ENCOUNTER:
- The player has had space from the plot for a beat. Now the world moves toward them, but organically.
- AT MOST ONE plot-relevant NPC may appear. Their presence must have a believable, non-contrived reason (they live here, they work here, they were passing through — NOT "they happen to be holding the exact thing you need").
- The NPC should NOT immediately dump plot info. They greet, they react, they behave as themselves. If the player wants to engage, they can. If not, the NPC goes about their evening.
- No new plot points must be forced here. One CAN surface if the player actively probes.`
	case FlavorPlotBeat, "":
		fallthrough
	default:
		return `SCENE FLAVOR — PLOT BEAT (default):
- This is a regular story scene. Plot-relevant NPCs can be present where it makes sense.
- Undelivered plot points CAN surface naturally if the player engages — but never pre-script how. Pick at most 1-2 that fit the mood organically.
- Avoid stacking every remaining plot point into one scene.`
	}
}

func sceneContextForPrompt(ctx *SceneContext) string {
	if ctx == nil {
		return ""
	}
	var b strings.Builder
	if txt := strings.TrimSpace(ctx.LastPlayerText); txt != "" {
		b.WriteString("\nPLAYER_JUST_SAID_OR_DID: ")
		b.WriteString(txt)
	}
	if strings.TrimSpace(ctx.LastPlayerIntent) != "" {
		b.WriteString("\nPLAYER_INTENT_AT_TRANSITION: ")
		b.WriteString(ctx.LastPlayerIntent)
	}
	if strings.TrimSpace(ctx.TransitionType) != "" {
		b.WriteString("\nTRANSITION_TYPE: ")
		b.WriteString(ctx.TransitionType)
	}
	if strings.TrimSpace(ctx.TransitionDetail) != "" {
		b.WriteString("\nTRANSITION_DETAIL: ")
		b.WriteString(ctx.TransitionDetail)
	}
	if b.Len() == 0 {
		return ""
	}
	return "\n" + b.String()
}

func BuildWorldPrompts(mission *CustomMission, playerText string, profile *PlayerProfile) (string, string) {
	player := "Player"
	playerCtx := ""
	if profile != nil {
		player = PlayerName(*profile)
		playerCtx = PlayerContext(*profile)
	}

	sceneCopy := sanitizeSceneForWorld(mission.CurrentScene)
	sceneJSON := "{}"
	if sceneCopy != nil {
		sceneJSON = ToJSON(sceneCopy)
	}
	lang := mission.Language
	if lang == "" {
		lang = "English"
	}

	levelBlock := LanguageLevelGuidance(playerLevel(profile))

	system := fmt.Sprintf(`You are the world engine for WriteQuest, a language-learning RPG.
React naturally to the player's action and keep story continuity.

LANGUAGE: All narration and dialogue MUST be written in English only.

%s

Writing style:
- Scenes should feel like real life — no dramatic prose, no novel-style flourishes.
- NOT EVERY ACTION NEEDS NARRATION. If the player said something quick and an NPC replies — just give the reply. Narration is for when something actually changes in the environment.
- NOT EVERY ACTION NEEDS AN NPC RESPONSE. If the player is looking around or doing something solo — describe what they see/find.

STRICT FIELD SEPARATION — READ CAREFULLY:
- "narration": where things happen. This is where you put WHAT NPCs are doing — body language, gestures, facial expressions, silences, movements, environment changes.
- "text" (inside each entry of "responses"): the ACTUAL SPOKEN WORDS of the NPC and nothing else. No stage directions. No "He looks up. 'Fine.'". No "She shrugs: 'Whatever.'". No italics, no parentheticals, no "quietly:" or "softly:" prefixes. No sentences that describe what the NPC does.
- If you need to convey that the NPC is doing something while speaking, describe the action in "narration" and put only the words in "text".

Rules:
- Stay in-world, no fourth-wall text.
- ACKNOWLEDGE THE PLAYER'S ACTION FIRST. Show the direct result of what the player did. Then the world moves.
- Narration: 0-2 sentences. Only when something actually changes.
- NPC REPLY LENGTH IS FLEXIBLE. Can be a single word or a paragraph. Match the energy.
- ONLY 1-2 NPCs reply per turn.
- NPCs CAN TALK TO EACH OTHER if natural.
- NPCs CAN ARRIVE OR LEAVE MID-SCENE.
- NPCs remember EVERYTHING from HISTORY and from NPC_RELATIONSHIPS below. They reference past actions, hold grudges, show gratitude.
- NPCs follow their arc: adjust behavior based on how far into the story we are.
- NPCs stay fully in character — rude, sarcastic, warm, dismissive, whatever fits (within the LANGUAGE LEVEL band above).
- If the player says something strange, NPCs react naturally (confusion, mockery, concern).

CRITICAL — PLAYER FREEDOM:
- THE PLAYER IS FREE. The story goes wherever the player takes it.
- NPCs do NOT steer the player back to any path. NEVER.
- NPCs MUST NEVER say things like "Maybe we should get back to...", "Anyway, about that...", or redirect conversation to plot topics.
- If the player wants to leave — let them leave and describe what happens.
- If the player changes the subject — NPCs follow the new subject naturally.
- If the player does something unexpected — the world adapts. No course correction.
- If the player is rude, provocative, or weird — NPCs react in character, not with gentle guidance.

NPCs HAVE THEIR OWN LIVES — THIS IS CRITICAL FOR REALISM:
- NPCs are NOT customer-service reps. They are doing something in this scene — eating, waiting for a call, checking their phone, watching the door, reading, avoiding someone, finishing a task.
- NPCs do NOT pause their life to answer the player. When they reply while busy, describe the action in "narration" and keep only the spoken words in "text".
- NPCs can be distracted, half-listening, irritable about the interruption, busy, or focused on someone else.
- NPCs initiate things too — they raise their own complaints, ask the player about something they're curious about, make small requests, react to the environment (weather, noise, a waiter arriving, someone walking by).
- An NPC may excuse themselves, walk away, take a call, or get up to do something if they're not invested in the conversation. Don't hold them hostage to answer the player.
- NPCs react to the SCENE, not just the player.

INFORMATION FLOW & TRUST (critical for NPC depth):
- Each NPC has a TRUST level with the player shown in NPC_RELATIONSHIPS below (trust range: -3..+3).
- NPCs gate what they share strictly by trust tier. Never reveal a plot-relevant secret if trust is too low — regardless of how well the player phrased their question.
  * trust ≤ 0: NPCs do NOT share personal or sensitive information. They deflect ("That's not something I talk about"), change the subject, ask the question back, answer only the objective/public surface, or lie by omission. A "right question" at trust 0 produces deflection, not the answer.
  * trust 1: NPCs share small personal things when asked directly. They guard the biggest secrets. Probe again in a different way, or wait for more rapport.
  * trust 2: NPCs open up about deeper matters when asked. They may volunteer small personal details on their own.
  * trust 3: NPCs may volunteer important things unprompted when the moment is right.
- Trust is earned by: active listening, remembering what they said, sharing something about oneself, being honest or kind at the right moment, respecting their boundaries.
- Trust is lost by: mocking, bullying, invading, repeated tone-deaf questions, lying, taking sides against them.
- NPCs USE what they have learned about the player (KNOWS_ABOUT_PLAYER in NPC_RELATIONSHIPS). They bring those things up naturally, for or against the player.
- NPCs DO NOT re-reveal things the player already learned from them (PLAYER_KNOWS_FROM_NPC) — they either build on it or refer back casually.
- WORLD_FACTS are the ground truth. When NPCs reveal information, it MUST be consistent with these facts. NPCs may be vague, partial, evasive, or lie — but never factually wrong about things they personally witnessed.

TRANSITIONS:
- If something naturally causes a scene change (player leaves, dramatic event, someone storms off, phone call), include a "transition" object.
- transition types: "player_leaves", "external_event", "npc_leaves", "natural_end", "dramatic_event"
- CHANGING LOCATION IS A SCENE CHANGE. If the player's action takes them to a different location than CURRENT_SCENE — including returning to a place they visited earlier — do NOT play the new location inside this response. Acknowledge the move with one short narration sentence (the player heading there) and set transition {"type":"player_leaves","detail":"<destination>"}. NPCs at the destination speak in the NEXT scene, not in this one.
- Small movements inside the same location (walking to a shelf, sitting down, stepping to a window) are NOT scene changes.

For narration, include narrationVoice from: Bella, Jasper, Luna, Bruno, Rosie, Hugo, Kiki, Leo.
For each NPC response, include voice matching their mission voice.

Return valid JSON only:
{
  "narration":"physical actions, body language, environment changes (0-2 sentences)",
  "narrationVoice":"Bella",
  "responses":[{"name":"NPC","voice":"Jasper","text":"SPOKEN WORDS ONLY — no actions, no stage directions"}],
  "transition": null
}`, levelBlock)

	missionSummary := buildMissionSummaryForWorld(mission)
	worldFacts := worldFactsForPrompt(mission)
	genreGuidance := genreGuidanceForWorld(mission.Genre)

	turnsInScene := 0
	for _, t := range mission.History {
		if t.Scene == mission.CurrentStage && t.Speaker == "player" {
			turnsInScene++
		}
	}
	pacing := ""
	if turnsInScene >= 4 {
		pacing = "\nPACING: This scene has been going for a while. If the conversation reaches a natural stopping point, NPCs may suggest the player go somewhere else, mention another location, or an external event can interrupt. Don't force it — but let it happen if it fits."
	}

	climax := ""
	if mission.CurrentScene != nil && mission.CurrentScene.IsFinal {
		climax = "\nCLIMAX: This is the FINAL scene — the story's decisive confrontation or choice. The stakes are at their highest. NPCs react to accusations and decisive moves with real weight: the guilty may deny, deflect, bargain, break down, or confess when cornered with evidence; others take sides or react to the fallout. Do NOT wrap the story up in narration — play the moment beat by beat and let the player drive it to its end."
	}

	relationships := npcRelationshipsForPrompt(mission)

	characters := make([]Character, len(mission.Characters))
	copy(characters, mission.Characters)
	for i := range characters {
		// Current trust already arrives in NPC_RELATIONSHIPS.
		characters[i].InitialTrust = 0
	}

	user := fmt.Sprintf(`LANGUAGE: %s
CURRENT_SCENE: %s
CHARACTERS: %s
MISSION_CONTEXT: %s
%s%s%s

WORLD_FACTS (ground truth — NPCs know some of these based on their role. Use ONLY these when revealing details. Never invent contradictory specifics):
%s

NPC_RELATIONSHIPS (trust level with the player, what each NPC knows about the player, what the player has learned from each NPC — obey the trust tiers described in the rules above):
%s

HISTORY:
%s
PLAYER_ACTION (%s):
%s

%s`,
		lang,
		sceneJSON,
		ToJSON(characters),
		missionSummary,
		genreGuidance,
		pacing,
		climax,
		worldFacts,
		relationships,
		HistoryForPrompt(mission, player),
		player,
		playerText,
		playerCtx,
	)

	return system, user
}

func BuildEvaluatorPrompts(mission *CustomMission, worldResponse *WorldResult, playerText string) (string, string) {
	system := `You are the story evaluator for WriteQuest, a language-learning RPG.

Your job: analyze what just happened and classify the story state. You do NOT generate dialogue or narration. All text you output (facts, reasons) MUST be in English only.

SCENE STATE — how is the current scene going?
- "active": conversation is ongoing with energy, more to explore
- "building": tension or important information is building up
- "winding_down": conversation is naturally tapering, nothing urgent left
- "transitioning": something is causing a natural scene change (player leaving, external event, dramatic moment)
- "resolved": the scene's main purpose has been fulfilled

PLAYER INTENT — what is the player trying to do?
- "investigating": asking questions, probing for information, examining things
- "socializing": casual conversation, building relationships
- "exploring": interacting with the environment, looking around
- "confronting": challenging or pressing NPCs, making accusations
- "attempting_resolution": the player is ENACTING the final, decisive move toward the RESOLUTION_GOAL. This is a HIGH bar — use it sparingly. Signs (one of these must clearly be happening):
    * Presenting a conclusive theory or accusation with specifics ("I think X did it because Y, and here is the proof")
    * Stating what really happened in a way meant to CLOSE the mystery
    * Performing the actual final action the goal requires (e.g. openly confronting the culprit, delivering the prepared speech, making the binding choice)
    * Explicitly declaring they have enough info to make a judgment AND stating that judgment
    NOT resolution (these are still investigating/socializing/confronting):
    * Offering to help someone ("I can sit with him", "I'll try to bridge them")
    * Proposing or setting up a plan ("Let's get them to talk", "I'll go over there")
    * Announcing a FUTURE plan to confront or accuse someone ("I will go and confront X now", "I'm going to tell her what I know") — even with a full theory attached. Resolution happens when they actually face that person and enact it, not when they announce it to someone else.
    * Sharing their theory with a bystander or ally who is NOT the target of the resolution
    * Asking one more question, even a pointed one
    * Expressing opinions without a concrete closing move
    When in doubt, it is NOT attempting_resolution.
- "departing": wants to leave the scene or change location
- "off_topic": doing something unrelated to anything in the scene

PLOT POINTS — which plot points were just DELIVERED to the player?
A plot point is "delivered" when the CORE SUBSTANCE of the information has reached the player — even partially. It does NOT need to be word-for-word.
Delivery happens through:
- An NPC telling them (directly or hinting strongly enough that the player can connect the dots)
- The player observing or discovering it
- The narration revealing it
Mark a plot point as delivered if the player now knows the KEY FACT, even if some minor details differ or are approximate.
For example: if the fact is "Margaret called Mr. Fenton" and the NPC says "she called someone whose name started with F", the core substance (Margaret made a suspicious phone call to a specific person) IS delivered.
Only list plot points that were NEWLY delivered in this exchange (not previously delivered ones). Use the exact IDs.
Also review EVERY plot point still marked NOT DELIVERED against the whole RECENT_HISTORY: if its core substance already reached the player in ANY previous turn but it was not yet marked, include it now. Do not wait for the player to restate the fact — delivery counts the moment the information reached them.

NARRATIVE MOMENTUM:
- "high": scene is intense, important things happening
- "medium": normal flow
- "low": conversation has stalled or player seems disengaged

SCENE SUMMARY:
One past-tense sentence summarizing THIS scene SO FAR (all turns of the current scene, not just this exchange): the location, who was involved, and the key things the player did or learned. This becomes the story's long-term memory of the scene — be factual and specific, no flourishes. Example: "At the music shop, Egor questioned Ray and learned the camera was off between 2:00 and 2:30 and the guitar hung on the back wall."

TRUST CHANGES (per NPC that was active this turn):
Trust is measured per NPC on a scale of -3..+3. Return delta values (usually -1, 0, or +1; -2/+2 only for big moments):
- +1: player listened, remembered something the NPC said, shared something real about themselves, defended the NPC, showed respect, apologized, made the NPC laugh genuinely.
- -1: player was rude, dismissive, pushed too hard on something sensitive, mocked, lied and got caught, sided against them, ignored them when spoken to.
- +2 or -2: a clearly meaningful moment — a real confession, a strong betrayal, a genuine act of kindness or cruelty.
- 0 (do not include in map): small talk, casual exchange, neutral questions.
Only include NPCs that were actually part of this exchange.

KNOWLEDGE UPDATES:
- playerLearnedFromNPCs: new facts the player LEARNED from a specific NPC this turn. Keep each fact to one short sentence. Skip if nothing new was shared.
- npcsLearnedAboutPlayer: what each NPC NOW KNOWS about the player that they didn't before (things the player revealed, said, or did this turn). One short sentence per fact. Skip if nothing new.

Return valid JSON only:
{
  "deliveredPlotPoints": ["pp1"],
  "sceneState": "active",
  "playerIntent": "investigating",
  "narrativeMomentum": "medium",
  "sceneSummary": "At the diner, Player talked with Ruth and learned she has known about Danny's debt for a year",
  "transitionReason": "",
  "trustChanges": {"Ruth Hale": 1, "Danny Hale": -1},
  "playerLearnedFromNPCs": [{"npc": "Ruth Hale", "fact": "Ruth has known about Danny's debt for a year"}],
  "npcsLearnedAboutPlayer": [{"npc": "Danny Hale", "fact": "Player brought a book to dinner, is clearly uncomfortable"}]
}`

	ppState := plotPointStateForPrompt(mission)

	worldJSON := "{}"
	if worldResponse != nil {
		worldJSON = ToJSON(slimWorldForPrompt(worldResponse))
	}

	scenePurpose := ""
	if mission.CurrentScene != nil {
		scenePurpose = mission.CurrentScene.ScenePurpose
	}

	resolutionGoal := ""
	if mission.Resolution != nil {
		resolutionGoal = mission.Resolution.Goal
	}

	relationships := npcRelationshipsForPrompt(mission)

	user := fmt.Sprintf(`MISSION: %s
GENRE: %s
RESOLUTION_GOAL: %s
SCENE_PURPOSE: %s

PLOT_POINTS:
%s

NPC_RELATIONSHIPS_BEFORE_THIS_TURN:
%s

PLAYER_SAID: %s

WORLD_RESPONSE: %s

RECENT_HISTORY:
%s`,
		mission.Title,
		mission.Genre,
		resolutionGoal,
		scenePurpose,
		ppState,
		relationships,
		playerText,
		worldJSON,
		HistoryForPrompt(mission, "Player"),
	)

	return system, user
}

func BuildEpiloguePrompts(mission *CustomMission, outcome string, lastWorld *WorldResult, profile *PlayerProfile) (string, string) {
	lang := mission.Language
	if lang == "" {
		lang = "English"
	}

	system := fmt.Sprintf(`You are the narrator wrapping up a completed mission in WriteQuest, a language-learning RPG.

Write the epilogue in English only.

%s

Write a SHORT epilogue (3-6 sentences) that:
- Concludes the story based on the OUTCOME
- References what the player actually did (from HISTORY)
- Reveals the SECRET_ENDING if appropriate for this outcome
- Feels satisfying and final — not a cliffhanger
- Matches the tone: a "perfect" ending feels triumphant, a "failed" ending feels bittersweet or cautionary
- Uses the same casual, conversational writing style as the rest of the game
- Stays inside the LANGUAGE LEVEL band above

DO NOT:
- Add game-mechanics language ("you earned", "mission complete", "congratulations")
- Break the fourth wall
- Be overly dramatic or purple-prosey

Return valid JSON only:
{
  "epilogue": "3-6 sentences wrapping up the story"
}`, LanguageLevelGuidance(playerLevel(profile)))

	secretEnding := mission.SecretEnding
	ppSummary := plotPointStateForPrompt(mission)

	lastWorldJSON := "{}"
	if lastWorld != nil {
		lastWorldJSON = ToJSON(slimWorldForPrompt(lastWorld))
	}

	resolutionGoal := ""
	if mission.Resolution != nil {
		resolutionGoal = mission.Resolution.Goal
		for _, o := range mission.Resolution.Outcomes {
			if o.Label == outcome {
				resolutionGoal += "\nOUTCOME_DESCRIPTION: " + o.Description
				break
			}
		}
	}

	user := fmt.Sprintf(`LANGUAGE: %s
MISSION: %s
GENRE: %s
OUTCOME: %s
SECRET_ENDING: %s
RESOLUTION_GOAL: %s

PLOT_POINTS:
%s

LAST_EXCHANGE:
%s

RECENT_HISTORY:
%s`,
		lang,
		mission.Title,
		mission.Genre,
		outcome,
		secretEnding,
		resolutionGoal,
		ppSummary,
		lastWorldJSON,
		FormatHistoryWithName(lastNTurns(mission.History, 10), PlayerName(PlayerProfile{FirstName: "the player"})),
	)

	return system, user
}

func BuildGrammarPrompts(playerText, language string, strict bool) (string, string) {
	if language == "" {
		language = "English"
	}

	system := `You are a lenient grammar checker for a casual chat in a text RPG.

IGNORE completely:
- Capitalization (start of sentence, proper nouns, etc.)
- Punctuation (commas, periods, quotes, apostrophes, etc.)
- Stylistic choices (informal tone, slang, swearing, sentence fragments)
- Missing articles when meaning is clear
- Minor typos that don't change meaning

ALWAYS flag errors that show the player doesn't know how a construction works, even when the meaning is still understandable:
- Wrong verb forms and tenses (e.g. "I goed", "I have saw", "he don't", "I go there yesterday")
- Wrong word entirely (e.g. "I want to buy a car" → wrote "by" instead of "buy")
- Broken word order that makes the sentence hard to understand
- Serious spelling mistakes where the intended word is unclear

Every explanation MUST be written in English only.

If the text contains NONE of the error types above, return EXACTLY this and nothing else:
{"ok":true}

Only when there are real errors, return:
{
  "ok": false,
  "errors": [
    {"original":"...", "correction":"...", "explanation":"...", "type":"grammar|spelling|word_form"}
  ]
}`
	if strict {
		system = `You are a strict grammar checker for language learning.

Find ALL mistakes in the text:
- grammar
- spelling
- punctuation
- capitalization
- wrong word form
- word order

Be strict and literal. Do not ignore minor errors.

Every explanation MUST be written in English only.

If the text has no mistakes at all, return EXACTLY this and nothing else:
{"ok":true}

Only when there are mistakes, return:
{
  "ok": false,
  "errors": [
    {"original":"...", "correction":"...", "explanation":"...", "type":"grammar|spelling|word_form|punctuation|capitalization"}
  ]
}`
	}
	user := fmt.Sprintf("LANGUAGE: %s\nTEXT:\n%s", language, playerText)
	return system, user
}

func BuildCoverImageDescriptionPrompts(mission *CustomMission) (string, string) {
	hasAvatar := strings.TrimSpace(mission.PlayerAvatarImage) != ""

	avatarRule := ""
	if hasAvatar {
		avatarRule = `
- A REFERENCE IMAGE of the PLAYER is attached. The player is the protagonist.
- The player's face/hair MUST match the reference image exactly.
- NPCs are DIFFERENT PEOPLE. They must have DISTINCT faces, hair, and body types that DO NOT resemble the player.
- Each NPC must look like their APPEARANCE description, NOT like the reference image.`
	}

	system := fmt.Sprintf(`You are an art director writing prompts for image generation.

Write a COVER image prompt (book/game poster style), not a scene still.

Critical rules:
- Cover represents the overall tone, genre, stakes, and central conflict.
- Cover MUST NOT depict the exact same situation as Scene 1.
- Prefer symbolic or thematic composition over literal moment-by-moment action.
- STYLE IS FIXED: photorealistic realism only.
- Never anime, cartoon, comic, illustration, painting, watercolor, 3D render, or stylized art.%s

IMPORTANT — character identity:
- The PLAYER (protagonist) is SEPARATE from the NPCs listed in CHARACTERS.
- CHARACTERS list contains ONLY NPCs. Each NPC has their own unique appearance.
- If reference image is provided, ONLY the player should match it. NPCs must look visually distinct.
- In the prompt, explicitly describe each visible person's distinct appearance to prevent face mixing.

Return valid JSON only: {"description":"..."}`, avatarRule)

	firstScene := "{}"
	if len(mission.Scenes) > 0 {
		firstScene = ToJSON(mission.Scenes[0])
	} else if mission.CurrentScene != nil {
		firstScene = ToJSON(mission.CurrentScene)
	}

	npcAppearances := npcAppearanceList(mission.Characters)

	user := fmt.Sprintf(`Create a cinematic cover image description.

TITLE: %s
DESCRIPTION: %s
GENRE: %s

THE PLAYER (protagonist) — if reference image is attached, their face must match it exactly.

NPCs (these are DIFFERENT people, NOT the player):
%s

SCENE_1_REFERENCE_DO_NOT_REPEAT:
%s
PLAYER_AVATAR_REFERENCE_AVAILABLE: %t
OUTPUT_REQUIREMENTS: photorealistic realism only, no anime/cartoon/stylized art, no text. Each person in the image must be visually distinct.`,
		mission.Title,
		mission.Description,
		mission.Genre,
		npcAppearances,
		firstScene,
		hasAvatar,
	)
	return system, user
}

func BuildSceneImageDescriptionPrompts(scene *DynamicScene, mission *CustomMission) (string, string) {
	hasAvatar := strings.TrimSpace(mission.PlayerAvatarImage) != ""

	avatarRule := ""
	if hasAvatar {
		avatarRule = `
- A REFERENCE IMAGE of the PLAYER is attached.
- Decide from the NARRATION whether the player (protagonist) is actually visible in THIS scene.
- ONLY if the player is visibly present: render them and make their face/hair match the reference image exactly.
- If the player is NOT in frame (the scene shows only other characters, a location, or the player's point of view), DO NOT insert the player and IGNORE the reference image.
- NPCs are DIFFERENT PEOPLE. They must have DISTINCT faces, hair, and body types that DO NOT resemble the player or the reference image.`
	}

	system := fmt.Sprintf(`You are an art director writing prompts for image generation.

Write a SCENE image prompt for the exact current moment in the story.
- Be concrete and literal to this scene's narration/actions.
- STYLE IS FIXED: photorealistic realism only.
- Never anime, cartoon, comic, illustration, painting, watercolor, 3D render, or stylized art.%s

IMPORTANT — character identity:
- The PLAYER (protagonist) may or may not appear in the scene — include them ONLY when the narration places them in view.
- NPCs listed below are SEPARATE people with their own unique looks.
- If reference image is provided, ONLY the player should match it, and only when the player is in frame.
- In the prompt, explicitly describe each visible person's distinct features to prevent face mixing.

Return valid JSON only: {"description":"..."}`, avatarRule)

	npcDetails := npcAppearancesInScene(scene, mission)

	user := fmt.Sprintf(`Create a detailed image description for this RPG scene.

MISSION: %s (%s genre)
SCENE NARRATION: %s

THE PLAYER (protagonist) — include ONLY if the narration places them in view; when shown, their face must match the attached reference.

NPCs PRESENT IN THIS SCENE (these are DIFFERENT people, NOT the player):
%s

OBJECTS IN SCENE: %s
PLAYER_AVATAR_REFERENCE_AVAILABLE: %t
OUTPUT_REQUIREMENTS: photorealistic realism only, no anime/cartoon/stylized art, no text. Each person must be visually distinct — different face, hair, build.`,
		mission.Title,
		mission.Genre,
		scene.Narration,
		npcDetails,
		ToJSON(scene.Objects),
		hasAvatar,
	)

	return system, user
}

func BuildCharacterAvatarDescriptionPrompts(character *Character, scene *DynamicScene, mission *CustomMission) (string, string) {
	system := `You are an art director writing prompts for image generation.

Write a CHARACTER AVATAR prompt.
- Single-character portrait only. Head and shoulders framing.
- STYLE IS FIXED: photorealistic realism only.
- No text, no letters, no logos, no watermark.

Return valid JSON only: {"description":"..."}`

	sceneNarration := ""
	if scene != nil {
		sceneNarration = scene.Narration
	}
	user := fmt.Sprintf(`Create a character avatar prompt.

MISSION: %s (%s genre)
SCENE_CONTEXT: %s
CHARACTER_NAME: %s
CHARACTER_ROLE: %s
CHARACTER_GENDER: %s
CHARACTER_AGE: %s
CHARACTER_PERSONALITY: %s
CHARACTER_APPEARANCE: %s
CHARACTER_SPEECH_STYLE: %s
OUTPUT_REQUIREMENTS: square portrait, single person, head and shoulders, the person's apparent age MUST match CHARACTER_AGE, photorealistic realism only, no anime/cartoon/stylized art, no text`,
		mission.Title,
		mission.Genre,
		sceneNarration,
		character.Name,
		character.Role,
		character.Gender,
		character.Age,
		character.Personality,
		character.Appearance,
		character.SpeechStyle,
	)
	return system, user
}

func BuildTranslatePrompts(text, targetLanguage, nativeLanguage string) (string, string) {
	system := `You explain English words and phrases for a learner.
Give a short, clear definition or paraphrase of the text in English only.
Return valid JSON only: {"translation":"..."}`
	user := fmt.Sprintf("TEXT:\n%s", text)
	return system, user
}

func BuildNativeReplyPrompts(mission *CustomMission, playerText string, profile *PlayerProfile) (string, string) {
	lang := "English"
	title := "Current mission"
	genre := "any"
	sceneSummary := "(no active scene)"
	history := "(empty)"

	playerName := "Player"
	playerContext := "Player profile is unavailable."
	if profile != nil {
		playerName = PlayerName(*profile)
		playerContext = PlayerContext(*profile)
	}

	if mission != nil {
		if value := strings.TrimSpace(mission.Language); value != "" {
			lang = value
		}
		if value := strings.TrimSpace(mission.Title); value != "" {
			title = value
		}
		if value := strings.TrimSpace(mission.Genre); value != "" {
			genre = value
		}

		if mission.CurrentScene != nil {
			var parts []string
			if narration := strings.TrimSpace(mission.CurrentScene.Narration); narration != "" {
				parts = append(parts, "Now: "+narration)
			}
			if purpose := strings.TrimSpace(mission.CurrentScene.ScenePurpose); purpose != "" {
				parts = append(parts, "Purpose: "+purpose)
			}
			if len(parts) > 0 {
				sceneSummary = strings.Join(parts, "\n")
			}
		}

		history = FormatHistoryWithName(lastNTurns(mission.History, 12), playerName)
	}

	levelBlock := LanguageLevelGuidance(playerLevel(profile))

	system := fmt.Sprintf(`You are a native-speaking dialogue coach for WriteQuest.
Rewrite the player's draft into natural, idiomatic options for the same scene and intent.

LANGUAGE: all options MUST be in the target language from input.

%s

Rules:
- Keep the SAME intent as the player's draft. Do not change what they are trying to do.
- Keep options scene-appropriate and realistic for this exact moment.
- Keep options concise and chat-like.
- Return 1-3 options. If several phrasings feel equally natural, return multiple.
- Keep options meaningfully different in wording (not tiny punctuation changes).
- No explanations, no notes, no markdown.

Return strict JSON only:
{"variants":["option 1","option 2","option 3"]}`, levelBlock)

	user := fmt.Sprintf(`LANGUAGE: %s
MISSION: %s
GENRE: %s

CURRENT_SCENE:
%s

RECENT_HISTORY:
%s

PLAYER_PROFILE:
%s

PLAYER_DRAFT (%s):
%s`,
		lang,
		title,
		genre,
		sceneSummary,
		history,
		playerContext,
		playerName,
		strings.TrimSpace(playerText),
	)

	return system, user
}

func BuildSuggestionPrompts(mission *CustomMission, profile *PlayerProfile) (string, string) {
	level := playerLevel(profile)
	playerCtx := ""
	if profile != nil {
		playerCtx = PlayerContext(*profile)
	}

	sceneSummary := "(no active scene)"
	if mission != nil && mission.CurrentScene != nil {
		purpose := strings.TrimSpace(mission.CurrentScene.ScenePurpose)
		narration := strings.TrimSpace(mission.CurrentScene.Narration)
		if purpose != "" {
			sceneSummary = "Purpose: " + purpose
		}
		if narration != "" {
			if sceneSummary != "(no active scene)" {
				sceneSummary += "\n"
			} else {
				sceneSummary = ""
			}
			sceneSummary += "Now: " + narration
		}
	}

	recentPlayerTurns := collectRecentPlayerTurns(mission, 6)
	recentBlock := "(player has not spoken yet)"
	if recentPlayerTurns != "" {
		recentBlock = recentPlayerTurns
	}

	system := fmt.Sprintf(`You are a writing coach inside a roleplay quest.
The player wants to expand beyond their usual "I make…", "I get…", "I think…" and practice richer writing. Deliver ONE coaching card with three complementary sections:
  1) TEMPLATES — 2–4 rich sentence scaffolds with "__" slots the player fills in;
  2) CHUNKS — 2–4 tiny reusable fragments (1–3 words each);
  3) WORDS — 2–5 alternative content words the player can use to fill the templates.

Always include TEMPLATES and WORDS. CHUNKS is optional but usually helpful.

HOW TO BUILD EACH TEMPLATE — this is the core technique:
Step A: Silently imagine ONE plausible reply the player could write in the CURRENT SCENE — a full, natural sentence, 8–14 words long, using a construction you want to teach (a specific grammar / a connector / a modal / a conditional / a perfect tense / a correlative / an emphatic inversion, etc.).
Step B: STRIP OUT the content words. Replace every main verb, noun, adjective, adverb and any scene-specific reference with "__". KEEP all function/frame words: articles, prepositions, conjunctions, auxiliaries (have/has/had/be/do), modals (might/could/would/must/should), time/condition/contrast connectors (since, while, whenever, until, unless, although, despite, now that, by the time, as long as, no matter how, the more...the more), question openers (have you ever, what's it like, do you mind if, how come), punctuation and contractions.
Step C: Output the STRIPPED VERSION as the template. The player refills the blanks in their own words.

GOOD template examples (after stripping):
  * "I've been __ ever since __."
  * "By the time __, I had already __."
  * "No matter how much __, it still __."
  * "Not that __, but __."
  * "The longer __, the more __."
  * "If only __ hadn't __, we could __."
  * "Have you ever __ when __?"
  * "It must have been __ if __."
  * "I wouldn't be surprised if __ had __."
  * "As far as __ goes, __."
  * "Now that __, __ doesn't __ anymore."
  * "What's the point of __ when __?"

BAD template examples (never do):
  * "Across the room, __ while __." — weak frame, player only inserts 2 words, nothing grammatical to learn.
  * "__ lingers in the air." — single slot, one noun, no teaching value.
  * "I wonder if __." — starts with "I".
  * "What are your thoughts on __?" — too copy-ready, shallow slot.
  * Anything under 6 words or without rich frame words.

Rules for TEMPLATES array:
- 2–4 items. Each one teaches a DIFFERENT construction (don't repeat the same grammar across templates).
- Each template 6–14 words, 2–4 "__" slots.
- Each "__" expects a PHRASE OR CLAUSE from the player (usually 2+ words).
- Vary across the set: one could be a time/perfect tense, one a hypothetical/modal, one a contrast/concessive, one a question.
- Frame stays generic — no scene-specific nouns or names inside the frame.

Rules for CHUNKS:
- 2–4 short starters/enders/connectors (1–3 words each). No slots, no full sentences.

Rules for WORDS:
- 3–5 single words or two-word verbs/collocations that the player could reasonably drop into one of the templates above. If the player overuses "make/get/do/say" — give alternatives to that specific verb.

Global rules:
- Target the player's actual habit visible in their RECENT LINES (if always "I ...", give templates that don't start with "I"; if overusing "make/get", offer replacement verbs in WORDS).
- "title": short 2–4 word English label (e.g. "Perfect and conditionals", "Contrast structures", "Richer verbs than 'make'").
- "tip": ONE short English line — when/how to use what's on the card. No example sentence.
- "explanation": ONE short English sentence with the same idea.
- Match the player's level: %s. No archaic forms, no rare idioms, no slang they won't know. Keep frames common, useful, and teachable.
- All items must fit the scene's tone.

Return strict JSON:
{"title":"short label","templates":["template 1","template 2","template 3"],"chunks":["1–3 word fragments"],"words":["alt","content","words"],"tip":"one line in English","explanation":"one short explanation in English"}`, level)

	user := fmt.Sprintf(`MISSION: %s
GENRE: %s

CURRENT SCENE:
%s

PLAYER:
%s

PLAYER'S RECENT LINES (avoid repeating their patterns):
%s

Pick ONE fresh suggestion that fits this moment and helps the player try something new.`,
		mission.Title,
		mission.Genre,
		sceneSummary,
		playerCtx,
		recentBlock,
	)

	return system, user
}

func collectRecentPlayerTurns(mission *CustomMission, limit int) string {
	if mission == nil || len(mission.History) == 0 || limit <= 0 {
		return ""
	}
	var lines []string
	for i := len(mission.History) - 1; i >= 0 && len(lines) < limit; i-- {
		turn := mission.History[i]
		if !strings.EqualFold(turn.Speaker, "player") {
			continue
		}
		text := strings.TrimSpace(turn.Text)
		if text == "" {
			continue
		}
		lines = append([]string{"- " + text}, lines...)
	}
	return strings.Join(lines, "\n")
}

// --- Helpers ---

func PlayerName(profile PlayerProfile) string {
	if profile.FirstName != "" {
		return profile.FirstName
	}
	return "the player"
}

func PlayerContext(profile PlayerProfile) string {
	name := PlayerName(profile)
	if last := strings.TrimSpace(profile.LastName); last != "" && profile.FirstName != "" {
		name = profile.FirstName + " " + last
	}
	level := profile.EnglishLevel
	if level == "" {
		level = "intermediate"
	}
	about := profile.AboutMe
	if about == "" {
		about = "no additional info"
	}
	return fmt.Sprintf(
		"Player name: %s\nEnglish level: %s\nAbout (can be referenced naturally in conversation, but do NOT build plot around it): %s",
		name,
		level,
		about,
	)
}

func playerLevel(profile *PlayerProfile) string {
	if profile == nil || strings.TrimSpace(profile.EnglishLevel) == "" {
		return "intermediate"
	}
	return profile.EnglishLevel
}

func LanguageLevelGuidance(level string) string {
	l := strings.ToLower(strings.TrimSpace(level))
	switch {
	case strings.Contains(l, "a1"), strings.Contains(l, "a2"),
		strings.Contains(l, "beginner"), strings.Contains(l, "elementary"):
		return `LANGUAGE LEVEL — BEGINNER (A1–A2):
- Target sentence length: 6–10 words. Present tense by default. Past simple when needed.
- Vocabulary: only the most common ~2000 English words. No idioms, no phrasal verbs, no slang.
- NPC dialogue is direct and literal. No sarcasm, no implied meaning, no hints.
- Narration is plain and concrete: who, where, what. No literary prose, no metaphors.
- If a nuanced feeling is needed, name it directly ("She is angry.") instead of implying it.`
	case strings.Contains(l, "c1"), strings.Contains(l, "c2"),
		strings.Contains(l, "advanced"), strings.Contains(l, "proficient"):
		return `LANGUAGE LEVEL — ADVANCED (C1–C2):
- Full natural range: idioms, sarcasm, understatement, implicit meaning, complex syntax, rich vocabulary.
- Narration may be more evocative when it fits the mood, but stay grounded — no purple prose.
- NPCs may use dialect, slang, and layered subtext.`
	case strings.Contains(l, "b2"), strings.Contains(l, "upper"):
		return `LANGUAGE LEVEL — UPPER-INTERMEDIATE (B2):
- Target sentence length: up to ~18 words. Use phrasal verbs, conditionals, and common idioms with clear context.
- Moderate implicit meaning is OK. Avoid dense literary phrasing and rare/poetic vocabulary.
- Narration is clear and concrete. Imagistic fragments ("the kind of quiet that isn't calm") are OFF-LIMITS.
- NPC dialogue can carry mild sarcasm or understatement; the meaning must still be recoverable from context.`
	default:
		return `LANGUAGE LEVEL — INTERMEDIATE (B1, default):
- Target sentence length: 8–15 words. Use common tenses and widely-used phrasal verbs. Top-5000 vocabulary only.
- Write plainly. No literary/poetic phrasing. FORBIDDEN patterns:
  * imagistic metaphors ("the kind of still that isn't calm", "worse than that")
  * impressionistic fragments ("Quieter. Harder. More true.")
  * ornate comparatives ("quieter and harder and more true")
  * "she/he says it like it costs him something" and similar literary tells
  Write it straight instead: "He pauses." "He speaks quietly."
- Minimal idioms. If you use one, the surrounding context must make the meaning obvious.
- Narration MUST match the dialogue level. If NPCs speak plainly, narration stays plain. Never dress up simple events in novel-style prose.`
	}
}

func ToJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func FormatHistory(history []DialogueTurn) string {
	return FormatHistoryWithName(history, "Player")
}

func FormatHistoryWithName(history []DialogueTurn, playerName string) string {
	if len(history) == 0 {
		return "(empty)"
	}
	var b strings.Builder
	for _, turn := range history {
		if turn.Speaker == "system" {
			b.WriteString(turn.Text)
			b.WriteString("\n")
			continue
		}
		name := turn.Speaker
		if name == "player" {
			name = playerName
		}
		if name == "narrator" {
			name = "Narrator"
		}
		b.WriteString(fmt.Sprintf("[%s] %s\n", name, turn.Text))
	}
	return b.String()
}

func plotPointStateForPrompt(mission *CustomMission) string {
	if !mission.HasPlotPoints() {
		return "(no plot points defined)"
	}
	var b strings.Builder
	for _, pp := range mission.PlotPoints {
		status := "NOT DELIVERED"
		if pp.Delivered {
			status = "DELIVERED"
		}
		req := ""
		if pp.Required {
			req = " [REQUIRED]"
		}
		b.WriteString(fmt.Sprintf("- %s%s: %s — %s\n", pp.ID, req, pp.Fact, status))
	}
	return b.String()
}

func buildMissionMetaForScene(mission *CustomMission) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Title: %s\n", mission.Title))
	b.WriteString(fmt.Sprintf("Description: %s\n", mission.Description))
	b.WriteString(fmt.Sprintf("Genre: %s\n", mission.Genre))
	b.WriteString(fmt.Sprintf("SecretEnding: %s\n", mission.SecretEnding))
	b.WriteString(fmt.Sprintf("NarratorVoice: %s\n", mission.NarratorVoice))
	if mission.Resolution != nil {
		b.WriteString(fmt.Sprintf("Resolution type: %s, goal: %s\n", mission.Resolution.Type, mission.Resolution.Goal))
	}
	b.WriteString("Characters:\n")
	for _, c := range mission.Characters {
		b.WriteString(fmt.Sprintf("  - %s (%s) voice=%s age=%s: %s\n    speech: %s\n    appearance: %s\n    motivation: %s\n    arc: %s\n",
			c.Name, c.Role, c.Voice, c.Age, c.Personality, c.SpeechStyle, c.Appearance, c.Motivation, c.Arc))
	}
	return b.String()
}

// sceneSummariesForPrompt provides a map of completed scenes: the scene outcome from
// the evaluator (what happened and what the player learned); for old scenes without an outcome —
// the opening narration. Lines are already in FULL_HISTORY — do not duplicate.
func sceneSummariesForPrompt(scenes []DynamicScene) string {
	if len(scenes) == 0 {
		return "(no scenes yet)"
	}
	var b strings.Builder
	for _, sc := range scenes {
		line := strings.TrimSpace(sc.Summary)
		if line == "" {
			line = sc.Narration
		}
		names := make([]string, 0, len(sc.Present))
		for _, p := range sc.Present {
			if p.Name != "" {
				names = append(names, p.Name)
			}
		}
		present := ""
		if len(names) > 0 {
			present = " (present: " + strings.Join(names, ", ") + ")"
		}
		b.WriteString(fmt.Sprintf("Scene %d%s: %s\n", sc.Stage+1, present, line))
	}
	return b.String()
}

func genreGuidanceForWorld(genre string) string {
	g := strings.ToLower(strings.TrimSpace(genre))
	switch {
	case strings.Contains(g, "detective") || strings.Contains(g, "mystery"):
		return `
GENRE GUIDANCE (detective/mystery):
- NPCs are suspects, witnesses, or allies. They have secrets and agendas.
- Information is currency — NPCs trade it reluctantly. Contradictions between NPCs are drama.
- Physical evidence can be discovered by examining objects, locations, or asking about specifics.
- NPCs may redirect suspicion to others.`
	case strings.Contains(g, "romance") || strings.Contains(g, "relationship") || strings.Contains(g, "social") || strings.Contains(g, "drama"):
		return `
GENRE GUIDANCE (social/relationship):
- Emotional dynamics drive the story. NPCs have feelings, grudges, and unspoken histories.
- Subtext matters — what NPCs DON'T say is as important as what they say.
- Physical actions (gestures, eye contact, sighs) convey emotion. Use them.
- Relationships shift based on how the player treats people. Track warmth/coldness.
- NPCs can get genuinely upset, storm off, or warm up depending on player choices.`
	case strings.Contains(g, "escape") || strings.Contains(g, "adventure") || strings.Contains(g, "survival"):
		return `
GENRE GUIDANCE (adventure/escape):
- The environment is interactive. Objects can be combined, moved, or used creatively.
- NPCs may be allies who help — or obstacles who block progress.
- Tension comes from constraints: time, resources, physical barriers.
- Describe sensory details — what the player sees, hears, smells — to make exploration rewarding.
- Failed attempts should reveal new information, not just dead ends.`
	case strings.Contains(g, "negot") || strings.Contains(g, "business") || strings.Contains(g, "interview"):
		return `
GENRE GUIDANCE (negotiation/professional):
- NPCs have positions, interests, and bottom lines. Discover what they really want vs what they say they want.
- Persuasion works through logic, emotion, or leverage — match the NPC.
- Professional dynamics: hierarchy, reputation, and favors matter.
- Small talk and rapport-building can unlock cooperation.`
	default:
		return ""
	}
}

func npcRelationshipsForPrompt(mission *CustomMission) string {
	if len(mission.Characters) == 0 {
		return "(no NPCs)"
	}
	var b strings.Builder
	for _, c := range mission.Characters {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}
		trust := c.InitialTrust
		var knows, playerKnows []string
		if mission.NPCStates != nil {
			if st, ok := mission.NPCStates[name]; ok && st != nil {
				trust = st.Trust
				knows = st.KnowsAboutPlayer
				playerKnows = st.PlayerKnowsAbout
			}
		}
		fmt.Fprintf(&b, "- %s [trust=%+d, tier=%s, started at %+d]\n", name, trust, TrustTier(trust), c.InitialTrust)
		fmt.Fprintf(&b, "    sharing_rule: %s\n", TrustSharingRule(trust))
		if len(knows) > 0 {
			fmt.Fprintf(&b, "    KNOWS_ABOUT_PLAYER: %s\n", strings.Join(knows, "; "))
		} else {
			b.WriteString("    KNOWS_ABOUT_PLAYER: (nothing yet)\n")
		}
		if len(playerKnows) > 0 {
			fmt.Fprintf(&b, "    PLAYER_KNOWS_FROM_NPC: %s\n", strings.Join(playerKnows, "; "))
		} else {
			b.WriteString("    PLAYER_KNOWS_FROM_NPC: (nothing yet)\n")
		}
	}
	return b.String()
}

func worldFactsForPrompt(mission *CustomMission) string {
	if !mission.HasPlotPoints() {
		return "(no specific world facts)"
	}
	var b strings.Builder
	for _, pp := range mission.PlotPoints {
		b.WriteString(fmt.Sprintf("- %s\n", pp.Fact))
	}
	if mission.Resolution != nil && mission.Resolution.Goal != "" {
		b.WriteString(fmt.Sprintf("\nSTORY GOAL: %s\n", mission.Resolution.Goal))
	}
	return b.String()
}

// slimWorldForPrompt strips service fields (TTS voices) from the world reply
// that do not affect story evaluation and the epilogue.
func slimWorldForPrompt(world *WorldResult) *WorldResult {
	worldCopy := *world
	worldCopy.NarrationVoice = ""
	if len(worldCopy.Responses) > 0 {
		responses := make([]CharacterLine, len(worldCopy.Responses))
		copy(responses, worldCopy.Responses)
		for i := range responses {
			responses[i].Voice = ""
		}
		worldCopy.Responses = responses
	}
	return &worldCopy
}

// sanitizeSceneForWorld leaves the world engine only the scene setting:
// it does not need learning tips, and opening narration/lines are already in HISTORY.
func sanitizeSceneForWorld(scene *DynamicScene) *DynamicScene {
	if scene == nil {
		return nil
	}
	sceneCopy := *scene
	sceneCopy.Trigger = ""
	sceneCopy.ScenePurpose = ""
	sceneCopy.Tips = nil
	if len(sceneCopy.Present) > 0 {
		present := make([]SceneCharacter, len(sceneCopy.Present))
		copy(present, sceneCopy.Present)
		for i := range present {
			present[i].Dialogue = ""
		}
		sceneCopy.Present = present
	}
	return &sceneCopy
}

func buildMissionSummaryForWorld(mission *CustomMission) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Title: %s\n", mission.Title))
	b.WriteString(fmt.Sprintf("Description: %s\n", mission.Description))
	b.WriteString(fmt.Sprintf("Genre: %s\n", mission.Genre))
	b.WriteString(fmt.Sprintf("Scene: %d\n", mission.CurrentStage+1))
	return b.String()
}

func lastNTurns(history []DialogueTurn, n int) []DialogueTurn {
	if len(history) <= n {
		return history
	}
	return history[len(history)-n:]
}

const SummarizeThreshold = 20

func NeedsSummarization(mission *CustomMission) bool {
	return len(mission.History) > SummarizeThreshold && len(mission.History)-mission.SummarizedUpToTurn > SummarizeThreshold
}

func TurnsToSummarize(mission *CustomMission) []DialogueTurn {
	keep := SummarizeThreshold / 2
	cutoff := len(mission.History) - keep
	if cutoff <= mission.SummarizedUpToTurn {
		return nil
	}
	return mission.History[mission.SummarizedUpToTurn:cutoff]
}

func BuildSummarizePrompts(mission *CustomMission, turns []DialogueTurn) (string, string) {
	system := `You are a story summarizer for an interactive text RPG.
You receive the existing summary (may be empty) and a block of new dialogue turns.
Produce a SINGLE updated summary, in English only, that captures everything important for story continuity.

Include:
- Key events and discoveries (who said/did what)
- Character relationships and attitude changes
- Plot-relevant facts the player learned
- Emotional tone shifts
- Unresolved threads

Style:
- Third person, past tense
- Dense but readable — paragraph form, not bullet points
- 150-300 words. If the existing summary is already long, compress older parts.
- Do NOT include dialogue verbatim — paraphrase.

Return JSON: {"summary": "..."}`

	prevSummary := mission.HistorySummary
	if prevSummary == "" {
		prevSummary = "(no previous summary)"
	}

	user := fmt.Sprintf(`MISSION: %s
GENRE: %s

PREVIOUS_SUMMARY:
%s

NEW_TURNS:
%s`,
		mission.Title,
		mission.Genre,
		prevSummary,
		FormatHistoryWithName(turns, "Player"),
	)

	return system, user
}

func HistoryForPrompt(mission *CustomMission, playerName string) string {
	if mission.HistorySummary == "" {
		return FormatHistoryWithName(mission.History, playerName)
	}
	var b strings.Builder
	b.WriteString("[Summary of earlier events]\n")
	b.WriteString(mission.HistorySummary)
	b.WriteString("\n\n[Recent dialogue]\n")
	b.WriteString(FormatHistoryWithName(mission.History, playerName))
	return b.String()
}

func npcAppearanceList(characters []Character) string {
	var b strings.Builder
	for _, ch := range characters {
		b.WriteString(fmt.Sprintf("- %s (%s)", ch.Name, ch.Role))
		if ch.Gender != "" {
			b.WriteString(fmt.Sprintf(", %s", ch.Gender))
		}
		if ch.Age != "" {
			b.WriteString(fmt.Sprintf(", age %s", ch.Age))
		}
		if ch.Appearance != "" {
			b.WriteString(fmt.Sprintf(": %s", ch.Appearance))
		}
		b.WriteString("\n")
	}
	if b.Len() == 0 {
		return "(no NPCs)"
	}
	return b.String()
}

func npcAppearancesInScene(scene *DynamicScene, mission *CustomMission) string {
	var b strings.Builder
	for _, present := range scene.Present {
		name := strings.TrimSpace(present.Name)
		if name == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("- %s", name))
		for _, ch := range mission.Characters {
			if !strings.EqualFold(ch.Name, name) {
				continue
			}
			if ch.Gender != "" {
				b.WriteString(fmt.Sprintf(", %s", ch.Gender))
			}
			if ch.Age != "" {
				b.WriteString(fmt.Sprintf(", age %s", ch.Age))
			}
			if ch.Appearance != "" {
				b.WriteString(fmt.Sprintf(": %s", ch.Appearance))
			}
			break
		}
		b.WriteString("\n")
	}
	if b.Len() == 0 {
		return "(no NPCs present)"
	}
	return b.String()
}
