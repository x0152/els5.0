import { useQuery } from '@tanstack/react-query'
import { speak } from '@els/ui'
import { api } from './api.ts'

export function useFilmUrl(filmId?: string) {
  const query = useQuery({
    queryKey: ['workout-film', filmId],
    queryFn: () => api.films.getFilm({ params: { path: { id: filmId! } } }),
    enabled: !!filmId,
    staleTime: Infinity,
  })
  return { film: query.data, videoUrl: query.data?.audio_tracks?.[0]?.url ?? '' }
}

let clipEl: HTMLVideoElement | null = null
let clipStop: (() => void) | null = null

// Plays the original film audio for a cue range through one shared hidden element;
// items without a cue fall back to browser TTS.
export function playClip(videoUrl: string, startMs: number, endMs: number, opts?: { rate?: number; onEnd?: () => void }) {
  stopClip()
  if (!clipEl) {
    clipEl = document.createElement('video')
    clipEl.preload = 'metadata'
    clipEl.style.display = 'none'
    document.body.appendChild(clipEl)
  }
  const el = clipEl
  if (el.src !== videoUrl) el.src = videoUrl
  el.playbackRate = opts?.rate ?? 1
  el.currentTime = startMs / 1000
  const onTime = () => {
    if (el.currentTime * 1000 >= endMs) stop()
  }
  const stop = () => {
    el.pause()
    el.removeEventListener('timeupdate', onTime)
    clipStop = null
    opts?.onEnd?.()
  }
  clipStop = stop
  el.addEventListener('timeupdate', onTime)
  void el.play().catch(() => stop())
}

export function stopClip() {
  clipStop?.()
}

export function playPhrase(videoUrl: string, phrase: { text: string; start_ms?: number; end_ms?: number }, rate = 1): Promise<HTMLAudioElement | null> {
  if (videoUrl && phrase.start_ms !== undefined && phrase.end_ms) {
    playClip(videoUrl, phrase.start_ms, phrase.end_ms, { rate })
    return Promise.resolve(null)
  }
  return speak(phrase.text, rate < 1 ? { rate } : undefined)
}
