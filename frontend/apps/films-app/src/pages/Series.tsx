import { useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Check, Film, Layers, Loader2, Pencil, Play, Trash2 } from 'lucide-react'
import { Button, cn, ConfirmDialog, Field, Input, LoadingState, Textarea, useAgentView } from '@els/ui'
import { watchProgress } from '../lib/progress.ts'
import { useDeleteFilm, useDeleteSeries, useFilms, useIsAdmin, useUpdateFilm } from '../lib/films.ts'
import type { FilmSummary } from '../lib/types.ts'

export const seriesLastKey = (key: string) => `els.films.series.${key}.last`

export function Series() {
  const { key = '' } = useParams()
  const navigate = useNavigate()
  const title = decodeURIComponent(key)
  const { data: films, isLoading } = useFilms()
  const { data: isAdmin } = useIsAdmin()
  const deleteFilm = useDeleteFilm()
  const deleteSeries = useDeleteSeries()
  const [editing, setEditing] = useState<FilmSummary | null>(null)
  const [deletingEpisode, setDeletingEpisode] = useState<FilmSummary | null>(null)
  const [deletingSeries, setDeletingSeries] = useState(false)

  const episodes = useMemo(
    () =>
      (films ?? [])
        .filter((f) => f.kind === 'series' && (f.series_title || f.title) === title)
        .sort((a, b) => a.season - b.season || a.episode - b.episode),
    [films, title],
  )

  const seasons = useMemo(() => {
    const map = new Map<number, FilmSummary[]>()
    for (const e of episodes) {
      const arr = map.get(e.season) ?? []
      arr.push(e)
      map.set(e.season, arr)
    }
    return [...map.entries()].sort((a, b) => a[0] - b[0])
  }, [episodes])

  let lastId = ''
  try {
    lastId = localStorage.getItem(seriesLastKey(key)) ?? ''
  } catch {
    /* ignore */
  }

  useAgentView({
    app: 'films',
    screen: 'series',
    title,
    info: 'The user is viewing a series episode list. Full film list — list_films; episode subtitles — read_film_subtitles.',
    state: { episodes: episodes.length },
  })

  if (isLoading) {
    return <LoadingState className="h-full items-center bg-neutral-50 py-0" />
  }

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-3xl space-y-6 p-6">
        <header className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => navigate('..')}
            className="rounded-lg p-1.5 text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-900"
          >
            <ArrowLeft size={18} />
          </button>
          <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
            <Layers className="h-6 w-6 text-brand-600" />
            {title}
          </h1>
          {isAdmin && episodes.length > 0 && (
            <button
              type="button"
              onClick={() => setDeletingSeries(true)}
              className="ml-auto inline-flex items-center gap-1.5 rounded-lg border border-red-200 bg-white px-3 py-1.5 text-sm font-medium text-red-600 transition-colors hover:bg-red-50"
            >
              <Trash2 size={15} /> Delete series
            </button>
          )}
        </header>

        {editing && isAdmin && <EpisodeEditForm episode={editing} onDone={() => setEditing(null)} />}

        {episodes.length === 0 ? (
          <p className="py-16 text-center text-sm text-neutral-500">No episodes yet.</p>
        ) : (
          <div className="space-y-8">
            {seasons.map(([season, eps]) => (
              <section key={season}>
                <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">Season {season}</h2>
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                  {eps.map((e) => (
                    <EpisodeTile
                      key={e.id}
                      episode={e}
                      current={e.id === lastId}
                      isAdmin={!!isAdmin}
                      onEdit={() => setEditing(e)}
                      onDelete={() => setDeletingEpisode(e)}
                    />
                  ))}
                </div>
              </section>
            ))}
          </div>
        )}

        {deletingSeries && (
          <ConfirmDialog
            title="Delete series"
            description={`Delete series "${title}" and all ${episodes.length} episodes? This cannot be undone.`}
            pending={deleteSeries.isPending}
            onConfirm={() =>
              deleteSeries.mutate(
                episodes.map((e) => e.id),
                { onSuccess: () => navigate('..') },
              )
            }
            onClose={() => setDeletingSeries(false)}
          />
        )}
        {deletingEpisode && (
          <ConfirmDialog
            title="Delete episode"
            description={`Delete S${deletingEpisode.season}E${deletingEpisode.episode}? This cannot be undone.`}
            onConfirm={() => {
              deleteFilm.mutate(deletingEpisode.id)
              setDeletingEpisode(null)
            }}
            onClose={() => setDeletingEpisode(null)}
          />
        )}
      </div>
    </div>
  )
}

