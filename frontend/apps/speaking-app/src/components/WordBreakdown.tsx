import { cn } from '@els/ui'
import type { PhonemeResult, WordResult } from '../lib/types.ts'

const VERDICT_STYLES: Record<string, string> = {
  good: 'bg-emerald-50 text-emerald-700 ring-emerald-200 hover:bg-emerald-100',
  close: 'bg-amber-50 text-amber-700 ring-amber-300 hover:bg-amber-100',
  wrong: 'bg-red-50 text-red-700 ring-red-300 hover:bg-red-100',
  missing: 'bg-neutral-100 text-neutral-400 ring-neutral-200 line-through hover:bg-neutral-200',
}

interface Props {
  words: WordResult[]
  onSelect: (word: WordResult, phoneme: PhonemeResult) => void
}

export function WordBreakdown({ words, onSelect }: Props) {
  return (
    <div className="flex flex-wrap gap-3">
      {words.map((word, i) => (
        <div
          key={`${word.word}-${i}`}
          className={cn(
            'rounded-2xl border bg-white px-4 py-3 shadow-sm',
            word.score >= 85 ? 'border-neutral-200' : word.score >= 60 ? 'border-amber-300' : 'border-red-300',
          )}
        >
          <div className="mb-2 flex items-baseline justify-between gap-4">
            <span className="font-semibold text-neutral-900">{word.word}</span>
            <span
              className={cn(
                'text-xs font-medium tabular-nums',
                word.score >= 85 ? 'text-emerald-600' : word.score >= 60 ? 'text-amber-600' : 'text-red-600',
              )}
            >
              {word.score}
            </span>
          </div>
          <div className="flex flex-wrap items-center gap-1">
            {(word.phonemes ?? []).map((p, j) => (
              <button
                key={j}
                type="button"
                onClick={() => onSelect(word, p)}
                title={p.verdict === 'good' ? `/${p.expected}/` : `expected /${p.expected}/, heard /${p.heard ?? '—'}/`}
                className={cn(
                  'rounded-lg px-2 py-1 font-mono text-sm ring-1 transition',
                  VERDICT_STYLES[p.verdict] ?? VERDICT_STYLES.good,
                )}
              >
                {p.expected}
              </button>
            ))}
            {(word.extra ?? []).map((sym, j) => (
              <span
                key={`extra-${j}`}
                title="Extra sound not in the word"
                className="rounded-lg bg-purple-50 px-2 py-1 font-mono text-sm text-purple-600 ring-1 ring-purple-200"
              >
                +{sym}
              </span>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
