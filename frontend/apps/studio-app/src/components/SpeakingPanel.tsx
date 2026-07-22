import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Check, Loader2, Mic, Sparkles, Square, X } from 'lucide-react'
import { SpeakButton, cn, useRecorder } from '@els/ui'
import { api } from '../lib/api.ts'
import type { Item } from '../lib/types.ts'

const PASS_SCORE = 75

type Assessment = NonNullable<Awaited<ReturnType<typeof api.speech.assessSpeech>>>

function weakSounds(assessment: Assessment) {
  const sounds: { symbol: string; ok: boolean }[] = []
  for (const w of assessment.words ?? [])
    for (const p of w.phonemes ?? [])
      sounds.push({ symbol: p.expected, ok: p.verdict === 'good' || p.verdict === 'close' })
  return sounds.slice(0, 18)
}

export function SpeakingPanel({ item, onDone }: { item: Item; onDone: () => void }) {
  const queryClient = useQueryClient()
  const [assessment, setAssessment] = useState<Assessment | null>(null)
  const [sentence, setSentence] = useState('')
  const target = sentence || item.text

  const sentenceM = useMutation({
    mutationFn: () => api.studio.studioRegenExample({ params: { path: { id: item.id } } }),
    onSuccess: (data) => {
      if (!data?.example) return
      setSentence(data.example)
      setAssessment(null)
      queryClient.invalidateQueries({ queryKey: ['studio', 'items'] })
    },
  })

  const assessM = useMutation({
    mutationFn: (blob: Blob) => {
      const form = new FormData()
      form.append('audio', blob, 'recording.webm')
      form.append('text', target)
      return api.speech.assessSpeech({ body: form as unknown as never })
    },
    onSuccess: (data) => {
      if (!data) return
      setAssessment(data)
      if (data.overall >= PASS_SCORE) onDone()
    },
  })

  const recorder = useRecorder((blob) => assessM.mutate(blob))
  const recording = recorder.state === 'recording'

  return (
    <div className="flex flex-1 flex-col rounded-2xl border border-emerald-200 bg-white shadow-sm">
      <div className="flex items-center justify-between border-b border-neutral-100 px-4 py-3">
        <span className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-brand-50 text-brand-600">
            <Mic className="h-4 w-4" />
          </span>
          Speaking
        </span>
        {assessment ? (
          <span
            className={cn(
              'flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold ring-1',
              assessment.overall >= PASS_SCORE
                ? 'bg-emerald-50 text-emerald-700 ring-emerald-200'
                : 'bg-amber-50 text-amber-700 ring-amber-200',
            )}
          >
            {Math.round(assessment.overall)}
          </span>
        ) : (
          item.spoken && (
            <span className="flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 ring-1 ring-emerald-200">
              <Check className="h-3 w-3" /> done
            </span>
          )
        )}
      </div>
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto p-4">
        {sentence && (
          <div className="flex items-start gap-1.5 rounded-xl bg-neutral-50 px-3 py-2">
            <p className="flex-1 text-sm leading-relaxed text-neutral-800">{sentence}</p>
            <SpeakButton title="Listen" className="shrink-0" iconClassName="h-3.5 w-3.5" text={sentence} />
            <button
              onClick={() => {
                setSentence('')
                setAssessment(null)
              }}
              title="Back to the phrase"
              className="shrink-0 rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-700"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        )}
        {assessment ? (
          <>
            <p className="text-xs text-neutral-400">Sounds, weakest highlighted:</p>
            <div className="flex flex-wrap gap-1">
              {weakSounds(assessment).map((s, i) => (
                <span
                  key={i}
                  className={cn(
                    'rounded-lg px-2 py-1 font-mono text-xs ring-1',
                    s.ok
                      ? 'bg-emerald-50 text-emerald-700 ring-emerald-200'
                      : 'bg-red-50 text-red-700 ring-red-300',
                  )}
                >
                  {s.symbol}
                </span>
              ))}
            </div>
          </>
        ) : (
          <p className="text-xs text-neutral-400">
            Read the {sentence ? 'sentence' : 'phrase'} aloud and get sound-by-sound feedback.
          </p>
        )}
        <div className="mt-auto flex gap-2">
          <button
            onClick={() => sentenceM.mutate()}
            disabled={sentenceM.isPending}
            title={sentence ? 'Another sentence' : 'Practice in a sentence'}
            className="inline-flex h-9 shrink-0 items-center justify-center rounded-lg px-2.5 text-neutral-500 ring-1 ring-inset ring-neutral-200 hover:bg-neutral-50 hover:text-brand-600 disabled:opacity-50"
          >
            {sentenceM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
          </button>
          <button
            onClick={() => (recording ? recorder.stop() : recorder.start())}
            disabled={assessM.isPending}
            className={cn(
              'inline-flex h-9 flex-1 items-center justify-center gap-2 rounded-lg px-4 text-sm font-semibold shadow-sm disabled:opacity-50',
              recording
                ? 'bg-red-600 text-white hover:bg-red-700'
                : 'bg-white text-neutral-800 ring-1 ring-inset ring-neutral-200 hover:bg-neutral-50',
            )}
          >
            {assessM.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : recording ? (
              <Square className="h-4 w-4" />
            ) : (
              <Mic className="h-4 w-4" />
            )}
            {assessM.isPending ? 'Scoring…' : recording ? 'Stop' : assessment ? 'Record again' : 'Record'}
          </button>
        </div>
      </div>
    </div>
  )
}
