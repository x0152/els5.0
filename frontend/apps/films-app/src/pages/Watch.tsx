import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import {
  ArrowLeft,
  BookPlus,
  MessageCircleQuestion,
  PanelRightClose,
  PanelRightOpen,
  PictureInPicture2,
  SkipBack,
  SkipForward,
} from 'lucide-react'
import {
  Button,
  cn,
  CueText,
  englishTrackIdx,
  ErrorState,
  FilmPlayer,
  LoadingState,
  Select,
  Spinner,
  useAgentView,
  useMiniPlayer,
} from '@els/ui'
import { saveProgress, useFilm, useFilms } from '../lib/films.ts'
import { emitListening, emitUnclear, requestAnalyze, requestAsk } from '../lib/events.ts'
import { seriesLastKey } from './Series.tsx'
import { SubtitlePanel } from '../components/SubtitlePanel.tsx'

const prefsKey = (id: string) => `els.films.prefs.${id}`

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
  const [isFullscreen, setIsFullscreen] = useState(false)
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

  const markUnclear = (cue: { index: number; text: string }) => {
    if (!emitEnabled) return
    emittedRef.current.add(cue.index)
    emitUnclear(cue.text, { app: 'films', film_id: id, lang: subLang })
    setMarkedCue(cue.index)
    setTimeout(() => setMarkedCue((c) => (c === cue.index ? null : c)), 1200)
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
    let prefs: { audio?: string; sub?: string } = {}
    try {
      prefs = JSON.parse(localStorage.getItem(prefsKey(id)) ?? '{}') as { audio?: string; sub?: string }
    } catch {
      /* ignore */
    }
    const a = audioTracks.findIndex((t) => t.lang === prefs.audio)
    const s = subtitleTracks.findIndex((t) => t.lang === prefs.sub)
    setAudioIdx(a >= 0 ? a : englishTrackIdx(audioTracks))
    setSubIdx(s >= 0 ? s : englishTrackIdx(subtitleTracks))
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

  const onLoadedMetadata = (v: HTMLVideoElement) => {
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
    <div className={cn('flex min-h-0 flex-col overflow-hidden bg-neutral-50', isFullscreen ? 'h-screen' : 'h-full')}>
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
                onChange={goEpisode}
                options={episodes.map((e) => ({
                  value: e.id,
                  label: `S${e.season}E${e.episode}${e.title ? ` · ${e.title}` : ''}`,
                }))}
                className="max-w-[12rem] rounded-md px-2 py-1 text-xs text-neutral-700"
              />
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
        <FilmPlayer
          videoUrl={videoUrl}
          audioTracks={audioTracks}
          subtitleTracks={subtitleTracks}
          audioIdx={audioIdx}
          subIdx={subIdx}
          onAudioChange={changeAudio}
          onSubChange={changeSub}
          onTimeChange={setCurrentMs}
          onLoadedMetadata={onLoadedMetadata}
          durationMs={film.duration_ms}
          videoRef={videoRef}
          renderCueOverlay={(cue) => (
            <div className="group/cue relative max-w-3xl">
              <span
                role={emitEnabled ? 'button' : undefined}
                title={emitEnabled ? 'Mark as not understood' : undefined}
                onClick={(e) => {
                  e.stopPropagation()
                  if (window.getSelection()?.isCollapsed !== false) markUnclear(cue)
                }}
                className={cn(
                  'pointer-events-auto block select-text rounded-xl px-4 py-2 text-center text-lg font-medium leading-snug text-white shadow-lg backdrop-blur-sm transition-colors',
                  emitEnabled && 'cursor-pointer',
                  markedCue === cue.index
                    ? 'bg-rose-600/80 ring-2 ring-rose-300'
                    : flashCue
                      ? 'bg-black/70 ring-2 ring-amber-300'
                      : 'bg-black/70',
                )}
              >
                <CueText text={cue.text} />
              </span>
              <div className="pointer-events-auto absolute -right-3 -top-3 flex gap-1.5 opacity-0 transition-opacity focus-within:opacity-100 group-hover/cue:opacity-100 [@media(hover:none)]:opacity-100">
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation()
                    requestAnalyze(cue.text)
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
                    requestAsk(cue.text)
                  }}
                  title="Ask the assistant"
                  className="rounded-full bg-black/70 p-1.5 text-white shadow-lg transition-colors hover:text-brand-400"
                >
                  <MessageCircleQuestion size={14} />
                </button>
              </div>
            </div>
          )}
          controlsStart={
            episodes.length > 1 && (
              <>
                <button
                  type="button"
                  onClick={() => prevEp && goEpisode(prevEp.id)}
                  disabled={!prevEp}
                  className="hidden transition-colors hover:text-brand-400 disabled:opacity-30 sm:block"
                  title="Previous episode"
                >
                  <SkipBack size={18} />
                </button>
                <button
                  type="button"
                  onClick={() => nextEp && goEpisode(nextEp.id)}
                  disabled={!nextEp}
                  className="hidden transition-colors hover:text-brand-400 disabled:opacity-30 sm:block"
                  title="Next episode"
                >
                  <SkipForward size={18} />
                </button>
              </>
            )
          }
          controlsEnd={
            <>
              {hasSubtitles && (
                <button type="button" onClick={() => setShowPanel((v) => !v)} className="transition-colors hover:text-brand-400">
                  {showPanel ? <PanelRightClose size={18} /> : <PanelRightOpen size={18} />}
                </button>
              )}
              <button
                type="button"
                onClick={popOut}
                title="Pop out (floating window)"
                className="hidden transition-colors hover:text-brand-400 sm:block"
              >
                <PictureInPicture2 size={18} />
              </button>
            </>
          }
        />

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
