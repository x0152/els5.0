import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Captions, Check, Film, FileVideo, Layers, Loader2, Pencil, Plus, Trash2 } from 'lucide-react'
import {
  AppInfoButton,
  Badge,
  Button,
  ConfirmDialog,
  EmptyState,
  Field,
  FileField,
  ImageField,
  Input,
  LoadingState,
  Tabs,
  Textarea,
  useAgentView,
} from '@els/ui'
import { watchProgress } from '../lib/progress.ts'
import {
  useDeleteFilm,
  useFilms,
  useIsAdmin,
  useSeries,
  useUpdateFilm,
  useUpdateSeries,
  useUploadFilm,
  type SeriesMeta,
} from '../lib/films.ts'
import type { FilmSummary } from '../lib/types.ts'

function formatDuration(ms: number): string {
  const total = Math.floor(ms / 1000)
  const m = Math.floor(total / 60)
  const s = total % 60
  return m > 0 ? `${m} min` : `${s} sec`
}

export function Library() {
  const { data: films, isLoading } = useFilms()
  const { data: isAdmin } = useIsAdmin()
  const { data: seriesMeta } = useSeries()
  const deleteFilm = useDeleteFilm()
  const [showUpload, setShowUpload] = useState(false)
  const [editing, setEditing] = useState<FilmSummary | null>(null)
  const [editingSeries, setEditingSeries] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<FilmSummary | null>(null)

  const list = films ?? []
  const movies = list.filter((f) => f.kind !== 'series')
  const series = groupSeries(list)
  const metaByTitle = useMemo(() => new Map((seriesMeta ?? []).map((m) => [m.title, m])), [seriesMeta])
  const isEmpty = movies.length === 0 && series.length === 0

  useAgentView({
    app: 'films',
    screen: 'library',
    info: 'The user is in the films catalog. Full list — list_films.',
  })

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-4xl space-y-6 p-6">
        <header className="flex items-end justify-between gap-4">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
              <Film className="h-6 w-6 text-brand-600" />
              Films <AppInfoButton />
            </h1>
            <p className="mt-1 text-sm text-neutral-500">Watch films with subtitles</p>
          </div>
          {isAdmin && (
            <Button variant="brand" onClick={() => setShowUpload((v) => !v)}>
              <Plus size={16} /> Upload
            </Button>
          )}
        </header>

        {showUpload && isAdmin && <UploadForm onDone={() => setShowUpload(false)} />}

        {editing && isAdmin && <EditForm film={editing} onDone={() => setEditing(null)} />}

        {editingSeries !== null && isAdmin && (
          <SeriesEditForm
            title={editingSeries}
            meta={metaByTitle.get(editingSeries)}
            onDone={() => setEditingSeries(null)}
          />
        )}

        {isLoading ? (
          <LoadingState />
        ) : isEmpty ? (
          <EmptyState icon={<Film className="h-8 w-8" />} title="No films yet" description="Uploaded films will appear here." />
        ) : (
          <div className="space-y-8">
            {series.length > 0 && (
              <section>
                <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">Series</h2>
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                  {series.map((g) => (
                    <SeriesCard
                      key={g.title}
                      group={g}
                      meta={metaByTitle.get(g.title)}
                      isAdmin={!!isAdmin}
                      onEdit={() => setEditingSeries(g.title)}
                    />
                  ))}
                </div>
              </section>
            )}
            {movies.length > 0 && (
              <section>
                {series.length > 0 && (
                  <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">Films</h2>
                )}
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                  {movies.map((f) => (
                    <FilmCard
                      key={f.id}
                      film={f}
                      isAdmin={!!isAdmin}
                      onEdit={() => setEditing(f)}
                      onDelete={() => setDeleting(f)}
                    />
                  ))}
                </div>
              </section>
            )}
          </div>
        )}

        {deleting && (
          <ConfirmDialog
            title="Delete film"
            description={`Delete "${deleting.title}"? This cannot be undone.`}
            onConfirm={() => {
              deleteFilm.mutate(deleting.id)
              setDeleting(null)
            }}
            onClose={() => setDeleting(null)}
          />
        )}
      </div>
    </div>
  )
}

interface SeriesGroup {
  title: string
  episodes: FilmSummary[]
  firstId: string
}

