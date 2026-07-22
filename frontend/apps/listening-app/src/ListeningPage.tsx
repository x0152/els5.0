import { useEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { emitTargetedEvents, emitTextEvents } from '@els/core-events'
import { AppInfoButton, Button, Input, Select, SpeakButton, Spinner, Textarea, VOICES, cn, speak, useAgentView } from '@els/ui'
import { CheckCheck, Headphones, Lightbulb, Play, RotateCcw, Sparkles, Turtle } from 'lucide-react'
import { api } from './lib/api'
import { alignWords, tokenize } from './diff'

const SOURCE = { app: 'listening' }

type Level = 'easy' | 'medium' | 'hard'

const LEVELS: { id: Level; label: string }[] = [
  { id: 'easy', label: 'Easy' },
  { id: 'medium', label: 'Medium' },
  { id: 'hard', label: 'Hard' },
]

const COUNTS = [3, 5, 7, 10] as const

type ClipResult = { heard: boolean[]; accuracy: number }

function checkAttempt(reference: string, attempt: string): ClipResult {
  const refWords = tokenize(reference)
  const heard = alignWords(refWords, tokenize(attempt))
  const ok = heard.filter(Boolean).length
  return { heard, accuracy: refWords.length ? Math.round((ok / refWords.length) * 100) : 0 }
}

function accuracyColor(accuracy: number) {
  if (accuracy === 100) return 'bg-emerald-500'
  if (accuracy >= 70) return 'bg-amber-400'
  return 'bg-red-400'
}

function HintSkeleton({ sentence }: { sentence: string }) {
  const words = sentence.match(/[A-Za-z][A-Za-z''-]*/g) ?? []
  return (
    <p className="flex flex-wrap gap-x-2 gap-y-1 font-mono text-sm text-neutral-500">
      {words.map((w, i) => (
        <span key={i}>
          {w[0]}
          {'·'.repeat(Math.max(w.length - 1, 1))}
        </span>
      ))}
    </p>
  )
}

function RevealedSentence({ sentence, heard }: { sentence: string; heard: boolean[] }) {
  let idx = -1
  const parts: { text: string; missed?: boolean }[] = []
  let pos = 0
  for (const m of sentence.matchAll(/[A-Za-z][A-Za-z''-]*/g)) {
    idx++
    if (m.index! > pos) parts.push({ text: sentence.slice(pos, m.index) })
    parts.push({ text: m[0], missed: !heard[idx] })
    pos = m.index! + m[0].length
  }
  if (pos < sentence.length) parts.push({ text: sentence.slice(pos) })
  return (
    <p className="text-lg leading-relaxed text-neutral-800">
      {parts.map((p, i) =>
        p.missed ? (
          <span key={i} className="rounded bg-red-100 px-0.5 font-medium text-red-700">
            {p.text}
          </span>
        ) : (
          <span key={i}>{p.text}</span>
        ),
      )}
    </p>
  )
}

export function ListeningPage() {
  const [topic, setTopic] = useState('')
  const [useVocab, setUseVocab] = useState(true)
  const [level, setLevel] = useState<Level>('medium')
  const [count, setCount] = useState<number>(5)
  const [current, setCurrent] = useState(0)
  const [attempt, setAttempt] = useState('')
  const [results, setResults] = useState<ClipResult[]>([])
  const [hint, setHint] = useState(false)
  const [voice, setVoice] = useState('')
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const play = (text: string, rate?: number) => speak(text, { ...(voice && { voice }), ...(rate && { rate }) })

  const generate = useMutation({
    mutationFn: () => api.listening.listeningGenerateDictation({ body: { topic, use_vocab: useVocab, level, count } }),
    onSuccess: () => restart(),
  })

  const restart = () => {
    setCurrent(0)
    setAttempt('')
    setResults([])
    setHint(false)
  }

  const sentences = generate.data?.sentences ?? []
  const sentence = sentences[current]
  const result = results[current]
  const done = sentences.length > 0 && results.filter(Boolean).length === sentences.length

  useEffect(() => {
    if (sentence && !result) {
      play(sentence)
      inputRef.current?.focus()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sentence])

  useAgentView({
    app: 'listening',
    screen: 'dictation',
    info: 'The user practices listening: sentences are spoken aloud and they type what they hear.',
    state: sentence ? { clip: current + 1, total: sentences.length, phase: result ? 'checked' : 'listening' } : undefined,
  })

  const check = () => {
    if (!sentence || attempt.trim().length < 3) return
    const res = checkAttempt(sentence, attempt)
    setResults((prev) => {
      const next = [...prev]
      next[current] = res
      return next
    })
    emitTextEvents(api, 'listening', [sentence], SOURCE)
    const refWords = tokenize(sentence)
    emitTargetedEvents(
      api,
      'listening',
      refWords.filter((_, i) => !res.heard[i]).map((w) => ({ target: w, outcome: 'fail' as const, context: sentence })),
      SOURCE,
    )
  }

  const next = () => {
    if (current < sentences.length - 1) {
      setCurrent((c) => c + 1)
      setAttempt('')
      setHint(false)
    }
  }

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (result) next()
      else check()
    }
  }

  const avg = results.length ? Math.round(results.reduce((s, r) => s + r.accuracy, 0) / results.length) : 0
  const missedWords = done
    ? [...new Set(sentences.flatMap((s, i) => tokenize(s).filter((_, j) => !results[i]!.heard[j])))]
    : []

  const showSetup = sentences.length === 0 || done

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-8">
        <header className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
            <Headphones className="h-6 w-6" />
          </div>
          <div>
            <h1 className="flex items-center gap-1.5 text-2xl font-bold text-neutral-900">
              Listening <AppInfoButton />
            </h1>
            <p className="text-sm text-neutral-500">Listen to the sentence and type exactly what you hear.</p>
          </div>
        </header>

        {sentences.length > 0 && (
          <div className="flex items-center gap-1.5">
            {sentences.map((_, i) => (
              <div
                key={i}
                className={cn(
                  'h-1.5 flex-1 rounded-full transition-colors',
                  results[i] ? accuracyColor(results[i]!.accuracy) : i === current ? 'bg-brand-300' : 'bg-neutral-200',
                )}
              />
            ))}
          </div>
        )}

        {sentence && !done && (
          <section className="flex flex-col gap-4 rounded-2xl border border-neutral-200 bg-white p-6 shadow-sm">
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-500">
                Clip {current + 1} of {sentences.length}
              </h2>
              <Select
                value={voice}
                onChange={setVoice}
                options={[
                  { value: '', label: 'Random voice' },
                  ...VOICES.map((v) => ({ value: v, label: v })),
                ]}
                className="w-auto rounded-md px-2 py-1 text-xs text-neutral-600"
                title="Voice"
              />
            </div>

            <div className="flex items-center gap-4">
              <SpeakButton variant="round" title="Play the sentence" onPlay={() => play(sentence)} />
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium text-neutral-800">
                  {result ? 'Here is how you did:' : 'Tap to listen, then type what you hear.'}
                </p>
                {!result && (
                  <div className="mt-2 flex flex-wrap gap-2">
                    <Button variant="ghost" size="sm" onClick={() => setHint(true)} disabled={hint}>
                      <Lightbulb className="h-4 w-4" /> Hint
                    </Button>
                    <SpeakButton
                      variant="button"
                      className="h-8 px-3"
                      icon={<Turtle className="h-4 w-4" />}
                      onPlay={() => play(sentence, 0.7)}
                    >
                      Slow
                    </SpeakButton>
                  </div>
                )}
              </div>
            </div>

            {hint && !result && (
              <div className="rounded-xl bg-neutral-50 px-4 py-3 ring-1 ring-neutral-100">
                <HintSkeleton sentence={sentence} />
              </div>
            )}

            <Textarea
              ref={inputRef}
              value={attempt}
              onChange={(e) => setAttempt(e.target.value)}
              onKeyDown={onKeyDown}
              rows={3}
              placeholder="Type what you hear… (Enter to check)"
              disabled={!!result}
            />

            {result ? (
              <div className={cn('rounded-xl border p-4', result.accuracy === 100 ? 'border-emerald-200 bg-emerald-50' : 'border-neutral-200 bg-neutral-50')}>
                <p className="mb-2 text-sm font-semibold text-neutral-700">
                  {result.accuracy === 100 ? 'Perfect!' : `Accuracy: ${result.accuracy}% — missed words are highlighted:`}
                </p>
                <RevealedSentence sentence={sentence} heard={result.heard} />
                {current < sentences.length - 1 && (
                  <Button variant="brand" size="sm" className="mt-3" onClick={next} autoFocus>
                    Next clip
                  </Button>
                )}
              </div>
            ) : (
              <Button variant="brand" className="self-end" onClick={check} disabled={attempt.trim().length < 3}>
                <CheckCheck className="h-4 w-4" /> Check
              </Button>
            )}
          </section>
        )}

        {done && (
          <section className="flex flex-col gap-4 rounded-2xl border border-neutral-200 bg-white p-6 shadow-sm">
            <div className="flex items-center gap-4">
              <div
                className={cn(
                  'flex h-16 w-16 shrink-0 items-center justify-center rounded-full text-xl font-bold text-white',
                  accuracyColor(avg),
                )}
              >
                {avg}%
              </div>
              <div>
                <p className="font-semibold text-neutral-900">Dictation finished</p>
                <p className="text-sm text-neutral-500">Results saved to your learning history.</p>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              {sentences.map((s, i) => (
                <div key={i} className="flex items-start gap-3 rounded-lg bg-neutral-50 p-3">
                  <span className={cn('mt-0.5 rounded-full px-2 py-0.5 text-xs font-semibold text-white', accuracyColor(results[i]!.accuracy))}>
                    {results[i]!.accuracy}%
                  </span>
                  <div className="min-w-0 flex-1">
                    <RevealedSentence sentence={s} heard={results[i]!.heard} />
                  </div>
                  <SpeakButton
                    variant="ghost"
                    className="h-9 w-9 px-0"
                    icon={<Play className="h-4 w-4" />}
                    onPlay={() => play(s)}
                  />
                </div>
              ))}
            </div>

            {missedWords.length > 0 && (
              <p className="text-sm text-neutral-500">
                Words to work on: <span className="font-medium text-neutral-700">{missedWords.join(', ')}</span>
              </p>
            )}

            <Button variant="secondary" className="self-start" onClick={restart}>
              <RotateCcw className="h-4 w-4" /> Repeat this dictation
            </Button>
          </section>
        )}

        {showSetup && (
          <section className="relative overflow-hidden rounded-2xl border border-brand-200 bg-gradient-to-br from-brand-50 to-white p-6 shadow-sm">
            <Headphones className="absolute -right-5 -top-5 h-32 w-32 text-brand-100" />
            <div className="relative flex flex-col gap-5">
              <div>
                <h2 className="text-lg font-bold text-neutral-900">{done ? 'Another round?' : 'New dictation'}</h2>
                <p className="mt-0.5 text-sm text-neutral-500">
                  AI writes fresh sentences and reads them aloud — you type what you hear.
                </p>
              </div>
              <div className="flex flex-wrap gap-x-8 gap-y-4">
                <div>
                  <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-neutral-400">Difficulty</p>
                  <div className="flex gap-1.5">
                    {LEVELS.map((l) => (
                      <button
                        key={l.id}
                        onClick={() => setLevel(l.id)}
                        className={cn(
                          'rounded-full px-3.5 py-1.5 text-sm font-medium ring-1 transition-colors',
                          level === l.id
                            ? 'bg-brand-600 text-white ring-brand-600 shadow-sm shadow-brand-600/25'
                            : 'bg-white text-neutral-600 ring-neutral-200 hover:bg-neutral-50',
                        )}
                      >
                        {l.label}
                      </button>
                    ))}
                  </div>
                </div>
                <div>
                  <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-neutral-400">Sentences</p>
                  <div className="flex gap-1.5">
                    {COUNTS.map((c) => (
                      <button
                        key={c}
                        onClick={() => setCount(c)}
                        className={cn(
                          'rounded-full px-3.5 py-1.5 text-sm font-medium ring-1 transition-colors',
                          count === c
                            ? 'bg-brand-600 text-white ring-brand-600 shadow-sm shadow-brand-600/25'
                            : 'bg-white text-neutral-600 ring-neutral-200 hover:bg-neutral-50',
                        )}
                      >
                        {c}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
              <div>
                <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-neutral-400">Topic</p>
                <Input
                  value={topic}
                  onChange={(e) => setTopic(e.target.value)}
                  placeholder="Optional — e.g. travel, work, small talk…"
                  className="max-w-sm"
                />
              </div>
              <div className="flex flex-wrap items-center gap-4">
                <Button variant="brand" size="lg" onClick={() => generate.mutate()} disabled={generate.isPending}>
                  {generate.isPending ? <Spinner className="h-4 w-4" /> : <Sparkles className="h-4 w-4" />}
                  {generate.isPending ? 'Preparing…' : 'Generate dictation'}
                </Button>
                {generate.isPending ? (
                  <p className="text-sm text-neutral-500">
                    Usually takes 10–30 seconds. Stay on this page — leaving will cancel the generation.
                  </p>
                ) : (
                  <label className="flex items-center gap-2 text-sm text-neutral-600">
                    <input type="checkbox" checked={useVocab} onChange={(e) => setUseVocab(e.target.checked)} />
                    Use words I'm learning
                  </label>
                )}
              </div>
              {generate.isError && (
                <p className="text-sm text-red-600">{isApiError(generate.error) ? generate.error.message : 'Generation failed'}</p>
              )}
            </div>
          </section>
        )}
      </div>
    </div>
  )
}
