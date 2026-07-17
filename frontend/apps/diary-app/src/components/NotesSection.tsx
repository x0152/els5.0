import { useState } from 'react'
import { Badge } from '@els/ui'
import { ChevronDown, Sparkles } from 'lucide-react'
import type { Correction } from '../lib/types'
import { CorrectionDiff } from './CorrectionDiff'

export function NotesSection({ notes, nativeSample, defaultOpen = false }: { notes: Correction[]; nativeSample?: string; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <div className="rounded-lg border border-neutral-200 bg-white">
      <button
        onClick={() => setOpen(!open)}
        className="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-neutral-700"
      >
        <span className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-brand-600" />
          Language notes {notes.length > 0 && <Badge>{notes.length}</Badge>}
        </span>
        <ChevronDown className={`h-4 w-4 transition-transform ${open ? 'rotate-180' : ''}`} />
      </button>
      {open && (
        <div className="flex flex-col gap-4 border-t border-neutral-100 px-4 py-4">
          {notes.length === 0 ? (
            <p className="text-sm text-emerald-700">No corrections — great entry!</p>
          ) : (
            notes.map((n, i) => (
              <div key={i}>
                <p className="text-neutral-800">
                  <CorrectionDiff item={n} />
                </p>
                <p className="mt-1 text-sm text-neutral-500">{n.description}</p>
              </div>
            ))
          )}
          {nativeSample && (
            <div className="rounded-md bg-brand-50 px-3 py-3">
              <p className="text-xs font-medium uppercase tracking-wide text-brand-700">How a native speaker might put it</p>
              <p className="mt-1 text-sm italic leading-relaxed text-neutral-700">{nativeSample}</p>
            </div>
          )}
          {notes.length > 0 && (
            <p className="text-xs text-neutral-400">These corrections will come back in tomorrow's warm-up and join your review system.</p>
          )}
        </div>
      )}
    </div>
  )
}
