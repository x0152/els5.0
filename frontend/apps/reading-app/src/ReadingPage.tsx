import { useMemo, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { emitTargetedEvents, emitTextEvents } from '@els/core-events'
import { Illustration, ImageApiCtx, type IllustrationStatus, type ImageApi } from '@els/blocks'
import { Button, Input, Spinner, cn, speak, useAgentView } from '@els/ui'
import { BookOpenText, CheckCheck, Clock, Sparkles, Volume2 } from 'lucide-react'
import { api } from './lib/api'

const SOURCE = { app: 'reading' }

type Level = 'easy' | 'medium' | 'hard'
type Length = 'short' | 'medium' | 'long'

const LEVELS: { id: Level; label: string }[] = [
  { id: 'easy', label: 'Easy' },
  { id: 'medium', label: 'Medium' },
  { id: 'hard', label: 'Hard' },
]

const LENGTHS: { id: Length; label: string }[] = [
  { id: 'short', label: 'Short' },
  { id: 'medium', label: 'Medium' },
  { id: 'long', label: 'Long' },
]

const imageApi: ImageApi = async (prompt, trigger, aspect) =>
  (await api.learn.ensureIllustration({ body: { prompt, trigger, aspect } })) as IllustrationStatus

type Chunk = { kind: 'paragraph'; text: string } | { kind: 'image'; prompt: string }

function stripInlineMarkdown(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, '$1')
    .replace(/__(.+?)__/g, '$1')
    .replace(/(?<!\w)\*(.+?)\*(?!\w)/g, '$1')
    .replace(/(?<!\w)_(.+?)_(?!\w)/g, '$1')
    .replace(/`(.+?)`/g, '$1')
}

function parseBody(body: string): Chunk[] {
  return body
    .split(/\n{2,}/)
    .map((block) => block.trim())
    .filter(Boolean)
    .flatMap((block): Chunk[] => {
      const m = block.match(/^\[image:\s*(.+?)\]$/i)
      if (m) return [{ kind: 'image', prompt: m[1]! }]
      return [{ kind: 'paragraph', text: stripInlineMarkdown(block.replace(/\n/g, ' ')) }]
    })
}

const WORD_RE = /[A-Za-z][A-Za-z''-]*/g

function wordKey(w: string) {
  return w.toLowerCase()
}

function sentenceOf(text: string, word: string): string {
  const sentence = text.split(/(?<=[.!?])\s+/).find((s) => s.toLowerCase().includes(word.toLowerCase()))
  return sentence ?? ''
}

function ClickableText({
  text,
  unknown,
  learner,
  onToggle,
  disabled,
}: {
  text: string
  unknown: Set<string>
  learner: Set<string>
  onToggle: (key: string) => void
  disabled: boolean
}) {
  const parts: { text: string; key?: string }[] = []
  let pos = 0
  for (const m of text.matchAll(WORD_RE)) {
    if (m.index! > pos) parts.push({ text: text.slice(pos, m.index) })
    parts.push({ text: m[0], key: wordKey(m[0]) })
    pos = m.index! + m[0].length
  }
  if (pos < text.length) parts.push({ text: text.slice(pos) })

  return (
    <p className="font-serif text-[18px] leading-8 text-neutral-800">
      {parts.map((p, i) =>
        p.key ? (
          <span
            key={i}
            onClick={disabled ? undefined : () => onToggle(p.key!)}
            className={cn(
              'rounded px-px transition-colors',
              !disabled && 'cursor-pointer hover:bg-amber-100',
              learner.has(p.key) && 'underline decoration-brand-300 decoration-2 underline-offset-4',
              unknown.has(p.key) && 'bg-amber-200 font-medium text-amber-900',
            )}
          >
            {p.text}
          </span>
        ) : (
          <span key={i}>{p.text}</span>
        ),
      )}
    </p>
  )
}

