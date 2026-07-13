import { cn } from './cn.ts'

export interface TabsProps<T extends string> {
  value: T
  onChange: (value: T) => void
  options: readonly { value: T; label: string }[]
  className?: string
}

export function Tabs<T extends string>({ value, onChange, options, className }: TabsProps<T>) {
  return (
    <div className={cn('flex gap-1 border-b border-neutral-200', className)}>
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          onClick={() => onChange(o.value)}
          className={cn(
            '-mb-px border-b-2 px-3 py-2 text-sm font-medium transition-colors',
            o.value === value ? 'border-brand-600 text-brand-700' : 'border-transparent text-neutral-500 hover:text-neutral-800',
          )}
        >
          {o.label}
        </button>
      ))}
    </div>
  )
}
