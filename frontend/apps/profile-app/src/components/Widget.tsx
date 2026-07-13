import type { ReactNode } from 'react'
import { cn } from '@els/ui'

export function Widget({
  title,
  icon,
  action,
  className,
  children,
}: {
  title: string
  icon?: ReactNode
  action?: ReactNode
  className?: string
  children: ReactNode
}) {
  return (
    <section
      className={cn(
        'rounded-2xl bg-white ring-1 ring-neutral-200 flex flex-col min-h-0',
        className,
      )}
    >
      <header className="flex items-center gap-2 px-5 py-3 border-b border-neutral-100">
        {icon && <span className="text-neutral-500">{icon}</span>}
        <h3 className="text-sm font-semibold text-neutral-800 flex-1">{title}</h3>
        {action}
      </header>
      <div className="flex-1 min-h-0">{children}</div>
    </section>
  )
}

export function StatCard({
  label,
  value,
  hint,
  icon,
  tone = 'neutral',
}: {
  label: string
  value: string | number
  hint?: string
  icon?: ReactNode
  tone?: 'neutral' | 'brand' | 'emerald' | 'amber' | 'rose'
}) {
  const toneCls = {
    neutral: 'bg-neutral-50 text-neutral-700 ring-neutral-200',
    brand: 'bg-brand-50 text-brand-700 ring-brand-200',
    emerald: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
    amber: 'bg-amber-50 text-amber-700 ring-amber-200',
    rose: 'bg-rose-50 text-rose-700 ring-rose-200',
  }[tone]
  return (
    <div className="rounded-xl bg-white ring-1 ring-neutral-200 p-4 flex items-start gap-3">
      {icon && (
        <div className={cn('h-9 w-9 rounded-lg flex items-center justify-center ring-1', toneCls)}>
          {icon}
        </div>
      )}
      <div className="flex-1 min-w-0">
        <div className="text-[11px] font-semibold uppercase tracking-wider text-neutral-500">
          {label}
        </div>
        <div className="mt-0.5 text-2xl font-bold text-neutral-900 leading-tight truncate">
          {value}
        </div>
        {hint && <div className="text-[11px] text-neutral-500 mt-0.5 truncate">{hint}</div>}
      </div>
    </div>
  )
}
