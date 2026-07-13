import { createContext, useContext } from 'react'

export type PracticeKind = string

export type PracticeVariant = {
  id: string
  title: string
  exercises: string
  status: 'generating' | 'ready' | 'error'
  error?: string
}

export type PracticeAnswer = {
  answer: string
  correct: boolean
  correction?: string
  explanation?: string
}

export type PracticeProgress = { answers: Record<string, PracticeAnswer>; completed: boolean }

export type CheckFreeInput = { kind: PracticeKind; number: number; instruction: string; answer: string }

export type CheckFreeResult = { correct: boolean; correction?: string; explanation?: string }

export type PracticeApi = {
  listVariants: (kind: PracticeKind, number: number) => Promise<PracticeVariant[]>
  generateVariant: (kind: PracticeKind, number: number) => Promise<PracticeVariant>
  deleteVariant: (id: string) => Promise<void>
  getProgress: (kind: PracticeKind, number: number, variant: string) => Promise<PracticeProgress>
  saveProgress: (kind: PracticeKind, number: number, variant: string, progress: PracticeProgress) => Promise<void>
  resetProgress: (kind: PracticeKind, number: number, variant: string) => Promise<void>
  checkFree: (input: CheckFreeInput) => Promise<CheckFreeResult>
}

export type PracticeKey = { kind: PracticeKind; number: number }

// Per-answer persistence consumed by Gap/Write components.
export type ProgressCtxValue = {
  enabled: boolean
  version: number
  get: (key: string) => PracticeAnswer | undefined
  set: (key: string, value: PracticeAnswer) => void
  register: (key: string) => void
  keys: () => string[]
  remove: (keys: string[]) => void
}

export const ProgressCtx = createContext<ProgressCtxValue>({
  enabled: false,
  version: 0,
  get: () => undefined,
  set: () => {},
  register: () => {},
  keys: () => [],
  remove: () => {},
})

export const useProgress = () => useContext(ProgressCtx)

export type PracticeMeta = { api: PracticeApi; kind: PracticeKind; number: number }

export const PracticeMetaCtx = createContext<PracticeMeta | null>(null)

export const usePracticeMeta = () => useContext(PracticeMetaCtx)

// Standalone free-answer checker, used when there is no chapter-bound PracticeApi.
export type CheckFreeFn = (input: { instruction: string; answer: string }) => Promise<CheckFreeResult>

export const CheckFreeCtx = createContext<CheckFreeFn | null>(null)

export const useCheckFree = () => useContext(CheckFreeCtx)

// Language the learner produced while solving an exercise (sent as a raw event).
export type ProduceEvent = (e: { skill: 'writing'; text: string; context?: string }) => void

export const ProduceCtx = createContext<ProduceEvent | null>(null)

export const useProduce = () => useContext(ProduceCtx)
