import { useMemo, useState } from 'react'
import { ArrowRight, BookOpen, Check, RotateCcw, Trophy, X } from 'lucide-react'
import { Mascot } from '@els/ui'
import { parseExercises, type Exercise } from './parse.ts'
import type { CheckFn } from './check.ts'
import { useProgress, type ProgressCtxValue } from './state.ts'
import { Inline } from './markdown.tsx'
import { BlockCtx } from './render/context.ts'
import { ExerciseCard } from './render/exercises.tsx'

export type PracticeSessionProps = {
  exercises: string
  checkAnswer: CheckFn
  onTheory: (s: string) => void
  onExit: () => void
}

const PRAISE = ['Perfect!', 'Nailed it!', 'Flawless!', 'Brilliant!', 'Spot on!']

function ownKeys(progress: ProgressCtxValue, ex: Exercise): string[] {
  return progress.keys().filter((k) => k === ex.id || k.startsWith(`${ex.id}.`) || k.startsWith(`${ex.id}:`))
}

type ExResult = { keys: string[]; answered: number; wrong: string[] }

function resultOf(progress: ProgressCtxValue, ex: Exercise): ExResult {
  const keys = ownKeys(progress, ex)
  return {
    keys,
    answered: keys.filter((k) => progress.get(k)).length,
    wrong: keys.filter((k) => {
      const a = progress.get(k)
      return a && !a.correct
    }),
  }
}

// One exercise per screen: intro interstitial → exercise → …, then a summary with retry.
export function PracticeSession({ exercises, checkAnswer, onTheory, onExit }: PracticeSessionProps) {
  const items = useMemo(() => parseExercises(exercises), [exercises])
  const progress = useProgress()
  const [queue, setQueue] = useState(() => items.map((_, i) => i))
  const [step, setStep] = useState(0)
  const [round, setRound] = useState(0)
  const [startedAt, setStartedAt] = useState(() => Date.now())

  const total = queue.length
  const pos = Math.min(Math.floor(step / 2), total - 1)
  const finished = step >= total * 2
  const exOf = (p: number) => items[queue[p]!]!

  const advance = () => setStep((s) => s + 1)

  const retry = () => {
    const failed = queue.filter((i) => resultOf(progress, items[i]!).wrong.length > 0)
    progress.remove(failed.flatMap((i) => ownKeys(progress, items[i]!)))
    setQueue(failed)
    setStep(0)
    setRound((r) => r + 1)
    setStartedAt(Date.now())
  }

  return (
    <div className="@container">
      <style>{`@keyframes pses-rise{from{opacity:0;transform:translateY(18px)}to{opacity:1;transform:none}}`}</style>
      <div className="mb-4 flex items-center gap-3">
        <div className="flex flex-1 gap-1">
          {queue.map((i, p) => {
            const r = resultOf(progress, items[i]!)
            const done = r.keys.length > 0 && r.answered === r.keys.length
            const cls = done
              ? r.wrong.length
                ? 'bg-rose-400'
                : 'bg-emerald-500'
              : p === pos && !finished
                ? 'bg-brand-500'
                : 'bg-neutral-200'
            return <span key={p} className={`h-1.5 flex-1 rounded-full transition-colors duration-500 ${cls}`} />
          })}
        </div>
        <button
          type="button"
          onClick={onExit}
          title="Back to the list"
          className="grid h-7 w-7 place-items-center rounded-full text-neutral-400 transition-colors hover:bg-neutral-100 hover:text-neutral-700"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div key={`${round}:${step}`} style={{ animation: 'pses-rise .4s ease-out' }}>
        {finished ? (
          <Summary items={items} queue={queue} progress={progress} startedAt={startedAt} onRetry={retry} onExit={onExit} />
        ) : step % 2 === 0 ? (
          <Interstitial
            prev={pos > 0 ? exOf(pos - 1) : undefined}
            next={exOf(pos)}
            pos={pos}
            total={total}
            progress={progress}
            onTheory={onTheory}
            onContinue={advance}
          />
        ) : (
          <div key={round}>
            <ExerciseStep ex={exOf(pos)} checkAnswer={checkAnswer} onTheory={onTheory} progress={progress} last={pos === total - 1} onContinue={advance} />
          </div>
        )}
      </div>
    </div>
  )
}

