import { useMemo, useState } from 'react'
import { emitTargetedEvents, emitTextEvents } from '@els/core-events'
import { SpeakButton, anchorOf, cn, speak, type PhonemeAnchor } from '@els/ui'
import { WordPopover } from '@els/lookup'
import { api } from '../lib/api.ts'
import type { Reading, StepOutcome } from '../lib/types.ts'
import { ContinueButton, StepShell } from './StepShell.tsx'

const WORD_RE = /[A-Za-z][A-Za-z''-]*/g

function stripInlineMarkdown(text: string): string {
  return text
    .replace(/\*\*(.+?)\*\*/g, '$1')
    .replace(/__(.+?)__/g, '$1')
    .replace(/(?<!\w)\*(.+?)\*(?!\w)/g, '$1')
    .replace(/(?<!\w)_(.+?)_(?!\w)/g, '$1')
    .replace(/`(.+?)`/g, '$1')
}

function sentenceOf(text: string, word: string): string {
  return text.split(/(?<=[.!?])\s+/).find((s) => s.toLowerCase().includes(word.toLowerCase())) ?? text
}

function selectionPhrase(): { phrase: string; context: string; anchor: PhonemeAnchor } | null {
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed) return null
  const phrase = sel
    .toString()
    .replace(/\s+/g, ' ')
    .replace(/^[^a-zA-Z''-]+|[^a-zA-Z''-]+$/g, '')
    .toLowerCase()
  const words = phrase.match(WORD_RE) ?? []
  if (words.length < 2 || words.length > 8) return null
  const paragraph = sel.anchorNode?.parentElement?.closest('p')?.textContent ?? ''
  const r = sel.getRangeAt(0).getBoundingClientRect()
  return {
    phrase,
    context: sentenceOf(paragraph, phrase),
    anchor: { top: r.top, bottom: r.bottom, left: r.left, width: r.width },
  }
}

export function ReadingStep({ reading, onDone }: { reading: Reading; onDone: (outcome: StepOutcome) => void }) {
  const [unknown, setUnknown] = useState<Set<string>>(new Set())
  const [selected, setSelected] = useState<{ word: string; context: string; anchor: PhonemeAnchor } | null>(null)

  const paragraphs = useMemo(
    () =>
      reading.body
        .split(/\n{2,}/)
        .map((p) => stripInlineMarkdown(p.replace(/^\[image:.*\]$/i, '').replace(/\n/g, ' ').trim()))
        .filter(Boolean),
    [reading.body],
  )
  const learner = useMemo(() => new Set((reading.words ?? []).map((w) => w.toLowerCase())), [reading.words])

  const mark = (key: string, isUnknown: boolean) =>
    setUnknown((prev) => {
      const next = new Set(prev)
      if (isUnknown) next.add(key)
      else next.delete(key)
      return next
    })

  const pickSelection = () => {
    const s = selectionPhrase()
    if (!s) return
    window.getSelection()?.removeAllRanges()
    setSelected({ word: s.phrase, context: s.context, anchor: s.anchor })
  }

  const finish = () => {
    emitTextEvents(api, 'reading', paragraphs, { app: 'workout' })
    emitTargetedEvents(
      api,
      'reading',
      [...unknown].map((w) => ({ target: w, outcome: 'fail' as const })),
      { app: 'workout' },
    )
    onDone({
      score: 100,
      results: [...unknown].map((w) => ({ kind: 'word' as const, text: w, score: 30 })),
    })
  }

  return (
    <StepShell>
      <div className="flex justify-end">
        <SpeakButton
          variant="pill"
          onPlay={() => speak(paragraphs.join(' '))}
          pendingText="Generating audio — a long text takes a while…"
        >
          Listen
        </SpeakButton>
      </div>

      <article
        className="mx-auto flex w-full max-w-3xl flex-col gap-4"
        onMouseUp={pickSelection}
        onTouchEnd={pickSelection}
      >
        {paragraphs.map((p, i) => (
          <TappableText
            key={i}
            text={p}
            unknown={unknown}
            learner={learner}
            onSelect={(word, anchor) => setSelected({ word, context: sentenceOf(p, word), anchor })}
          />
        ))}
      </article>

      <p className="text-xs text-neutral-400">Tap a word — or select a whole phrase — to look it up.</p>

      {unknown.size > 0 && <p className="text-sm text-amber-700">{unknown.size} unknown word(s) will go to your review spiral.</p>}
      <ContinueButton onClick={finish} label="Finished reading" />

      {selected && (
        <WordPopover
          api={api}
          word={selected.word}
          context={selected.context}
          anchor={selected.anchor}
          unknown={unknown.has(selected.word)}
          onMark={(isUnknown) => mark(selected.word, isUnknown)}
          onClose={() => setSelected(null)}
        />
      )}
    </StepShell>
  )
}

function phraseRanges(text: string, unknown: Set<string>): [number, number][] {
  const lower = text.toLowerCase()
  const ranges: [number, number][] = []
  for (const phrase of unknown) {
    if (!phrase.includes(' ')) continue
    let i = 0
    while ((i = lower.indexOf(phrase, i)) !== -1) {
      ranges.push([i, i + phrase.length])
      i += phrase.length
    }
  }
  return ranges
}

function TappableText({
  text,
  unknown,
  learner,
  onSelect,
}: {
  text: string
  unknown: Set<string>
  learner: Set<string>
  onSelect: (key: string, anchor: PhonemeAnchor) => void
}) {
  const parts: { text: string; start: number; key?: string }[] = []
  let pos = 0
  for (const m of text.matchAll(WORD_RE)) {
    if (m.index! > pos) parts.push({ text: text.slice(pos, m.index), start: pos })
    parts.push({ text: m[0], start: m.index!, key: m[0].toLowerCase() })
    pos = m.index! + m[0].length
  }
  if (pos < text.length) parts.push({ text: text.slice(pos), start: pos })

  const ranges = phraseRanges(text, unknown)
  const inPhrase = (p: { start: number; text: string }) =>
    ranges.some(([from, to]) => p.start >= from && p.start + p.text.length <= to)

  return (
    <p className="font-serif text-[17px] leading-8 text-neutral-800">
      {parts.map((p, i) =>
        p.key ? (
          <span
            key={i}
            onClick={(e) => {
              if (window.getSelection()?.isCollapsed === false) return
              onSelect(p.key!, anchorOf(e.currentTarget))
            }}
            className={cn(
              'cursor-pointer rounded px-px transition-colors hover:bg-amber-100',
              learner.has(p.key) && 'underline decoration-brand-300 decoration-2 underline-offset-4',
              (unknown.has(p.key) || inPhrase(p)) && 'bg-amber-200 font-medium text-amber-900',
            )}
          >
            {p.text}
          </span>
        ) : (
          <span key={i} className={cn(inPhrase(p) && 'bg-amber-200 text-amber-900')}>
            {p.text}
          </span>
        ),
      )}
    </p>
  )
}
