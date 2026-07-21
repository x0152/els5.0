import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, ErrorState, LoadingState, Mascot, cn } from '@els/ui'
import { ArrowLeft, Check } from 'lucide-react'
import { api } from './lib/api.ts'
import { STEP_META, stepSubtitle, stepTitle } from './lib/steps.ts'
import type { Lesson, Step, StepOutcome } from './lib/types.ts'
import { GrammarStep } from './components/GrammarStep.tsx'
import { PhrasesStep } from './components/PhrasesStep.tsx'
import { QuestionsStep } from './components/QuestionsStep.tsx'
import { ReadingStep } from './components/ReadingStep.tsx'
import { VocabStep } from './components/VocabStep.tsx'
import { WatchStep } from './components/WatchStep.tsx'
import { WritingStep } from './components/WritingStep.tsx'

export function LessonPage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeId, setActiveId] = useState<string | null>(null)

  const lessonQ = useQuery({
    queryKey: ['workout-lesson', id],
    queryFn: () => api.workout.workoutGetLesson({ params: { path: { id } } }),
  })
  const lesson = lessonQ.data as Lesson | undefined

  const submit = useMutation({
    mutationFn: ({ stepId, outcome }: { stepId: string; outcome: StepOutcome }) =>
      api.workout.workoutSubmitStep({
        params: { path: { id, step: stepId } },
        body: { score: outcome.score, results: outcome.results },
      }),
    onSuccess: (updated) => {
      queryClient.setQueryData(['workout-lesson', id], updated)
      void queryClient.invalidateQueries({ queryKey: ['workout-today'] })
      setActiveId(null)
    },
  })

  const current = useMemo(() => {
    if (!lesson) return undefined
    if (activeId) return lesson.steps.find((s) => s.id === activeId)
    return lesson.steps.find((s) => !s.done)
  }, [lesson, activeId])

  if (lessonQ.isPending) return <LoadingState className="h-full" />
  if (lessonQ.isError || !lesson)
    return (
      <ErrorState
        className="h-full"
        action={
          <Button variant="secondary" onClick={() => lessonQ.refetch()}>
            Retry
          </Button>
        }
      />
    )

  const doneCount = lesson.steps.filter((s) => s.done).length
  const finished = lesson.status === 'completed'
  const currentIdx = current ? lesson.steps.indexOf(current) : -1

  const onDone = (step: Step) => (outcome: StepOutcome) => submit.mutate({ stepId: step.id, outcome })

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex min-h-full w-full max-w-3xl flex-col gap-5 px-4 py-6 sm:px-6 lg:max-w-5xl">
        <header className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => navigate('..')}>
            <ArrowLeft className="h-5 w-5" />
          </Button>
          <div className="min-w-0 flex-1">
            <h1 className="text-lg font-bold text-neutral-900">
              Lesson {lesson.number}
              {lesson.review && <span className="ml-2 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-800">Review day</span>}
            </h1>
          </div>
          <span className="text-sm font-medium text-neutral-400">
            {doneCount}/{lesson.steps.length}
          </span>
        </header>

        <StepTrack steps={lesson.steps} currentId={current?.id} onSelect={(s) => s.done && setActiveId(s.id)} />

        {finished && !current ? (
          <Finished lesson={lesson} onHome={() => navigate('..')} onReplay={(s) => setActiveId(s.id)} />
        ) : current ? (
          <div key={current.id} className="flex flex-1 flex-col gap-4">
            <StepHero step={current} index={currentIdx} total={lesson.steps.length} />
            <StepView step={current} onDone={onDone(current)} />
          </div>
        ) : null}
      </div>
    </div>
  )
}

function StepTrack({ steps, currentId, onSelect }: { steps: Step[]; currentId?: string; onSelect: (s: Step) => void }) {
  return (
    <div className="flex items-center">
      {steps.map((s, i) => {
        const meta = STEP_META[s.kind]
        const active = s.id === currentId
        return (
          <div key={s.id} className={cn('flex items-center', i > 0 && 'flex-1')}>
            {i > 0 && <span className={cn('h-px flex-1', s.done || active ? 'bg-brand-300' : 'bg-neutral-200')} />}
            <button
              type="button"
              title={meta.label}
              onClick={() => onSelect(s)}
              className={cn(
                'grid h-9 w-9 shrink-0 place-items-center rounded-full transition-all',
                s.done
                  ? 'bg-emerald-500 text-white shadow-sm'
                  : active
                    ? `bg-gradient-to-br ${meta.grad} scale-110 text-white shadow-md`
                    : 'bg-white text-neutral-300 ring-1 ring-neutral-200',
              )}
            >
              {s.done ? <Check className="h-4 w-4" /> : <meta.icon className="h-4 w-4" />}
            </button>
          </div>
        )
      })}
    </div>
  )
}

