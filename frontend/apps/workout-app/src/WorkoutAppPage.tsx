import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { Button, ErrorState, LoadingState, Mascot, Spinner, cn } from '@els/ui'
import { CalendarDays, Check, Dumbbell, Flame, Play, Sparkles, Trophy } from 'lucide-react'
import { api } from './lib/api.ts'
import { STEP_META, stepDetail } from './lib/steps.ts'
import type { Lesson, Step } from './lib/types.ts'

export function WorkoutAppPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const todayQ = useQuery({
    queryKey: ['workout-today'],
    queryFn: () => api.workout.workoutToday(),
  })

  const start = useMutation({
    mutationFn: () => api.workout.workoutStartLesson(),
    onSuccess: (lesson) => {
      void queryClient.invalidateQueries({ queryKey: ['workout-today'] })
      if (lesson) navigate(`lesson/${lesson.id}`)
    },
  })

  if (todayQ.isPending) return <LoadingState className="h-full" />
  if (todayQ.isError)
    return (
      <ErrorState
        className="h-full"
        action={
          <Button variant="secondary" onClick={() => todayQ.refetch()}>
            Retry
          </Button>
        }
      />
    )

  const today = todayQ.data
  const lesson = today?.lesson as Lesson | undefined
  const streak = today?.streak ?? 0
  const dayList = today?.days ?? []
  const days = new Set(dayList)
  const completedToday = today?.completed ?? false

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex max-w-2xl flex-col gap-5 px-4 py-8 sm:px-6 lg:max-w-3xl">
        <Hero streak={streak} completedToday={completedToday} />
        <StatsRow streak={streak} dayList={dayList} />

        {completedToday && (!lesson || !lesson.steps.some((s) => s.done)) ? (
          <DoneToday lesson={lesson} onOpen={(id) => navigate(`lesson/${id}`)} />
        ) : lesson ? (
          <TodayLesson lesson={lesson} onOpen={() => navigate(`lesson/${lesson.id}`)} />
        ) : (
          <StartCard
            starting={start.isPending}
            error={start.isError ? (isApiError(start.error) ? start.error.message : 'Failed to build the lesson') : undefined}
            onStart={() => start.mutate()}
          />
        )}

        <Calendar days={days} />
      </div>
    </div>
  )
}

function greeting(): string {
  const h = new Date().getHours()
  if (h < 6) return 'Burning the midnight oil'
  if (h < 12) return 'Good morning'
  if (h < 18) return 'Good afternoon'
  return 'Good evening'
}

function Hero({ streak, completedToday }: { streak: number; completedToday: boolean }) {
  const dateLine = new Date().toLocaleDateString('en-US', { weekday: 'long', day: 'numeric', month: 'long' })
  return (
    <section className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-brand-600 via-brand-700 to-emerald-900 p-6 text-white shadow-lg sm:p-8">
      <div className="absolute -right-10 -top-14 h-44 w-44 rounded-full bg-white/10" />
      <div className="absolute -bottom-16 right-24 h-36 w-36 rounded-full bg-white/5" />
      <Dumbbell className="absolute right-6 top-6 h-10 w-10 text-white/20" />
      <div className="relative">
        <p className="text-[11px] font-bold uppercase tracking-widest text-white/60">{dateLine}</p>
        <h1 className="mt-1 text-3xl font-extrabold tracking-tight">{greeting()}.</h1>
        <p className="mt-1.5 max-w-md text-sm text-white/80">
          {completedToday
            ? 'Today’s workout is in the bag — the streak lives another day.'
            : 'One lesson a day: watch, answer, speak, write. Twenty minutes, all four skills.'}
        </p>
        {streak > 0 && (
          <div className="mt-4 inline-flex items-center gap-2 rounded-full bg-white/15 px-4 py-1.5 backdrop-blur-sm">
            <Flame className="h-4 w-4 text-orange-300" />
            <span className="text-sm font-bold">{streak}-day streak</span>
          </div>
        )}
      </div>
    </section>
  )
}

function bestStreak(dayList: string[]): number {
  const sorted = [...new Set(dayList)].sort()
  let best = 0
  let cur = 0
  let prev = ''
  for (const key of sorted) {
    const prevDate = new Date(key)
    prevDate.setDate(prevDate.getDate() - 1)
    cur = prev === dateKey(prevDate) ? cur + 1 : 1
    best = Math.max(best, cur)
    prev = key
  }
  return best
}

