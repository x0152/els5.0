import type { QuestComponents } from '@els/api-client'

type Schemas = QuestComponents['schemas']

export type Mission = Schemas['CustomMission']
export type MissionSummary = Schemas['MissionSummary']
export type ActiveReply = Schemas['RespondJobStatusResponse']
export type RespondResult = Schemas['RespondJobResult']
export type DialogueTurn = Schemas['DialogueTurn']
export type DynamicScene = Schemas['DynamicScene']
export type SceneCharacter = Schemas['SceneCharacter']
export type Character = Schemas['Character']
export type GrammarError = Schemas['GrammarError']
export type PartialWorld = Schemas['PartialWorld']
export type LanguageTip = Schemas['LanguageTip']
export type PlotPoint = Schemas['PlotPoint']

export type GenerationStatus = 'generating' | 'ready' | 'error'
export type ImageStatus = 'generating' | 'ready' | 'error' | string
