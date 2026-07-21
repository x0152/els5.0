import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './api.ts'
import type { Film, FilmSummary } from './types.ts'

export function useIsAdmin() {
  return useQuery({
    queryKey: ['account', 'me', 'isAdmin'],
    staleTime: 60_000,
    queryFn: async (): Promise<boolean> => {
      const me = await api.account.accountMe()
      return !!me?.is_global_admin
    },
  })
}

export function useFilms() {
  return useQuery({
    queryKey: ['films', 'list'],
    queryFn: async (): Promise<FilmSummary[]> => {
      const res = await api.films.listFilms()
      return (res?.items ?? []) as FilmSummary[]
    },
    refetchInterval: (q) => {
      const list = q.state.data as FilmSummary[] | undefined
      return list?.some((f) => f.status === 'processing') ? 3000 : false
    },
  })
}

export function useFilm(id: string) {
  return useQuery({
    queryKey: ['films', 'film', id],
    enabled: !!id,
    queryFn: async (): Promise<Film> => {
      const res = await api.films.getFilm({ params: { path: { id } } })
      return { ...res, audio_tracks: res?.audio_tracks ?? [], subtitles: res?.subtitles ?? [] } as Film
    },
    refetchInterval: (q) => ((q.state.data as Film | undefined)?.status === 'processing' ? 3000 : false),
  })
}

export function saveProgress(id: string, positionMs: number) {
  return api.films
    .saveFilmProgress({ params: { path: { id } }, body: { position_ms: Math.max(0, Math.round(positionMs)) } })
    .catch((e) => {
      console.error('failed to save film progress', e)
      return undefined
    })
}

export function useUploadFilm() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (args: {
      title: string
      video: File
      subtitles?: File
      poster?: File
      kind: 'film' | 'series'
      level: string
      seriesTitle?: string
      season?: number
      episode?: number
    }) => {
      const form = new FormData()
      form.append('title', args.title)
      form.append('video', args.video)
      form.append('kind', args.kind)
      form.append('level', args.level)
      if (args.kind === 'series') {
        form.append('series_title', args.seriesTitle ?? '')
        form.append('season', String(args.season ?? 1))
        form.append('episode', String(args.episode ?? 1))
      }
      if (args.subtitles) form.append('subtitles', args.subtitles)
      if (args.poster) form.append('poster', args.poster)
      return api.films.uploadFilm({ body: form as unknown as never })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['films', 'list'] }),
  })
}

export function useUpdateFilm() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (args: { id: string; title: string; description: string; level: string; poster?: File }) => {
      const form = new FormData()
      form.append('title', args.title)
      form.append('description', args.description)
      form.append('level', args.level)
      if (args.poster) form.append('poster', args.poster)
      return api.films.updateFilm({ params: { path: { id: args.id } }, body: form as unknown as never })
    },
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['films', 'list'] })
      qc.invalidateQueries({ queryKey: ['films', 'film', vars.id] })
    },
  })
}

export function useDeleteFilm() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.films.deleteFilm({ params: { path: { id } } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['films', 'list'] }),
  })
}

export function useDeleteSeries() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (ids: string[]) => {
      for (const id of ids) await api.films.deleteFilm({ params: { path: { id } } })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['films', 'list'] }),
  })
}

export interface SeriesMeta {
  title: string
  description?: string
  poster_url?: string
}

export function useSeries() {
  return useQuery({
    queryKey: ['films', 'series'],
    queryFn: async (): Promise<SeriesMeta[]> => {
      const res = await api.films.listSeries()
      return res?.items ?? []
    },
  })
}

export function useUpdateSeries() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (args: { title: string; newTitle?: string; description: string; poster?: File }) => {
      const form = new FormData()
      form.append('title', args.title)
      if (args.newTitle && args.newTitle !== args.title) form.append('new_title', args.newTitle)
      form.append('description', args.description)
      if (args.poster) form.append('poster', args.poster)
      return api.films.updateSeries({ body: form as unknown as never })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['films', 'series'] })
      qc.invalidateQueries({ queryKey: ['films', 'list'] })
    },
  })
}