function StatsRow({ streak, dayList }: { streak: number; dayList: string[] }) {
  const monthPrefix = dateKey(new Date()).slice(0, 7)
  const stats = [
    { label: 'Day streak', value: streak, icon: Flame, cls: 'bg-orange-100 text-orange-600' },
    { label: 'This month', value: dayList.filter((d) => d.startsWith(monthPrefix)).length, icon: CalendarDays, cls: 'bg-sky-100 text-sky-600' },
    { label: 'Best streak', value: bestStreak(dayList), icon: Trophy, cls: 'bg-amber-100 text-amber-600' },
    { label: 'Total lessons', value: dayList.length, icon: Dumbbell, cls: 'bg-brand-100 text-brand-700' },
  ]
  return (
    <section className="grid grid-cols-2 gap-3 sm:grid-cols-4">
      {stats.map((s) => (
        <div key={s.label} className="flex items-center gap-3 rounded-2xl border border-neutral-200 bg-white p-3.5 shadow-sm">
          <span className={cn('grid h-9 w-9 shrink-0 place-items-center rounded-xl', s.cls)}>
            <s.icon className="h-4.5 w-4.5" />
          </span>
          <span>
            <span className="block text-lg font-bold leading-tight tabular-nums text-neutral-900">{s.value}</span>
            <span className="block text-[11px] font-medium text-neutral-400">{s.label}</span>
          </span>
        </div>
      ))}
    </section>
  )
}

function TodayLesson({ lesson, onOpen }: { lesson: Lesson; onOpen: () => void }) {
  const done = lesson.steps.filter((s) => s.done).length
  const started = done > 0
  return (
    <section className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm">
      <div className="flex items-center gap-4">
        <div className="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-brand-600 text-white shadow-md">
          <Play className="ml-0.5 h-6 w-6" />
        </div>
        <div className="min-w-0 flex-1">
          <h2 className="text-lg font-bold text-neutral-900">
            Lesson {lesson.number}
            {lesson.review && <span className="ml-2 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-800">Review day</span>}
          </h2>
          <p className="text-sm text-neutral-500">
            {started ? `${done} of ${lesson.steps.length} steps done — keep going.` : `${lesson.steps.length} steps, all four skills.`}
          </p>
        </div>
        <Button variant="brand" onClick={onOpen}>
          {started ? 'Continue' : 'Start'}
        </Button>
      </div>

      {started && (
        <div className="mt-4 h-1.5 overflow-hidden rounded-full bg-neutral-100">
          <div className="h-full rounded-full bg-brand-500 transition-[width]" style={{ width: `${(done / lesson.steps.length) * 100}%` }} />
        </div>
      )}

      <ol className="mt-5 space-y-0.5">
        {lesson.steps.map((s, i) => (
          <StepRow key={s.id} step={s} last={i === lesson.steps.length - 1} active={!s.done && lesson.steps.findIndex((x) => !x.done) === i} />
        ))}
      </ol>
    </section>
  )
}

function StepRow({ step, last, active }: { step: Step; last: boolean; active: boolean }) {
  const meta = STEP_META[step.kind]
  const detail = stepDetail(step)
  return (
    <li className="flex gap-3">
      <div className="flex flex-col items-center">
        <span
          className={cn(
            'grid h-8 w-8 shrink-0 place-items-center rounded-full',
            step.done ? 'bg-emerald-500 text-white' : active ? `${meta.chip} ring-2 ring-current` : meta.chip,
            !step.done && !active && 'opacity-60',
          )}
        >
          {step.done ? <Check className="h-4 w-4" /> : <meta.icon className="h-4 w-4" />}
        </span>
        {!last && <span className={cn('w-px flex-1', step.done ? 'bg-emerald-200' : 'bg-neutral-200')} />}
      </div>
      <div className={cn('flex min-w-0 flex-1 items-baseline justify-between gap-3 pb-4', !step.done && !active && 'opacity-60')}>
        <div className="min-w-0">
          <span className="text-sm font-semibold text-neutral-800">{meta.label}</span>
          {detail && <span className="ml-2 truncate text-xs text-neutral-400">{detail}</span>}
        </div>
        {step.done && <span className={cn('text-xs font-bold tabular-nums', step.score >= 70 ? 'text-emerald-600' : 'text-rose-500')}>{step.score}</span>}
      </div>
    </li>
  )
}