function groupSeries(list: FilmSummary[]): SeriesGroup[] {
  const map = new Map<string, FilmSummary[]>()
  for (const f of list) {
    if (f.kind !== 'series') continue
    const key = f.series_title || f.title
    const arr = map.get(key) ?? []
    arr.push(f)
    map.set(key, arr)
  }
  const groups: SeriesGroup[] = []
  for (const [title, episodes] of map) {
    episodes.sort((a, b) => a.season - b.season || a.episode - b.episode)
    groups.push({
      title,
      episodes,
      firstId: episodes[0]?.id ?? '',
    })
  }
  return groups.sort((a, b) => a.title.localeCompare(b.title))
}

function FilmCard({
  film: f,
  isAdmin,
  onEdit,
  onDelete,
}: {
  film: FilmSummary
  isAdmin: boolean
  onEdit: () => void
  onDelete: () => void
}) {
  const { percent, done } = watchProgress(f.position_ms, f.duration_ms)
  return (
    <div className="group relative overflow-hidden rounded-2xl bg-white ring-1 ring-neutral-200 transition-colors hover:ring-brand-300">
      <Link to={f.id} className="block">
        <div className="relative flex aspect-[2/3] items-center justify-center overflow-hidden bg-brand-50">
          {f.poster_url ? (
            <>
              <img src={f.poster_url} alt="" aria-hidden className="absolute inset-0 h-full w-full scale-110 object-cover blur-xl" />
              <img src={f.poster_url} alt={f.title} className="relative h-full w-full object-contain" />
            </>
          ) : (
            <Film className="h-8 w-8 text-brand-300" />
          )}
          {done ? (
            <span className="absolute left-2 top-2 inline-flex items-center gap-1 rounded-full bg-emerald-600 px-2 py-0.5 text-xs font-medium text-white">
              <Check size={11} /> Watched
            </span>
          ) : (
            percent > 0 && (
              <div className="absolute inset-x-0 bottom-0 h-1 bg-black/40">
                <div className="h-full bg-brand-500" style={{ width: `${percent}%` }} />
              </div>
            )
          )}
        </div>
        <div className="p-4">
          <p className="truncate text-sm font-semibold text-neutral-900">{f.title}</p>
          {f.status === 'processing' ? (
            <Badge tone="warning" className="mt-1">
              <Loader2 size={12} className="animate-spin" /> Processing…
            </Badge>
          ) : f.status === 'failed' ? (
            <Badge tone="danger" className="mt-1">
              Failed
            </Badge>
          ) : (
            <Badge tone="brand" className="mt-1">
              {formatDuration(f.duration_ms)}
            </Badge>
          )}
        </div>
      </Link>
      {isAdmin && (
        <div className="absolute right-2 top-2 flex gap-1.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
          <button
            type="button"
            onClick={onEdit}
            className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-brand-600 hover:text-white"
          >
            <Pencil size={14} />
          </button>
          <button
            type="button"
            onClick={onDelete}
            className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-red-600 hover:text-white"
          >
            <Trash2 size={14} />
          </button>
        </div>
      )}
    </div>
  )
}

