import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Brain, Check, Eye, X } from 'lucide-react'
import { cn } from '@els/ui'
import { api } from '../lib/api.ts'
import { alignWords, tokenize } from '../lib/diff.ts'
import type { Item } from '../lib/types.ts'

export function RecallPanel({
  item,
  hidden,
  onHiddenChange,
  onDone,
}: {
  item: Item
  hidden: boolean
  onHiddenChange: (hidden: boolean) => void
  onDone: () => void
}) {
  const [attempt, setAttempt] = useState('')
  const [result, setResult] = useState<boolean[] | null>(null)

  const meQ = useQuery({
    queryKey: ['studio', 'me'],
    queryFn: () => api.account.accountMe(),
    staleTime: 10_000,
  })
  const showTranslations = meQ.data?.show_translations ?? true

  const prompt =
    (showTranslations ? item.translation || item.explanation_native : '') ||
    item.explanation ||
    'No hint on this card — write the phrase you were just studying.'

  const start = () => {
    setAttempt('')
    setResult(null)
    onHiddenChange(true)
  }

  const check = () => {
    const ref = tokenize(item.text)
    const got = tokenize(attempt)
    const marks = alignWords(ref, got)
    setResult(marks)
    onHiddenChange(false)
    if (marks.every(Boolean) && ref.length === got.length) onDone()
  }

  const words = item.text.split(/\s+/)
  let k = 0
  const marks = result ? words.map((w) => (/[a-z]/i.test(w) ? (result[k++] ?? true) : true)) : []
  const perfect = result !== null && marks.every(Boolean) && tokenize(attempt).length === tokenize(item.text).length

  return (
    <div className="flex flex-1 flex-col rounded-2xl border border-emerald-200 bg-white shadow-sm">
      <div className="flex items-center justify-between border-b border-neutral-100 px-4 py-3">
        <span className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-brand-50 text-brand-600">
            <Brain className="h-4 w-4" />
          </span>
          Recall
        </span>
        {item.recalled && (
          <span className="flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 ring-1 ring-emerald-200">
            <Check className="h-3 w-3" /> done
          </span>
        )}
      </div>
      <div className="flex min-h-0 flex-1 flex-col gap-2.5 overflow-y-auto p-4">
        {!hidden && result === null ? (
          <>
            <p className="text-xs text-neutral-400">
              The card hides and you get a hint — write the phrase from memory.
            </p>
            <button
              onClick={start}
              className="mt-auto inline-flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-white px-4 text-sm font-semibold text-neutral-800 shadow-sm ring-1 ring-inset ring-neutral-200 hover:bg-neutral-50"
            >
              <Eye className="h-4 w-4" /> Hide &amp; recall
            </button>
          </>
        ) : hidden ? (
          <>
            <div className="flex items-start gap-1.5 rounded-xl bg-neutral-50 px-3 py-2">
              <div className="min-w-0 flex-1">
                <p className="text-xs font-semibold uppercase tracking-wide text-neutral-400">
                  Hint — say it in English
                </p>
                <p className="mt-1 text-sm font-medium leading-relaxed text-neutral-800">{prompt}</p>
              </div>
              <button
                onClick={() => onHiddenChange(false)}
                title="Cancel and show the card"
                className="shrink-0 rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-700"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
            <form
              className="mt-auto flex items-center gap-2"
              onSubmit={(e) => {
                e.preventDefault()
                if (attempt.trim()) check()
              }}
            >
              <input
                value={attempt}
                onChange={(e) => setAttempt(e.target.value)}
                autoFocus
                placeholder="Write the phrase in English…"
                className="w-full rounded-lg border border-neutral-200 px-3 py-2 text-sm placeholder:text-neutral-400 focus:border-brand-400 focus:outline-none focus:ring-2 focus:ring-brand-100"
              />
              <button
                type="submit"
                disabled={!attempt.trim()}
                className="inline-flex h-9 shrink-0 items-center rounded-lg bg-brand-600 px-4 text-sm font-semibold text-white shadow-sm shadow-brand-600/25 hover:bg-brand-700 disabled:opacity-50"
              >
                Check
              </button>
            </form>
          </>
        ) : (
          <>
            <div className="min-w-0">
              <p className="text-xs text-neutral-400">
                {perfect ? 'Perfect — recalled word for word.' : 'Not quite — missed words are highlighted:'}
              </p>
              <p className="mt-1 leading-relaxed text-neutral-900">
                {words.map((w, i) => (
                  <span key={i} className={cn(!marks[i] && 'rounded bg-red-100 px-0.5 font-medium text-red-700')}>
                    {w}{' '}
                  </span>
                ))}
              </p>
            </div>
            <button
              onClick={start}
              className="mt-auto inline-flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-brand-600 px-4 text-sm font-semibold text-white shadow-sm shadow-brand-600/25 hover:bg-brand-700"
            >
              Try again
            </button>
          </>
        )}
      </div>
    </div>
  )
}
