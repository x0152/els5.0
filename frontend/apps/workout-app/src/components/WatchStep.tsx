import { useMemo, useRef, useState } from 'react'
import { useFilmUrl } from '../lib/audio.ts'
import type { Watch } from '../lib/types.ts'
import { ContinueButton, StepShell } from './StepShell.tsx'

export function WatchStep({ watch, onDone }: { watch: Watch; onDone: (score: number) => void }) {
  const { film, videoUrl } = useFilmUrl(watch.film_id)
  const videoRef = useRef<HTMLVideoElement>(null)
  const [currentMs, setCurrentMs] = useState(watch.start_ms)
  const [reachedEnd, setReachedEnd] = useState(false)

  const cues = useMemo(() => {
    const track = film?.subtitles?.find((t) => t.lang?.toLowerCase().startsWith('en')) ?? film?.subtitles?.[0]
    return (track?.cues ?? []).filter((c) => c.start_ms >= watch.start_ms && c.end_ms <= watch.end_ms)
  }, [film, watch])

  const activeCue = useMemo(() => cues.find((c) => currentMs >= c.start_ms && currentMs <= c.end_ms), [cues, currentMs])
  const progress = Math.min(Math.max((currentMs - watch.start_ms) / Math.max(watch.end_ms - watch.start_ms, 1), 0), 1)

  return (
    <StepShell>
      <div className="overflow-hidden rounded-xl bg-black">
        {videoUrl ? (
          <video
            ref={videoRef}
            src={videoUrl}
            controls
            playsInline
            className="aspect-video w-full"
            onLoadedMetadata={(e) => {
              e.currentTarget.currentTime = watch.start_ms / 1000
            }}
            onSeeking={(e) => {
              const el = e.currentTarget
              const ms = el.currentTime * 1000
              if (ms < watch.start_ms) el.currentTime = watch.start_ms / 1000
              else if (ms > watch.end_ms) el.currentTime = watch.end_ms / 1000
            }}
            onTimeUpdate={(e) => {
              const ms = Math.round(e.currentTarget.currentTime * 1000)
              setCurrentMs(ms)
              if (ms >= watch.end_ms) {
                e.currentTarget.pause()
                setReachedEnd(true)
              }
            }}
          />
        ) : (
          <div className="flex aspect-video items-center justify-center text-sm text-neutral-400">Loading video…</div>
        )}
      </div>

      <div className="min-h-6 text-center text-[15px] font-medium text-neutral-700">{activeCue?.text.replace(/\n/g, ' ')}</div>

      <div className="h-1.5 overflow-hidden rounded-full bg-neutral-100">
        <div className="h-full rounded-full bg-brand-500 transition-[width]" style={{ width: `${progress * 100}%` }} />
      </div>

      <ContinueButton onClick={() => onDone(100)} label={reachedEnd ? 'Continue' : 'I watched it'} />
    </StepShell>
  )
}