function StepHero({ step, index, total }: { step: Step; index: number; total: number }) {
  const meta = STEP_META[step.kind]
  return (
    <section className={cn('relative overflow-hidden rounded-3xl bg-gradient-to-br p-6 text-white shadow-lg sm:p-7', meta.grad)}>
      <div className="absolute -right-10 -top-12 h-40 w-40 rounded-full bg-white/10" />
      <div className="absolute -bottom-14 right-20 h-32 w-32 rounded-full bg-white/5" />
      <meta.icon className="absolute bottom-4 right-5 h-16 w-16 text-white/20" />
      <div className="relative pr-16">
        <p className="text-[11px] font-bold uppercase tracking-widest text-white/70">
          Step {index + 1} of {total} · {meta.label}
        </p>
        <h2 className="mt-1 text-2xl font-extrabold tracking-tight">{stepTitle(step)}</h2>
        <p className="mt-1.5 max-w-xl text-sm leading-relaxed text-white/85">{stepSubtitle(step)}</p>
      </div>
    </section>
  )
}

function StepView({ step, onDone }: { step: Step; onDone: (outcome: StepOutcome) => void }) {
  switch (step.kind) {
    case 'warmup':
      return <PhrasesStep items={step.warmup ?? []} onDone={onDone} />
    case 'watch':
      return step.watch ? <WatchStep watch={step.watch} onDone={(score) => onDone({ score })} /> : null
    case 'questions':
      return <QuestionsStep questions={step.questions ?? []} onDone={(score) => onDone({ score })} />
    case 'speak':
      return <PhrasesStep items={(step.phrases ?? []).map((p) => ({ ...p, mode: 'speak' as const }))} onDone={onDone} />
    case 'dictation':
      return <PhrasesStep items={(step.phrases ?? []).map((p) => ({ ...p, mode: 'dictation' as const }))} onDone={onDone} />
    case 'reading':
      return step.reading ? <ReadingStep reading={step.reading} onDone={onDone} /> : null
    case 'writing':
      return step.writing ? <WritingStep writing={step.writing} onDone={onDone} /> : null
    case 'grammar':
      return step.grammar ? <GrammarStep grammar={step.grammar} onDone={onDone} /> : null
    case 'vocab':
      return <VocabStep words={step.vocab ?? []} onDone={onDone} />
  }
}

function Finished({ lesson, onHome, onReplay }: { lesson: Lesson; onHome: () => void; onReplay: (s: Step) => void }) {
  const avg = lesson.steps.length ? Math.round(lesson.steps.reduce((s, x) => s + x.score, 0) / lesson.steps.length) : 0
  return (
    <section className="flex flex-col items-center gap-5 rounded-3xl border border-neutral-200 bg-gradient-to-b from-emerald-50/70 to-white p-8 text-center shadow-sm">
      <Mascot className="h-28 w-28" />
      <div>
        <h2 className="text-2xl font-extrabold text-neutral-900">Workout complete!</h2>
        <p className="mt-1 text-sm text-neutral-500">Great job — your streak keeps burning. The next lesson will be waiting for you.</p>
      </div>
      <div className="rounded-2xl border border-neutral-200 bg-white px-6 py-3 shadow-sm">
        <span className={cn('text-2xl font-bold tabular-nums', avg >= 70 ? 'text-emerald-600' : 'text-amber-500')}>{avg}</span>
        <span className="block text-[10px] font-semibold uppercase tracking-wider text-neutral-400">Average score</span>
      </div>
      <div className="flex w-full max-w-md flex-col gap-1.5">
        {lesson.steps.map((s) => {
          const meta = STEP_META[s.kind]
          return (
            <button
              key={s.id}
              type="button"
              onClick={() => onReplay(s)}
              title="Replay this step"
              className="flex items-center gap-3 rounded-xl border border-neutral-200/80 bg-white px-3 py-2 text-left transition-colors hover:bg-neutral-50"
            >
              <span className={cn('grid h-7 w-7 shrink-0 place-items-center rounded-lg', meta.chip)}>
                <meta.icon className="h-3.5 w-3.5" />
              </span>
              <span className="flex-1 truncate text-sm font-medium text-neutral-700">{stepTitle(s)}</span>
              <span className={cn('text-sm font-bold tabular-nums', s.score >= 70 ? 'text-emerald-600' : 'text-rose-500')}>{s.score}</span>
            </button>
          )
        })}
      </div>
      <Button variant="brand" onClick={onHome}>
        Back to calendar
      </Button>
    </section>
  )
}
