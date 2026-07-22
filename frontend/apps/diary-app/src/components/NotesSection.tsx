import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Badge, Spinner } from '@els/ui'
import { Check, ChevronDown, Plus, Sparkles } from 'lucide-react'
import { api } from '../lib/api'
import type { Correction } from '../lib/types'
import { CorrectionDiff } from './CorrectionDiff'

export function NotesSection({ notes, nativeSample, defaultOpen = false }: { notes: Correction[]; nativeSample?: string; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen)
  const [added, setAdded] = useState<Set<string>>(new Set())

  const addToStudio = useMutation({
    mutationFn: (text: string) =>
      api.studio.studioCaptureItem({ body: { text, area: 'My mistakes', icon: 'pen-line' } }),
    onSuccess: (_, text) => setAdded((prev) => new Set(prev).add(text)),
  })
  return (
    <div className="rounded-2xl border border-neutral-200 bg-white shadow-sm">
      <button
        onClick={() => setOpen(!open)}
        className="flex w-full items-center justify-between rounded-2xl px-5 py-4 text-sm font-semibold text-neutral-900 transition-colors hover:bg-neutral-50"
      >
        <span className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 items-center justify-center rounded-full bg-brand-50 text-brand-600">
            <Sparkles className="h-4 w-4" />
          </span>
          Language notes {notes.length > 0 && <Badge tone="brand">{notes.length}</Badge>}
        </span>
        <ChevronDown className={`h-4 w-4 text-neutral-400 transition-transform ${open ? 'rotate-180' : ''}`} />
      </button>
      {open && (
        <div className="flex flex-col gap-4 border-t border-neutral-100 px-5 py-4">
          {notes.length === 0 ? (
            <p className="text-sm text-emerald-700">No corrections — great entry!</p>
          ) : (
            notes.map((n, i) => (
              <div key={i} className="border-l-2 border-brand-200 pl-3.5">
                <p className="leading-relaxed text-neutral-800">
                  <CorrectionDiff items={[n]} />
                </p>
                <p className="mt-1 text-sm text-neutral-500">{n.description}</p>
                <button
                  onClick={() => addToStudio.mutate(n.correction)}
                  disabled={added.has(n.correction) || (addToStudio.isPending && addToStudio.variables === n.correction)}
                  className="mt-1.5 inline-flex items-center gap-1 rounded-full bg-brand-50 px-2.5 py-1 text-xs font-medium text-brand-700 ring-1 ring-brand-100 transition-colors hover:bg-brand-100 disabled:opacity-60"
                >
                  {added.has(n.correction) ? (
                    <>
                      <Check className="h-3 w-3" /> Added to Studio
                    </>
                  ) : addToStudio.isPending && addToStudio.variables === n.correction ? (
                    <>
                      <Spinner className="h-3 w-3" /> Adding…
                    </>
                  ) : (
                    <>
                      <Plus className="h-3 w-3" /> Train in Studio
                    </>
                  )}
                </button>
              </div>
            ))
          )}
          {nativeSample && (
            <div className="rounded-xl bg-gradient-to-br from-brand-50 to-emerald-50 px-4 py-3.5">
              <p className="text-xs font-semibold uppercase tracking-wide text-brand-700">How a native speaker might put it</p>
              <p className="mt-1.5 text-sm italic leading-relaxed text-neutral-700">{nativeSample}</p>
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
