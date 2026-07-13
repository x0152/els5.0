export interface Cue {
  index: number
  start_ms: number
  end_ms: number
  text: string
}

export interface SubtitleTrack {
  lang: string
  label: string
  cues: Cue[]
}

export interface AudioTrack {
  lang: string
  label: string
  url: string
}

export type FilmStatus = 'processing' | 'ready' | 'failed'

export type FilmKind = 'film' | 'series'

export interface FilmSummary {
  id: string
  title: string
  description?: string
  poster_url?: string
  duration_ms: number
  position_ms: number
  status: FilmStatus
  kind: FilmKind
  series_title?: string
  season: number
  episode: number
  created_at: string
}

export interface Film {
  id: string
  title: string
  description?: string
  poster_url?: string
  duration_ms: number
  position_ms: number
  status: FilmStatus
  error?: string
  kind: FilmKind
  series_title?: string
  season: number
  episode: number
  audio_tracks: AudioTrack[]
  subtitles: SubtitleTrack[]
  created_at: string
}