export function ReadingPage() {
  const [topic, setTopic] = useState('')
  const [useVocab, setUseVocab] = useState(true)
  const [level, setLevel] = useState<Level>('medium')
  const [length, setLength] = useState<Length>('medium')
  const [unknown, setUnknown] = useState<Set<string>>(new Set())
  const [finished, setFinished] = useState(false)
  const [readSeconds, setReadSeconds] = useState(0)
  const startedAt = useRef(0)

  const generate = useMutation({
    mutationFn: () => api.reading.readingGenerateText({ body: { topic, use_vocab: useVocab, level, length } }),
    onSuccess: () => {
      setUnknown(new Set())
      setFinished(false)
      startedAt.current = Date.now()
    },
  })

  const text = generate.data
  const chunks = useMemo(() => (text ? parseBody(text.body) : []), [text])
  const paragraphs = useMemo(() => chunks.filter((c) => c.kind === 'paragraph').map((c) => c.text), [chunks])
  const learnerWords = useMemo(() => new Set((text?.words ?? []).map(wordKey)), [text])
  const wordCount = useMemo(() => paragraphs.join(' ').match(WORD_RE)?.length ?? 0, [paragraphs])

  useAgentView({
    app: 'reading',
    screen: 'text',
    info: 'The user reads a generated text and marks unknown words by tapping them.',
    state: text ? { title: text.title, unknown: unknown.size, phase: finished ? 'finished' : 'reading' } : undefined,
  })

  const toggle = (key: string) =>
    setUnknown((prev) => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })

  const finish = () => {
    const fullText = paragraphs.join('\n\n')
    emitTextEvents(api, 'reading', paragraphs, SOURCE)
    emitTargetedEvents(
      api,
      'reading',
      [...unknown].map((w) => ({ target: w, outcome: 'fail' as const, context: sentenceOf(fullText, w) })),
      SOURCE,
    )
    setReadSeconds(Math.max(Math.round((Date.now() - startedAt.current) / 1000), 1))
    setFinished(true)
  }

  const wpm = readSeconds ? Math.round((wordCount / readSeconds) * 60) : 0

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-8">
        <header className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
            <BookOpenText className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-neutral-900">Reading</h1>
            <p className="text-sm text-neutral-500">
              Read the text and tap the words you don't know. Finishing the page counts the rest as known.
            </p>
          </div>
        </header>

        <section className="flex flex-col gap-3 rounded-2xl border border-neutral-200 bg-white p-5">
          <div className="flex flex-col gap-2 sm:flex-row">
            <Input
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              placeholder="Topic (optional) — e.g. space, cooking, startups…"
              className="flex-1"
            />
            <Button variant="brand" onClick={() => generate.mutate()} disabled={generate.isPending}>
              {generate.isPending ? <Spinner className="h-4 w-4" /> : <Sparkles className="h-4 w-4" />}
              {text ? 'New text' : 'Generate text'}
            </Button>
          </div>
          <div className="flex flex-wrap items-center gap-4 text-sm">
            <div className="flex gap-1">
              {LEVELS.map((l) => (
                <button
                  key={l.id}
                  onClick={() => setLevel(l.id)}
                  className={cn(
                    'rounded-full px-3 py-1 text-xs font-medium ring-1 transition-colors',
                    level === l.id ? 'bg-brand-600 text-white ring-brand-600' : 'bg-white text-neutral-600 ring-neutral-200 hover:bg-neutral-50',
                  )}
                >
                  {l.label}
                </button>
              ))}
            </div>
            <div className="flex gap-1">
              {LENGTHS.map((l) => (
                <button
                  key={l.id}
                  onClick={() => setLength(l.id)}
                  className={cn(
                    'rounded-full px-3 py-1 text-xs font-medium ring-1 transition-colors',
                    length === l.id ? 'bg-neutral-900 text-white ring-neutral-900' : 'bg-white text-neutral-600 ring-neutral-200 hover:bg-neutral-50',
                  )}
                >
                  {l.label}
                </button>
              ))}
            </div>
            <label className="flex items-center gap-2 text-sm text-neutral-600">
              <input type="checkbox" checked={useVocab} onChange={(e) => setUseVocab(e.target.checked)} />
              Weave in words I'm learning
            </label>
          </div>
          {generate.isError && (
            <p className="text-sm text-red-600">{isApiError(generate.error) ? generate.error.message : 'Generation failed'}</p>
          )}
        </section>

        {text && (
          <ImageApiCtx.Provider value={imageApi}>
            <article className="flex flex-col gap-4 rounded-2xl border border-neutral-200 bg-white p-6 shadow-sm sm:p-8">
              <div>
                <h2 className="font-serif text-2xl font-bold text-neutral-900">{text.title}</h2>
                <div className="mt-2 flex flex-wrap items-center gap-3 text-xs text-neutral-400">
                  <span className="flex items-center gap-1">
                    <Clock className="h-3.5 w-3.5" /> ~{Math.max(Math.round(wordCount / 180), 1)} min · {wordCount} words
                  </span>
                  <button
                    type="button"
                    onClick={() => speak(paragraphs.join(' '))}
                    className="flex items-center gap-1 text-brand-600 hover:underline"
                  >
                    <Volume2 className="h-3.5 w-3.5" /> Listen
                  </button>
                </div>
                {(text.words?.length ?? 0) > 0 && (
                  <p className="mt-2 text-xs text-neutral-400">
                    Your words (underlined in the text): <span className="text-neutral-600">{text.words!.join(', ')}</span>
                  </p>
                )}
              </div>
              {chunks.map((chunk, i) =>
                chunk.kind === 'image' ? (
                  <Illustration
                    key={i}
                    prompt={chunk.prompt}
                    aspect="landscape"
                    className="mx-auto w-full max-w-md rounded-xl"
                    style={{ aspectRatio: '16/9' }}
                  />
                ) : (
                  <ClickableText
                    key={i}
                    text={chunk.text}
                    unknown={unknown}
                    learner={learnerWords}
                    onToggle={toggle}
                    disabled={finished}
                  />
                ),
              )}
            </article>

            {finished ? (
              <section className="flex flex-col gap-3 rounded-2xl border border-emerald-200 bg-emerald-50 p-5">
                <p className="flex items-center gap-2 font-medium text-emerald-800">
                  <CheckCheck className="h-5 w-5" /> Page saved to your learning history.
                </p>
                <div className="flex flex-wrap gap-4 text-sm text-emerald-700">
                  <span>
                    <span className="font-semibold">{wordCount - unknown.size}</span> of {wordCount} words known
                  </span>
                  <span>
                    <span className="font-semibold">
                      {Math.floor(readSeconds / 60)}:{String(readSeconds % 60).padStart(2, '0')}
                    </span>{' '}
                    reading time
                  </span>
                  <span>
                    <span className="font-semibold">{wpm}</span> words/min
                  </span>
                </div>
                {unknown.size > 0 && (
                  <p className="text-sm text-emerald-700">Marked as unknown: {[...unknown].join(', ')}.</p>
                )}
                <Button variant="brand" className="self-start" onClick={() => generate.mutate()} disabled={generate.isPending}>
                  {generate.isPending ? <Spinner className="h-4 w-4" /> : <Sparkles className="h-4 w-4" />}
                  Read another text
                </Button>
              </section>
            ) : (
              <div className="flex items-center justify-between rounded-2xl border border-neutral-200 bg-white p-4">
                <p className="text-sm text-neutral-500">
                  {unknown.size > 0 ? `Unknown words: ${unknown.size}` : 'Tap words you don\u2019t know, then finish the page.'}
                </p>
                <Button variant="brand" onClick={finish}>
                  <CheckCheck className="h-4 w-4" /> I've read it
                </Button>
              </div>
            )}
          </ImageApiCtx.Provider>
        )}
      </div>
    </div>
  )
}
