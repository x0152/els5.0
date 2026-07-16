import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import {
  ArrowLeft,
  BookPlus,
  Captions,
  CaptionsOff,
  Maximize,
  MessageCircleQuestion,
  Minimize,
  PanelRightClose,
  PanelRightOpen,
  Pause,
  PictureInPicture2,
  Play,
  SkipBack,
  SkipForward,
  Volume2,
  VolumeX,
} from 'lucide-react'
import { Button, cn, ErrorState, LoadingState, Select, Spinner, useAgentView, useMiniPlayer } from '@els/ui'
import { saveProgress, useFilm, useFilms } from '../lib/films.ts'
import { emitListening, emitUnclear, requestAnalyze, requestAsk } from '../lib/events.ts'
import { seriesLastKey } from './Series.tsx'
import { SubtitlePanel } from '../components/SubtitlePanel.tsx'
import { CueText } from '../components/CueText.tsx'

const prefsKey = (id: string) => `els.films.prefs.${id}`

function formatTime(ms: number): string {
  const total = Math.max(0, Math.floor(ms / 1000))
  const s = total % 60
  const m = Math.floor(total / 60) % 60
  const h = Math.floor(total / 3600)
  const mm = `${m}`.padStart(h > 0 ? 2 : 1, '0')
  const ss = `${s}`.padStart(2, '0')
  return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`
}

const darkSelectClass =
  'w-auto rounded-md border-white/20 bg-white/10 px-2 py-1 text-xs text-white focus:border-white/40 focus:ring-white/20 [&>option]:text-neutral-900'

export function Watch() {
  const { id = '' } = useParams()
  return <WatchInner key={id} id={id} />
}

function WatchInner({ id }: { id: string }) {
  const navigate = useNavigate()
  const location = useLocation()
  const [searchParams, setSearchParams] = useSearchParams()
  const mini = useMiniPlayer()
  const { data: film, isLoading, error } = useFilm(id)
  const { data: allFilms } = useFilms()

  const videoRef = useRef<HTMLVideoElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const resumeRef = useRef<{ ms: number; play: boolean } | null>(null)
  const initialSeek = Number(searchParams.get('t'))
  const pendingSeekRef = useRef<number | null>(
    searchParams.has('t') && Number.isFinite(initialSeek) && initialSeek >= 0 ? initialSeek : null,
  )
  const seekGuardRef = useRef<number | null>(null)
  const emittedRef = useRef<Set<number>>(new Set())
  const watchRef = useRef<{ index: number; enteredAt: number } | null>(null)
  const [flashCue, setFlashCue] = useState(false)

  const [currentMs, setCurrentMs] = useState(0)
  const [durationMs, setDurationMs] = useState(0)
  const [paused, setPaused] = useState(true)
  const [volume, setVolume] = useState(1)
  const [muted, setMuted] = useState(false)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [showOverlay, setShowOverlay] = useState(true)
  const [showPanel, setShowPanel] = useState(true)
  const [autoScroll, setAutoScroll] = useState(true)
  const [audioIdx, setAudioIdx] = useState(0)
  const [subIdx, setSubIdx] = useState(0)
  const [markedCue, setMarkedCue] = useState<number | null>(null)

  const episodes = useMemo(() => {
    if (!film || film.kind !== 'series') return []
    const series = film.series_title || film.title
    return (allFilms ?? [])
      .filter((f) => f.kind === 'series' && (f.series_title || f.title) === series)
      .sort((a, b) => a.season - b.season || a.episode - b.episode)
  }, [allFilms, film])
  const curIdx = episodes.findIndex((e) => e.id === id)
  const prevEp = curIdx > 0 ? episodes[curIdx - 1] : null
  const nextEp = curIdx >= 0 && curIdx < episodes.length - 1 ? episodes[curIdx + 1] : null
  const goEpisode = useCallback((epId: string) => navigate(`../${epId}`), [navigate])
  const seriesKey = film?.kind === 'series' ? encodeURIComponent(film.series_title || film.title) : ''
  const backPath = seriesKey ? `../series/${seriesKey}` : '..'

  useEffect(() => {
    if (!id || !seriesKey) return
    try {
      localStorage.setItem(seriesLastKey(seriesKey), id)
    } catch {
      /* ignore */
    }
  }, [id, seriesKey])

  const audioTracks = useMemo(() => film?.audio_tracks ?? [], [film])
  const subtitleTracks = useMemo(() => film?.subtitles ?? [], [film])
  const cues = useMemo(() => subtitleTracks[subIdx]?.cues ?? [], [subtitleTracks, subIdx])
  const hasSubtitles = cues.length > 0
  const videoUrl = audioTracks[audioIdx]?.url ?? ''

  const activeIdx = useMemo(() => {
    for (let i = cues.length - 1; i >= 0; i--) {
      const cue = cues[i]
      if (cue && currentMs >= cue.start_ms && currentMs <= cue.end_ms) return i
    }
    return -1
  }, [cues, currentMs])
  const activeCue = activeIdx >= 0 ? cues[activeIdx] : null

  const subLang = (subtitleTracks[subIdx]?.lang ?? '').toLowerCase()
  const emitEnabled = subLang === '' || subLang.startsWith('en') || subLang.startsWith('und')

  useAgentView(
    film
      ? {
          app: 'films',
          screen: 'watch',
          title: film.title,
          info: 'The user is watching a film with subtitles. To read cues at the current or any moment — read_film_subtitles with filmId and at_ms (positionMs).',
          ids: { filmId: id },
          state: {
            positionMs: currentMs,
            durationMs,
            subtitleLang: subLang,
            currentSubtitle: activeCue?.text ?? '',
          },
        }
      : null,
  )

  const markUnclear = () => {
    if (!activeCue || !emitEnabled) return
    emittedRef.current.add(activeCue.index)
    emitUnclear(activeCue.text, { app: 'films', film_id: id, lang: subLang })
    setMarkedCue(activeCue.index)
    setTimeout(() => setMarkedCue((c) => (c === activeCue.index ? null : c)), 1200)
  }

  useEffect(() => {
    emittedRef.current = new Set()
    watchRef.current = null
  }, [film?.id, subIdx])

  useEffect(() => {
    const prev = watchRef.current
    if (prev && prev.index !== activeIdx) {
      const cue = cues[prev.index]
      if (cue && emitEnabled && !emittedRef.current.has(cue.index)) {
        const dwell = performance.now() - prev.enteredAt
        if (dwell >= (cue.end_ms - cue.start_ms) * 0.5) {
          emittedRef.current.add(cue.index)
          emitListening(cue.text, { app: 'films', film_id: id, lang: subLang })
        }
      }
    }
    if (activeIdx < 0) watchRef.current = null
    else if (!prev || prev.index !== activeIdx) watchRef.current = { index: activeIdx, enteredAt: performance.now() }
  }, [activeIdx, cues, emitEnabled, id, subLang])

  const panelVisible = showPanel && !isFullscreen && hasSubtitles

  const seekTo = useCallback((ms: number) => {
    if (videoRef.current) videoRef.current.currentTime = ms / 1000
  }, [])

  const togglePlay = useCallback(() => {
    const v = videoRef.current
    if (!v) return
    if (v.paused) void v.play()
    else v.pause()
  }, [])

  const toggleFullscreen = useCallback(() => {
    if (document.fullscreenElement) void document.exitFullscreen()
    else void containerRef.current?.requestFullscreen()
  }, [])

  const savePrefs = useCallback(
    (a: number, s: number) => {
      if (!id) return
      try {
        localStorage.setItem(
          prefsKey(id),
          JSON.stringify({ audio: audioTracks[a]?.lang ?? null, sub: subtitleTracks[s]?.lang ?? null }),
        )
      } catch {
        /* ignore */
      }
    },
    [id, audioTracks, subtitleTracks],
  )

  const changeAudio = (idx: number) => {
    const v = videoRef.current
    resumeRef.current = { ms: currentMs, play: v ? !v.paused : false }
    setAudioIdx(idx)
    savePrefs(idx, subIdx)
  }

  const changeSub = (idx: number) => {
    setSubIdx(idx)
    savePrefs(audioIdx, idx)
  }

  useEffect(() => {
    const onChange = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onChange)
    return () => document.removeEventListener('fullscreenchange', onChange)
  }, [])

  useEffect(() => {
    if (!film) return
    try {
      const raw = localStorage.getItem(prefsKey(id))
      if (!raw) return
      const prefs = JSON.parse(raw) as { audio?: string; sub?: string }
      const a = audioTracks.findIndex((t) => t.lang === prefs.audio)
      const s = subtitleTracks.findIndex((t) => t.lang === prefs.sub)
      if (a >= 0) setAudioIdx(a)
      if (s >= 0) setSubIdx(s)
    } catch {
      /* ignore */
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [film?.id, audioTracks, subtitleTracks])

  const playedMsRef = useRef<number | null>(null)

  useEffect(() => {
    if (!id) return
    playedMsRef.current = null
    const interval = setInterval(() => {
      const v = videoRef.current
      if (!v || v.paused) return
      const ms = Math.round(v.currentTime * 1000)
      if (seekGuardRef.current != null) {
        if (ms < seekGuardRef.current) return
        seekGuardRef.current = null
      }
      playedMsRef.current = ms
      void saveProgress(id, ms)
    }, 5000)
    return () => {
      clearInterval(interval)
      if (playedMsRef.current != null) void saveProgress(id, playedMsRef.current)
    }
  }, [id])

  useEffect(() => {
    mini.close()
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
          if (videoRef.current) videoRef.current.currentTime = Math.max(0, videoRef.current.currentTime - 5)
          break
        case 'arrowright':
          e.preventDefault()
          if (videoRef.current) videoRef.current.currentTime += 5
          break
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [togglePlay, toggleFullscreen])

  const onLoadedMetadata = () => {
    const v = videoRef.current
    if (!v) return
    setDurationMs(Math.round(v.duration * 1000))
    if (resumeRef.current) {
      v.currentTime = resumeRef.current.ms / 1000
      if (resumeRef.current.play) void v.play()
      resumeRef.current = null
      return
    }
    if (pendingSeekRef.current != null) {
      const ms = pendingSeekRef.current
      pendingSeekRef.current = null
      v.currentTime = ms / 1000
      seekGuardRef.current = ms + 30000
      setFlashCue(true)
      setTimeout(() => setFlashCue(false), 2500)
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev)
          next.delete('t')
          return next
        },
        { replace: true },
      )
      return
    }
    const saved = film?.position_ms ?? 0
    if (saved > 0 && saved < v.duration * 1000 - 5000) v.currentTime = saved / 1000
  }

  const backToLibrary = (
    <Button variant="secondary" onClick={() => navigate('..')}>
      Back to library
    </Button>
  )

  if (isLoading) {
    return <LoadingState className="h-full items-center bg-neutral-50 py-0" />
  }
  if (error || !film) {
    return (
      <div className="flex h-full items-center justify-center bg-neutral-50 p-8">
        <ErrorState title="Failed to load the film" action={backToLibrary} className="w-full max-w-md" />
      </div>
    )
  }

  if (film.status !== 'ready') {
    const failed = film.status === 'failed'
    return (
      <div className="flex h-full items-center justify-center bg-neutral-50 p-8">
        {failed ? (
          <ErrorState title="Failed to process the film" description={film.error} action={backToLibrary} className="w-full max-w-md" />
        ) : (
          <div className="flex flex-col items-center gap-4">
            <Spinner className="h-8 w-8 text-brand-500" />
            <p className="text-sm text-neutral-500">Processing tracks and subtitles…</p>
            {backToLibrary}
          </div>
        )}
      </div>
    )
  }

  const popOut = () => {
    const v = videoRef.current
    const wasPlaying = v ? !v.paused : true
    if (v) {
      playedMsRef.current = null
      void saveProgress(id, Math.round(v.currentTime * 1000))
      v.pause()
    }
    mini.open({
      id,
      src: videoUrl,
      title: film.kind === 'series' ? `${film.series_title ?? ''} · S${film.season}E${film.episode}` : film.title,
      startMs: currentMs,
      playing: wasPlaying,
      returnTo: location.pathname,
      onProgress: (ms) => void saveProgress(id, ms),
    })
    navigate(backPath)
  }

  return (
    <div ref={containerRef} className={cn('flex min-h-0 flex-col overflow-hidden bg-neutral-50', isFullscreen ? 'h-screen' : 'h-full')}>
      {!isFullscreen && (
        <header className="flex items-center gap-3 border-b border-neutral-200 bg-white px-4 py-3">
          <button
            type="button"
            onClick={() => navigate(backPath)}
            className="rounded-lg p-1.5 text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-900"
          >
            <ArrowLeft size={18} />
          </button>
          <div className="min-w-0">
            <h1 className="truncate text-sm font-semibold text-neutral-900">
              {film.kind === 'series' ? film.series_title : film.title}
            </h1>
            {film.kind === 'series' && (
              <p className="truncate text-xs text-neutral-500">
                S{film.season}E{film.episode}
                {film.title ? ` · ${film.title}` : ''}
              </p>
            )}
          </div>
          {episodes.length > 1 && (
            <div className="ml-auto flex items-center gap-1.5">
              <button
                type="button"
                onClick={() => prevEp && goEpisode(prevEp.id)}
                disabled={!prevEp}
                className="rounded-lg p-1.5 text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-900 disabled:opacity-30"
                title="Previous episode"
              >
                <SkipBack size={18} />
              </button>
              <Select
                value={id}
                onChange={(e) => goEpisode(e.target.value)}
                className="max-w-[12rem] truncate rounded-md px-2 py-1 text-xs text-neutral-700"
              >
                {episodes.map((e) => (
                  <option key={e.id} value={e.id}>
                    S{e.season}E{e.episode}
                    {e.title ? ` · ${e.title}` : ''}
                  </option>
                ))}
              </Select>
              <button
                type="button"
                onClick={() => nextEp && goEpisode(nextEp.id)}
                disabled={!nextEp}
                className="rounded-lg p-1.5 text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-900 disabled:opacity-30"
                title="Next episode"
              >
                <SkipForward size={18} />
              </button>
            </div>
          )}
        </header>
      )}

      <div className={cn('flex min-h-0 flex-1', panelVisible ? 'flex-col lg:flex-row' : 'flex-col')}>
        <div className="group relative flex min-h-0 flex-1 items-center justify-center bg-black">
          <video
            ref={videoRef}
            src={videoUrl}
            playsInline
            className="max-h-full w-full"
            onClick={togglePlay}
            onTimeUpdate={(e) => setCurrentMs(Math.round(e.currentTarget.currentTime * 1000))}
            onLoadedMetadata={onLoadedMetadata}
            onPlay={() => setPaused(false)}
            onPause={() => setPaused(true)}
            onVolumeChange={(e) => {
              setVolume(e.currentTarget.volume)
              setMuted(e.currentTarget.muted)
            }}
          />

          {showOverlay && activeCue && (
            <div className="pointer-events-none absolute inset-x-0 bottom-20 flex justify-center px-6">
              <div className="group/cue relative max-w-3xl">
                <span
                  role={emitEnabled ? 'button' : undefined}
                  title={emitEnabled ? "Mark as not understood" : undefined}
                  onClick={(e) => {
                    e.stopPropagation()
                    if (window.getSelection()?.isCollapsed !== false) markUnclear()
                  }}
                  className={cn(
                    'pointer-events-auto block select-text rounded-xl px-4 py-2 text-center text-lg font-medium leading-snug text-white shadow-lg backdrop-blur-sm transition-colors',
                    emitEnabled && 'cursor-pointer',
                    markedCue === activeCue.index
                      ? 'bg-rose-600/80 ring-2 ring-rose-300'
                      : flashCue
                        ? 'bg-black/70 ring-2 ring-amber-300'
                        : 'bg-black/70',
                  )}
                >
                  <CueText text={activeCue.text} />
                </span>
                <div className="pointer-events-auto absolute -right-3 -top-3 flex gap-1.5 opacity-0 transition-opacity focus-within:opacity-100 group-hover/cue:opacity-100 [@media(hover:none)]:opacity-100">
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation()
                      requestAnalyze(activeCue.text)
                    }}
                    title="Analyze this line"
                    className="rounded-full bg-black/70 p-1.5 text-white shadow-lg transition-colors hover:text-brand-400"
                  >
                    <BookPlus size={14} />
                  </button>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation()
                      requestAsk(activeCue.text)
                    }}
                    title="Ask the assistant"
                    className="rounded-full bg-black/70 p-1.5 text-white shadow-lg transition-colors hover:text-brand-400"
                  >
                    <MessageCircleQuestion size={14} />
                  </button>
                </div>
              </div>
            </div>
          )}

          <div className="absolute inset-x-0 bottom-0 flex flex-col gap-2 bg-gradient-to-t from-black/80 to-transparent px-4 pb-[max(0.75rem,env(safe-area-inset-bottom))] pt-10 transition-opacity [@media(hover:hover)]:opacity-0 [@media(hover:hover)]:group-hover:opacity-100">
            <input
              type="range"
              min={0}
              max={durationMs || film.duration_ms || 0}
              value={currentMs}
              onChange={(e) => seekTo(Number(e.target.value))}
              className="h-1 w-full cursor-pointer accent-brand-500"
            />
            <div className="flex flex-wrap items-center gap-x-3 gap-y-2 text-white">
              {episodes.length > 1 && (
                <button
                  type="button"
                  onClick={() => prevEp && goEpisode(prevEp.id)}
                  disabled={!prevEp}
                  className="transition-colors hover:text-brand-400 disabled:opacity-30"
                  title="Previous episode"
                >
                  <SkipBack size={18} />
                </button>
              )}
              <button type="button" onClick={togglePlay} className="transition-colors hover:text-brand-400">
                {paused ? <Play size={20} /> : <Pause size={20} />}
              </button>
              {episodes.length > 1 && (
                <button
                  type="button"
                  onClick={() => nextEp && goEpisode(nextEp.id)}
                  disabled={!nextEp}
                  className="transition-colors hover:text-brand-400 disabled:opacity-30"
                  title="Next episode"
                >
                  <SkipForward size={18} />
                </button>
              )}
              <button
                type="button"
                onClick={() => videoRef.current && (videoRef.current.muted = !videoRef.current.muted)}
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
                onChange={(e) => videoRef.current && (videoRef.current.volume = Number(e.target.value))}
                className="hidden h-1 w-20 cursor-pointer accent-brand-500 sm:block"
              />
              <span className="text-xs tabular-nums text-neutral-300">
                {formatTime(currentMs)} / {formatTime(durationMs || film.duration_ms)}
              </span>
              <div className="ml-auto flex items-center gap-3">
                {audioTracks.length > 1 && (
                  <Select
                    value={audioIdx}
                    onChange={(e) => changeAudio(Number(e.target.value))}
                    className={darkSelectClass}
                    title="Audio track"
                  >
                    {audioTracks.map((t, i) => (
                      <option key={i} value={i}>
                        {t.label}
                      </option>
                    ))}
                  </Select>
                )}
                {subtitleTracks.length > 0 && (
                  <Select
                    value={subIdx}
                    onChange={(e) => changeSub(Number(e.target.value))}
                    className={darkSelectClass}
                    title="Subtitles"
                  >
                    {subtitleTracks.map((t, i) => (
                      <option key={i} value={i}>
                        {t.label}
                      </option>
                    ))}
                  </Select>
                )}
                {hasSubtitles && (
                  <button type="button" onClick={() => setShowOverlay((v) => !v)} className="transition-colors hover:text-brand-400">
                    {showOverlay ? <Captions size={18} /> : <CaptionsOff size={18} />}
                  </button>
                )}
                {hasSubtitles && !isFullscreen && (
                  <button type="button" onClick={() => setShowPanel((v) => !v)} className="transition-colors hover:text-brand-400">
                    {showPanel ? <PanelRightClose size={18} /> : <PanelRightOpen size={18} />}
                  </button>
                )}
                {!isFullscreen && (
                  <button
                    type="button"
                    onClick={popOut}
                    title="Pop out (floating window)"
                    className="transition-colors hover:text-brand-400"
                  >
                    <PictureInPicture2 size={18} />
                  </button>
                )}
                <button type="button" onClick={toggleFullscreen} className="transition-colors hover:text-brand-400">
                  {isFullscreen ? <Minimize size={18} /> : <Maximize size={18} />}
                </button>
              </div>
            </div>
          </div>
        </div>

        {panelVisible && (
          <SubtitlePanel
            cues={cues}
            activeIdx={activeIdx}
            currentMs={currentMs}
            autoScroll={autoScroll}
            onToggleAutoScroll={() => setAutoScroll((v) => !v)}
            onSeek={seekTo}
          />
        )}
      </div>
    </div>
  )
}
