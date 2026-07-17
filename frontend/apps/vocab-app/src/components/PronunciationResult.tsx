import { cn, anchorOf, type PhonemeAnchor } from '@els/ui'
import type { SpeechComponents } from '@els/api-client'

export type Assessment = SpeechComponents['schemas']['AssessOutput']

const VERDICT_STYLES: Record<string, string> = {
  good: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
  close: 'bg-amber-50 text-amber-700 ring-amber-300',
  wrong: 'bg-red-50 text-red-700 ring-red-300',
  missing: 'bg-neutral-100 text-neutral-400 ring-neutral-200 line-through',
}

export function PronunciationResult({
  assessment,
  onSelect,
  className,
}: {
  assessment: Assessment
  onSelect: (symbol: string, anchor: PhonemeAnchor) => void
  className?: string
}) {
  const score = assessment.overall
  return (
    <div className={cn('rounded-xl bg-white/70 p-3 ring-1 ring-neutral-200', className)}>
      <p className="text-sm font-medium text-neutral-900">
        Pronunciation:{' '}
        <span className={score >= 85 ? 'text-emerald-600' : score >= 60 ? 'text-amber-600' : 'text-red-600'}>
          {score}/100
        </span>
      </p>
      <div className="mt-2 flex flex-wrap items-center gap-1">
        {(assessment.words ?? []).flatMap((w, i) => [
          ...(w.phonemes ?? []).map((p, j) => (
            <button
              key={`${i}-${j}`}
              type="button"
              onClick={(e) => onSelect(p.expected, anchorOf(e.currentTarget))}
              title={p.verdict === 'good' ? `/${p.expected}/` : `expected /${p.expected}/, heard /${p.heard ?? '—'}/`}
              className={cn('rounded-lg px-2 py-1 font-mono text-sm ring-1', VERDICT_STYLES[p.verdict] ?? VERDICT_STYLES.good)}
            >
              {p.expected}
            </button>
          )),
          ...(w.extra ?? []).map((sym, j) => (
            <span
              key={`${i}-extra-${j}`}
              title="Extra sound"
              className="rounded-lg bg-purple-50 px-2 py-1 font-mono text-sm text-purple-600 ring-1 ring-purple-200"
            >
              +{sym}
            </span>
          )),
        ])}
      </div>
    </div>
  )
}
