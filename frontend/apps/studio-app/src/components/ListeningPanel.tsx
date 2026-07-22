import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Check, Headphones, Loader2, Sparkles, Turtle } from 'lucide-react'
import { SpeakButton, cn, speak } from '@els/ui'
import { api } from '../lib/api.ts'
import { alignWords, tokenize } from '../lib/diff.ts'
import type { Item } from '../lib/types.ts'

export function ListeningPanel({ item, onDone }: { item: Item; onDone: () => void }) {
  const queryClient = useQueryClient()
  const [sentence, setSentence] = useState('')
  const [attempt, setAttempt] = useState('')
  const [heard, setHeard] = useState<boolean[] | null>(null)

  const sentenceM = useMutation({
    mutationFn: () => api.studio.studioRegenExample({ params: { path: { id: item.id } } }),
    onSuccess: (data) => {
      if (!data?.example) return
      setSentence(data.example)
      setAttempt('')
      setHeard(null)
      speak(data.example)
    },
  })

  const check = () => {
    const result = alignWords(tokenize(sentence), tokenize(attempt))
    setHeard(result)
    queryClient.invalidateQueries({ queryKey: ['studio', 'items'] })
    if (result.every(Boolean)) onDone()
  }

  const words = sentence.split(/\s+/)
  let k = 0
  const marks = heard ? words.map((w) => (/[a-z]/i.test(w) ? (heard[k++] ?? true) : true)) : []

  const newSentenceButton = (
    <button
      onClick={() => sentenceM.mutate()}
      disabled={sentenceM.isPending}
      className="mt-auto inline-flex h-9 w-full shrink-0 items-center justify-center gap-2 rounded-lg bg-brand-600 px-4 text-sm font-semibold text-white shadow-sm shadow-brand-600/25 hover:bg-brand-700 disabled:opacity-50"
    >
      {sentenceM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
      {sentence ? 'Next sentence' : 'New sentence'}
    </button>
  )

  return (
    <div className="flex flex-[2] flex-col rounded-2xl border border-emerald-200 bg-white shadow-sm">
      <div className="flex items-center justify-between border-b border-neutral-100 px-5 py-3">
        <span className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-brand-50 text-brand-600">
            <Headphones className="h-4 w-4" />
          </span>
          Listening — dictation
        </span>
        {item.listened && (
          <span className="flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 ring-1 ring-emerald-200">
            <Check className="h-3 w-3" /> done
          </span>
        )}
      </div>
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-5 py-4">
        {!sentence ? (
          <>
            <p className="text-xs text-neutral-400">
              AI will make a fresh sentence with your phrase and read it aloud — type what you hear.
            </p>
            {newSentenceButton}
          </>
        ) : heard === null ? (
          <>
            <div className="flex flex-wrap items-center gap-2">
              <SpeakButton variant="button" text={sentence}>
                Replay
              </SpeakButton>
              <SpeakButton variant="ghost" icon={<Turtle className="h-4 w-4" />} text={sentence} rate={0.7}>
                Slow
              </SpeakButton>
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
                placeholder="Type what you hear…"
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
                {marks.every(Boolean) ? 'Perfect — every word caught.' : 'Missed words are highlighted:'}
              </p>
              <p className="mt-1 leading-relaxed text-neutral-900">
                {words.map((w, i) => (
                  <span key={i} className={cn(!marks[i] && 'rounded bg-red-100 px-0.5 font-medium text-red-700')}>
                    {w}{' '}
                  </span>
                ))}
              </p>
            </div>
            {newSentenceButton}
          </>
        )}
      </div>
    </div>
  )
}
