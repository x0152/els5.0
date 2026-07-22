import { useEffect, useState, type FormEvent } from 'react'
import { useMutation } from '@tanstack/react-query'
import type { VocabComponents } from '@els/api-client'
import { emitTargetedEvents } from '@els/core-events'
import { Button, Input, SpeakButton, Spinner, cn } from '@els/ui'
import { Check, X } from 'lucide-react'
import { api } from '../lib/api.ts'
import type { ItemResult, StepOutcome, VocabWord } from '../lib/types.ts'
import { ContinueButton, StepShell } from './StepShell.tsx'

type Card = VocabComponents['schemas']['CardOutput']
type CardAnswer = VocabComponents['schemas']['AnswerCardOutput']

export function VocabStep({ words, onDone }: { words: VocabWord[]; onDone: (outcome: StepOutcome) => void }) {
  const deck = useMutation({ mutationFn: () => api.vocab.generateVocabCards({ body: {} }) })

  useEffect(() => {
    deck.mutate()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (deck.isPending || deck.isIdle) {
    return (
      <StepShell>
        <div className="flex flex-1 items-center justify-center py-10 text-neutral-400">
          <Spinner className="h-6 w-6" />
        </div>
      </StepShell>
    )
  }
  const cards = deck.data?.cards ?? []
  if (deck.isError || cards.length === 0) return <Flashcards words={words} onDone={onDone} />
  return <CardsQuiz cards={cards} onDone={onDone} />
}

function CardsQuiz({ cards, onDone }: { cards: Card[]; onDone: (outcome: StepOutcome) => void }) {
  const [index, setIndex] = useState(0)
  const [results, setResults] = useState<ItemResult[]>([])
  const [feedback, setFeedback] = useState<CardAnswer | null>(null)
  const [chosen, setChosen] = useState('')
  const [typed, setTyped] = useState('')
  const card = cards[index]

  const answer = useMutation({
    mutationFn: (value: string) => api.vocab.answerVocabCard({ body: { unit_id: card!.unit_id, answer: value } }),
    onSuccess: (res) => {
      if (!res) return
      setFeedback(res)
      emitTargetedEvents(api, 'reading', [{ target: res.unit.text, outcome: res.correct ? 'ok' : 'fail' }], { app: 'workout' })
      setResults((r) => [...r, { kind: 'word', text: res.unit.text, score: res.correct ? 100 : 0 }])
    },
  })

  const next = () => {
    setFeedback(null)
    setChosen('')
    setTyped('')
    setIndex((i) => i + 1)
  }

  const submitTyped = (e: FormEvent) => {
    e.preventDefault()
    if (typed.trim() && !answer.isPending && !feedback) answer.mutate(typed)
  }

  if (!card) {
    const correct = results.filter((r) => r.score >= 70).length
    const score = results.length ? Math.round((correct / results.length) * 100) : 100
    return (
      <StepShell>
        <div className="flex flex-1 flex-col items-center justify-center gap-2 py-6">
          <span className="text-4xl font-bold text-neutral-900">
            {correct} / {results.length}
          </span>
          <span className="text-sm text-neutral-500">correct — answers move your words through the spaced repetition.</span>
        </div>
        <ContinueButton onClick={() => onDone({ score, results })} />
      </StepShell>
    )
  }

  return (
    <StepShell>
      <div className="flex items-center gap-2">
        <div className="flex flex-1 gap-1">
          {cards.map((_, i) => (
            <span key={i} className={cn('h-1 flex-1 rounded-full', i < index ? 'bg-brand-400' : i === index ? 'bg-brand-200' : 'bg-neutral-100')} />
          ))}
        </div>
        <span className="text-xs font-medium tabular-nums text-neutral-400">
          {index + 1} / {cards.length}
        </span>
      </div>

      <div className="mx-auto flex w-full max-w-lg flex-1 flex-col justify-center gap-5">
        {card.image_url && (
          <div className="mx-auto aspect-[4/3] w-full max-w-xs overflow-hidden rounded-2xl bg-neutral-100">
            <img src={card.image_url} alt="" className="h-full w-full object-cover" />
          </div>
        )}
        {card.direction === 'translation' ? (
          <div className="text-center">
            <p className="text-2xl font-bold text-neutral-900">{card.word}</p>
            {card.transcription && <p className="mt-1 text-sm text-neutral-400">/{card.transcription}/</p>}
            <p className="mt-2 text-sm text-neutral-500">Pick the right translation.</p>
          </div>
        ) : (
          <p className="text-center text-base text-neutral-700">{card.definition || 'Guess the word shown on the picture.'}</p>
        )}

        {card.mode === 'choice' ? (
          <div className="grid gap-2 sm:grid-cols-2">
            {(card.options ?? []).map((opt) => {
              const correctValue = card.direction === 'translation' ? feedback?.unit.translation : feedback?.unit.text
              return (
                <button
                  key={opt}
                  type="button"
                  disabled={!!feedback || answer.isPending}
                  onClick={() => {
                    setChosen(opt)
                    answer.mutate(opt)
                  }}
                  className={cn(
                    'rounded-xl px-4 py-3 text-left text-sm font-medium ring-1 transition',
                    feedback && correctValue === opt
                      ? 'bg-emerald-50 text-emerald-700 ring-emerald-300'
                      : feedback && chosen === opt
                        ? 'bg-rose-50 text-rose-700 ring-rose-300'
                        : 'bg-white text-neutral-800 ring-neutral-200 enabled:hover:bg-neutral-50',
                  )}
                >
                  {opt}
                </button>
              )
            })}
          </div>
        ) : (
          <form onSubmit={submitTyped} className="mx-auto flex w-full max-w-sm gap-2">
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
                {answer.isPending ? <Spinner className="h-4 w-4" /> : 'Check'}
              </Button>
            )}
          </form>
        )}

        {feedback && (
          <div className={cn('rounded-2xl p-4 ring-1', feedback.correct ? 'bg-emerald-50 ring-emerald-200' : 'bg-rose-50 ring-rose-200')}>
            <p className={cn('flex items-center gap-1.5 text-sm font-semibold', feedback.correct ? 'text-emerald-700' : 'text-rose-700')}>
              {feedback.correct ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
              <span className="flex-1">
                {feedback.correct
                  ? 'Correct!'
                  : `Correct answer: ${card.direction === 'translation' ? feedback.unit.translation : feedback.unit.text}`}
              </span>
              <SpeakButton className="p-0 hover:bg-transparent hover:text-neutral-700" text={feedback.unit.text} />
            </p>
            {feedback.unit.example && <p className="mt-1.5 text-sm italic text-neutral-500">{feedback.unit.example}</p>}
            <div className="mt-3 flex justify-end">
              <Button variant="brand" size="sm" autoFocus onClick={next}>
                Next
              </Button>
            </div>
          </div>
        )}
      </div>
    </StepShell>
  )
}

function Flashcards({ words, onDone }: { words: VocabWord[]; onDone: (outcome: StepOutcome) => void }) {
  const [index, setIndex] = useState(0)
  const [revealed, setRevealed] = useState(false)
  const [results, setResults] = useState<ItemResult[]>([])
  const word = words[index]

  const grade = (known: boolean) => {
    if (!word) return
    emitTargetedEvents(api, 'reading', [{ target: word.text, outcome: known ? 'ok' : 'fail' }], { app: 'workout' })
    setResults((r) => [...r, { kind: 'word', text: word.text, score: known ? 100 : 20 }])
    setRevealed(false)
    setIndex((i) => i + 1)
  }

  if (!word) {
    const known = results.filter((r) => r.score >= 70).length
    const score = results.length ? Math.round((known / results.length) * 100) : 100
    return (
      <StepShell>
        <p className="py-6 text-center text-lg text-neutral-700">
          You knew <span className="font-bold text-neutral-900">{known}</span> of {results.length} words.
        </p>
        <ContinueButton onClick={() => onDone({ score, results })} />
      </StepShell>
    )
  }

  return (
    <StepShell>
      <div className="flex items-center gap-2">
        <div className="flex flex-1 gap-1">
          {words.map((_, i) => (
            <span key={i} className={cn('h-1 flex-1 rounded-full', i < index ? 'bg-brand-400' : i === index ? 'bg-brand-200' : 'bg-neutral-100')} />
          ))}
        </div>
        <span className="text-xs font-medium tabular-nums text-neutral-400">
          {index + 1} / {words.length}
        </span>
      </div>
      <div className="flex flex-1 flex-col items-center justify-center gap-4">
        <button
          type="button"
          onClick={() => setRevealed(true)}
          className={cn(
            'mx-auto flex min-h-56 w-full max-w-lg flex-col items-center justify-center gap-3 rounded-2xl border p-6 text-center transition-colors',
            revealed ? 'border-brand-200 bg-brand-50' : 'cursor-pointer border-neutral-200 bg-white hover:bg-neutral-50',
          )}
        >
          <span className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
            {word.text}
            <SpeakButton
              className="p-0 hover:bg-transparent"
              iconClassName="h-5 w-5"
              text={word.text}
              onClick={(e) => e.stopPropagation()}
            />
          </span>
          {revealed ? (
            <span className="flex flex-col gap-1 text-[15px] text-neutral-600">
              {word.translation && <span className="font-medium text-neutral-800">{word.translation}</span>}
              {word.definition && <span>{word.definition}</span>}
              {word.example && <span className="italic text-neutral-500">“{word.example}”</span>}
            </span>
          ) : (
            <span className="text-sm text-neutral-400">Tap to reveal</span>
          )}
        </button>
        {revealed && (
          <div className="flex justify-center gap-2">
            <Button variant="secondary" onClick={() => grade(false)}>
              <X className="h-4 w-4 text-rose-500" /> Still learning
            </Button>
            <Button variant="brand" onClick={() => grade(true)}>
              <Check className="h-4 w-4" /> I know it
            </Button>
          </div>
        )}
      </div>
    </StepShell>
  )
}