function DoneToday({ lesson, onOpen }: { lesson?: Lesson; onOpen: (id: string) => void }) {
  return (
    <section className="flex items-center gap-5 rounded-3xl border border-emerald-200/70 bg-gradient-to-r from-emerald-50 to-brand-50/50 p-6 shadow-sm">
      <Mascot className="h-20 w-20 shrink-0" />
      <div className="min-w-0 flex-1">
        <h2 className="text-lg font-bold text-neutral-900">Today is done!</h2>
        <p className="text-sm text-neutral-600">Come back tomorrow to keep the streak alive.</p>
      </div>
      {lesson && (
        <Button variant="secondary" onClick={() => onOpen(lesson.id)}>
          One more
        </Button>
      )}
    </section>
  )
}

function StartCard({ starting, error, onStart }: { starting: boolean; error?: string; onStart: () => void }) {
  return (
    <section className="flex flex-col items-center gap-4 rounded-3xl border border-neutral-200 bg-white p-8 text-center shadow-sm">
      <div className="grid h-14 w-14 place-items-center rounded-2xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-md">
        <Sparkles className="h-7 w-7" />
      </div>
      <div>
        <h2 className="text-lg font-bold text-neutral-900">No lesson yet today</h2>
        <p className="mx-auto mt-1 max-w-sm text-sm text-neutral-500">
          Built from a film scene when one is ready — otherwise from speaking, reading, writing and your recent mistakes.
        </p>
      </div>
      <Button variant="brand" onClick={onStart} disabled={starting}>
        {starting ? (
          <>
            <Spinner className="h-4 w-4" /> Building your lesson…
          </>
        ) : (
          'Start today’s workout'
        )}
      </Button>
      {error && <p className="text-sm text-red-600">{error}</p>}
    </section>
  )
}

function Calendar({ days }: { days: Set<string> }) {
  const { weeks, monthLabel } = useMemo(() => buildMonth(new Date()), [])
  const todayKey = dateKey(new Date())
  const monthCount = [...days].filter((d) => d.startsWith(todayKey.slice(0, 7))).length

  return (
    <section className="rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm">
      <div className="mb-4 flex items-baseline justify-between">
        <h2 className="text-sm font-semibold text-neutral-700">{monthLabel}</h2>
        <span className="text-xs text-neutral-400">
          {monthCount} workout{monthCount === 1 ? '' : 's'} this month
        </span>
      </div>
      <div className="grid grid-cols-7 gap-1.5 text-center">
        {['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'].map((d) => (
          <span key={d} className="text-xs font-medium text-neutral-400">
            {d}
          </span>
        ))}
        {weeks.flat().map((date, i) =>
          date ? (
            <div
              key={i}
              className={cn(
                'mx-auto flex h-8 w-8 items-center justify-center rounded-full text-sm',
                days.has(dateKey(date))
                  ? 'bg-gradient-to-br from-brand-500 to-emerald-600 font-semibold text-white shadow-sm'
                  : 'text-neutral-600',
                dateKey(date) === todayKey && !days.has(dateKey(date)) && 'ring-2 ring-brand-300',
                date > new Date() && 'text-neutral-300',
              )}
            >
              {date.getDate()}
            </div>
          ) : (
            <span key={i} />
          ),
        )}
      </div>
    </section>
  )
}

function dateKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function buildMonth(now: Date): { weeks: (Date | null)[][]; monthLabel: string } {
  const first = new Date(now.getFullYear(), now.getMonth(), 1)
  const daysInMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0).getDate()
  const offset = (first.getDay() + 6) % 7
  const cells: (Date | null)[] = Array.from({ length: offset }, () => null)
  for (let d = 1; d <= daysInMonth; d++) cells.push(new Date(now.getFullYear(), now.getMonth(), d))
  while (cells.length % 7 !== 0) cells.push(null)
  const weeks: (Date | null)[][] = []
  for (let i = 0; i < cells.length; i += 7) weeks.push(cells.slice(i, i + 7))
  return { weeks, monthLabel: now.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }) }
}
