export type WarmupItem = { mode: 'speak' | 'dictation'; text: string; film_id?: string; start_ms?: number; end_ms?: number }
export type Watch = { film_id: string; title: string; start_ms: number; end_ms: number; recap?: string; summary?: string }
export type Question = { text: string; options: string[]; answer: number }
export type Phrase = { text: string; film_id?: string; start_ms?: number; end_ms?: number }
export type Reading = { title: string; body: string; words?: string[] }
export type Writing = { scenario: string; dialogue: string }
export type Grammar = { topic: string; theory?: string; exercises: string }
export type VocabWord = { text: string; translation?: string; definition?: string; example?: string }

export type Step = {
  id: string
  kind: 'warmup' | 'watch' | 'questions' | 'speak' | 'dictation' | 'reading' | 'writing' | 'grammar' | 'vocab'
  title: string
  done: boolean
  score: number
  warmup?: WarmupItem[]
  watch?: Watch
  questions?: Question[]
  phrases?: Phrase[]
  reading?: Reading
  writing?: Writing
  grammar?: Grammar
  vocab?: VocabWord[]
}

export type Lesson = {
  id: string
  number: number
  cycle_index: number
  review: boolean
  film_id?: string
  start_ms: number
  end_ms: number
  status: 'active' | 'completed'
  steps: Step[]
  created_at: string
}

export type ItemResult = { kind: 'phrase' | 'word'; text: string; film_id?: string; start_ms?: number; end_ms?: number; score: number }

export type StepOutcome = { score: number; results?: ItemResult[] }
