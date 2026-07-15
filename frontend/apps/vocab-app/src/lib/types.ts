import type { VocabComponents } from '@els/api-client'

type Schemas = VocabComponents['schemas']

export type Unit = Schemas['UnitOutput']
export type AddUnitResult = Schemas['AddUnitOutput']
export type Occurrences = Schemas['OccurrencesOutput']
export type Card = Schemas['CardOutput']
export type CardAnswer = Schemas['AnswerCardOutput']

export type UnitStatus = 'new' | 'learning' | 'learned'
export type UnitKind = 'word' | 'phrase' | 'phrasal_verb' | 'idiom'

export const KIND_LABELS: Record<string, string> = {
  word: 'Word',
  phrase: 'Phrase',
  phrasal_verb: 'Phrasal verb',
  idiom: 'Idiom',
}

export const STATUS_LABELS: Record<UnitStatus, string> = {
  new: 'New',
  learning: 'Learning',
  learned: 'Learned',
}
