import { useEffect, useRef, useState, type FormEvent, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, ArrowRight, Check, CirclePlay, Image, Loader2, Mic, RotateCcw, Square, Volume2, X } from 'lucide-react'
import { Badge, Button, cn, Input, IpaText, PhonemePopover, canonicalPhoneme, speak, useAgentView, useRecorder, type PhonemeAnchor } from '@els/ui'
import { isApiError } from '@els/api-client'
import { KindGlyph } from '../components/KindGlyph.tsx'
import { PronunciationResult } from '../components/PronunciationResult.tsx'
import { api } from '../lib/api.ts'
import { pronounced, reviewed } from '../lib/events.ts'
import { usePhonemeGuide } from '../hooks/usePhonemeGuide.ts'
import { statusPill } from '../lib/ui.ts'
import { useShowTranslations } from '../store/me.ts'
import { KIND_LABELS, STATUS_LABELS } from '../lib/types.ts'
import type { Card, CardAnswer, UnitStatus } from '../lib/types.ts'

const IMAGES_ONLY_KEY = 'els.vocab.cards.imagesOnly'

interface Result {
  card: Card
  answer: CardAnswer
}

export function Cards() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const showTranslations = useShowTranslations()
  const [imagesOnly, setImagesOnly] = useState(() => localStorage.getItem(IMAGES_ONLY_KEY) === '1')
  const [queue, setQueue] = useState<Card[]>([])
  const [index, setIndex] = useState(0)
  const [results, setResults] = useState<Result[]>([])
  const [feedback, setFeedback] = useState<CardAnswer | null>(null)
  const [chosen, setChosen] = useState('')
  const [typed, setTyped] = useState('')
  const [sound, setSound] = useState<{ symbol: string; anchor: PhonemeAnchor } | null>(null)
  const openSound = (symbol: string, anchor: PhonemeAnchor) => setSound({ symbol, anchor })
  const guide = usePhonemeGuide()
  const retried = useRef(new Set<string>())

  const assessM = useMutation({
    mutationFn: ({ text, blob }: { text: string; blob: Blob }) => {
      const form = new FormData()
      form.append('audio', blob, 'recording.webm')
      form.append('text', text)
      return api.speech.assessSpeech({ body: form as unknown as never })
    },
    onSuccess: (data, vars) => {
      if (data) pronounced(vars.text, data.overall >= 60 ? 'ok' : 'fail')
    },
  })
  const recorder = useRecorder((blob) => {
    const text = feedback?.unit.text
    if (text) assessM.mutate({ text, blob })
  })

  const start = useMutation({
    mutationFn: (imgs: boolean) => api.vocab.generateVocabCards({ body: { images_only: imgs } }),
    onSuccess: (d) => {
      setQueue(d?.cards ?? [])
      setIndex(0)
      setResults([])
      setFeedback(null)
      setChosen('')
      setTyped('')
      retried.current.clear()
    },
  })

  useEffect(() => {
    start.mutate(imagesOnly)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [imagesOnly])

  const answer = useMutation({
    mutationFn: (vars: { card: Card; value: string }) =>
      api.vocab.answerVocabCard({ body: { unit_id: vars.card.unit_id, answer: vars.value } }),
    onSuccess: (res, vars) => {
      if (!res) return
      assessM.reset()
      recorder.clear()
      setFeedback(res)
      reviewed(res.unit.text, res.correct ? 'ok' : 'fail', vars.card.mode === 'input' ? 'writing' : 'reading')
      setResults((r) => (r.some((x) => x.card.unit_id === vars.card.unit_id) ? r : [...r, { card: vars.card, answer: res }]))
      if (!res.correct && !retried.current.has(vars.card.unit_id)) {
        retried.current.add(vars.card.unit_id)
        setQueue((q) => [...q, vars.card])
      }
    },
  })

  const card = queue[index]
  const done = queue.length > 0 && index >= queue.length

  useEffect(() => {
    if (done) void qc.invalidateQueries({ queryKey: ['vocab', 'units'] })
  }, [done, qc])

  useAgentView({
    app: 'vocab',
    screen: 'cards',
    info: 'The user is guessing words by image and definition; correct answers advance word status.',
    state: { card: index + 1, total: queue.length, imagesOnly: imagesOnly ? 'on' : 'off' },
  })

  const toggleImagesOnly = () => {
    setImagesOnly((v) => {
      localStorage.setItem(IMAGES_ONLY_KEY, v ? '0' : '1')
      return !v
    })
  }

  const next = () => {
    setFeedback(null)
    setChosen('')
    setTyped('')
    assessM.reset()
    recorder.clear()
    setIndex((i) => i + 1)
  }

  const submitTyped = (e: FormEvent) => {
    e.preventDefault()
    if (!card || !typed.trim() || answer.isPending) return
    answer.mutate({ card, value: typed })
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (sound) return
      if (feedback) {
        if (e.key === 'Enter') {
          e.preventDefault()
          next()
        }
        return
      }
      if (!card || card.mode !== 'choice' || answer.isPending) return
      const opt = card.options?.[Number(e.key) - 1]
      if (e.key >= '1' && e.key <= '4' && opt) {
        setChosen(opt)
        answer.mutate({ card, value: opt })
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [feedback, card, answer.isPending, sound])

  const header = (
    <div className="flex items-center justify-between gap-3">
      <button
        type="button"
        onClick={toggleImagesOnly}
        className={cn(
          'inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-sm font-medium ring-1 transition',
          imagesOnly ? 'bg-brand-600 text-white ring-brand-600' : 'bg-white text-neutral-700 ring-neutral-200 hover:bg-neutral-50',
        )}
      >
        <Image className="h-4 w-4" /> Images only
      </button>
      <Button variant="secondary" disabled={start.isPending} onClick={() => start.mutate(imagesOnly)}>
        {start.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <RotateCcw className="h-4 w-4" />} New session
      </Button>
    </div>
  )

  if (start.isError) {
    const raw = isApiError(start.error) ? start.error.message : ''
    const message = raw.replace(/^validation failed:\s*/i, '') || 'Could not build a deck. Try again.'
    return (
      <Shell onBack={() => navigate('..')}>
        {header}
        <Panel>
          <div className="py-14 text-center">
            <h3 className="text-lg font-semibold text-neutral-900">Not enough words</h3>
            <p className="mx-auto mt-1 max-w-sm text-sm text-neutral-500">{message}</p>
          </div>
        </Panel>
      </Shell>
    )
  }

  if (start.isPending || queue.length === 0) {
    return (
      <Shell onBack={() => navigate('..')}>
        {header}
        <Panel>
          <div className="grid place-items-center py-20 text-neutral-400">
            <Loader2 className="h-6 w-6 animate-spin" />
          </div>
        </Panel>
      </Shell>
    )
  }

  if (done) {
    const correct = results.filter((r) => r.answer.correct).length
    return (
      <Shell onBack={() => navigate('..')}>
        {header}
        <Panel>
          <div className="text-center">
            <h3 className="text-2xl font-bold text-neutral-900">
              {correct} / {results.length} correct
            </h3>
            <p className="mt-1 text-sm text-neutral-500">Answer correctly on different days to move words forward.</p>
          </div>
          <ul className="mt-6 space-y-2">
            {results.map(({ card: c, answer: a }) => (
              <li key={a.unit.id} className="flex items-center justify-between rounded-xl bg-neutral-50 px-4 py-2.5">
                <span className="flex items-center gap-2">
                  {a.correct ? <Check className="h-4 w-4 text-emerald-600" /> : <X className="h-4 w-4 text-rose-500" />}
                  <span className="font-medium text-neutral-900">{a.unit.text}</span>
                </span>
                <span className="flex items-center gap-1.5">
                  {c.status !== a.unit.status && (
                    <>
                      <span className="text-xs text-neutral-400">{STATUS_LABELS[c.status as UnitStatus] ?? c.status}</span>
                      <ArrowRight className="h-3 w-3 text-neutral-400" />
                    </>
                  )}
                  <span className={cn('rounded-full px-2.5 py-0.5 text-xs font-medium ring-1', statusPill[a.unit.status as UnitStatus])}>
                    {STATUS_LABELS[a.unit.status as UnitStatus] ?? a.unit.status}
                  </span>
                </span>
              </li>
            ))}
          </ul>
          <div className="mt-6 flex justify-center">
            <Button variant="brand" onClick={() => start.mutate(imagesOnly)}>
              <RotateCcw className="h-4 w-4" /> New session
            </Button>
          </div>
        </Panel>
      </Shell>
    )
  }

  if (!card) return null

  return (
    <Shell onBack={() => navigate('..')}>
      {header}
      <div className="h-1.5 w-full overflow-hidden rounded-full bg-neutral-200">
        <div className="h-full rounded-full bg-brand-600 transition-all" style={{ width: `${(index / queue.length) * 100}%` }} />
      </div>
      <Panel>
        <div className="mb-4 flex items-center justify-between">
          <Badge className="text-[11px]">
            <KindGlyph kind={card.kind} className="h-3 w-3" /> {KIND_LABELS[card.kind] ?? card.kind}
          </Badge>
          <span className="text-xs font-medium text-neutral-400">
            {index + 1} / {queue.length}
          </span>
        </div>

        {card.image_url && (
          <div className="mx-auto mb-4 aspect-[4/3] w-full max-w-sm overflow-hidden rounded-2xl bg-neutral-100">
            <img src={card.image_url} alt="" className="h-full w-full object-cover" />
          </div>
        )}
        {card.direction === 'translation' ? (
          <div className="text-center">
            <p className="text-3xl font-bold text-neutral-900">{card.word}</p>
            {card.transcription && (
              <p className="mt-1 text-sm text-neutral-400">
                /<IpaText ipa={card.transcription} onSelect={openSound} />/
              </p>
            )}
            <p className="mt-2 text-sm text-neutral-500">Pick the right translation.</p>
          </div>
        ) : (
          <p className="text-center text-base text-neutral-700">{card.definition || 'Guess the word shown on the picture.'}</p>
        )}

        <div className="mt-6">
          {card.mode === 'choice' ? (
            <div className="grid gap-2 sm:grid-cols-2">
              {(card.options ?? []).map((opt, i) => {
                const correctValue = card.direction === 'translation' ? feedback?.unit.translation : feedback?.unit.text
                const isCorrect = correctValue === opt
                const isChosen = chosen === opt
                return (
                  <button
                    key={opt}
                    type="button"
                    disabled={!!feedback || answer.isPending}
                    onClick={() => {
                      setChosen(opt)
                      answer.mutate({ card, value: opt })
                    }}
                    className={cn(
                      'flex items-center gap-2.5 rounded-xl px-4 py-3 text-left text-sm font-medium ring-1 transition',
                      feedback && isCorrect
                        ? 'bg-emerald-50 text-emerald-700 ring-emerald-300'
                        : feedback && isChosen
                          ? 'bg-rose-50 text-rose-700 ring-rose-300'
                          : 'bg-white text-neutral-800 ring-neutral-200 enabled:hover:bg-neutral-50',
                    )}
                  >
                    <span className="hidden h-5 w-5 shrink-0 place-items-center rounded bg-neutral-100 text-[11px] text-neutral-400 sm:grid">
                      {i + 1}
                    </span>
                    {opt}
                  </button>
                )
              })}
            </div>
          ) : (
            <form onSubmit={submitTyped} className="mx-auto flex max-w-sm gap-2">
              <Input
                value={typed}
                onChange={(e) => setTyped(e.target.value)}
                placeholder="Type the word…"
                disabled={!!feedback || answer.isPending}
                autoFocus
                autoComplete="off"
              />
              {!feedback && (
                <Button type="submit" variant="brand" disabled={!typed.trim() || answer.isPending}>
                  {answer.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Check'}
                </Button>
              )}
            </form>
          )}
        </div>

        {feedback && (
          <div
            className={cn(
              'mt-6 rounded-2xl p-4 ring-1',
              feedback.correct ? 'bg-emerald-50 ring-emerald-200' : 'bg-rose-50 ring-rose-200',
            )}
          >
            <p className={cn('flex items-center gap-1.5 text-sm font-semibold', feedback.correct ? 'text-emerald-700' : 'text-rose-700')}>
              {feedback.correct ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
              <span className="flex-1">
                {feedback.correct
                  ? 'Correct!'
                  : `Correct answer: ${card.direction === 'translation' ? feedback.unit.translation : feedback.unit.text}`}
              </span>
              <button
                type="button"
                onClick={() => speak(feedback.unit.text)}
                title="Pronounce"
                className="rounded-full p-1.5 text-neutral-500 transition hover:bg-white hover:text-neutral-800"
              >
                <Volume2 className="h-4 w-4" />
              </button>
              <button
                type="button"
                onClick={recorder.state === 'recording' ? recorder.stop : recorder.start}
                disabled={assessM.isPending || recorder.state === 'unsupported'}
                title={recorder.state === 'recording' ? 'Stop recording' : 'Check my pronunciation'}
                className={cn(
                  'rounded-full p-1.5 transition disabled:opacity-50',
                  recorder.state === 'recording'
                    ? 'bg-white text-red-600 hover:bg-red-50'
                    : 'text-neutral-500 hover:bg-white hover:text-neutral-800',
                )}
              >
                {recorder.state === 'recording' ? (
                  <Square className="h-4 w-4" />
                ) : assessM.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Mic className="h-4 w-4" />
                )}
              </button>
              {recorder.blob && (
                <button
                  type="button"
                  onClick={recorder.play}
                  disabled={recorder.state === 'recording'}
                  title="Play my recording"
                  className="rounded-full p-1.5 text-neutral-500 transition hover:bg-white hover:text-neutral-800 disabled:opacity-50"
                >
                  <CirclePlay className="h-4 w-4" />
                </button>
              )}
            </p>
            {recorder.state === 'recording' && (
              <p className="mt-1.5 text-xs text-red-600">Recording… {recorder.elapsed}s — say “{feedback.unit.text}” and press stop.</p>
            )}
            {assessM.isError && <p className="mt-1.5 text-xs text-red-600">The pronunciation service did not respond. Try again.</p>}
            {assessM.data && <PronunciationResult assessment={assessM.data} onSelect={openSound} className="mt-2" />}
            {(card.direction === 'translation' || feedback.unit.transcription) && (
              <p className="mt-1.5 text-sm">
                {card.direction === 'translation' && <span className="font-medium text-neutral-900">{feedback.unit.text} </span>}
                {feedback.unit.transcription && (
                  <span className="text-neutral-400">
                    /<IpaText ipa={feedback.unit.transcription} onSelect={openSound} />/
                  </span>
                )}
              </p>
            )}
            {showTranslations && feedback.unit.translation && card.direction !== 'translation' && (
              <p className="mt-1 text-sm text-neutral-700">{feedback.unit.translation}</p>
            )}
            {feedback.unit.example && <p className="mt-1 text-sm italic text-neutral-500">{feedback.unit.example}</p>}
            <div className="mt-3 flex justify-end">
              <Button variant="brand" autoFocus onClick={next}>
                Next
              </Button>
            </div>
          </div>
        )}
      </Panel>

      {sound && (
        <PhonemePopover
          symbol={canonicalPhoneme(sound.symbol)}
          info={guide(sound.symbol)}
          anchor={sound.anchor}
          onClose={() => setSound(null)}
        />
      )}
    </Shell>
  )
}

function Shell({ children, onBack }: { children: ReactNode; onBack: () => void }) {
  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-2xl space-y-4 p-6">
        <button onClick={onBack} className="inline-flex items-center gap-1.5 text-sm font-medium text-neutral-500 hover:text-neutral-800">
          <ArrowLeft className="h-4 w-4" /> Back to vocabulary
        </button>
        <h1 className="text-2xl font-bold text-neutral-900">Cards</h1>
        {children}
      </div>
    </div>
  )
}

function Panel({ children }: { children: ReactNode }) {
  return <div className="rounded-3xl bg-white p-6 ring-1 ring-neutral-200">{children}</div>
}
