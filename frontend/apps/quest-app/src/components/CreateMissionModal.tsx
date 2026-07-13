import { useState } from 'react'
import { Loader2, Sparkles } from 'lucide-react'
import { Button, cn, Modal, Textarea } from '@els/ui'
import { GENRES } from '../lib/helpers.ts'

interface Props {
  submitting: boolean
  error?: string | null
  onClose: () => void
  onCreate: (payload: { prompt: string; genre: string; practiceGoals: string }) => void
}

export function CreateMissionModal({ submitting, error, onClose, onCreate }: Props) {
  const [genre, setGenre] = useState('random')
  const [prompt, setPrompt] = useState('')
  const [practiceGoals, setPracticeGoals] = useState('')

  function submit(randomMode: boolean) {
    const text = randomMode ? '' : prompt.trim()
    if (!randomMode && !text) return
    onCreate({ prompt: text, genre, practiceGoals: practiceGoals.trim() })
  }

  return (
    <Modal
      onClose={() => !submitting && onClose()}
      title={
        <>
          <Sparkles className="h-5 w-5 text-brand-600" />
          New adventure
        </>
      }
    >
      <label className="mb-1.5 block text-xs font-medium text-neutral-500">Genre</label>
      <div className="mb-4 flex flex-wrap gap-2">
        {GENRES.map((g) => (
          <button
            key={g.id}
            type="button"
            onClick={() => setGenre(g.id)}
            className={cn(
              'rounded-full px-3 py-1.5 text-sm font-medium ring-1 transition',
              genre === g.id
                ? 'bg-brand-600 text-white ring-brand-600'
                : 'bg-white text-neutral-700 ring-neutral-200 hover:bg-neutral-50',
            )}
          >
            {g.emoji} {g.label}
          </button>
        ))}
      </div>

      <label className="mb-1.5 block text-xs font-medium text-neutral-500">What is the story about?</label>
      <Textarea
        rows={3}
        value={prompt}
        onChange={(e) => setPrompt(e.target.value)}
        placeholder="e.g. I'm a detective investigating a disappearance in a small town…"
        className="mb-4 rounded-xl p-3"
      />

      <label className="mb-1.5 block text-xs font-medium text-neutral-500">
        What would you like to practice? <span className="text-neutral-400">(optional)</span>
      </label>
      <Textarea
        rows={2}
        value={practiceGoals}
        onChange={(e) => setPracticeGoals(e.target.value)}
        placeholder="e.g. past tenses, phrasal verbs…"
        className="mb-6 rounded-xl p-3"
      />

      {error && <p className="mb-3 text-sm text-rose-600">{error}</p>}

      <div className="flex gap-3">
        <Button variant="secondary" className="flex-1" onClick={() => submit(true)} disabled={submitting}>
          🎲 Random
        </Button>
        <Button
          variant="brand"
          className="flex-1"
          onClick={() => submit(false)}
          disabled={submitting || !prompt.trim()}
        >
          {submitting ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Create'}
        </Button>
      </div>
    </Modal>
  )
}
