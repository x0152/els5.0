import { useMemo, useState } from 'react'
import { emitTextEvents } from '@els/core-events'
import { englishTrackIdx, FilmPlayer } from '@els/ui'
import { api } from '../lib/api.ts'
import { useFilmUrl } from '../lib/audio.ts'
import type { Watch } from '../lib/types.ts'
import { ContinueButton, StepShell } from './StepShell.tsx'

export function WatchStep({ watch, onDone }: { watch: Watch; onDone: (score: number) => void }) {
  const { film } = useFilmUrl(watch.film_id)
  const [currentMs, setCurrentMs] = useState(watch.start_ms)
  const [reachedEnd, setReachedEnd] = useState(false)
  const [audioIdx, setAudioIdx] = useState<number | null>(null)
  const [subIdx, setSubIdx] = useState<number | null>(null)

  const audioTracks = useMemo(() => film?.audio_tracks ?? [], [film])
  const subtitleTracks = useMemo(
    () => (film?.subtitles ?? []).map((t) => ({ ...t, cues: t.cues ?? [] })),
    [film],
  )
  const videoUrl = audioTracks[audioIdx ?? englishTrackIdx(audioTracks)]?.url ?? ''

  const progress = Math.min(Math.max((currentMs - watch.start_ms) / Math.max(watch.end_ms - watch.start_ms, 1), 0), 1)

  return (
    <StepShell>
      <div className="flex aspect-video overflow-hidden rounded-xl bg-black">
        {videoUrl ? (
          <FilmPlayer
            videoUrl={videoUrl}
            audioTracks={audioTracks}
            subtitleTracks={subtitleTracks}
            audioIdx={audioIdx ?? englishTrackIdx(audioTracks)}
            subIdx={subIdx ?? englishTrackIdx(subtitleTracks)}
            onAudioChange={setAudioIdx}
            onSubChange={setSubIdx}
            startMs={watch.start_ms}
            endMs={watch.end_ms}
            onTimeChange={setCurrentMs}
            onWindowEnd={() => setReachedEnd(true)}
          />
        ) : (
          <div className="flex flex-1 items-center justify-center text-sm text-neutral-400">Loading video…</div>
        )}
      </div>

      <div className="h-1.5 overflow-hidden rounded-full bg-neutral-100">
        <div className="h-full rounded-full bg-brand-500 transition-[width]" style={{ width: `${progress * 100}%` }} />
      </div>

      <ContinueButton
        onClick={() => {
          const cues = subtitleTracks[englishTrackIdx(subtitleTracks)]?.cues ?? []
          const text = cues
            .filter((c) => c.start_ms >= watch.start_ms && c.end_ms <= watch.end_ms)
            .map((c) => c.text.replace(/<[^>]+>/g, ' '))
            .join(' ')
          emitTextEvents(api, 'listening', [text], { app: 'workout', film_id: watch.film_id })
          onDone(100)
        }}
        label={reachedEnd ? 'Continue' : 'I watched it'}
      />
    </StepShell>
  )
}
