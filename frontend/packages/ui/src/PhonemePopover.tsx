import { useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import { phonemeImage, type PhonemeGuideInfo } from './phonemes.ts'
import { SpeakButton } from './SpeakButton.tsx'

export interface PhonemeAnchor {
  top: number
  bottom: number
  left: number
  width: number
}

export interface PhonemePopoverProps {
  symbol: string
  info?: PhonemeGuideInfo
  anchor: PhonemeAnchor
  onClose: () => void
}

export function anchorOf(el: Element): PhonemeAnchor {
  const r = el.getBoundingClientRect()
  return { top: r.top, bottom: r.bottom, left: r.left, width: r.width }
}

const WIDTH = 264

export function PhonemePopover({ symbol, info, anchor, onClose }: PhonemePopoverProps) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const onDown = (e: PointerEvent) => {
      if (ref.current?.contains(e.target as Node)) return
      e.stopPropagation()
      onClose()
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return
      e.stopPropagation()
      onClose()
    }
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

  const image = phonemeImage(symbol)
  const margin = 8
  const center = anchor.left + anchor.width / 2
  const left = Math.min(Math.max(center - WIDTH / 2, margin), window.innerWidth - WIDTH - margin)
  const below = anchor.top < 320
  const exampleWord = info?.examples?.split(',')[0]?.trim()

  return createPortal(
    <div
      ref={ref}
      style={{
        position: 'fixed',
        left,
        width: WIDTH,
        zIndex: 2147483647,
        ...(below ? { top: anchor.bottom + margin } : { bottom: window.innerHeight - anchor.top + margin }),
      }}
      className="rounded-2xl bg-white p-3 shadow-2xl ring-1 ring-neutral-200"
    >
      <div className="flex items-center gap-2">
        <span className="font-mono text-lg font-bold text-neutral-900">/{symbol}/</span>
        {info?.kind && (
          <span className="rounded-full bg-brand-50 px-1.5 py-0.5 text-[10px] font-medium text-brand-700">{info.kind}</span>
        )}
        {exampleWord && (
          <SpeakButton
            title={`Hear “${exampleWord}”`}
            className="ml-auto rounded-lg p-1.5 hover:text-neutral-600"
            text={exampleWord}
          />
        )}
      </div>
      {image && <img src={image} alt={`Tongue and lip position for /${symbol}/`} className="mx-auto my-1.5 h-28 w-auto" />}
      {info?.description && <p className="text-xs leading-relaxed text-neutral-700">{info.description}</p>}
      {info?.examples && (
        <p className="mt-1 text-xs text-neutral-500">
          As in: <span className="italic">{info.examples}</span>
        </p>
      )}
      {info?.pitfall && <p className="mt-1.5 rounded-lg bg-amber-50 px-2 py-1.5 text-xs text-amber-800">{info.pitfall}</p>}
      {!info && !image && <p className="text-xs text-neutral-500">No articulation notes for this sound.</p>}
    </div>,
    document.body,
  )
}
