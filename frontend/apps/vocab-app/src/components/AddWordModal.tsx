import { useState } from 'react'
import { ArrowRight, Check, Loader2, Sparkles, TriangleAlert } from 'lucide-react'
import { Button, cn, Input, Modal } from '@els/ui'
import { isApiError } from '@els/api-client'
import { useAddUnit } from '../store/units.ts'
import { KindGlyph } from './KindGlyph.tsx'
import type { AddUnitResult } from '../lib/types.ts'

interface Props {
  onClose: () => void
}

export function AddWordModal({ onClose }: Props) {
  const addM = useAddUnit()
  const [text, setText] = useState('')
  const [result, setResult] = useState<AddUnitResult | null>(null)
  const [error, setError] = useState('')

  async function submit() {
    const value = text.trim()
    if (!value) return
    setError('')
    setResult(null)
    try {
      const res = await addM.mutateAsync(value)
      setResult(res)
      if (res.correct) setText('')
    } catch (e) {
      setError(isApiError(e) && e.status === 409 ? 'This item is already in your collection.' : 'Could not check the word. Try again.')
    }
  }

  function addAnother() {
    setResult(null)
    setText('')
  }

  const added = result?.correct ? result.unit : null
  const correction = result && !result.correct ? result : null

  return (
    <Modal
      onClose={onClose}
      title={
        <>
          <Sparkles className="h-5 w-5 text-brand-600" />
          Add a word
        </>
      }
    >
      {added ? (
        <div className="space-y-4">
          <div className="rounded-2xl bg-brand-50 p-4 ring-1 ring-brand-200">
            <div className="flex items-center gap-2 text-sm font-medium text-brand-700">
              <Check className="h-4 w-4" /> Added to your collection
            </div>
            <p className="mt-2 flex items-center gap-2 text-lg font-semibold text-neutral-900">
              <KindGlyph kind={added.kind} className="h-4 w-4 text-brand-600" />
              {added.text}
            </p>
            {added.translation && <p className="text-sm text-neutral-600">{added.translation}</p>}
            {added.definition && <p className="mt-1 text-sm text-neutral-500">{added.definition}</p>}
          </div>
          <div className="flex gap-3">
            <Button variant="secondary" className="flex-1" onClick={addAnother}>
              Add another
            </Button>
            <Button variant="brand" className="flex-1" onClick={onClose}>
              Done
            </Button>
          </div>
        </div>
      ) : (
        <>
          <label className="mb-1.5 block text-xs font-medium text-neutral-500">
            Word, phrase, phrasal verb or idiom
          </label>
          <div className="flex gap-2">
            <Input
              autoFocus
              value={text}
              onChange={(e) => {
                setText(e.target.value)
                setResult(null)
                setError('')
              }}
              onKeyDown={(e) => e.key === 'Enter' && submit()}
              placeholder="e.g. give up, serendipity, a piece of cake…"
              className="rounded-xl p-3"
            />
            <Button variant="brand" onClick={submit} disabled={addM.isPending || !text.trim()}>
              {addM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Check'}
            </Button>
          </div>
          <p className="mt-2 text-xs text-neutral-400">
            The assistant checks the spelling, then writes a translation, definition and example.
          </p>

          {correction && (
            <div className="mt-4 rounded-2xl bg-amber-50 p-4 ring-1 ring-amber-200">
              <div className="flex items-center gap-2 text-sm font-medium text-amber-700">
                <TriangleAlert className="h-4 w-4" /> Needs a fix
              </div>
              {correction.explanation && <p className="mt-1 text-sm text-amber-800">{correction.explanation}</p>}
              {correction.correction && (
                <button
                  type="button"
                  onClick={() => {
                    setText(correction.correction ?? '')
                    setResult(null)
                  }}
                  className={cn(
                    'mt-3 inline-flex items-center gap-2 rounded-lg bg-white px-3 py-1.5 text-sm font-medium',
                    'text-amber-800 ring-1 ring-amber-200 transition hover:bg-amber-100',
                  )}
                >
                  Use “{correction.correction}” <ArrowRight className="h-3.5 w-3.5" />
                </button>
              )}
            </div>
          )}

          {error && (
            <div className="mt-4 rounded-2xl bg-rose-50 p-4 text-sm text-rose-700 ring-1 ring-rose-200">{error}</div>
          )}
        </>
      )}
    </Modal>
  )
}
