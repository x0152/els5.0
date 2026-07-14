import { useState, type ReactNode } from 'react'
import { BookOpen, Film, ImageIcon, Loader2, Trash2, TriangleAlert, X } from 'lucide-react'
import { Badge, Button, cn, CefrBadge, FrequencyBars, Modal } from '@els/ui'
import { SpotsDialog } from '@els/lookup'
import { useWordImage } from '../hooks/useWordImage.ts'
import { useDeleteUnit, useUnitOccurrences, useUpdateStatus } from '../store/units.ts'
import { KindGlyph } from './KindGlyph.tsx'
import { statusPill } from '../lib/ui.ts'
import { KIND_LABELS, STATUS_LABELS } from '../lib/types.ts'
import type { Unit, UnitStatus } from '../lib/types.ts'

const STATUSES: UnitStatus[] = ['new', 'learning', 'learned']

interface Props {
  unit: Unit
  onClose: () => void
}

export function WordDetailModal({ unit, onClose }: Props) {
  const updateM = useUpdateStatus()
  const deleteM = useDeleteUnit()
  const image = useWordImage(unit.text)
  const occ = useUnitOccurrences(unit.text).data
  type Media = NonNullable<NonNullable<typeof occ>['media']>[number]
  const [places, setPlaces] = useState<Media | null>(null)
  const status = (updateM.variables?.id === unit.id ? updateM.variables.status : unit.status) as UnitStatus

  const spotHref = (mediaType: string, mediaId: string, ref: number) =>
    mediaType === 'film' ? `/v1/films/${mediaId}?t=${ref}` : `/v1/reader/${mediaId}?pos=${ref}`

  async function remove() {
    await deleteM.mutateAsync(unit.id)
    onClose()
  }

  return (
    <Modal onClose={() => (places ? setPlaces(null) : onClose())}>
      <div className="mb-4 flex items-start justify-between">
        <div>
          <div className="flex flex-wrap items-center gap-1.5">
            <Badge className="text-[11px]">
              <KindGlyph kind={unit.kind} className="h-3 w-3" /> {KIND_LABELS[unit.kind] ?? unit.kind}
            </Badge>
            <CefrBadge level={unit.cefr} />
            <FrequencyBars value={unit.frequency} />
            {image.status !== 'ready' && (
              <button
                type="button"
                onClick={image.generate}
                disabled={image.status === 'generating'}
                className="inline-flex items-center gap-1 rounded-full bg-brand-600 px-2.5 py-0.5 text-[11px] font-semibold text-white transition-colors hover:bg-brand-700 disabled:opacity-70"
              >
                {image.status === 'generating' ? (
                  <Loader2 className="h-3 w-3 animate-spin" />
                ) : image.status === 'error' ? (
                  <TriangleAlert className="h-3 w-3" />
                ) : (
                  <ImageIcon className="h-3 w-3" />
                )}
                {image.status === 'generating' ? 'Generating…' : image.status === 'error' ? 'Retry image' : 'Image'}
              </button>
            )}
          </div>
          <h2 className="mt-2 text-2xl font-bold text-neutral-900">{unit.text}</h2>
          {unit.transcription && <p className="text-sm text-neutral-400">/{unit.transcription}/</p>}
        </div>
        <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close">
          <X className="h-4 w-4" />
        </Button>
      </div>

      {image.status === 'ready' && image.url && (
        <img
          src={image.url}
          alt={unit.text}
          className="mb-4 aspect-square w-full rounded-xl object-cover ring-1 ring-neutral-200"
        />
      )}
      {image.status === 'generating' && (
        <div className="mb-4 grid aspect-video place-items-center rounded-xl bg-neutral-100 text-neutral-500 ring-1 ring-neutral-200">
          <div className="flex items-center gap-2 text-sm">
            <Loader2 className="h-4 w-4 animate-spin text-brand-500" /> Generating image…
          </div>
        </div>
      )}

      <div className="space-y-3">
        {unit.translation && (
          <Field label="Translation">
            <p className="text-base text-neutral-800">{unit.translation}</p>
          </Field>
        )}
        {unit.definition && (
          <Field label="Definition">
            <p className="text-sm text-neutral-700">{unit.definition}</p>
          </Field>
        )}
        {unit.example && (
          <Field label="Example">
            <p className="text-sm italic text-neutral-600">“{unit.example}”</p>
          </Field>
        )}
        {occ && (occ.common ? occ.total > 0 : (occ.media?.length ?? 0) > 0) && (
          <Field label="Found in">
            {occ.common ? (
              <p className="text-sm text-neutral-500">Common word · seen {occ.total}×</p>
            ) : (
              <div className="flex flex-wrap gap-1.5">
                {(occ.media ?? []).map((m, i) => {
                  const chip = (
                    <>
                      {m.media_type === 'film' ? <Film className="h-3 w-3" /> : <BookOpen className="h-3 w-3" />}
                      <span className="max-w-[220px] truncate">{m.title || 'Untitled'}</span>
                      {m.count > 1 && <span className="text-neutral-400">×{m.count}</span>}
                    </>
                  )
                  const chipClass =
                    'inline-flex items-center gap-1 rounded-md bg-neutral-100 px-2 py-1 text-xs text-neutral-600 hover:bg-neutral-200'
                  return m.count <= 1 ? (
                    <a key={`${m.title}-${i}`} href={spotHref(m.media_type, m.media_id, m.spots?.[0]?.ref ?? 0)} className={chipClass}>
                      {chip}
                    </a>
                  ) : (
                    <button
                      key={`${m.title}-${i}`}
                      type="button"
                      onClick={() => setPlaces(m)}
                      className={chipClass}
                    >
                      {chip}
                    </button>
                  )
                })}
              </div>
            )}
          </Field>
        )}
      </div>

      <div className="mt-5">
        <p className="mb-1.5 text-xs font-medium text-neutral-500">Memorization status</p>
        <div className="flex gap-2">
          {STATUSES.map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => updateM.mutate({ id: unit.id, status: s })}
              className={cn(
                'flex-1 rounded-xl px-3 py-2 text-sm font-medium ring-1 transition',
                status === s ? statusPill[s] : 'bg-white text-neutral-600 ring-neutral-200 hover:bg-neutral-50',
              )}
            >
              {STATUS_LABELS[s]}
            </button>
          ))}
        </div>
      </div>

      <button
        type="button"
        onClick={remove}
        disabled={deleteM.isPending}
        className="mt-5 inline-flex items-center gap-2 text-sm font-medium text-rose-600 transition hover:text-rose-700 disabled:opacity-50"
      >
        <Trash2 className="h-4 w-4" /> Delete from collection
      </button>

      {places && (
        <SpotsDialog
          title={places.title}
          mediaType={places.media_type}
          kind={places.kind}
          seriesTitle={places.series_title}
          season={places.season}
          episode={places.episode}
          author={places.author}
          spots={(places.spots ?? []).map((s) => ({ ref: s.ref, example: s.example ?? '' }))}
          hrefFor={(ref) => spotHref(places.media_type, places.media_id, ref)}
          onClose={() => setPlaces(null)}
        />
      )}
    </Modal>
  )
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div>
      <p className="mb-0.5 text-xs font-medium uppercase tracking-wide text-neutral-400">{label}</p>
      {children}
    </div>
  )
}
