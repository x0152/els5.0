import { cn } from './cn.ts'

const BAR_HEIGHTS = ['h-1.5', 'h-2', 'h-2.5', 'h-3', 'h-3.5']

function freqColor(value: number): string {
  if (value >= 4) return 'bg-emerald-500'
  if (value === 3) return 'bg-amber-500'
  return 'bg-rose-500'
}

export function FrequencyBars({ value, className }: { value: number; className?: string }) {
  if (!value) return null
  const fill = freqColor(value)
  return (
    <span
      className={cn('inline-flex items-end gap-0.5', className)}
      title={`Frequency ${value}/5`}
      aria-label={`Frequency ${value} of 5`}
    >
      {BAR_HEIGHTS.map((h, i) => (
        <span key={i} className={cn('w-1 rounded-sm', h, i < value ? fill : 'bg-neutral-200')} />
      ))}
    </span>
  )
}

const CEFR_CLASS: Record<string, string> = {
  A1: 'bg-emerald-100 text-emerald-700',
  A2: 'bg-emerald-100 text-emerald-700',
  B1: 'bg-sky-100 text-sky-700',
  B2: 'bg-sky-100 text-sky-700',
  C1: 'bg-violet-100 text-violet-700',
  C2: 'bg-violet-100 text-violet-700',
}

export function CefrBadge({ level, className }: { level: string; className?: string }) {
  if (!level || !CEFR_CLASS[level]) return null
  return (
    <span
      className={cn('inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-semibold', CEFR_CLASS[level], className)}
      title={`CEFR level ${level}`}
    >
      {level}
    </span>
  )
}
