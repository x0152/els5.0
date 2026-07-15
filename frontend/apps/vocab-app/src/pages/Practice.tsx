import { useEffect, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Eye, EyeOff, Loader2, RotateCcw } from 'lucide-react'
import { Button, useAgentView } from '@els/ui'
import { PracticeSheet, PracticeSkeleton, type PracticeAnswer } from '@els/blocks'
import { isApiError } from '@els/api-client'
import { api } from '../lib/api.ts'
import { produce } from '../lib/events.ts'
import { useShowTranslations } from '../store/me.ts'
import type { Unit } from '../lib/types.ts'

const QUERY_KEY = ['vocab', 'practice']

const checkFree = async (input: { instruction: string; answer: string }) => {
  const r = await api.vocab.checkVocabPractice({ body: input })
  return { correct: !!r?.correct, correction: r?.correction ?? '', explanation: r?.explanation ?? '' }
}

export function Practice() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [showWords, setShowWords] = useState(false)

  const sessionQ = useQuery({
    queryKey: QUERY_KEY,
    queryFn: () => api.vocab.getVocabPractice({}),
    refetchInterval: (q) => (q.state.data?.status === 'generating' ? 2000 : false),
    refetchOnWindowFocus: false,
    retry: false,
  })

  const generate = useMutation({
    mutationFn: () => api.vocab.generateVocabPractice({}),
    onSuccess: (s) => qc.setQueryData(QUERY_KEY, s),
  })

  const notFound = isApiError(sessionQ.error) && sessionQ.error.status === 404

  useEffect(() => {
    if (notFound && !generate.isPending && !generate.isError) generate.mutate()
  }, [notFound, generate])

  const session = sessionQ.data
  const words = (session?.words ?? []) as Unit[]

  useAgentView({
    app: 'vocab',
    screen: 'practice',
    info: 'The user is practicing words with learning status from their collection.',
    state: { words: words.length, status: session?.status ?? 'loading' },
  })

  const saveProgress = (answers: Record<string, PracticeAnswer>, completed: boolean) => {
    if (!session) return
    void api.vocab
      .saveVocabPracticeProgress({ body: { session_id: session.id, answers, completed } })
      .catch((e) => console.error('failed to save practice progress', e))
  }

  if (generate.isError) {
    const message = isApiError(generate.error) ? generate.error.message : 'Could not generate practice. Try again.'
    return (
      <Shell onBack={() => navigate('..')}>
        <div className="grid place-items-center rounded-3xl bg-white py-20 text-center ring-1 ring-neutral-200">
          <h3 className="text-lg font-semibold text-neutral-900">Not enough words to practice</h3>
          <p className="mx-auto mt-1 max-w-sm text-sm text-neutral-500">{message}</p>
          <Button variant="brand" className="mt-6" onClick={() => generate.mutate()}>
            <RotateCcw className="h-4 w-4" /> Try again
          </Button>
        </div>
      </Shell>
    )
  }

  if (!session) {
    if (sessionQ.isError && !notFound) {
      return (
        <Shell onBack={() => navigate('..')}>
          <div className="grid place-items-center rounded-3xl bg-white py-20 text-center ring-1 ring-neutral-200">
            <h3 className="text-lg font-semibold text-neutral-900">Could not load practice</h3>
            <Button variant="brand" className="mt-6" onClick={() => sessionQ.refetch()}>
              <RotateCcw className="h-4 w-4" /> Retry
            </Button>
          </div>
        </Shell>
      )
    }
    return (
      <Shell onBack={() => navigate('..')}>
        <div className="rounded-3xl bg-white p-6 ring-1 ring-neutral-200">
          <PracticeSkeleton />
        </div>
      </Shell>
    )
  }

  const busy = generate.isPending || session.status === 'generating'

  return (
    <Shell onBack={() => navigate('..')} aside={showWords ? <WordList units={words} /> : undefined}>
      <div className="mb-4 flex items-center justify-end gap-3">
        {session.status === 'generating' && (
          <span className="inline-flex items-center gap-1.5 text-sm text-neutral-400">
            <Loader2 className="h-4 w-4 animate-spin" /> Generating…
          </span>
        )}
        <Button variant="secondary" onClick={() => setShowWords((v) => !v)}>
          {showWords ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />} {showWords ? 'Hide words' : 'Show words'}
        </Button>
        <Button variant="secondary" disabled={busy} onClick={() => generate.mutate()}>
          {generate.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RotateCcw className="h-4 w-4" />} New session
        </Button>
      </div>
      <div className="rounded-3xl bg-white p-6 ring-1 ring-neutral-200">
        {session.status === 'error' ? (
          <p className="py-10 text-center text-sm text-rose-500">{session.error || 'Generation failed.'}</p>
        ) : session.exercises ? (
          <>
            <PracticeSheet
              key={session.id}
              exercises={session.exercises}
              adapters={{
                checkFree,
                produce,
                progress: { answers: session.answers as Record<string, PracticeAnswer>, onChange: saveProgress },
              }}
            />
            {session.status === 'generating' && (
              <div className="mt-6">
                <PracticeSkeleton count={1} />
              </div>
            )}
          </>
        ) : (
          <PracticeSkeleton />
        )}
      </div>
    </Shell>
  )
}

function Shell({ children, onBack, aside }: { children: ReactNode; onBack: () => void; aside?: ReactNode }) {
  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className={`mx-auto space-y-4 p-6 ${aside ? 'max-w-6xl' : 'max-w-2xl'}`}>
        <button onClick={onBack} className="inline-flex items-center gap-1.5 text-sm font-medium text-neutral-500 hover:text-neutral-800">
          <ArrowLeft className="h-4 w-4" /> Back to vocabulary
        </button>
        <h1 className="text-2xl font-bold text-neutral-900">Practice</h1>
        {aside ? (
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
            <div>{children}</div>
            {aside}
          </div>
        ) : (
          children
        )}
      </div>
    </div>
  )
}

function WordList({ units }: { units: Unit[] }) {
  const showTranslations = useShowTranslations()
  return (
    <aside
      onCopy={(e) => e.preventDefault()}
      onCut={(e) => e.preventDefault()}
      onContextMenu={(e) => e.preventDefault()}
      onDragStart={(e) => e.preventDefault()}
      className="select-none self-start rounded-3xl bg-white p-5 ring-1 ring-neutral-200 lg:sticky lg:top-6"
    >
      <p className="mb-3 text-xs font-medium uppercase tracking-wide text-neutral-400">Words in this session</p>
      <ul className="max-h-[70vh] space-y-3 overflow-y-auto pr-1">
        {units.map((u) => (
          <li key={u.id} className="border-b border-neutral-100 pb-3 last:border-0 last:pb-0">
            <div className="flex items-baseline gap-2">
              <span className="font-semibold text-neutral-900">{u.text}</span>
              {u.transcription && <span className="text-xs text-neutral-400">/{u.transcription}/</span>}
            </div>
            {(u.definition || (showTranslations && u.translation)) && (
              <p className="mt-0.5 text-sm text-neutral-500">{u.definition || u.translation}</p>
            )}
          </li>
        ))}
      </ul>
    </aside>
  )
}
