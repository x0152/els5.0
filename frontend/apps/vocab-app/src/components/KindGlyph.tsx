import { createElement } from 'react'
import type { LucideProps } from 'lucide-react'
import { getKindIcon } from '../lib/ui.ts'

export function KindGlyph({ kind, ...props }: { kind: string } & LucideProps) {
  return createElement(getKindIcon(kind), props)
}
