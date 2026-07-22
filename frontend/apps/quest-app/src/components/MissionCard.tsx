import { CheckCircle2, ImageIcon, Loader2, Trash2, TriangleAlert } from 'lucide-react'
import { cn } from '@els/ui'
import { genreEmoji } from '../lib/helpers.ts'
import type { MissionSummary } from '../lib/types.ts'

interface Props {
  mission: MissionSummary
  onOpen: (id: string) => void
  onDelete: (mission: MissionSummary) => void
}

export function MissionCard({ mission, onOpen, onDelete }: Props) {
  const generating = mission.generationStatus === 'generating'
  const failed = mission.generationStatus === 'error'
  const clickable = !generating && !failed

  return (
    <div
      className={cn(
        'group relative aspect-[3/4] overflow-hidden rounded-2xl ring-1 ring-neutral-200 transition',
        clickable && 'hover:-translate-y-1 hover:shadow-lg',
      )}
    >
      {mission.coverImage ? (
        <img src={mission.coverImage} alt={mission.title} className="absolute inset-0 h-full w-full object-cover" />
      ) : (
        <div className="absolute inset-0 flex items-center justify-center bg-gradient-to-br from-brand-100 to-brand-300 text-6xl">
          {genreEmoji(mission.genre)}
        </div>
      )}

      <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/10 to-transparent" />

      {clickable && (
        <button
          type="button"
          aria-label={mission.title || 'Open adventure'}
          onClick={() => onOpen(mission.id)}
          className="absolute inset-0 z-10 cursor-pointer"
        />
      )}

      <div className="pointer-events-none absolute left-2.5 top-2.5 z-20 flex items-center gap-1.5">
        {mission.coverImageStatus === 'generating' && !generating && (
          <span className="flex items-center gap-1 rounded-full bg-black/60 px-2 py-0.5 text-[10px] font-medium text-white backdrop-blur-sm">
            <ImageIcon className="h-2.5 w-2.5" />
            Artwork…
          </span>
        )}
        {mission.isComplete && (
          <span className="flex items-center gap-1 rounded-full bg-brand-600/90 px-2 py-0.5 text-[10px] font-medium text-white">
            <CheckCircle2 className="h-2.5 w-2.5" />
            Completed
          </span>
        )}
      </div>

      <button
        type="button"
        aria-label="Delete"
        onClick={() => onDelete(mission)}
        className="absolute right-2.5 top-2.5 z-30 grid h-7 w-7 place-items-center rounded-full bg-black/50 text-white backdrop-blur-sm transition hover:bg-rose-600 sm:opacity-0 sm:group-hover:opacity-100"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>

      {generating && (
        <div className="pointer-events-none absolute inset-0 z-20 flex flex-col items-center justify-center gap-2 bg-black/55 px-4 text-center text-white">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span className="text-xs font-medium">Creating adventure…</span>
          <span className="text-[10px] text-white/70">Usually takes 5–10 minutes — check back soon</span>
        </div>
      )}
      {failed && (
        <div className="pointer-events-none absolute inset-0 z-20 flex flex-col items-center justify-center gap-2 bg-rose-950/60 px-4 text-center text-white">
          <TriangleAlert className="h-6 w-6" />
          <span className="text-xs font-medium">Generation failed</span>
        </div>
      )}

      <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 p-3">
        <div className="mb-1 flex items-center gap-1.5">
          <span className="rounded-full bg-white/20 px-2 py-0.5 text-[10px] font-medium text-white backdrop-blur-sm">
            {genreEmoji(mission.genre)} {mission.genre || 'story'}
          </span>
        </div>
        <h3 className="line-clamp-2 text-sm font-semibold leading-tight text-white">
          {mission.title || 'New adventure'}
        </h3>
      </div>
    </div>
  )
}