function ExerciseStep({
  ex,
  checkAnswer,
  onTheory,
  progress,
  last,
  onContinue,
}: {
  ex: Exercise
  checkAnswer: CheckFn
  onTheory: (s: string) => void
  progress: ProgressCtxValue
  last: boolean
  onContinue: () => void
}) {
  const r = resultOf(progress, ex)
  const ready = r.keys.length === 0 || r.answered === r.keys.length
  return (
    <div className="space-y-4">
      <BlockCtx.Provider
        value={{ section: ex.section, dense: false, check: checkAnswer, onTheory, keyBase: ex.id, instruction: ex.instruction }}
      >
        <ExerciseCard ex={ex} />
      </BlockCtx.Provider>
      <div className="flex items-center justify-between">
        <span className="text-xs text-neutral-400">
          {r.keys.length > 0 && `${r.answered} / ${r.keys.length} answered`}
        </span>
        <button
          type="button"
          disabled={!ready}
          onClick={onContinue}
          className="inline-flex h-10 items-center gap-1.5 rounded-full bg-brand-600 px-6 text-sm font-semibold text-white shadow-md transition-all hover:bg-brand-700 disabled:opacity-40 disabled:shadow-none"
        >
          {last ? 'Finish' : 'Continue'} <ArrowRight className="h-4 w-4" />
        </button>
      </div>
    </div>
  )
}

function Interstitial({
  prev,
  next,
  pos,
  total,
  progress,
  onTheory,
  onContinue,
}: {
  prev?: Exercise
  next: Exercise
  pos: number
  total: number
  progress: ProgressCtxValue
  onTheory: (s: string) => void
  onContinue: () => void
}) {
  const prevR = prev ? resultOf(progress, prev) : undefined
  const perfect = prevR && prevR.keys.length > 0 && prevR.wrong.length === 0
  return (
    <div className="flex min-h-72 flex-col items-center justify-center gap-5 rounded-2xl border border-neutral-200/90 bg-gradient-to-b from-brand-50/60 to-white p-6 text-center shadow-sm">
      {prev && prevR && (
        perfect ? (
          <div className="flex flex-col items-center gap-1">
            <Mascot className="h-24 w-24" />
            <p className="text-lg font-bold text-emerald-600">{PRAISE[pos % PRAISE.length]}</p>
            <p className="text-xs text-neutral-500">Everything correct in exercise {prev.id}.</p>
          </div>
        ) : (
          <MistakeRecap ex={prev} result={prevR} progress={progress} onTheory={onTheory} />
        )
      )}
      <div className="space-y-1.5">
        <p className="text-[11px] font-semibold uppercase tracking-widest text-brand-500">
          Exercise {pos + 1} of {total}
        </p>
        {next.lead && <p className="mx-auto max-w-md text-lg font-semibold leading-snug text-neutral-800">{next.lead}</p>}
        {!next.lead && next.instruction && (
          <p className="mx-auto max-w-md text-sm leading-snug text-neutral-500">
            <Inline text={next.instruction} />
          </p>
        )}
      </div>
      <button
        type="button"
        onClick={onContinue}
        className="inline-flex h-10 items-center gap-1.5 rounded-full bg-brand-600 px-7 text-sm font-semibold text-white shadow-md transition-colors hover:bg-brand-700"
      >
        {prev ? 'Continue' : 'Start'} <ArrowRight className="h-4 w-4" />
      </button>
    </div>
  )
}

function MistakeRecap({
  ex,
  result,
  progress,
  onTheory,
}: {
  ex: Exercise
  result: ExResult
  progress: ProgressCtxValue
  onTheory: (s: string) => void
}) {
  return (
    <div className="w-full max-w-md space-y-2 text-left">
      <p className="text-center text-sm font-semibold text-neutral-700">
        {result.keys.length - result.wrong.length} of {result.keys.length} correct — let's look at the slips:
      </p>
      <MistakeList wrongKeys={result.wrong} progress={progress} />
      {ex.section && (
        <div className="text-center">
          <button
            type="button"
            onClick={() => onTheory(ex.section!)}
            className="inline-flex items-center gap-1 rounded-md bg-brand-50 px-2 py-1 text-xs font-medium text-brand-700 transition-colors hover:bg-brand-100"
          >
            <BookOpen className="h-3.5 w-3.5" /> Review section {ex.section}
          </button>
        </div>
      )}
    </div>
  )
}

