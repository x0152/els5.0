import { useMemo, useState } from 'react'
import { Check } from 'lucide-react'
import type { Node } from '../parse.ts'
import { useBlockProgress } from './blockProgress.ts'

type HighlightBlock = Extract<Node, { t: 'highlight' }>

type Seg = { kind: 'target'; id: number; text: string } | { kind: 'word'; text: string } | { kind: 'space'; text: string }

function segment(lines: string[]): { segs: Seg[][]; total: number } {
  let id = 0
  const segs = lines.map((line) => {
    const out: Seg[] = []
    for (const part of line.split(/(==[^=]+==)/g)) {
      if (!part) continue
      const m = /^==([^=]+)==$/.exec(part)
      if (m) {
        out.push({ kind: 'target', id: id++, text: (m[1] ?? '').trim() })
        continue
      }
      for (const tok of part.split(/(\s+)/)) {
        if (!tok) continue
        out.push(/^\s+$/.test(tok) ? { kind: 'space', text: tok } : { kind: 'word', text: tok })
      }
    }
    return out
  })
  return { segs, total: id }
}

export function HighlightText({ block }: { block: HighlightBlock }) {
  const { segs, total } = useMemo(() => segment(block.lines), [block])
  const bp = useBlockProgress<number[]>('highlight')
  const [found, setFound] = useState<Set<number>>(() => new Set(bp.saved ?? []))
  const [wrong, setWrong] = useState<string | null>(null)
  const done = found.size === total

  const mark = (id: number) => {
    const next = new Set(found).add(id)
    setFound(next)
    bp.save([...next], next.size === total)
  }

  const miss = (key: string) => {
    setWrong(key)
    setTimeout(() => setWrong((w) => (w === key ? null : w)), 600)
  }

  return (
    <div className="space-y-2 rounded-xl border border-neutral-200/90 bg-white p-4 shadow-sm">
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-medium text-neutral-400">Tap the words in the text</span>
        <span className={`flex items-center gap-1 text-xs font-semibold ${done ? 'text-emerald-600' : 'text-neutral-500'}`}>
          {done && <Check className="h-3.5 w-3.5" />}
          {found.size} / {total}
        </span>
      </div>
      <div className="text-sm leading-[1.9] text-neutral-800">
        {segs.map((line, li) => (
          <p key={li} className="mb-1.5 last:mb-0">
            {line.map((s, si) => {
              if (s.kind === 'space') return <span key={si}> </span>
              if (s.kind === 'target') {
                const isFound = found.has(s.id)
                return (
                  <button
                    key={si}
                    type="button"
                    onClick={() => mark(s.id)}
                    className={`rounded px-0.5 transition-colors ${
                      isFound ? 'bg-emerald-100 font-semibold text-emerald-800' : 'hover:bg-brand-50'
                    }`}
                  >
                    {s.text}
                  </button>
                )
              }
              const key = `${li}:${si}`
              return (
                <button
                  key={si}
                  type="button"
                  onClick={() => miss(key)}
                  className={`rounded px-0.5 transition-colors ${wrong === key ? 'bg-rose-100 text-rose-600' : 'hover:bg-brand-50'}`}
                >
                  {s.text}
                </button>
              )
            })}
          </p>
        ))}
      </div>
    </div>
  )
}
