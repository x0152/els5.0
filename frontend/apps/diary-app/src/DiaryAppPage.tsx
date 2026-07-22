import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { AppInfoButton, Button, ErrorState, LoadingState, Spinner, Textarea, cn } from '@els/ui'
import { CheckCircle2, Eye, Flame, GitCompareArrows, MessageCircleHeart, NotebookPen, Quote, TriangleAlert } from 'lucide-react'
import { api } from './lib/api'
import { diffWords } from './lib/diff'
import type { Correction, GrammarError } from './lib/types'
import { NotesSection } from './components/NotesSection'
import { WarmupCard } from './components/WarmupCard'

function GrammarErrorCard({ error }: { error: GrammarError }) {
  const [revealed, setRevealed] = useState(false)
  return (
    <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm">
      <p className="font-medium leading-relaxed text-amber-900">
        <span className="rounded bg-red-100 px-1 text-red-700 line-through decoration-2">{error.original}</span>
        {revealed && <span className="ml-2 rounded bg-emerald-100 px-1 font-medium text-emerald-700">{error.correction}</span>}
      </p>
      <p className="mt-1.5 text-amber-800">{error.explanation}</p>
      {!revealed && (
        <button
          type="button"
          onClick={() => setRevealed(true)}
          className="mt-2 inline-flex items-center gap-1.5 rounded-full bg-white px-2.5 py-1 text-xs font-medium text-amber-700 ring-1 ring-amber-200 transition-colors hover:bg-amber-100"
        >
          <Eye className="h-3.5 w-3.5" /> Show the fix
        </button>
      )}
    </div>
  )
}

function EntryDiff({ draft, text }: { draft: string; text: string }) {
  const tokens = diffWords(draft, text)
  return (
    <div className="rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
      <div className="flex items-center gap-2.5">
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-brand-50 text-brand-600">
          <GitCompareArrows className="h-4 w-4" />
        </div>
        <span className="text-sm font-semibold text-neutral-900">What you fixed</span>
      </div>
      <p className="mt-3 leading-relaxed text-neutral-800">
        {tokens.map((t, i) => (
          <span
            key={i}
            className={cn(
              t.kind === 'removed' && 'rounded bg-red-100 px-0.5 text-red-700 line-through decoration-2',
              t.kind === 'added' && 'rounded bg-emerald-100 px-0.5 font-medium text-emerald-700',
            )}
          >
            {t.text}
            {i < tokens.length - 1 && ' '}
          </span>
        ))}
      </p>
    </div>
  )
}

