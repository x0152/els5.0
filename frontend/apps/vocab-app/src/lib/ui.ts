import type { ComponentType } from 'react'
import { MessageSquare, Puzzle, Sparkles, Type, type LucideProps } from 'lucide-react'
import type { UnitStatus } from './types.ts'

export type KindIcon = ComponentType<LucideProps>

export const kindIcon: Record<string, KindIcon> = {
  word: Type,
  phrase: MessageSquare,
  phrasal_verb: Puzzle,
  idiom: Sparkles,
}

export function getKindIcon(kind: string): KindIcon {
  return kindIcon[kind] ?? Type
}

export const statusPill: Record<UnitStatus, string> = {
  new: 'bg-sky-50 text-sky-700 ring-sky-200',
  learning: 'bg-amber-50 text-amber-700 ring-amber-200',
  learned: 'bg-brand-50 text-brand-700 ring-brand-200',
}

export const statusDot: Record<UnitStatus, string> = {
  new: 'bg-sky-400',
  learning: 'bg-amber-400',
  learned: 'bg-brand-500',
}