function SeriesCard({
  group,
  meta,
  isAdmin,
  onEdit,
}: {
  group: SeriesGroup
  meta?: SeriesMeta
  isAdmin: boolean
  onEdit: () => void
}) {
  const processing = group.episodes.some((e) => e.status === 'processing')
  const total = group.episodes.length
  const watched = group.episodes.filter((e) => watchProgress(e.position_ms, e.duration_ms).done).length
  return (
    <div className="group relative overflow-hidden rounded-2xl bg-white ring-1 ring-neutral-200 transition-colors hover:ring-brand-300">
      <Link to={`series/${encodeURIComponent(group.title)}`} className="block">
        <div className="relative flex aspect-[2/3] items-center justify-center overflow-hidden bg-brand-50">
          {meta?.poster_url ? (
            <>
              <img src={meta.poster_url} alt="" aria-hidden className="absolute inset-0 h-full w-full scale-110 object-cover blur-xl" />
              <img src={meta.poster_url} alt={group.title} className="relative h-full w-full object-contain" />
            </>
          ) : (
            <Layers className="h-8 w-8 text-brand-300" />
          )}
          <span className="absolute left-2 top-2 inline-flex items-center gap-1 rounded-full bg-black/60 px-2 py-0.5 text-xs font-medium text-white">
            <Layers size={11} /> Series
          </span>
          {watched > 0 && (
            <div className="absolute inset-x-0 bottom-0 h-1 bg-black/30">
              <div className="h-full bg-brand-500" style={{ width: `${(watched / total) * 100}%` }} />
            </div>
          )}
        </div>
        <div className="p-4">
          <p className="truncate text-sm font-semibold text-neutral-900">{group.title}</p>
          <Badge tone="brand" className="mt-1">
            {processing && <Loader2 size={12} className="animate-spin" />}
            {watched > 0 ? `${watched}/${total} watched` : `${total} ${total === 1 ? 'episode' : 'episodes'}`}
          </Badge>
        </div>
      </Link>
      {isAdmin && (
        <button
          type="button"
          onClick={onEdit}
          title="Edit series"
          className="absolute right-2 top-2 rounded-lg bg-white/90 p-1.5 text-neutral-600 opacity-100 ring-1 ring-neutral-200 transition-opacity hover:bg-brand-600 hover:text-white sm:opacity-0 sm:group-hover:opacity-100"
        >
          <Pencil size={14} />
        </button>
      )}
    </div>
  )
}

function SeriesEditForm({ title, meta, onDone }: { title: string; meta?: SeriesMeta; onDone: () => void }) {
  const update = useUpdateSeries()
  const [newTitle, setNewTitle] = useState(title)
  const [description, setDescription] = useState(meta?.description ?? '')
  const [poster, setPoster] = useState<File | null>(null)

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTitle.trim()) return
    update.mutate(
      { title, newTitle: newTitle.trim(), description, poster: poster ?? undefined },
      { onSuccess: onDone },
    )
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-brand-200">
      <p className="text-sm font-semibold text-neutral-900">Edit series "{title}"</p>
      <div className="flex gap-4">
        <div className="shrink-0">
          <Field label="Cover">
            <ImageField value={poster} onChange={setPoster} initialUrl={meta?.poster_url} aspect="aspect-[2/3]" placeholder="Add cover" className="w-28" />
          </Field>
        </div>
        <div className="flex-1 space-y-3">
          <Field label="Title">
            <Input value={newTitle} onChange={(e) => setNewTitle(e.target.value)} />
          </Field>
          <Field label="Description">
            <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} />
          </Field>
        </div>
      </div>
      {update.isError && <p className="text-sm text-red-600">Failed to save. Please try again.</p>}
      <div className="flex items-center gap-3">
        <Button type="submit" variant="brand" disabled={!newTitle.trim() || update.isPending}>
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

function EditForm({ film, onDone }: { film: FilmSummary; onDone: () => void }) {
  const update = useUpdateFilm()
  const [title, setTitle] = useState(film.title)
  const [description, setDescription] = useState(film.description ?? '')
  const [level, setLevel] = useState(film.level || 'B1')
  const [poster, setPoster] = useState<File | null>(null)

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    update.mutate({ id: film.id, title, description, level, poster: poster ?? undefined }, { onSuccess: onDone })
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-brand-200">
      <p className="text-sm font-semibold text-neutral-900">Edit "{film.title}"</p>
      <div className="flex gap-4">
        <div className="shrink-0">
          <Field label="Poster">
            <ImageField value={poster} onChange={setPoster} initialUrl={film.poster_url} aspect="aspect-[2/3]" placeholder="Add poster" className="w-28" />
          </Field>
        </div>
        <div className="flex-1 space-y-3">
          <Field label="Title">
            <Input value={title} onChange={(e) => setTitle(e.target.value)} />
          </Field>
          <Field label="Description">
            <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} />
          </Field>
          <Field label="Level">
            <LevelPicker value={level} onChange={setLevel} />
          </Field>
        </div>
      </div>
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

const LEVELS = ['A1', 'A2', 'B1', 'B2', 'C1', 'C2']

