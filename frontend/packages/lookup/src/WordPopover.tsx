import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { useQuery } from '@tanstack/react-query'
import { Check, X } from 'lucide-react'
import { Button, SpeakButton, Spinner, type PhonemeAnchor } from '@els/ui'
import type { Api } from '@els/api-client'

const POPOVER_WIDTH = 320

export function WordPopover({
  api,
  word,
  context,
  anchor,
  unknown,
  onMark,
  onClose,
}: {
  api: Pick<Api, 'vocab' | 'account'>
  word: string
  context: string
  anchor: PhonemeAnchor
  unknown: boolean
  onMark: (unknown: boolean) => void
  onClose: () => void
}) {
  const ref = useRef<HTMLDivElement>(null)
  const meQ = useQuery({ queryKey: ['lookup', 'me'], queryFn: () => api.account.accountMe(), staleTime: 60_000 })
  const showTranslations = meQ.data?.show_translations ?? true
  const lookup = useQuery({
    queryKey: ['lookup', 'word', word],
    queryFn: () => api.vocab.analyzeVocab({ body: { text: word, context } }),
    staleTime: Infinity,
    retry: false,
  })
  const item = lookup.data?.items?.[0]

  useEffect(() => {
    const onDown = (e: PointerEvent) => {
      if (ref.current?.contains(e.target as Node)) return
      onClose()
    }
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    const onScroll = (e: Event) => {
      if (ref.current?.contains(e.target as Node)) return
      onClose()
    }
    document.addEventListener('pointerdown', onDown, true)
    window.addEventListener('keydown', onKey, true)
    window.addEventListener('scroll', onScroll, true)
    return () => {
      document.removeEventListener('pointerdown', onDown, true)
      window.removeEventListener('keydown', onKey, true)
      window.removeEventListener('scroll', onScroll, true)
    }
  }, [onClose])

  const margin = 8
  const center = anchor.left + anchor.width / 2
  const left = Math.min(Math.max(center - POPOVER_WIDTH / 2, margin), window.innerWidth - POPOVER_WIDTH - margin)
  const below = anchor.top < 300

  return createPortal(
    <div
      ref={ref}
      style={{
        position: 'fixed',
        left,
        width: POPOVER_WIDTH,
        zIndex: 2147483647,
        ...(below ? { top: anchor.bottom + margin } : { bottom: window.innerHeight - anchor.top + margin }),
      }}
      className="rounded-2xl bg-white p-3.5 shadow-2xl ring-1 ring-neutral-200"
    >
      <p className="flex items-center gap-2 font-semibold text-neutral-900">
        {word}
        <SpeakButton className="p-0 hover:bg-transparent" text={word} />
      </p>
      {lookup.isPending ? (
        <div className="flex items-center gap-2 py-2 text-sm text-neutral-400">
          <Spinner className="h-4 w-4" /> Looking it up…
        </div>
      ) : item ? (
        <div className="mt-1 space-y-1">
          <p className="text-[11px] font-medium uppercase tracking-wide text-neutral-400">
            {item.kind}
            {item.cefr && ` · ${item.cefr}`}
          </p>
          {item.description && <p className="text-sm text-neutral-700">{item.description}</p>}
          {showTranslations && item.translation && <p className="text-sm font-medium text-neutral-900">{item.translation}</p>}
        </div>
      ) : (
        <p className="py-2 text-sm text-neutral-400">Could not look this word up.</p>
      )}
      <div className="mt-3 flex gap-2">
        <Button
          size="sm"
          variant={unknown ? 'brand' : 'secondary'}
          className="flex-1"
          onClick={() => {
            onMark(true)
            onClose()
          }}
        >
          <X className="h-4 w-4" /> Don't know
        </Button>
        <Button
          size="sm"
          variant={unknown ? 'secondary' : 'brand'}
          className="flex-1"
          onClick={() => {
            onMark(false)
            onClose()
          }}
        >
          <Check className="h-4 w-4" /> I know it
        </Button>
      </div>
    </div>,
    document.body,
  )
}
