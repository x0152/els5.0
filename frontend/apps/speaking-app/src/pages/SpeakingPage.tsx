import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { BookOpen, Loader2, Mic, CirclePlay, Sparkles, Square } from 'lucide-react'
import { AppInfoButton, Button, ErrorState, SpeakButton, Textarea, cn, useAgentView, useRecorder } from '@els/ui'
import { api } from '../lib/api.ts'
import { spoken } from '../lib/events.ts'
import { buildIssues, type Assessment, type PhonemeResult, type WordResult } from '../lib/types.ts'
import { ScoreRing } from '../components/ScoreRing.tsx'
import { WordBreakdown } from '../components/WordBreakdown.tsx'
import { PhonemeDetailModal } from '../components/PhonemeDetailModal.tsx'
import { FeedbackPanel } from '../components/FeedbackPanel.tsx'

const PRESETS = [
  'I think this is the third Thursday of the month.',
  'The weather was really very good this evening.',
  'Would you like a cup of water with your vegetables?',
  'She usually works with three other girls in the village.',
]

function weakSounds(assessment: Assessment): string[] {
  const sounds = new Set<string>()
  for (const w of assessment.words ?? [])
    for (const p of w.phonemes ?? [])
      if (p.verdict === 'wrong' || p.verdict === 'missing') sounds.add(p.expected)
  return [...sounds].slice(0, 6)
}

