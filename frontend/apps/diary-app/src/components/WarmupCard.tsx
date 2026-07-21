import { useState } from 'react'
import { Button } from '@els/ui'
import type { Correction } from '../lib/types'
import { CorrectionDiff } from './CorrectionDiff'

export function WarmupCard({ items }: { items: Correction[] }) {
  const [answer, setAnswer] = useState('')
  const [revealed, setRevealed] = useState(false)
  const ok = revealed && items.every((it) => answer.toLowerCase().includes(it.correction.toLowerCase()))

  return (
    <div className="rounded-lg border border-neutral-200 bg-white p-4">
      <p className="text-sm text-neutral-500">You wrote earlier:</p>
      <p className="mt-1 text-neutral-800">
        <CorrectionDiff items={items} revealed={revealed} />
      </p>
      {!revealed ? (
        <div className="mt-3 flex gap-2">
          <input
            value={answer}
            onChange={(e) => setAnswer(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && answer && setRevealed(true)}
            placeholder="Rewrite the sentence correctly…"
            className="flex-1 rounded-md border border-neutral-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
          />
          <Button size="sm" onClick={() => setRevealed(true)} disabled={!answer.trim()}>
            Check
          </Button>
        </div>
      ) : (
        <div className={`mt-3 flex flex-col gap-1 rounded-md px-3 py-2 text-sm ${ok ? 'bg-emerald-50 text-emerald-800' : 'bg-amber-50 text-amber-800'}`}>
          {items.map((it, i) => (
            <p key={i}>
              {i === 0 && ok ? 'Exactly! ' : ''}
              {it.description}
            </p>
          ))}
        </div>
      )}
    </div>
  )
}
