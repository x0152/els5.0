import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Check, Loader2, Play, Plus, RotateCcw, Trash2, TriangleAlert } from 'lucide-react'
import { Button } from '@els/ui'
import { PracticeSession } from './PracticeSession.tsx'
import { mockCheckAnswer } from './check.ts'
import {
  PracticeMetaCtx,
  ProgressCtx,
  type PracticeAnswer,
  type PracticeApi,
  type PracticeKey,
  type PracticeKind,
  type PracticeVariant,
} from './state.ts'
import { PROSE_CLS } from './markdown.tsx'
import { BlocksProvider, type BlocksAdapters } from './Blocks.tsx'
import { BlockCtx } from './render/context.ts'
import { ExercisesList, Theory } from './render/exercises.tsx'
import { PracticeSkeleton } from './PracticeSheet.tsx'

export type BookSpreadProps = {
  heading?: string
  backLabel?: string
  onBack?: () => void
  theory: string
  exercises: string
  page: number
  footer?: string
  exercisesTitle?: string
  adapters?: BlocksAdapters
  practiceApi?: PracticeApi
  practiceKey?: PracticeKey
}

export function BookSpread({
  heading,
  backLabel = 'Back',
  onBack,
  theory,
  exercises,
  page,
  footer,
  exercisesTitle = 'Exercises',
  adapters = {},
  practiceApi,
  practiceKey,
}: BookSpreadProps) {
  const checkAnswer = adapters.check ?? mockCheckAnswer
  const meta = practiceApi && practiceKey ? { api: practiceApi, kind: practiceKey.kind, number: practiceKey.number } : null
  function scrollToTheory(section: string) {
    document.getElementById(`theory-${section}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }

  return (
    <BlocksProvider adapters={adapters}>
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-200">
      <div className="mx-auto w-full max-w-[1700px] p-3 md:p-6">
        {onBack && (
          <header className="mb-3 flex items-center">
            <Button variant="ghost" size="sm" onClick={onBack}>
              <ArrowLeft className="h-4 w-4" /> {backLabel}
            </Button>
          </header>
        )}

        <div className="grid rounded-xl shadow-2xl ring-1 ring-neutral-300 lg:grid-cols-2">
          <section className="flex flex-col rounded-t-xl border-neutral-200 bg-[#f4faf7] p-4 sm:p-6 lg:rounded-l-xl lg:rounded-tr-none lg:border-r lg:p-10">
            {heading && (
              <h1 className="mb-5 border-b border-brand-200/80 pb-3 text-xl font-bold tracking-tight text-brand-800 md:text-2xl">{heading}</h1>
            )}
            <div className={`flex-1 ${PROSE_CLS}`}>
              <BlockCtx.Provider value={{ dense: false, check: checkAnswer, onTheory: scrollToTheory }}>
                <Theory markdown={theory} />
              </BlockCtx.Provider>
            </div>
            <footer className="mt-6 flex items-end justify-between gap-4 border-t border-neutral-300/80 pt-3">
              <span className="text-lg font-semibold tabular-nums text-neutral-400">{page}</span>
              {footer && <span className="max-w-[75%] text-right text-[11px] leading-snug text-neutral-400">{footer}</span>}
            </footer>
          </section>

          <section className="flex flex-col rounded-b-xl bg-white p-4 sm:p-6 lg:rounded-r-xl lg:rounded-bl-none lg:p-10">
            <h2 className="mb-5 border-b border-brand-200/80 pb-3 text-xl font-bold tracking-tight text-brand-800 md:text-2xl">{exercisesTitle}</h2>
            <ExercisesSection mainExercises={exercises} meta={meta} checkAnswer={checkAnswer} onTheory={scrollToTheory} />
            <footer className="mt-6 flex justify-end border-t border-neutral-200 pt-3">
              <span className="text-lg font-semibold tabular-nums text-neutral-400">{page + 1}</span>
            </footer>
          </section>
        </div>
      </div>
    </div>
    </BlocksProvider>
  )
}

type Meta = { api: PracticeApi; kind: PracticeKind; number: number }
type CheckAnswerFn = typeof mockCheckAnswer

function ExercisesSection({
  mainExercises,
  meta,
  checkAnswer,
  onTheory,
}: {
  mainExercises: string
  meta: Meta | null
  checkAnswer: CheckAnswerFn
  onTheory: (s: string) => void
}) {
  if (!meta) {
    return (
      <div className="flex-1">
        <ExercisesList exercises={mainExercises} checkAnswer={checkAnswer} onTheory={onTheory} />
      </div>
    )
  }
  return <PracticeExercises mainExercises={mainExercises} meta={meta} checkAnswer={checkAnswer} onTheory={onTheory} />
}

function PracticeExercises({
  mainExercises,
  meta,
  checkAnswer,
  onTheory,
}: {
  mainExercises: string
  meta: Meta
  checkAnswer: CheckAnswerFn
  onTheory: (s: string) => void
}) {
  const qc = useQueryClient()
  const variantsKey = ['practice', 'variants', meta.kind, meta.number]
  const variantsQ = useQuery({
    queryKey: variantsKey,
    queryFn: () => meta.api.listVariants(meta.kind, meta.number),
    refetchInterval: (q) => (q.state.data?.some((v) => v.status === 'generating') ? 2000 : false),
  })
  const variants = variantsQ.data ?? []
  const [activeId, setActiveId] = useState('main')
  const active = activeId === 'main' ? undefined : variants.find((v) => v.id === activeId)

  const tabIds = ['main', ...variants.map((v) => v.id)]
  const progressQs = useQueries({
    queries: tabIds.map((id) => ({
      queryKey: ['practice', 'progress', meta.kind, meta.number, id],
      queryFn: () => meta.api.getProgress(meta.kind, meta.number, id),
      staleTime: 15_000,
    })),
  })
  const completed: Record<string, boolean> = {}
  tabIds.forEach((id, i) => {
    if (progressQs[i]?.data?.completed) completed[id] = true
  })

  const generate = useMutation({
    mutationFn: () => meta.api.generateVariant(meta.kind, meta.number),
    onSuccess: (v) => {
      qc.setQueryData<PracticeVariant[]>(variantsKey, (old = []) => [...old, v])
      setActiveId(v.id)
    },
  })
  const del = useMutation({
    mutationFn: (id: string) => meta.api.deleteVariant(id),
    onSuccess: (_d, id) => {
      qc.setQueryData<PracticeVariant[]>(variantsKey, (old = []) => old.filter((v) => v.id !== id))
      setActiveId((cur) => (cur === id ? 'main' : cur))
    },
  })

  return (
    <div className="flex-1">
      <VariantBar
        variants={variants}
        activeId={activeId}
        completed={completed}
        onSelect={setActiveId}
        onGenerate={() => generate.mutate()}
        onDelete={(id) => del.mutate(id)}
        generating={generate.isPending}
        deletingId={del.isPending ? (del.variables as string) : undefined}
      />
      {active?.status === 'error' ? (
        <div className="grid h-40 place-items-center gap-2 text-center text-sm text-rose-500">
          <TriangleAlert className="mx-auto h-6 w-6" />
          {active.error || 'Failed to generate exercises'}
        </div>
      ) : active?.status === 'generating' && !active.exercises ? (
        <PracticeSkeleton />
      ) : (
        <ProgressProvider key={activeId} meta={meta} variant={activeId} generating={active?.status === 'generating'}>
          <SessionOrList
            exercises={active ? active.exercises : mainExercises}
            generating={active?.status === 'generating'}
            checkAnswer={checkAnswer}
            onTheory={onTheory}
          />
        </ProgressProvider>
      )}
    </div>
  )
}

function SessionOrList({
  exercises,
  generating,
  checkAnswer,
  onTheory,
}: {
  exercises: string
  generating?: boolean
  checkAnswer: CheckAnswerFn
  onTheory: (s: string) => void
}) {
  const [session, setSession] = useState(false)
  if (session) {
    return <PracticeSession exercises={exercises} checkAnswer={checkAnswer} onTheory={onTheory} onExit={() => setSession(false)} />
  }
  return (
    <>
      {!generating && (
        <button
          type="button"
          onClick={() => setSession(true)}
          className="mb-4 flex w-full items-center gap-3 rounded-2xl border border-brand-200/80 bg-gradient-to-r from-brand-50 to-emerald-50/40 px-4 py-3 text-left shadow-sm transition-colors hover:border-brand-300 hover:from-brand-100/70"
        >
          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-brand-600 text-white shadow-md">
            <Play className="ml-0.5 h-4 w-4" />
          </span>
          <span className="min-w-0">
            <span className="block text-sm font-semibold text-brand-800">Practice as a session</span>
            <span className="block text-xs text-neutral-500">One exercise at a time, with a summary and retry at the end</span>
          </span>
        </button>
      )}
      <ExercisesList exercises={exercises} checkAnswer={checkAnswer} onTheory={onTheory} />
      {generating && (
        <div className="mt-6">
          <PracticeSkeleton count={1} />
        </div>
      )}
    </>
  )
}

function VariantBar({
  variants,
  activeId,
  completed,
  onSelect,
  onGenerate,
  onDelete,
  generating,
  deletingId,
}: {
  variants: PracticeVariant[]
  activeId: string
  completed: Record<string, boolean>
  onSelect: (id: string) => void
  onGenerate: () => void
  onDelete: (id: string) => void
  generating: boolean
  deletingId?: string
}) {
  const pages = [
    { id: 'main', title: 'Main', status: 'ready' as const },
    ...variants.map((v) => ({ id: v.id, title: v.title, status: v.status })),
  ]
  return (
    <div className="mb-4 flex flex-wrap items-center gap-1.5">
      {pages.map((p, i) => {
        const active = p.id === activeId
        const deleting = p.id === deletingId
        return (
          <span key={p.id} className="inline-flex items-center">
            <button
              type="button"
              onClick={() => onSelect(p.id)}
              className={`inline-flex h-7 items-center gap-1 rounded-full px-3 text-xs font-semibold transition-colors ${
                active ? 'bg-brand-600 text-white' : 'bg-neutral-100 text-neutral-500 hover:bg-neutral-200'
              }`}
            >
              {p.status === 'generating' && <Loader2 className="h-3 w-3 animate-spin" />}
              {completed[p.id] && <Check className={`h-3 w-3 ${active ? 'text-white' : 'text-emerald-500'}`} />}
              {i === 0 ? p.title : i + 1}
              {p.id !== 'main' && active && (
                <span
                  role="button"
                  tabIndex={0}
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete(p.id)
                  }}
                  className="-mr-1 ml-0.5 grid h-4 w-4 place-items-center rounded-full text-white/80 hover:bg-white/20 hover:text-white"
                >
                  {deleting ? <Loader2 className="h-3 w-3 animate-spin" /> : <Trash2 className="h-3 w-3" />}
                </span>
              )}
            </button>
          </span>
        )
      })}
      <button
        type="button"
        onClick={onGenerate}
        disabled={generating}
        title="Generate more exercises"
        className="inline-flex h-7 items-center gap-1 rounded-full border border-dashed border-brand-300 px-3 text-xs font-semibold text-brand-600 transition-colors hover:bg-brand-50 disabled:opacity-60"
      >
        {generating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
        {generating ? 'Generating…' : 'More'}
      </button>
    </div>
  )
}

function ProgressProvider({
  meta,
  variant,
  generating = false,
  children,
}: {
  meta: Meta
  variant: string
  generating?: boolean
  children: React.ReactNode
}) {
  const qc = useQueryClient()
  const progressKey = ['practice', 'progress', meta.kind, meta.number, variant]
  const progressQ = useQuery({
    queryKey: progressKey,
    queryFn: () => meta.api.getProgress(meta.kind, meta.number, variant),
  })

  const mapRef = useRef<Record<string, PracticeAnswer>>({})
  const registry = useRef<Set<string>>(new Set())
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [completed, setCompleted] = useState(false)
  const [ready, setReady] = useState(false)
  const [nonce, setNonce] = useState(0)
  const [version, setVersion] = useState(0)

  useEffect(() => {
    if (!progressQ.data) return
    mapRef.current = { ...progressQ.data.answers }
    setCompleted(progressQ.data.completed)
    setReady(true)
  }, [progressQ.data])

  const computeCompleted = useCallback(() => {
    if (generating) return false
    const keys = [...registry.current]
    return keys.length > 0 && keys.every((k) => mapRef.current[k]?.correct)
  }, [generating])

  const scheduleSave = useCallback(() => {
    if (saveTimer.current) clearTimeout(saveTimer.current)
    saveTimer.current = setTimeout(() => {
      const done = computeCompleted()
      setCompleted(done)
      qc.setQueryData(progressKey, { answers: mapRef.current, completed: done })
      void meta.api.saveProgress(meta.kind, meta.number, variant, { answers: mapRef.current, completed: done }).catch((err) => console.error(err))
    }, 600)
  }, [meta, variant, computeCompleted])

  useEffect(
    () => () => {
      if (saveTimer.current) clearTimeout(saveTimer.current)
    },
    [],
  )

  const value = useMemo(
    () => ({
      enabled: true,
      version,
      get: (k: string) => mapRef.current[k],
      set: (k: string, v: PracticeAnswer) => {
        mapRef.current = { ...mapRef.current, [k]: v }
        setVersion((n) => n + 1)
        scheduleSave()
      },
      register: (k: string) => {
        if (!registry.current.has(k)) {
          registry.current.add(k)
          setVersion((n) => n + 1)
        }
      },
      keys: () => [...registry.current],
      remove: (ks: string[]) => {
        if (!ks.length) return
        const next = { ...mapRef.current }
        for (const k of ks) delete next[k]
        mapRef.current = next
        setVersion((n) => n + 1)
        scheduleSave()
      },
    }),
    [scheduleSave, version],
  )

  function reset() {
    mapRef.current = {}
    registry.current = new Set()
    setCompleted(false)
    setNonce((n) => n + 1)
    qc.setQueryData(progressKey, { answers: {}, completed: false })
    void meta.api.resetProgress(meta.kind, meta.number, variant).catch((err) => console.error(err))
  }

  if (!ready) {
    return (
      <div className="grid h-40 place-items-center text-neutral-300">
        <Loader2 className="h-5 w-5 animate-spin" />
      </div>
    )
  }

  return (
    <PracticeMetaCtx.Provider value={meta}>
      <ProgressCtx.Provider value={value}>
        <div className="mb-3 flex items-center justify-between">
          <span className={`text-xs font-medium ${completed ? 'text-emerald-600' : 'text-neutral-400'}`}>
            {completed ? 'Completed' : 'In progress'}
          </span>
          <button
            type="button"
            onClick={reset}
            className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-neutral-400 transition-colors hover:bg-neutral-100 hover:text-neutral-700"
          >
            <RotateCcw className="h-3.5 w-3.5" /> Reset
          </button>
        </div>
        <div key={nonce}>{children}</div>
      </ProgressCtx.Provider>
    </PracticeMetaCtx.Provider>
  )
}
