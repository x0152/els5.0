import { Loader2, Trash2 } from 'lucide-react'
import { Badge, cn, CefrBadge, FrequencyBars } from '@els/ui'
import { useWordImage } from '../hooks/useWordImage.ts'
import { useShowTranslations } from '../store/me.ts'
import { KindGlyph } from './KindGlyph.tsx'
import { statusDot } from '../lib/ui.ts'
import { KIND_LABELS } from '../lib/types.ts'
import type { Unit, UnitStatus } from '../lib/types.ts'

interface Props {
  unit: Unit
  onOpen: (unit: Unit) => void
  onDelete: (unit: Unit) => void
}

export function WordCard({ unit, onOpen, onDelete }: Props) {
  const status = unit.status as UnitStatus
  const image = useWordImage(unit.text)
  const showTranslations = useShowTranslations()
  const hasImage = image.status === 'ready' && !!image.url
  return (
    <div className="group relative flex flex-col overflow-hidden rounded-2xl bg-white text-left ring-1 ring-neutral-200 transition hover:-translate-y-0.5 hover:shadow-md">
      <button type="button" onClick={() => onOpen(unit)} className="absolute inset-0 z-10 cursor-pointer" aria-label={unit.text} />

      {hasImage ? (
        <div className="aspect-[4/3] w-full overflow-hidden bg-neutral-100">
          <img src={image.url} alt="" className="h-full w-full object-cover" />
        </div>
      ) : image.status === 'generating' ? (
        <div className="flex aspect-[4/3] w-full animate-pulse items-center justify-center gap-2 bg-neutral-100 text-neutral-400">
          <Loader2 className="h-4 w-4 animate-spin" />
          <span className="text-xs font-medium">Generating image…</span>
        </div>
      ) : null}

      <div className="flex flex-1 flex-col p-4">
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <Badge className="text-[11px]">
              <KindGlyph kind={unit.kind} className="h-3 w-3" /> {KIND_LABELS[unit.kind] ?? unit.kind}
            </Badge>
            <CefrBadge level={unit.cefr} />
            <FrequencyBars value={unit.frequency} />
          </div>
          <div className="flex items-center gap-1.5">
            {status !== 'learned' && (
              <span className="flex items-center gap-0.5" title={`Progress ${unit.correct_streak}/3`}>
                {[0, 1, 2].map((i) => (
                  <span key={i} className={cn('h-1.5 w-1.5 rounded-full', i < unit.correct_streak ? 'bg-brand-500' : 'bg-neutral-200')} />
                ))}
              </span>
            )}
            <span className={cn('h-2 w-2 rounded-full', statusDot[status] ?? 'bg-neutral-300')} title={status} />
          </div>
        </div>

        <h3 className="truncate text-base font-semibold text-neutral-900">{unit.text}</h3>
        {unit.transcription && <p className="mt-0.5 truncate text-xs text-neutral-400">/{unit.transcription}/</p>}
        {unit.definition && <p className="mt-1 line-clamp-2 text-sm text-neutral-600">{unit.definition}</p>}
        {showTranslations && unit.translation && <p className="mt-0.5 line-clamp-1 text-sm text-neutral-500">{unit.translation}</p>}
      </div>

      <button
        type="button"
        aria-label="Delete"
        onClick={() => onDelete(unit)}
        className="absolute right-2 top-2 z-20 grid h-7 w-7 place-items-center rounded-full bg-white/90 text-neutral-500 shadow-sm ring-1 ring-black/5 transition hover:bg-rose-600 hover:text-white sm:opacity-0 sm:group-hover:opacity-100"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>
    </div>
  )
}
