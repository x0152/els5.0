import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { Badge, Button, ErrorState, LoadingState, Spinner, Textarea } from '@els/ui'
import { Flame, MessageCircleHeart, NotebookPen } from 'lucide-react'
import { api } from './lib/api'
import { NotesSection } from './components/NotesSection'
import { WarmupCard } from './components/WarmupCard'

export function DiaryAppPage() {
  const [text, setText] = useState('')
  const qc = useQueryClient()

  const today = useQuery({
    queryKey: ['diary', 'today'],
    queryFn: () => api.diary.diaryToday(),
  })

  const submit = useMutation({
    mutationFn: (body: { text: string; question?: string }) => api.diary.diarySubmitEntry({ body }),
    onSuccess: () => {
      setText('')
      void qc.invalidateQueries({ queryKey: ['diary'] })
    },
  })

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
  const warmupItems = warmup ?? []

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-8">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
              <NotebookPen className="h-6 w-6 text-brand-600" /> Diary
            </h1>
            <p className="text-sm text-neutral-500">
              {new Date().toLocaleDateString('en-US', { weekday: 'long', day: 'numeric', month: 'long' })}
            </p>
          </div>
          {streak > 0 && (
            <Badge tone="warning" className="gap-1">
              <Flame className="h-3.5 w-3.5" /> {streak}-day streak
            </Badge>
          )}
        </header>

        {!entry && warmupItems.length > 0 && (
          <section className="flex flex-col gap-3">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-500">Warm-up · your past corrections</h2>
            {warmupItems.map((item, i) => (
              <WarmupCard key={i} item={item} />
            ))}
          </section>
        )}

        <section className="flex flex-col gap-3">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-500">Question of the day</h2>
          <div className="rounded-lg border border-brand-200 bg-brand-50 p-4">
            <p className="font-medium text-neutral-900">{entry?.question || question}</p>
          </div>

          {!entry ? (
            <>
              <Textarea
                value={text}
                onChange={(e) => setText(e.target.value)}
                rows={6}
                placeholder="Write 3–5 sentences in English. Don't worry about mistakes — they are reviewed separately."
                disabled={submit.isPending}
              />
              {submit.isError && (
                <p className="text-sm text-red-600">
                  {isApiError(submit.error) ? submit.error.message : 'Failed to submit the entry'}
                </p>
              )}
              <div className="flex items-center justify-end gap-3">
                {submit.isPending && (
                  <span className="flex items-center gap-2 text-sm text-neutral-500">
                    <Spinner className="h-4 w-4" /> Your friend is reading your entry…
                  </span>
                )}
                <Button variant="brand" onClick={() => submit.mutate({ text, question })} disabled={submit.isPending || text.trim().length < 10}>
                  Send
                </Button>
              </div>
            </>
          ) : (
            <div className="whitespace-pre-wrap rounded-lg border border-neutral-200 bg-white p-4 leading-relaxed text-neutral-800">
              {entry.text}
            </div>
          )}
        </section>

        {entry && (
          <section className="flex flex-col gap-3">
            <div className="rounded-lg border border-neutral-200 bg-white p-4">
              <div className="flex items-center gap-2 text-sm font-medium text-neutral-500">
                <MessageCircleHeart className="h-4 w-4 text-rose-500" /> Friend's reply
              </div>
              <p className="mt-2 whitespace-pre-wrap leading-relaxed text-neutral-800">{entry.reply}</p>
              {entry.next_question && (
                <p className="mt-3 rounded-md bg-neutral-50 px-3 py-2 text-sm text-neutral-600">
                  Question for tomorrow: <span className="font-medium text-neutral-800">{entry.next_question}</span>
                </p>
              )}
            </div>
            <NotesSection notes={entry.corrections ?? []} nativeSample={entry.native_sample} />
            <p className="text-center text-sm text-neutral-400">Today's entry is done — come back tomorrow.</p>
          </section>
        )}
      </div>
    </div>
  )
}
