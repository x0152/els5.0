import type { LucideIcon } from 'lucide-react'
import { BookOpenText, Clapperboard, Ear, Flame, HelpCircle, Layers, Mic, PenLine, SpellCheck } from 'lucide-react'
import type { Step } from './types.ts'

export type StepMeta = {
  label: string
  icon: LucideIcon
  grad: string
  chip: string
  blurb: string
}

export const STEP_META: Record<Step['kind'], StepMeta> = {
  warmup: {
    label: 'Warm-up',
    icon: Flame,
    grad: 'from-orange-500 to-amber-500',
    chip: 'bg-orange-100 text-orange-700',
    blurb: 'A quick pass over what you met in recent lessons.',
  },
  watch: {
    label: 'Watch',
    icon: Clapperboard,
    grad: 'from-neutral-800 to-neutral-600',
    chip: 'bg-neutral-200 text-neutral-700',
    blurb: 'Watch the scene with subtitles and catch every line.',
  },
  questions: {
    label: 'Questions',
    icon: HelpCircle,
    grad: 'from-violet-500 to-purple-600',
    chip: 'bg-violet-100 text-violet-700',
    blurb: 'Answer the questions about the scene you just watched.',
  },
  speak: {
    label: 'Speak',
    icon: Mic,
    grad: 'from-sky-500 to-blue-600',
    chip: 'bg-sky-100 text-sky-700',
    blurb: 'Listen to the original line, then say it yourself.',
  },
  dictation: {
    label: 'Dictation',
    icon: Ear,
    grad: 'from-indigo-500 to-blue-700',
    chip: 'bg-indigo-100 text-indigo-700',
    blurb: 'Play each line and type it word for word.',
  },
  reading: {
    label: 'Reading',
    icon: BookOpenText,
    grad: 'from-emerald-500 to-teal-600',
    chip: 'bg-emerald-100 text-emerald-700',
    blurb: "Read and tap the words you don't know.",
  },
  writing: {
    label: 'Writing',
    icon: PenLine,
    grad: 'from-rose-500 to-pink-600',
    chip: 'bg-rose-100 text-rose-700',
    blurb: 'Write a natural reply in English.',
  },
  grammar: {
    label: 'Grammar',
    icon: SpellCheck,
    grad: 'from-amber-500 to-orange-600',
    chip: 'bg-amber-100 text-amber-700',
    blurb: 'Fill everything in, then check yourself.',
  },
  vocab: {
    label: 'Words',
    icon: Layers,
    grad: 'from-brand-500 to-emerald-700',
    chip: 'bg-brand-100 text-brand-700',
    blurb: 'Flip each card and be honest with yourself.',
  },
}

export const STEP_LABELS: Record<Step['kind'], string> = Object.fromEntries(
  Object.entries(STEP_META).map(([k, m]) => [k, m.label]),
) as Record<Step['kind'], string>

export function stepTitle(step: Step): string {
  switch (step.kind) {
    case 'watch':
      return step.watch?.title || 'Watch the scene'
    case 'reading':
      return step.reading?.title || 'Read'
    case 'writing':
      return 'Write back'
    case 'grammar':
      return 'Grammar drill'
    case 'questions':
      return 'Did you get it?'
    case 'vocab':
      return 'Your words'
    default:
      return step.title || STEP_META[step.kind].label
  }
}

export function stepSubtitle(step: Step): string {
  switch (step.kind) {
    case 'watch':
      return step.watch?.recap ? `Previously: ${step.watch.recap}` : STEP_META.watch.blurb
    case 'writing':
      return step.writing?.scenario || STEP_META.writing.blurb
    case 'grammar':
      return step.grammar?.topic ? `${step.grammar.topic}. ${STEP_META.grammar.blurb}` : STEP_META.grammar.blurb
    default:
      return STEP_META[step.kind].blurb
  }
}

export function stepDetail(step: Step): string | undefined {
  switch (step.kind) {
    case 'warmup':
      return step.warmup?.length ? `${step.warmup.length} phrases` : undefined
    case 'watch':
      return step.watch?.title
    case 'questions':
      return step.questions?.length ? `${step.questions.length} questions` : undefined
    case 'speak':
    case 'dictation':
      return step.phrases?.length ? `${step.phrases.length} lines` : undefined
    case 'reading':
      return step.reading?.title
    case 'writing':
      return step.writing?.scenario
    case 'grammar':
      return step.grammar?.topic
    case 'vocab':
      return step.vocab?.length ? `${step.vocab.length} words` : undefined
  }
}
