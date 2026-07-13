import { createContext } from 'react'
import { mockCheckAnswer, type CheckFn } from '../check.ts'

export const ACCENT = 'bg-brand-50 ring-1 ring-brand-100'
export const BADGE = 'bg-brand-600'

export const BlockCtx = createContext<{
  section?: string
  dense: boolean
  check: CheckFn
  onTheory: (s: string) => void
  keyBase?: string
  instruction?: string
}>({
  dense: false,
  check: mockCheckAnswer,
  onTheory: () => {},
})

export function scopeKey(base: string | undefined, i: number): string | undefined {
  return base === undefined ? undefined : `${base}.${i}`
}

// Responsive grid columns based on the CONTAINER width (the panel), not the viewport,
// so columns only split when the panel itself is wide enough.
export function gridColsClass(cols: number): string {
  if (cols >= 4) return 'grid-cols-1 @xl:grid-cols-2 @3xl:grid-cols-4'
  if (cols === 3) return 'grid-cols-1 @xl:grid-cols-2 @3xl:grid-cols-3'
  if (cols === 2) return 'grid-cols-1 @xl:grid-cols-2'
  return 'grid-cols-1'
}
