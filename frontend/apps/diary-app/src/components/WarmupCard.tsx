import { useState } from 'react'
import { Button, cn } from '@els/ui'
import { CheckCircle2, Dumbbell, TriangleAlert } from 'lucide-react'
import type { Correction } from '../lib/types'
import { CorrectionDiff } from './CorrectionDiff'

export function WarmupCard({ items }: { items: Correction[] }) {
  const [answer, setAnswer] = useState('')
  const [revealed, setRevealed] = useState(false)
  const ok = revealed && items.every((it) => answer.toLowerCase().includes(it.correction.toLowerCase()))

  return (
    <div className="rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
      <p className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-neutral-400">
        <Dumbbell className="h-3.5 w-3.5" /> You wrote earlier
      </p>
      <p className="mt-2 leading-relaxed text-neutral-800">
        <CorrectionDiff items={items} revealed={revealed} />
      </p>
      {!revealed ? (
        <div className="mt-3 flex gap-2">
          <input
            value={answer}
            onChange={(e) => setAnswer(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && answer && setRevealed(true)}
            placeholder="Rewrite the sentence correctly…"
            className="flex-1 rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm outline-none transition-colors placeholder:text-neutral-400 focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
          <Button size="sm" variant="brand" onClick={() => setRevealed(true)} disabled={!answer.trim()}>
            Check
          </Button>
        </div>
      ) : (
        <div
          className={cn(
            'mt-3 flex gap-2.5 rounded-xl border px-3.5 py-3 text-sm',
            ok ? 'border-emerald-200 bg-emerald-50 text-emerald-800' : 'border-amber-200 bg-amber-50 text-amber-800',
          )}
        >
          {ok ? (
            <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-500" />
          ) : (
            <TriangleAlert className="mt-0.5 h-4 w-4 shrink-0 text-amber-500" />
          )}
          <div className="flex flex-col gap-1">
            {items.map((it, i) => (
              <p key={i}>
                {i === 0 && ok ? 'Exactly! ' : ''}
                {it.description}
              </p>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