export function SpeakingPage() {
  const navigate = useNavigate()
  const [presets, setPresets] = useState<string[]>(PRESETS)
  const [text, setText] = useState<string>(PRESETS[0]!)
  const [assessment, setAssessment] = useState<Assessment | null>(null)
  const [assessedText, setAssessedText] = useState('')
  const [selected, setSelected] = useState<{ word: WordResult; phoneme: PhonemeResult } | null>(null)

  const meQ = useQuery({
    queryKey: ['speaking', 'me'],
    queryFn: () => api.account.accountMe(),
    staleTime: 60_000,
  })
  const native = meQ.data?.native_language || 'Russian'
  const strictness = meQ.data?.speech_strictness ?? 0.5
  const strictnessLabel = strictness >= 2 ? 'Strict' : strictness >= 1 ? 'Normal' : 'Easy'

  const guideQ = useQuery({
    queryKey: ['speech', 'phonemes'],
    queryFn: () => api.speech.listSpeechPhonemes(),
    staleTime: Infinity,
  })
  const guide = useMemo(
    () => new Map((guideQ.data?.items ?? []).map((p) => [p.symbol, p])),
    [guideQ.data],
  )

  const assessM = useMutation({
    mutationFn: (blob: Blob) => {
      const form = new FormData()
      form.append('audio', blob, 'recording.webm')
      form.append('text', text.trim())
      form.append('strictness', String(strictness))
      return api.speech.assessSpeech({ body: form as unknown as never })
    },
    onSuccess: (data) => {
      setAssessment(data ?? null)
      setAssessedText(text.trim())
      feedbackM.reset()
      if (data) spoken(text.trim(), data.overall)
    },
  })

  const feedbackM = useMutation({
    mutationFn: () => {
      if (!assessment) throw new Error('no assessment')
      return api.speech.speechFeedback({
        body: {
          text: assessedText,
          heard: assessment.heard,
          native_language: native,
          issues: buildIssues(assessment),
        },
      })
    },
  })

  const practiceM = useMutation({
    mutationFn: (sounds: string[]) => api.speech.speechGeneratePractice({ body: { sounds } }),
    onSuccess: (data) => {
      const sentences = data?.sentences ?? []
      if (!sentences.length) return
      setPresets(sentences)
      setText(sentences[0]!)
    },
  })

  const recorder = useRecorder((blob) => assessM.mutate(blob))
  const busy = assessM.isPending || recorder.state === 'recording'

  useAgentView({
    app: 'speaking',
    screen: 'practice',
    info: 'The user practices English pronunciation: they read a text aloud and get phoneme-level scoring.',
    state: assessment ? { text, score: assessment.overall } : { text },
  })

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-4xl space-y-6 p-6">
        <header className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
              <Mic className="h-6 w-6" />
            </div>
            <div>
              <h1 className="flex items-center gap-1.5 text-2xl font-bold text-neutral-900">
                Speaking <AppInfoButton />
              </h1>
              <p className="text-sm text-neutral-500">
                Read the text aloud and get sound-by-sound pronunciation feedback.
              </p>
            </div>
          </div>
          <Button variant="secondary" onClick={() => navigate('sounds')}>
            <BookOpen className="h-4 w-4" />
            Sound guide
          </Button>
        </header>

        <section className="relative space-y-4 overflow-hidden rounded-2xl border border-brand-200 bg-gradient-to-br from-brand-50 to-white p-5 shadow-sm">
          <Mic className="pointer-events-none absolute -right-5 -top-5 h-28 w-28 text-brand-100" />
          <Textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            rows={3}
            maxLength={500}
            placeholder="Type or pick a sentence to read aloud…"
            disabled={recorder.state === 'recording'}
            className="relative"
          />
          <div className="relative flex flex-wrap gap-2">
            {presets.map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => setText(p)}
                className={cn(
                  'rounded-full px-3 py-1.5 text-xs ring-1 transition',
                  p === text
                    ? 'bg-brand-600 text-white ring-brand-600'
                    : 'bg-white text-neutral-600 ring-neutral-200 hover:bg-neutral-50',
                )}
              >
                {p}
              </button>
            ))}
            <button
              type="button"
              onClick={() => practiceM.mutate([])}
              disabled={practiceM.isPending}
              className="flex items-center gap-1 rounded-full px-3 py-1.5 text-xs text-brand-600 ring-1 ring-brand-200 transition hover:bg-brand-50 disabled:opacity-50"
            >
              {practiceM.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Sparkles className="h-3 w-3" />}
              New sentences
            </button>
          </div>

          <div className="relative flex flex-wrap items-center gap-4 border-t border-brand-100 pt-4">
            <button
              type="button"
              onClick={recorder.state === 'recording' ? recorder.stop : recorder.start}
              disabled={!text.trim() || assessM.isPending}
              title={recorder.state === 'recording' ? 'Stop recording' : 'Record'}
              className={cn(
                'flex h-16 w-16 shrink-0 items-center justify-center rounded-full text-white shadow-lg transition active:scale-95 disabled:opacity-50',
                recorder.state === 'recording'
                  ? 'animate-pulse bg-red-600 shadow-red-600/30'
                  : 'bg-gradient-to-br from-brand-500 to-brand-700 shadow-brand-600/30 hover:brightness-110',
              )}
            >
              {assessM.isPending ? (
                <Loader2 className="h-7 w-7 animate-spin" />
              ) : recorder.state === 'recording' ? (
                <Square className="h-6 w-6" fill="currentColor" />
              ) : (
                <Mic className="h-7 w-7" />
              )}
            </button>
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-neutral-800">
                {assessM.isPending
                  ? 'Scoring your pronunciation…'
                  : recorder.state === 'recording'
                    ? `Recording · ${recorder.elapsed}s — tap to stop`
                    : 'Tap to record yourself reading the text.'}
              </p>
              <div className="mt-2 flex flex-wrap items-center gap-2">
                <SpeakButton variant="button" className="h-8 px-3" text={text} disabled={!text.trim()}>
                  Example
                </SpeakButton>
                {recorder.blob && (
                  <Button variant="secondary" size="sm" onClick={recorder.play} disabled={recorder.state === 'recording'}>
                    <CirclePlay className="h-4 w-4" />
                    My recording
                  </Button>
                )}
                <p className="text-xs text-neutral-400">
                  {strictnessLabel} scoring ·
                  <button
                    type="button"
                    onClick={() => navigate('/v1/profile')}
                    className="ml-1 text-brand-600 hover:underline"
                    disabled={busy}
                  >
                    change in profile
                  </button>
                </p>
              </div>
            </div>
          </div>
          {recorder.state === 'unsupported' && (
            <p className="text-sm text-red-600">Microphone access is unavailable. Allow it in browser settings.</p>
          )}
        </section>

        {assessM.isError && (
          <ErrorState
            title="Scoring failed"
            description="The pronunciation service did not respond. Try again."
            action={
              <Button variant="secondary" onClick={() => assessM.reset()}>
                Dismiss
              </Button>
            }
          />
        )}

        {assessment && (
          <section className="space-y-5 rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
            <div className="flex flex-wrap items-center gap-6">
              <ScoreRing score={assessment.overall} />
              <div className="min-w-0 flex-1">
                <p className="text-xs font-semibold uppercase tracking-wide text-neutral-500">What the model heard</p>
                <p className="mt-1.5 break-words rounded-lg bg-neutral-50 px-3 py-2 font-mono text-sm text-neutral-600">/{assessment.heard}/</p>
                <p className="mt-2.5 text-xs text-neutral-400">
                  Click any sound below to see how to articulate it.
                </p>
              </div>
            </div>

            <WordBreakdown
              words={assessment.words ?? []}
              onSelect={(word, phoneme) => setSelected({ word, phoneme })}
            />

            <div className="flex flex-wrap items-center gap-3 border-t border-neutral-100 pt-4">
              <Button variant="brand" onClick={() => feedbackM.mutate()} disabled={feedbackM.isPending}>
                {feedbackM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                {feedbackM.isPending ? 'Analyzing…' : 'Explain with AI'}
              </Button>
              {weakSounds(assessment).length > 0 && (
                <Button variant="secondary" onClick={() => practiceM.mutate(weakSounds(assessment))} disabled={practiceM.isPending}>
                  {practiceM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Mic className="h-4 w-4" />}
                  Practice {weakSounds(assessment).map((s) => `/${s}/`).join(' ')}
                </Button>
              )}
              <span className="text-sm text-neutral-500">Advice is tailored for {native} speakers (from your profile).</span>
            </div>

            {feedbackM.isError && (
              <p className="text-sm text-red-600">The AI coach is unavailable right now. Try again later.</p>
            )}
            {feedbackM.data && <FeedbackPanel feedback={feedbackM.data} />}
          </section>
        )}

        {selected && (
          <PhonemeDetailModal
            word={selected.word}
            phoneme={selected.phoneme}
            guide={guide}
            onClose={() => setSelected(null)}
          />
        )}
      </div>
    </div>
  )
}
