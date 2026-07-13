import { cn } from './cn.ts'

export interface VersionBadgeProps {
  version: string
  className?: string
}

type Variant = 'dev' | 'stage' | 'prod'

function parse(raw: string): { label: string; title: string; variant: Variant } {
  if (!raw || raw === 'dev' || raw === 'unknown') {
    return { label: 'dev', title: 'Local development build', variant: 'dev' }
  }
  if (raw.startsWith('stage-')) {
    const sha = raw.slice('stage-'.length)
    return {
      label: `stage · ${sha}`,
      title: `Staging build · commit ${sha}`,
      variant: 'stage',
    }
  }
  if (/^v\d/.test(raw)) {
    return { label: raw, title: `Production release ${raw}`, variant: 'prod' }
  }
  return { label: raw, title: raw, variant: 'dev' }
}

const variants: Record<Variant, string> = {
  dev: 'bg-neutral-200/60 text-neutral-500 ring-neutral-300/60 hover:bg-neutral-200 hover:text-neutral-700',
  stage:
    'bg-amber-100/70 text-amber-800 ring-amber-200/70 hover:bg-amber-100 hover:text-amber-900',
  prod: 'bg-neutral-900/5 text-neutral-500 ring-neutral-900/10 hover:bg-neutral-900/10 hover:text-neutral-800',
}

export function VersionBadge({ version, className }: VersionBadgeProps) {
  const { label, title, variant } = parse(version)

  return (
    <div
      className={cn(
        'pointer-events-none fixed bottom-2 right-2 z-40 select-none',
        'hidden sm:block',
        className,
      )}
    >
      <span
        title={title}
        className={cn(
          'pointer-events-auto inline-flex items-center rounded-full px-2 py-0.5',
          'font-mono text-[10px] leading-none tracking-tight',
          'ring-1 backdrop-blur transition-colors',
          variants[variant],
        )}
      >
        {label}
      </span>
    </div>
  )
}