function MistakeList({ wrongKeys, progress }: { wrongKeys: string[]; progress: ProgressCtxValue }) {
  return (
    <ul className="space-y-1.5">
      {wrongKeys.map((k) => {
        const a = progress.get(k)
        if (!a) return null
        return (
          <li key={k} className="rounded-xl border border-rose-100 bg-white px-3 py-2 text-sm shadow-sm">
            <span className="text-rose-500 line-through decoration-rose-300">{a.answer || '—'}</span>
            {a.correction && <span className="ml-2 font-semibold text-emerald-700">{a.correction}</span>}
            {a.explanation && <span className="mt-0.5 block text-xs leading-relaxed text-neutral-500">{a.explanation}</span>}
          </li>
        )
      })}
    </ul>
  )
}

function Summary({
  items,
  queue,
  progress,
  startedAt,
  onRetry,
  onExit,
}: {
  items: Exercise[]
  queue: number[]
  progress: ProgressCtxValue
  startedAt: number
  onRetry: () => void
  onExit: () => void
}) {
  const results = queue.map((i) => ({ ex: items[i]!, r: resultOf(progress, items[i]!) }))
  const totalKeys = results.reduce((n, x) => n + x.r.keys.length, 0)
  const wrong = results.flatMap((x) => x.r.wrong)
  const accuracy = totalKeys ? Math.round(((totalKeys - wrong.length) / totalKeys) * 100) : 100
  const secs = Math.max(1, Math.round((Date.now() - startedAt) / 1000))
  const time = `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`
  const perfect = wrong.length === 0

  return (
    <div className="flex flex-col items-center gap-5 rounded-2xl border border-neutral-200/90 bg-gradient-to-b from-brand-50/60 to-white p-6 text-center shadow-sm">
      {perfect ? <Mascot className="h-28 w-28" /> : <Trophy className="h-12 w-12 text-amber-400" />}
      <div>
        <h3 className="text-xl font-bold text-neutral-900">{perfect ? 'Perfect session!' : 'Session complete'}</h3>
        <p className="mt-1 text-sm text-neutral-500">
          {queue.length} exercise{queue.length === 1 ? '' : 's'} · {time} min
        </p>
      </div>
      <div className="flex gap-3">
        <Stat label="Accuracy" value={`${accuracy}%`} tone={perfect ? 'good' : accuracy >= 70 ? 'mid' : 'bad'} />
        <Stat label="Correct" value={`${totalKeys - wrong.length}`} tone="good" />
        <Stat label="Mistakes" value={`${wrong.length}`} tone={wrong.length ? 'bad' : 'good'} />
      </div>
      {!perfect && (
        <div className="w-full max-w-md space-y-2 text-left">
          <p className="text-center text-xs font-semibold uppercase tracking-widest text-neutral-400">Worth another look</p>
          <MistakeList wrongKeys={wrong} progress={progress} />
        </div>
      )}
      <div className="flex flex-wrap items-center justify-center gap-2">
        {!perfect && (
          <button
            type="button"
            onClick={onRetry}
            className="inline-flex h-10 items-center gap-1.5 rounded-full bg-brand-600 px-6 text-sm font-semibold text-white shadow-md transition-colors hover:bg-brand-700"
          >
            <RotateCcw className="h-4 w-4" /> Retry mistakes
          </button>
        )}
        <button
          type="button"
          onClick={onExit}
          className={`inline-flex h-10 items-center gap-1.5 rounded-full px-6 text-sm font-semibold transition-colors ${
            perfect
              ? 'bg-brand-600 text-white shadow-md hover:bg-brand-700'
              : 'border border-neutral-200 bg-white text-neutral-600 hover:bg-neutral-50'
          }`}
        >
          {perfect && <Check className="h-4 w-4" />} Back to the list
        </button>
      </div>
    </div>
  )
}

function Stat({ label, value, tone }: { label: string; value: string; tone: 'good' | 'mid' | 'bad' }) {
  const cls = tone === 'good' ? 'text-emerald-600' : tone === 'mid' ? 'text-amber-500' : 'text-rose-500'
  return (
    <div className="min-w-20 rounded-xl border border-neutral-200/90 bg-white px-4 py-2.5 shadow-sm">
      <div className={`text-lg font-bold tabular-nums ${cls}`}>{value}</div>
      <div className="text-[10px] font-semibold uppercase tracking-wider text-neutral-400">{label}</div>
    </div>
  )
}