export function DiaryAppPage() {
  const [text, setText] = useState('')
  const [draft, setDraft] = useState<string | null>(null)
  const [errors, setErrors] = useState<GrammarError[] | null>(null)
  const qc = useQueryClient()

  const today = useQuery({
    queryKey: ['diary', 'today'],
    queryFn: () => api.diary.diaryToday(),
    refetchInterval: (q) => (q.state.data?.entry?.status === 'pending' ? 3000 : false),
  })

  const submit = useMutation({
    mutationFn: (body: { text: string; question?: string; draft?: string }) => api.diary.diarySubmitEntry({ body }),
    onSuccess: () => {
      setText('')
      setDraft(null)
      setErrors(null)
      void qc.invalidateQueries({ queryKey: ['diary'] })
    },
  })

  const check = useMutation({
    mutationFn: (value: string) => api.diary.diaryCheckEntry({ body: { text: value } }),
    onSuccess: (res, value) => {
      if (!res) return
      if (res.ok) {
        setErrors(null)
        submit.mutate({ text: value, question: today.data?.question, draft: draft ?? undefined })
      } else {
        setErrors(res.errors ?? [])
        if (draft === null) setDraft(value)
      }
    },
  })

  const busy = check.isPending || submit.isPending
  const wordCount = text.trim() ? text.trim().split(/\s+/).length : 0

  if (today.isError) {
    return (
      <ErrorState
        title="Failed to load the diary"
        description={isApiError(today.error) ? today.error.message : String(today.error)}
        action={<Button variant="secondary" onClick={() => today.refetch()}>Retry</Button>}
      />
    )
  }
  if (today.isPending || !today.data) return <LoadingState className="py-24" />

  const { question, entry, warmup, streak } = today.data
  const warmupGroups = [
    ...(warmup ?? []).reduce((m, it) => m.set(it.sentence, [...(m.get(it.sentence) ?? []), it]), new Map<string, Correction[]>()).values(),
  ]

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-8">
        <header className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
              <NotebookPen className="h-6 w-6" />
            </div>
            <div>
              <h1 className="flex items-center gap-1.5 text-2xl font-bold text-neutral-900">
                Diary <AppInfoButton />
              </h1>
              <p className="text-sm text-neutral-500">
                {new Date().toLocaleDateString('en-US', { weekday: 'long', day: 'numeric', month: 'long' })}
              </p>
            </div>
          </div>
          {streak > 0 && (
            <div className="flex items-center gap-1.5 rounded-full bg-gradient-to-r from-amber-400 to-orange-500 px-3.5 py-1.5 text-sm font-semibold text-white shadow-sm shadow-orange-500/25">
              <Flame className="h-4 w-4" /> {streak}-day streak
            </div>
          )}
        </header>

        {!entry && warmupGroups.length > 0 && (
          <section className="flex flex-col gap-3">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-500">Warm-up · your past corrections</h2>
            {warmupGroups.map((items, i) => (
              <WarmupCard key={i} items={items} />
            ))}
          </section>
        )}

        <section className="flex flex-col gap-3">
          <div className="relative overflow-hidden rounded-2xl border border-brand-200 bg-gradient-to-br from-brand-50 to-white p-5 shadow-sm">
            <Quote className="absolute -right-1 -top-1 h-16 w-16 -scale-x-100 text-brand-100" />
            <p className="text-xs font-semibold uppercase tracking-wide text-brand-700">Question of the day</p>
            <p className="relative mt-1.5 text-lg font-medium leading-snug text-neutral-900">{entry?.question || question}</p>
          </div>

          {!entry ? (
            <div className="flex flex-col gap-3 rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
              <Textarea
                value={text}
                onChange={(e) => setText(e.target.value)}
                rows={7}
                placeholder="Write 3–5 sentences in English. Fix any real grammar slips before sending — typos and commas don't count."
                disabled={busy}
              />
              {errors && errors.length > 0 && (
                <div className="flex flex-col gap-2">
                  <p className="flex items-center gap-1.5 text-sm font-medium text-amber-800">
                    <TriangleAlert className="h-4 w-4" /> Fix these before sending — edit your text above and check again:
                  </p>
                  {errors.map((e, i) => (
                    <GrammarErrorCard key={`${e.original}-${i}`} error={e} />
                  ))}
                </div>
              )}
              {(check.isError || submit.isError) && (
                <p className="text-sm text-red-600">
                  {check.isError
                    ? isApiError(check.error)
                      ? check.error.message
                      : 'Failed to check the entry'
                    : isApiError(submit.error)
                      ? submit.error.message
                      : 'Failed to submit the entry'}
                </p>
              )}
              <div className="flex items-center justify-between gap-3">
                <span className="text-xs text-neutral-400">{wordCount > 0 && `${wordCount} ${wordCount === 1 ? 'word' : 'words'}`}</span>
                <div className="flex items-center gap-3">
                  {check.isPending && (
                    <span className="flex items-center gap-2 text-sm text-neutral-500">
                      <Spinner className="h-4 w-4" /> Checking your grammar…
                    </span>
                  )}
                  {submit.isPending && (
                    <span className="flex items-center gap-2 text-sm text-neutral-500">
                      <Spinner className="h-4 w-4" /> Your friend is reading your entry…
                    </span>
                  )}
                  <Button variant="brand" onClick={() => check.mutate(text)} disabled={busy || text.trim().length < 10}>
                    {errors ? 'Check again & send' : 'Send'}
                  </Button>
                </div>
              </div>
            </div>
          ) : (
            <div className="whitespace-pre-wrap rounded-2xl border border-neutral-200 bg-white p-5 leading-relaxed text-neutral-800 shadow-sm">
              {entry.text}
            </div>
          )}
        </section>

        {entry && entry.status === 'pending' && (
          <section className="flex items-center gap-3 rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-rose-50 text-rose-500">
              <MessageCircleHeart className="h-4 w-4" />
            </div>
            <div className="min-w-0">
              <p className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
                <Spinner className="h-4 w-4" /> Your friend is reading your entry…
              </p>
              <p className="mt-0.5 text-sm text-neutral-500">
                Feel free to close the app — the reply will be waiting here when you come back.
              </p>
            </div>
          </section>
        )}

        {entry && entry.status !== 'pending' && (
          <section className="flex flex-col gap-3">
            {entry.draft && entry.draft !== entry.text && <EntryDiff draft={entry.draft} text={entry.text} />}
            <div className="rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
              <div className="flex items-center gap-2.5">
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-rose-50 text-rose-500">
                  <MessageCircleHeart className="h-4 w-4" />
                </div>
                <span className="text-sm font-semibold text-neutral-900">Friend's reply</span>
              </div>
              <p className="mt-3 whitespace-pre-wrap leading-relaxed text-neutral-800">{entry.reply}</p>
              {entry.next_question && (
                <p className="mt-4 rounded-xl bg-brand-50 px-3.5 py-2.5 text-sm text-brand-800">
                  Question for tomorrow: <span className="font-medium">{entry.next_question}</span>
                </p>
              )}
            </div>
            <NotesSection notes={entry.corrections ?? []} nativeSample={entry.native_sample} />
            <p className="flex items-center justify-center gap-1.5 py-2 text-sm text-neutral-400">
              <CheckCircle2 className="h-4 w-4 text-emerald-500" /> Today's entry is done — come back tomorrow.
            </p>
          </section>
        )}
      </div>
    </div>
  )
}
