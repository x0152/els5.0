import { useEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import type { SpeechComponents } from '@els/api-client'
import { emitTextEvents } from '@els/core-events'
import { Button, SpeakButton, Spinner, cn, useRecorder } from '@els/ui'
import { Check, Mic, Play, Snail, Square } from 'lucide-react'
import { api } from '../lib/api.ts'
import { playPhrase, stopClip, useFilmUrl } from '../lib/audio.ts'
import { accuracy } from '../lib/diff.ts'
import type { ItemResult, StepOutcome, WarmupItem } from '../lib/types.ts'
import { ContinueButton, StepShell } from './StepShell.tsx'

export type PhraseTask = WarmupItem

// One runner for speak / dictation / warm-up steps: a queue of phrases, each either
// spoken aloud (scored by speech assess) or typed from hearing (scored by word diff).
export function PhrasesStep({ items, onDone }: { items: PhraseTask[]; onDone: (outcome: StepOutcome) => void }) {
  const [index, setIndex] = useState(0)
  const [results, setResults] = useState<ItemResult[]>([])
  const item = items[index]
  const finished = index >= items.length

  useEffect(() => () => stopClip(), [])

  const record = (score: number) => {
    if (!item) return
    setResults((r) => [...r, { kind: 'phrase', text: item.text, film_id: item.film_id, start_ms: item.start_ms, end_ms: item.end_ms, score }])
    setIndex((i) => i + 1)
  }

  if (finished || !item) {
    const avg = results.length ? Math.round(results.reduce((s, r) => s + r.score, 0) / results.length) : 100
    return (
      <StepShell>
        <div className="flex flex-1 flex-col items-center justify-center gap-2 py-6">
          <span className="text-4xl font-bold text-neutral-900">{avg}</span>
          <span className="text-sm text-neutral-500">average score</span>
          <ul className="mt-3 w-full max-w-md space-y-1">
            {results.map((r, i) => (
              <li key={i} className="flex items-center justify-between gap-3 rounded-lg bg-neutral-50 px-3 py-1.5 text-sm">
                <span className="truncate text-neutral-700">{r.text}</span>
                <span className={cn('font-semibold', r.score >= 70 ? 'text-emerald-600' : 'text-rose-500')}>{r.score}</span>
              </li>
            ))}
          </ul>
        </div>
        <ContinueButton onClick={() => onDone({ score: avg, results })} />
      </StepShell>
    )
  }

  return (
    <StepShell>
      <div className="flex items-center gap-2">
        <div className="flex flex-1 gap-1">
          {items.map((_, i) => (
            <span key={i} className={cn('h-1 flex-1 rounded-full', i < index ? 'bg-brand-400' : i === index ? 'bg-brand-200' : 'bg-neutral-100')} />
          ))}
        </div>
        <span className="text-xs font-medium tabular-nums text-neutral-400">
          {index + 1} / {items.length}
        </span>
      </div>
      {item.mode === 'speak' ? (
        <SpeakTask key={index} item={item} onScore={record} />
      ) : (
        <DictationTask key={index} item={item} onScore={record} />
      )}
      <button type="button" onClick={() => record(0)} className="self-center text-xs text-neutral-400 hover:text-neutral-600 hover:underline">
        Skip this one
      </button>
    </StepShell>
  )
}

type Assessment = SpeechComponents['schemas']['AssessOutput']

const VERDICT_STYLES: Record<string, string> = {
  good: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
  close: 'bg-amber-50 text-amber-700 ring-amber-300',
  wrong: 'bg-red-50 text-red-700 ring-red-300',
  missing: 'bg-neutral-100 text-neutral-400 ring-neutral-200 line-through',
}

function WordBreakdown({ assessment }: { assessment: Assessment }) {
  return (
    <div className="flex flex-wrap justify-center gap-2">
      {(assessment.words ?? []).map((word, i) => (
        <div
          key={i}
          className={cn(
            'rounded-xl border bg-white px-3 py-2',
            word.score >= 85 ? 'border-neutral-200' : word.score >= 60 ? 'border-amber-300' : 'border-red-300',
          )}
        >
          <div className="mb-1.5 flex items-baseline justify-between gap-3">
            <span className="text-sm font-semibold text-neutral-900">{word.word}</span>
            <span
              className={cn(
                'text-xs font-medium tabular-nums',
                word.score >= 85 ? 'text-emerald-600' : word.score >= 60 ? 'text-amber-600' : 'text-red-600',
              )}
            >
              {word.score}
            </span>
          </div>
          <div className="flex flex-wrap items-center gap-1">
            {(word.phonemes ?? []).map((p, j) => (
              <span
                key={j}
                title={p.verdict === 'good' ? `/${p.expected}/` : `expected /${p.expected}/, heard /${p.heard ?? '—'}/`}
                className={cn('rounded-md px-1.5 py-0.5 font-mono text-xs ring-1', VERDICT_STYLES[p.verdict] ?? VERDICT_STYLES.good)}
              >
                {p.expected}
              </span>
            ))}
            {(word.extra ?? []).map((sym, j) => (
              <span key={`extra-${j}`} title="Extra sound not in the word" className="rounded-md bg-purple-50 px-1.5 py-0.5 font-mono text-xs text-purple-600 ring-1 ring-purple-200">
                +{sym}
              </span>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

function SpeakTask({ item, onScore }: { item: PhraseTask; onScore: (score: number) => void }) {
  const { videoUrl } = useFilmUrl(item.film_id)
  const [assessment, setAssessment] = useState<Assessment | null>(null)

  const assess = useMutation({
    mutationFn: (blob: Blob) => {
      const form = new FormData()
      form.append('audio', blob, 'recording.webm')
      form.append('text', item.text)
      return api.speech.assessSpeech({ body: form as unknown as never })
    },
    onSuccess: (data) => setAssessment(data ?? null),
  })
  const recorder = useRecorder((blob) => assess.mutate(blob))
  const recording = recorder.state === 'recording'

  const wordScores = (assessment?.words ?? []).map((w) => w.score)
  const worst = wordScores.length ? Math.min(...wordScores) : (assessment?.overall ?? 0)
  // One failed word must not drown in the average: it caps the score and blocks Next.
  const score = assessment ? Math.min(assessment.overall, worst + 50) : null
  const passed = worst >= 50

  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-5 py-4">
      <p className="max-w-2xl text-center text-xl font-semibold text-neutral-900 sm:text-2xl">“{item.text}”</p>
      <div className="flex items-center gap-2">
        <SpeakButton variant="button" onPlay={() => playPhrase(videoUrl, item)}>
          Listen
        </SpeakButton>
        <Button
          variant={recording ? 'danger' : 'brand'}
          onClick={() => (recording ? recorder.stop() : recorder.start())}
          disabled={assess.isPending}
        >
          {assess.isPending ? <Spinner className="h-4 w-4" /> : recording ? <Square className="h-4 w-4" /> : <Mic className="h-4 w-4" />}
          {recording ? 'Stop' : 'Record'}
        </Button>
      </div>
      {assessment && score !== null && (
        <div className="flex w-full flex-col items-center gap-3">
          <span className={cn('text-3xl font-bold', score >= 70 ? 'text-emerald-600' : 'text-rose-500')}>{score}</span>
          <WordBreakdown assessment={assessment} />
          {!passed && <p className="text-sm text-rose-600">Some words didn't come through — record again to move on.</p>}
          <div className="mt-1 flex gap-2">
            <Button variant="secondary" size="sm" onClick={() => setAssessment(null)}>
              Try again
            </Button>
            {passed && (
              <Button variant="brand" size="sm" onClick={() => onScore(score)}>
                Next <Check className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function DictationTask({ item, onScore }: { item: PhraseTask; onScore: (score: number) => void }) {
  const { videoUrl } = useFilmUrl(item.film_id)
  const [attempt, setAttempt] = useState('')
  const [checked, setChecked] = useState<ReturnType<typeof accuracy> | null>(null)
  const played = useRef(false)

  useEffect(() => {
    if (videoUrl || !item.film_id) {
      if (!played.current) {
        played.current = true
        playPhrase(videoUrl, item)
      }
    }
  }, [videoUrl, item])

  const check = () => {
    const result = accuracy(item.text, attempt)
    setChecked(result)
    emitTextEvents(api, 'listening', [item.text], { app: 'workout' })
  }

  return (
    <div className="mx-auto flex w-full max-w-xl flex-1 flex-col justify-center gap-4 py-2">
      <div className="flex items-center gap-2">
        <SpeakButton variant="button" icon={<Play className="h-4 w-4" />} onPlay={() => playPhrase(videoUrl, item)}>
          Play
        </SpeakButton>
        <SpeakButton variant="button" icon={<Snail className="h-4 w-4" />} onPlay={() => playPhrase(videoUrl, item, 0.7)}>
          Slow
        </SpeakButton>
      </div>
      {!checked ? (
        <div className="flex gap-2">
          <input
            value={attempt}
            onChange={(e) => setAttempt(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && attempt.trim() && check()}
            placeholder="Type what you hear…"
            autoFocus
            className="flex-1 rounded-xl border border-neutral-200 bg-white px-3.5 py-2.5 text-[15px] outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100"
          />
          <Button variant="brand" onClick={check} disabled={!attempt.trim()}>
            Check
          </Button>
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          <p className="flex flex-wrap gap-x-1.5 gap-y-1 text-lg leading-7">
            {checked.words.map((w, i) => (
              <span key={i} className={checked.heard[i] ? 'text-emerald-700' : 'rounded bg-rose-100 px-1 font-medium text-rose-700'}>
                {w}
              </span>
            ))}
          </p>
          <div className="flex items-center gap-3">
            <span className={cn('text-2xl font-bold', checked.percent >= 70 ? 'text-emerald-600' : 'text-rose-500')}>{checked.percent}%</span>
            <Button variant="brand" size="sm" onClick={() => onScore(checked.percent)}>
              Next <Check className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
