import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
  type ReactNode,
  type RefObject,
} from 'react'
import {
  Captions,
  CaptionsOff,
  Maximize,
  Minimize,
  Pause,
  Play,
  Volume2,
  VolumeX,
} from 'lucide-react'
import { cn } from './cn.ts'
import { Select } from './Select.tsx'
import { CueText } from './CueText.tsx'
import { Spinner } from './Spinner.tsx'

export interface PlayerCue {
  index: number
  start_ms: number
  end_ms: number
  text: string
}

export interface PlayerAudioTrack {
  lang: string
  label: string
  url: string
}

export interface PlayerSubtitleTrack {
  lang: string
  label: string
  cues: PlayerCue[]
}

export function englishTrackIdx(tracks: { lang: string; label: string }[]): number {
  let idx = tracks.findIndex((t) => t.lang.toLowerCase().startsWith('en'))
  if (idx < 0) idx = tracks.findIndex((t) => /english|orig/i.test(t.label))
  return idx >= 0 ? idx : 0
}

const darkSelectClass = 'w-auto rounded-md px-2 py-1 text-xs'

function formatTime(ms: number): string {
  const total = Math.max(0, Math.floor(ms / 1000))
  const s = total % 60
  const m = Math.floor(total / 60) % 60
  const h = Math.floor(total / 3600)
  const mm = `${m}`.padStart(h > 0 ? 2 : 1, '0')
  const ss = `${s}`.padStart(2, '0')
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`
}

const clamp = (v: number, min: number, max: number) => Math.min(Math.max(v, min), max)

export interface FilmPlayerProps {
  videoUrl: string
  audioTracks: PlayerAudioTrack[]
  subtitleTracks: PlayerSubtitleTrack[]
  audioIdx?: number
  subIdx?: number
  onAudioChange?: (idx: number) => void
  onSubChange?: (idx: number) => void
  /** Playback window: seeking and the scrubber are clamped to [startMs, endMs]. */
  startMs?: number
  endMs?: number
  onWindowEnd?: () => void
  onTimeChange?: (ms: number) => void
  onLoadedMetadata?: (video: HTMLVideoElement) => void
  durationMs?: number
  videoRef?: RefObject<HTMLVideoElement | null>
  renderCueOverlay?: (cue: PlayerCue) => ReactNode
  controlsStart?: ReactNode
  controlsEnd?: ReactNode
  className?: string
}

export function FilmPlayer({
  videoUrl,
  audioTracks,
  subtitleTracks,
  audioIdx = 0,
  subIdx = 0,
  onAudioChange,
  onSubChange,
  startMs,
  endMs,
  onWindowEnd,
  onTimeChange,
  onLoadedMetadata,
  durationMs = 0,
  videoRef,
  renderCueOverlay,
  controlsStart,
  controlsEnd,
  className,
}: FilmPlayerProps) {
  const innerRef = useRef<HTMLVideoElement | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const setVideoRef = useCallback(
    (el: HTMLVideoElement | null) => {
      innerRef.current = el
      if (videoRef) videoRef.current = el
    },
    [videoRef],
  )

  const [currentMs, setCurrentMs] = useState(startMs ?? 0)
  const [measuredMs, setMeasuredMs] = useState(0)
  const [bufferedMs, setBufferedMs] = useState(0)
  const [paused, setPaused] = useState(true)
  const [volume, setVolume] = useState(1)
  const [muted, setMuted] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [showOverlay, setShowOverlay] = useState(true)

  // Stall tracking. `seeking` and `waiting` mirror the media element state via
  // events; the spinner is derived from them (with a short delay to avoid
  // flashing on fast seeks) instead of being toggled ad hoc, so it can never
  // stick while playback is actually progressing.
  const [seeking, setSeeking] = useState(false)
  const [waiting, setWaiting] = useState(false)
  const stalled = seeking || (!paused && waiting)
  const [spinner, setSpinner] = useState(false)
  useEffect(() => {
    if (!stalled) return
    const t = window.setTimeout(() => setSpinner(true), 250)
    return () => {
      window.clearTimeout(t)
      setSpinner(false)
    }
  }, [stalled])

  const cues = useMemo(() => subtitleTracks[subIdx]?.cues ?? [], [subtitleTracks, subIdx])
  const hasSubtitles = cues.length > 0
  const activeCue = useMemo(() => {
    for (let i = cues.length - 1; i >= 0; i--) {
      const cue = cues[i]
      if (cue && currentMs >= cue.start_ms && currentMs <= cue.end_ms) return cue
    }
    return null
  }, [cues, currentMs])

  const rangeMin = startMs ?? 0
  const rangeMax = endMs ?? (measuredMs || durationMs)

  // On slow networks a seek into an unbuffered range would keep playing stale
  // audio while the new fragment loads; pause for the load and resume after.
  const resumeAfterSeekRef = useRef(false)
  const isBuffered = (v: HTMLVideoElement, t: number) => {
    for (let i = 0; i < v.buffered.length; i++) {
      if (v.buffered.start(i) <= t && t < v.buffered.end(i)) return true
    }
    return false
  }

  const updateBuffered = useCallback((v: HTMLVideoElement) => {
    const t = v.currentTime
    for (let i = 0; i < v.buffered.length; i++) {
      if (v.buffered.start(i) <= t + 0.5 && t <= v.buffered.end(i)) {
        setBufferedMs(Math.round(v.buffered.end(i) * 1000))
        return
      }
    }
    setBufferedMs(0)
  }, [])

  const seekTo = useCallback(
    (ms: number) => {
      const v = innerRef.current
      if (!v) return
      v.currentTime = clamp(ms, rangeMin, rangeMax || ms) / 1000
    },
    [rangeMin, rangeMax],
  )

  const togglePlay = useCallback(() => {
    const v = innerRef.current
    if (!v) return
    if (v.paused) v.play().catch(() => {})
    else {
      resumeAfterSeekRef.current = false
      v.pause()
    }
  }, [])

  const toggleFullscreen = useCallback(() => {
    if (document.fullscreenElement) void document.exitFullscreen()
    else void containerRef.current?.requestFullscreen()
  }, [])

  useEffect(() => {
    const onChange = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onChange)
    return () => document.removeEventListener('fullscreenchange', onChange)
  }, [])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return
      switch (e.key.toLowerCase()) {
        case ' ':
          e.preventDefault()
          togglePlay()
          break
        case 'f':
          e.preventDefault()
          toggleFullscreen()
          break
        case 'c':
          e.preventDefault()
          setShowOverlay((v) => !v)
          break
        case 'arrowleft':
          e.preventDefault()
          if (innerRef.current) seekTo(innerRef.current.currentTime * 1000 - 5000)
          break
        case 'arrowright':
          e.preventDefault()
          if (innerRef.current) seekTo(innerRef.current.currentTime * 1000 + 5000)
          break
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [togglePlay, toggleFullscreen, seekTo])

  // Timeline scrubbing: the position is previewed while dragging and the
  // actual seek happens once, on release.
  const barRef = useRef<HTMLDivElement>(null)
  const [scrubMs, setScrubMs] = useState<number | null>(null)
  const [hoverMs, setHoverMs] = useState<number | null>(null)

  const msFromPointer = useCallback(
    (e: ReactPointerEvent) => {
      const rect = barRef.current!.getBoundingClientRect()
      const frac = clamp((e.clientX - rect.left) / rect.width, 0, 1)
      return Math.round(rangeMin + frac * (rangeMax - rangeMin))
    },
    [rangeMin, rangeMax],
  )

  const shownMs = scrubMs ?? currentMs
  const span = Math.max(rangeMax - rangeMin, 1)
  const playedPct = clamp(((shownMs - rangeMin) / span) * 100, 0, 100)
  const bufferedPct = clamp(((bufferedMs - rangeMin) / span) * 100, 0, 100)
  const tooltipMs = scrubMs ?? hoverMs

  return (
    <div
      ref={containerRef}
      className={cn('group relative flex min-h-0 flex-1 items-center justify-center bg-black', className)}
    >
      <video
        ref={setVideoRef}
        src={videoUrl}
        playsInline
        className="max-h-full w-full"
        onClick={togglePlay}
        onTimeUpdate={(e) => {
          const v = e.currentTarget
          const ms = Math.round(v.currentTime * 1000)
          setCurrentMs(ms)
          onTimeChange?.(ms)
          if (!v.seeking && v.readyState >= 3) setWaiting(false)
          updateBuffered(v)
          if (endMs != null && ms >= endMs) {
            v.pause()
            onWindowEnd?.()
          }
        }}
        onProgress={(e) => updateBuffered(e.currentTarget)}
        onSeeking={(e) => {
          setSeeking(true)
          const el = e.currentTarget
          const ms = el.currentTime * 1000
          if (startMs != null && ms < startMs) el.currentTime = startMs / 1000
          else if (endMs != null && ms > endMs) el.currentTime = endMs / 1000
          if (!el.paused && !isBuffered(el, el.currentTime)) {
            el.pause()
            resumeAfterSeekRef.current = true
          }
        }}
        onSeeked={(e) => {
          setSeeking(false)
          updateBuffered(e.currentTarget)
          if (resumeAfterSeekRef.current) {
            resumeAfterSeekRef.current = false
            e.currentTarget.play().catch(() => {})
          }
        }}
        onWaiting={() => setWaiting(true)}
        onCanPlay={() => setWaiting(false)}
        onPlaying={() => setWaiting(false)}
        onLoadStart={() => {
          setSeeking(false)
          setWaiting(false)
          setBufferedMs(0)
          resumeAfterSeekRef.current = false
        }}
        onError={() => {
          setSeeking(false)
          setWaiting(false)
        }}
        onLoadedMetadata={(e) => {
          const v = e.currentTarget
          setMeasuredMs(Math.round(v.duration * 1000))
          if (onLoadedMetadata) onLoadedMetadata(v)
          else if (startMs != null) v.currentTime = startMs / 1000
        }}
        onPlay={() => setPaused(false)}
        onPause={() => setPaused(true)}
        onVolumeChange={(e) => {
          setVolume(e.currentTarget.volume)
          setMuted(e.currentTarget.muted)
        }}
      />

      {spinner && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center bg-black/30">
          <Spinner className="h-10 w-10 text-white" />
        </div>
      )}

      {paused && !stalled && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <div className="rounded-full bg-black/50 p-5 backdrop-blur-sm">
            <Play size={36} className="text-white" fill="currentColor" />
          </div>
        </div>
      )}

      {showOverlay && activeCue && (
        <div className="pointer-events-none absolute inset-x-0 bottom-20 flex justify-center px-6 transition-[bottom] duration-200 [@media(hover:hover)]:bottom-4 [@media(hover:hover)]:group-hover:bottom-20">
          {renderCueOverlay ? (
            renderCueOverlay(activeCue)
          ) : (
            <span className="pointer-events-auto block max-w-3xl select-text rounded-xl bg-black/70 px-4 py-2 text-center text-lg font-medium leading-snug text-white shadow-lg backdrop-blur-sm">
              <CueText text={activeCue.text} />
            </span>
          )}
        </div>
      )}

      <div className="absolute inset-x-0 bottom-0 flex flex-col gap-2 bg-gradient-to-t from-black/80 to-transparent px-4 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-10 transition-opacity [@media(hover:hover)]:opacity-0 [@media(hover:hover)]:group-hover:opacity-100">
        <div
          ref={barRef}
          className="group/bar relative flex h-4 cursor-pointer touch-none items-center"
          onPointerDown={(e) => {
            try {
              e.currentTarget.setPointerCapture(e.pointerId)
            } catch {
              /* synthetic events have no active pointer */
            }
            setScrubMs(msFromPointer(e))
          }}
          onPointerMove={(e) => {
            const ms = msFromPointer(e)
            if (scrubMs != null) setScrubMs(ms)
            else setHoverMs(ms)
          }}
          onPointerUp={() => {
            if (scrubMs != null) {
              seekTo(scrubMs)
              setCurrentMs(scrubMs)
              setScrubMs(null)
            }
          }}
          onPointerCancel={() => setScrubMs(null)}
          onPointerLeave={() => setHoverMs(null)}
        >
          {tooltipMs != null && (
            <span
              style={{ left: `${clamp(((tooltipMs - rangeMin) / span) * 100, 0, 100)}%` }}
              className="pointer-events-none absolute bottom-full mb-1 -translate-x-1/2 rounded bg-black/80 px-1.5 py-0.5 text-[11px] tabular-nums text-white"
            >
              {formatTime(tooltipMs - rangeMin)}
            </span>
          )}
          <div className="relative h-1 w-full overflow-hidden rounded-full bg-white/25 transition-[height] group-hover/bar:h-1.5">
            <div className="absolute inset-y-0 left-0 bg-white/30" style={{ width: `${bufferedPct}%` }} />
            <div className="absolute inset-y-0 left-0 bg-brand-500" style={{ width: `${playedPct}%` }} />
          </div>
          <div
            style={{ left: `${playedPct}%` }}
            className={cn(
              'pointer-events-none absolute h-3 w-3 -translate-x-1/2 rounded-full bg-brand-500 shadow transition-opacity',
              scrubMs != null ? 'opacity-100' : 'opacity-0 group-hover/bar:opacity-100',
            )}
          />
        </div>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2 text-white">
          {controlsStart}
          <button type="button" onClick={togglePlay} className="transition-colors hover:text-brand-400">
            {paused ? <Play size={20} /> : <Pause size={20} />}
          </button>
          <button
            type="button"
            onClick={() => innerRef.current && (innerRef.current.muted = !innerRef.current.muted)}
            className="transition-colors hover:text-brand-400"
          >
            {muted || volume === 0 ? <VolumeX size={18} /> : <Volume2 size={18} />}
          </button>
          <input
            type="range"
            min={0}
            max={1}
            step={0.05}
            value={muted ? 0 : volume}
            onChange={(e) => innerRef.current && (innerRef.current.volume = Number(e.target.value))}
            className="hidden h-1 w-20 cursor-pointer accent-brand-500 sm:block"
          />
          <span className="text-xs tabular-nums text-neutral-300">
            {formatTime(shownMs - rangeMin)} / {formatTime(rangeMax - rangeMin)}
          </span>
          <div className="ml-auto flex items-center gap-3">
            {audioTracks.length > 1 && (
              <Select
                dark
                value={String(audioIdx)}
                onChange={(v) => onAudioChange?.(Number(v))}
                options={audioTracks.map((t, i) => ({ value: String(i), label: t.label }))}
                className={darkSelectClass}
                title="Audio track"
              />
            )}
            {subtitleTracks.length > 0 && (
              <Select
                dark
                value={String(subIdx)}
                onChange={(v) => onSubChange?.(Number(v))}
                options={subtitleTracks.map((t, i) => ({ value: String(i), label: t.label }))}
                className={darkSelectClass}
                title="Subtitles"
              />
            )}
            {hasSubtitles && (
              <button type="button" onClick={() => setShowOverlay((v) => !v)} className="transition-colors hover:text-brand-400">
                {showOverlay ? <Captions size={18} /> : <CaptionsOff size={18} />}
              </button>
            )}
            {!isFullscreen && controlsEnd}
            <button type="button" onClick={toggleFullscreen} className="transition-colors hover:text-brand-400">
              {isFullscreen ? <Minimize size={18} /> : <Maximize size={18} />}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
