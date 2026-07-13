import type { Node } from '../parse.ts'
import { Inline } from '../markdown.tsx'
import { ItemText } from './gaps.tsx'

type ForkBlock = Extract<Node, { t: 'fork' }>

export function Fork({ block }: { block: ForkBlock }) {
  return (
    <div className="flex w-fit max-w-full items-center gap-0">
      <span className="z-10 shrink-0 rounded-xl bg-brand-600 px-3 py-1.5 text-sm font-semibold text-white shadow-sm">
        <Inline text={block.stem} />
      </span>
      <div className="relative flex min-w-0 flex-col gap-1.5 pl-7">
        <span className="absolute bottom-4 left-3 top-4 w-px bg-neutral-300" aria-hidden />
        {block.branches.map((b, i) => (
          <span key={i} className="relative flex min-w-0 items-center py-0.5">
            <span className="absolute -left-4 h-px w-4 bg-neutral-300" aria-hidden />
            <span className="min-w-0 rounded-lg bg-white px-2.5 py-1 text-sm text-neutral-800 shadow-sm ring-1 ring-neutral-200">
              {b.includes('{{') ? <ItemText text={b} lineIdx={i} /> : <Inline text={b} />}
            </span>
          </span>
        ))}
      </div>
    </div>
  )
}