function LevelPicker({ value, onChange }: { value: string; onChange: (level: string) => void }) {
  return (
    <div className="flex gap-1">
      {LEVELS.map((l) => (
        <button
          key={l}
          type="button"
          onClick={() => onChange(l)}
          className={`rounded-full px-2.5 py-1 text-xs font-semibold ring-1 transition-colors ${
            value === l ? 'bg-brand-600 text-white ring-brand-600' : 'bg-white text-neutral-600 ring-neutral-200 hover:bg-neutral-50'
          }`}
        >
          {l}
        </button>
      ))}
    </div>
  )
}

function UploadForm({ onDone }: { onDone: () => void }) {
  const upload = useUploadFilm()
  const [kind, setKind] = useState<'film' | 'series'>('film')
  const [title, setTitle] = useState('')
  const [seriesTitle, setSeriesTitle] = useState('')
  const [season, setSeason] = useState(1)
  const [episode, setEpisode] = useState(1)
  const [level, setLevel] = useState('B1')
  const [video, setVideo] = useState<File | null>(null)
  const [subtitles, setSubtitles] = useState<File | null>(null)
  const [poster, setPoster] = useState<File | null>(null)

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!video) return
    if (kind === 'series' && !seriesTitle.trim()) return
    upload.mutate(
      {
        title,
        video,
        subtitles: subtitles ?? undefined,
        poster: poster ?? undefined,
        kind,
        level,
        seriesTitle: seriesTitle.trim(),
        season,
        episode,
      },
      {
        onSuccess: () => {
          setTitle('')
          setVideo(null)
          setSubtitles(null)
          setPoster(null)
          onDone()
        },
      },
    )
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-neutral-200">
      <Tabs
        value={kind}
        onChange={setKind}
        options={[
          { value: 'film', label: 'Film' },
          { value: 'series', label: 'Series' },
        ]}
      />
      {kind === 'series' && (
        <>
          <Field label="Series name *">
            <Input
              value={seriesTitle}
              onChange={(e) => setSeriesTitle(e.target.value)}
              placeholder="e.g. Breaking Bad"
            />
          </Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label="Season">
              <Input
                type="number"
                min={1}
                value={season}
                onChange={(e) => setSeason(Math.max(1, Number(e.target.value)))}
              />
            </Field>
            <Field label="Episode">
              <Input
                type="number"
                min={1}
                value={episode}
                onChange={(e) => setEpisode(Math.max(1, Number(e.target.value)))}
              />
            </Field>
          </div>
        </>
      )}
      <div className="flex gap-4">
        <div className="shrink-0">
          <Field label="Poster">
            <ImageField value={poster} onChange={setPoster} aspect="aspect-[2/3]" placeholder="Add poster" className="w-28" />
          </Field>
        </div>
        <div className="flex-1 space-y-3">
          <Field label={kind === 'series' ? 'Episode title' : 'Title'}>
            <Input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={kind === 'series' ? 'Optional — defaults to S1E1' : 'Optional — defaults to the file name'}
            />
          </Field>
          <Field label="Video (MKV / MP4 / WebM) *">
            <FileField value={video} onChange={setVideo} accept="video/*,.mkv,.avi" placeholder="Choose a video or drop it here" icon={<FileVideo className="h-4 w-4" />} />
          </Field>
          <Field label="Subtitles (.srt)">
            <FileField value={subtitles} onChange={setSubtitles} accept=".srt" placeholder="Optional subtitles" icon={<Captions className="h-4 w-4" />} />
          </Field>
          <Field label="Level (CEFR)">
            <LevelPicker value={level} onChange={setLevel} />
          </Field>
        </div>
      </div>
      <p className="text-xs text-neutral-400">Tracks and subtitles are extracted automatically; the status will update on its own.</p>
      {upload.isError && <p className="text-sm text-red-600">Upload failed. Check the files and try again.</p>}
      <div className="flex items-center gap-3">
        <Button
          type="submit"
          variant="brand"
          disabled={!video || upload.isPending || (kind === 'series' && !seriesTitle.trim())}
        >
          {upload.isPending ? (
            <>
              <Loader2 size={16} className="animate-spin" /> Uploading…
            </>
          ) : (
            kind === 'series' ? 'Upload episode' : 'Upload film'
          )}
        </Button>
        <button type="button" onClick={onDone} className="text-sm text-neutral-500 hover:text-neutral-700">
          Cancel
        </button>
      </div>
    </form>
  )
}
