import { Volume2 } from 'lucide-react'
import { Badge, Button, Modal, phonemeImage, speak } from '@els/ui'
import { VERDICT_LABELS, type PhonemeInfo, type PhonemeResult, type Verdict, type WordResult } from '../lib/types.ts'

const VERDICT_BADGES: Record<Verdict, 'success' | 'warning' | 'danger' | 'neutral'> = {
  good: 'success',
  close: 'warning',
  wrong: 'danger',
  missing: 'neutral',
}

interface Props {
  word: WordResult
  phoneme: PhonemeResult
  guide: Map<string, PhonemeInfo>
  onClose: () => void
}

function SoundCard({ title, info, symbol }: { title: string; info?: PhonemeInfo; symbol: string }) {
  const image = phonemeImage(symbol)
  return (
    <div className="rounded-2xl bg-neutral-50 p-4">
      <div className="mb-2 flex items-center gap-3">
        <span className="text-xs font-medium uppercase tracking-wide text-neutral-500">{title}</span>
        <span className="font-mono text-2xl font-bold text-neutral-900">/{symbol}/</span>
        {info && <Badge tone="brand">{info.kind}</Badge>}
      </div>
      <div className="flex items-start gap-4">
        {image && (
          <img
            src={image}
            alt={`Tongue and lip position for /${symbol}/`}
            className="h-32 w-auto shrink-0 rounded-xl bg-white ring-1 ring-neutral-200"
          />
        )}
        <div className="min-w-0">
          {info ? (
            <>
              <p className="text-sm text-neutral-700">{info.description}</p>
              <p className="mt-1 text-sm text-neutral-500">
                As in: <span className="italic">{info.examples}</span>
              </p>
              {info.pitfall && (
                <p className="mt-2 rounded-xl bg-amber-50 px-3 py-2 text-sm text-amber-800">{info.pitfall}</p>
              )}
            </>
          ) : (
            <p className="text-sm text-neutral-500">No articulation notes for this sound.</p>
          )}
        </div>
      </div>
    </div>
  )
}

export function PhonemeDetailModal({ word, phoneme, guide, onClose }: Props) {
  const verdict = phoneme.verdict as Verdict
  return (
    <Modal
      onClose={onClose}
      title={
        <>
          <span className="font-mono">/{phoneme.expected}/</span>
          <span className="text-neutral-400">in</span>
          <span>{word.word}</span>
        </>
      }
    >
      <div className="space-y-4">
        <div className="flex items-center gap-3">
          <Badge tone={VERDICT_BADGES[verdict] ?? 'neutral'}>{VERDICT_LABELS[verdict] ?? verdict}</Badge>
          <span className="text-sm tabular-nums text-neutral-500">match {Math.round(phoneme.score * 100)}%</span>
          <Button variant="secondary" size="sm" onClick={() => speak(word.word)}>
            <Volume2 className="h-4 w-4" />
            Hear the word
          </Button>
        </div>

        <SoundCard title="Target sound" symbol={phoneme.expected} info={guide.get(phoneme.expected)} />

        {phoneme.heard && phoneme.heard !== phoneme.expected && (
          <SoundCard title="What you said" symbol={phoneme.heard} info={guide.get(phoneme.heard)} />
        )}
        {verdict === 'missing' && (
          <p className="text-sm text-neutral-600">
            This sound was not detected in your recording. Slow down and articulate it deliberately.
          </p>
        )}
      </div>
    </Modal>
  )
}
