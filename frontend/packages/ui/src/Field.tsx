import type { ReactNode } from 'react'
import { cn } from './cn.ts'

export interface FieldProps {
  label: ReactNode
  hint?: ReactNode
  error?: ReactNode
  className?: string
  children: ReactNode
}

export function Field({ label, hint, error, className, children }: FieldProps) {
  return (
    <label className={cn('block', className)}>
      <span className="mb-1 block text-xs font-medium text-neutral-500">{label}</span>
      {children}
      {error ? (
        <span className="mt-1 block text-xs text-red-600">{error}</span>
      ) : hint ? (
        <span className="mt-1 block text-xs text-neutral-400">{hint}</span>
      ) : null}
    </label>
  )
}