function EpisodeTile({
  episode: e,
  current,
  isAdmin,
  onEdit,
  onDelete,
}: {
  episode: FilmSummary
  current: boolean
  isAdmin: boolean
  onEdit: () => void
  onDelete: () => void
}) {
  const ready = e.status === 'ready'
  const { percent, done } = watchProgress(e.position_ms, e.duration_ms)
  const content = (
    <>
      <div className="relative flex aspect-video items-center justify-center overflow-hidden bg-brand-50">
        {e.poster_url ? (
          <img src={e.poster_url} alt={e.title} className="h-full w-full object-cover" />
        ) : (
          <Film className="h-8 w-8 text-brand-300" />
        )}
        <span className="absolute left-2 top-2 rounded-full bg-black/60 px-2 py-0.5 text-xs font-medium text-white">
          S{e.season}E{e.episode}
        </span>
        {current ? (
          <span className="absolute right-2 top-2 rounded-full bg-brand-600 px-2 py-0.5 text-xs font-medium text-white">
            Current
          </span>
        ) : (
          done && (
            <span className="absolute right-2 top-2 inline-flex items-center rounded-full bg-emerald-600 p-1 text-white">
              <Check size={12} />
            </span>
          )
        )}
        {ready && (
          <span className="absolute inset-0 flex items-center justify-center bg-black/0 opacity-0 transition-opacity group-hover:bg-black/30 group-hover:opacity-100">
            <Play size={28} className="text-white" />
          </span>
        )}
        {!done && percent > 0 && (
          <div className="absolute inset-x-0 bottom-0 h-1 bg-black/40">
            <div className="h-full bg-brand-500" style={{ width: `${percent}%` }} />
          </div>
        )}
      </div>
      <div className="p-3">
        <p className="truncate text-sm font-semibold text-neutral-900">{e.title || `Episode ${e.episode}`}</p>
        {e.status === 'processing' ? (
          <span className="mt-1 inline-flex items-center gap-1 text-xs text-amber-600">
            <Loader2 size={12} className="animate-spin" /> Processing…
          </span>
        ) : e.status === 'failed' ? (
          <span className="mt-1 inline-block text-xs text-red-600">Failed</span>
        ) : null}
      </div>
    </>
  )

  const cls = cn(
    'group relative overflow-hidden rounded-2xl bg-white ring-1 transition-colors',
    current ? 'ring-2 ring-brand-400' : 'ring-neutral-200 hover:ring-brand-300',
    !ready && 'opacity-60',
  )

  return (
    <div className={cls}>
      {ready ? (
        <Link to={`../${e.id}`} className="block">
          {content}
        </Link>
      ) : (
        content
      )}
      {isAdmin && (
        <div className="absolute right-2 top-2 z-10 flex gap-1.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
          <button
            type="button"
            onClick={onEdit}
            title="Edit episode"
            className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-brand-600 hover:text-white"
          >
            <Pencil size={14} />
          </button>
          <button
            type="button"
            onClick={onDelete}
            title="Delete episode"
            className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-red-600 hover:text-white"
          >
            <Trash2 size={14} />
          </button>
        </div>
      )}
    </div>
  )
}

function EpisodeEditForm({ episode, onDone }: { episode: FilmSummary; onDone: () => void }) {
  const update = useUpdateFilm()
  const [title, setTitle] = useState(episode.title)
  const [description, setDescription] = useState(episode.description ?? '')

  const submit = (ev: React.FormEvent) => {
    ev.preventDefault()
    update.mutate({ id: episode.id, title, description }, { onSuccess: onDone })
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-brand-200">
      <p className="text-sm font-semibold text-neutral-900">
        Edit S{episode.season}E{episode.episode}
      </p>
      <Field label="Title">
        <Input value={title} onChange={(e) => setTitle(e.target.value)} />
      </Field>
      <Field label="Description">
        <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} />
      </Field>
      {update.isError && <p className="text-sm text-red-600">Failed to save. Please try again.</p>}
      <div className="flex items-center gap-3">
        <Button type="submit" variant="brand" disabled={!title.trim() || update.isPending}>
          {update.isPending ? (
            <>
              <Loader2 size={16} className="animate-spin" /> Saving…
            </>
          ) : (
            'Save'
          )}
        </Button>
        <button type="button" onClick={onDone} className="text-sm text-neutral-500 hover:text-neutral-700">
          Cancel
        </button>
      </div>
    </form>
  )
}
