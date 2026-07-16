import { Fragment, useEffect, useMemo, useRef } from 'react'
import { BookPlus, MessageCircleQuestion } from 'lucide-react'
import { cn } from '@els/ui'
import type { Cue } from '../lib/types.ts'
import { requestAnalyze, requestAsk } from '../lib/events.ts'
import { CueText } from './CueText.tsx'

interface Props {
  cues: Cue[]
  activeIdx: number
  currentMs: number
  autoScroll: boolean
  onToggleAutoScroll: () => void
  onSeek: (ms: number) => void
}

function formatTime(ms: number): string {
  const total = Math.max(0, Math.floor(ms / 1000))
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${m}:${`${s}`.padStart(2, '0')}`
}

export function SubtitlePanel({ cues, activeIdx, currentMs, autoScroll, onToggleAutoScroll, onSeek }: Props) {
  const activeRef = useRef<HTMLDivElement>(null)

  const inGap = activeIdx < 0
  const gapIdx = useMemo(() => {
    if (!inGap) return -1
    const idx = cues.findIndex((c) => c.start_ms > currentMs)
    return idx === -1 ? cues.length : idx
  }, [inGap, cues, currentMs])

  useEffect(() => {
    if (autoScroll && activeRef.current) {
      activeRef.current.scrollIntoView({ block: 'center', behavior: 'smooth' })
    }
  }, [activeIdx, gapIdx, autoScroll])

  return (
    <aside className="flex min-h-0 w-full flex-1 flex-col border-t border-neutral-200 bg-white lg:w-96 lg:flex-none lg:border-l lg:border-t-0">
      <div className="flex shrink-0 items-center justify-between border-b border-neutral-100 px-4 py-3">
        <span className="text-xs font-bold uppercase tracking-wider text-neutral-400">Subtitles</span>
        <button
          type="button"
          onClick={onToggleAutoScroll}
          className={cn(
            'rounded-full px-2.5 py-0.5 text-[11px] font-medium ring-1 transition-colors',
            autoScroll
              ? 'bg-brand-600 text-white ring-brand-600'
              : 'bg-white text-neutral-500 ring-neutral-200 hover:bg-neutral-50',
          )}
        >
          auto-scroll
        </button>
      </div>
      <div className="min-h-0 flex-1 space-y-0.5 overflow-y-auto px-2 py-2">
        {cues.map((cue, i) => {
          const active = i === activeIdx
          return (
            <Fragment key={cue.index}>
              {inGap && gapIdx === i && (
                <GapMarker ref={activeRef} secondsToNext={Math.max(0, Math.ceil((cue.start_ms - currentMs) / 1000))} />
              )}
            <div
              ref={active ? activeRef : undefined}
              className={cn(
                'group flex w-full gap-3 rounded-lg px-3 py-2 transition-colors',
                active ? 'bg-brand-50 ring-1 ring-brand-200' : 'hover:bg-neutral-50',
              )}
            >
              <button
                type="button"
                onClick={() => onSeek(cue.start_ms)}
                title="Jump to this line"
                className={cn(
                  'mt-0.5 shrink-0 text-xs tabular-nums transition-colors hover:text-brand-600',
                  active ? 'font-semibold text-brand-600' : 'text-neutral-400',
                )}
              >
                {formatTime(cue.start_ms)}
              </button>
              <span
                className={cn(
                  'select-text cursor-text text-sm leading-snug',
                  active ? 'font-medium text-brand-800' : 'text-neutral-600',
                )}
              >
                <CueText text={cue.text} />
              </span>
              <div className="flex shrink-0 self-center gap-1 opacity-0 transition-opacity focus-within:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100">
                <button
                  type="button"
                  onClick={() => requestAnalyze(cue.text)}
                  title="Analyze this line"
                  className="rounded-full bg-white p-1 text-neutral-400 shadow ring-1 ring-neutral-200 transition-colors hover:text-brand-600"
                >
                  <BookPlus size={14} />
                </button>
                <button
                  type="button"
                  onClick={() => requestAsk(cue.text)}
                  title="Ask the assistant"
                  className="rounded-full bg-white p-1 text-neutral-400 shadow ring-1 ring-neutral-200 transition-colors hover:text-brand-600"
                >
                  <MessageCircleQuestion size={14} />
                </button>
              </div>
            </div>
            </Fragment>
          )
        })}
        {inGap && gapIdx === cues.length && <GapMarker ref={activeRef} secondsToNext={null} />}
      </div>
    </aside>
  )
}

function GapMarker({ secondsToNext, ref }: { secondsToNext: number | null; ref?: React.Ref<HTMLDivElement> }) {
  return (
    <div ref={ref} className="flex items-center gap-2 px-3 py-1.5">
      <span className="relative flex h-2 w-2 shrink-0">
        <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-400 opacity-75" />
        <span className="relative inline-flex h-2 w-2 rounded-full bg-brand-500" />
      </span>
      <span className="shrink-0 text-[10px] tabular-nums text-neutral-400">
        {secondsToNext != null ? `next in ${secondsToNext}s` : 'end'}
      </span>
      <span className="h-px flex-1 bg-gradient-to-r from-brand-300 to-neutral-200" />
    </div>
  )
}
