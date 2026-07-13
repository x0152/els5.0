import { Loader2, RotateCcw } from 'lucide-react'
import { cn } from '@els/ui'
import { useRegenerateImage } from '../store/missions.ts'

export function RegenButton({
  missionId,
  kind,
  itemKey,
  className,
}: {
  missionId: string
  kind: 'cover' | 'scene' | 'avatar'
  itemKey?: string
  className?: string
}) {
  const regen = useRegenerateImage(missionId)
  return (
    <button
      type="button"
      onClick={(e) => {
        e.stopPropagation()
        regen.mutate({ kind, key: itemKey })
      }}
      disabled={regen.isPending}
      className={cn(
        'inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-rose-700 hover:bg-rose-100 disabled:opacity-60',
        className,
      )}
    >
      {regen.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <RotateCcw className="h-3 w-3" />}
      Retry
    </button>
  )
}
