import { useMemo, useState } from 'react'
import { Check } from 'lucide-react'
import type { Node } from '../parse.ts'
import { useBlockProgress } from './blockProgress.ts'

type SortBlock = Extract<Node, { t: 'sort' }>

// Deterministic shuffle so the pool order is stable across re-renders.
function shuffled(items: string[]): string[] {
  const out = [...items]
  let seed = items.join('|').split('').reduce((h, c) => (h * 31 + c.charCodeAt(0)) % 2147483647, 7)
  for (let i = out.length - 1; i > 0; i--) {
    seed = (seed * 48271) % 2147483647
    const j = seed % (i + 1)
    ;[out[i], out[j]] = [out[j]!, out[i]!]
  }
  return out
}

export function SortBuckets({ block }: { block: SortBlock }) {
  const answers = useMemo(() => {
    const m = new Map<string, string>()
    for (const c of block.cats) for (const it of c.items) m.set(it, c.name)
    return m
  }, [block])
  const pool = useMemo(() => shuffled([...answers.keys()]), [answers])
  const bp = useBlockProgress<Record<string, string>>('sort')
  const [placed, setPlaced] = useState<Record<string, string>>(() => bp.saved ?? {})
  const [selected, setSelected] = useState<string | null>(null)

  const unplaced = pool.filter((it) => !placed[it])
  const allCorrect = unplaced.length === 0 && Object.entries(placed).every(([it, cat]) => answers.get(it) === cat)

  const update = (next: Record<string, string>) => {
    setPlaced(next)
    const done = pool.every((it) => next[it]) && Object.entries(next).every(([it, cat]) => answers.get(it) === cat)
    bp.save(next, done)
  }

  const place = (cat: string) => {
    if (!selected) return
    update({ ...placed, [selected]: cat })
    setSelected(null)
  }

  return (
    <div className="space-y-3">
      {unplaced.length > 0 ? (
        <div className="flex flex-wrap gap-2 rounded-xl bg-brand-50/80 p-3.5 ring-1 ring-brand-200/80">
          {unplaced.map((it) => (
            <button
              key={it}
              type="button"
              onClick={() => setSelected((s) => (s === it ? null : it))}
              className={`rounded-lg px-2.5 py-1 text-sm font-medium shadow-sm ring-1 transition-colors ${
                selected === it
                  ? 'bg-brand-600 text-white ring-brand-600'
                  : 'bg-white text-neutral-800 ring-neutral-200 hover:ring-brand-300'
              }`}
            >
              {it}
            </button>
          ))}
        </div>
      ) : (
        <div className={`flex items-center gap-1.5 text-sm font-medium ${allCorrect ? 'text-emerald-600' : 'text-neutral-500'}`}>
          {allCorrect && <Check className="h-4 w-4" />}
          {allCorrect ? 'All sorted correctly!' : 'All placed — fix the red ones.'}
        </div>
      )}

      <div className={`grid gap-3 ${block.cats.length >= 3 ? 'grid-cols-1 @xl:grid-cols-3' : block.cats.length === 2 ? 'grid-cols-1 @xl:grid-cols-2' : 'grid-cols-1'}`}>
        {block.cats.map((cat) => (
          <button
            key={cat.name}
            type="button"
            onClick={() => place(cat.name)}
            className={`min-h-24 rounded-xl border bg-white p-3 text-left shadow-sm transition-colors ${
              selected ? 'cursor-pointer border-brand-300 ring-2 ring-brand-100 hover:border-brand-400' : 'cursor-default border-neutral-200/90'
            }`}
          >
            <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-neutral-500">{cat.name}</div>
            <div className="flex flex-wrap gap-1.5">
              {pool
                .filter((it) => placed[it] === cat.name)
                .map((it) => {
                  const correct = answers.get(it) === cat.name
                  return (
                    <span
                      key={it}
                      role="button"
                      tabIndex={0}
                      onClick={(e) => {
                        e.stopPropagation()
                        const { [it]: _, ...rest } = placed
                        update(rest)
                      }}
                      className={`cursor-pointer rounded-lg px-2 py-0.5 text-sm font-medium ring-1 transition-colors ${
                        correct ? 'bg-emerald-50 text-emerald-700 ring-emerald-300' : 'bg-rose-50 text-rose-600 ring-rose-300'
                      }`}
                    >
                      {it}
                      <span className="ml-1 opacity-50">×</span>
                    </span>
                  )
                })}
            </div>
          </button>
        ))}
      </div>
    </div>
  )
}
