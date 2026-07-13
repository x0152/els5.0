import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { BookOpen } from 'lucide-react'
import type { Node } from '../parse.ts'
import { Inline } from '../markdown.tsx'

type GlossBlock = Extract<Node, { t: 'gloss' }>

function parse(raw: string): { text: string[]; defs: Map<string, string> } {
  const lines = raw.split('\n')
  const sep = lines.findIndex((l) => l.trim() === '---')
  const text = (sep === -1 ? lines : lines.slice(0, sep)).map((l) => l.trim()).filter(Boolean)
  const defs = new Map<string, string>()
  for (const l of sep === -1 ? [] : lines.slice(sep + 1)) {
    const m = /^(\d+)\.\s+(.*)$/.exec(l.trim())
    if (m) defs.set(m[1] ?? '', m[2] ?? '')
  }
  return { text, defs }
}

export function GlossText({ block }: { block: GlossBlock }) {
  const { text, defs } = parse(block.raw)
  const [open, setOpen] = useState<string | null>(null)

  // Keep the popover inside the viewport (markers near the right edge on phones).
  const popRef = useRef<HTMLSpanElement>(null)
  const [shift, setShift] = useState(0)
  useLayoutEffect(() => {
    if (!open) {
      setShift(0)
      return
    }
    const r = popRef.current?.getBoundingClientRect()
    if (!r) return
    const over = r.right - window.innerWidth + 12
    if (over > 0) setShift(-Math.min(over, Math.max(r.left - 12, 0)))
  }, [open])

  useEffect(() => {
    if (!open) return
    const onDown = (e: PointerEvent) => {
      if (!(e.target as Element).closest('[data-gloss]')) setOpen(null)
    }
    document.addEventListener('pointerdown', onDown)
    return () => document.removeEventListener('pointerdown', onDown)
  }, [open])

  return (
    <div className="rounded-xl border border-neutral-200/90 bg-white p-4 shadow-sm">
      <div className="text-[15px] leading-[1.85] text-neutral-800">
        {text.map((line, li) => (
          <p key={li} className="mb-2 last:mb-0">
            {line.split(/(\[\d+\])/g).map((part, pi) => {
              const m = /^\[(\d+)\]$/.exec(part)
              if (!m) return <Inline key={pi} text={part} />
              const n = m[1] ?? ''
              const active = open === `${li}:${pi}`
              return (
                <span key={pi} data-gloss={active ? '' : undefined} className={`relative ${active ? 'z-40' : ''}`}>
                  <button
                    type="button"
                    onClick={() => setOpen(active ? null : `${li}:${pi}`)}
                    className={`-top-1.5 mx-0.5 inline-grid h-4 min-w-4 place-items-center rounded-full px-0.5 align-super text-[10px] font-bold leading-none transition-colors ${
                      active ? 'bg-brand-600 text-white' : 'bg-brand-100 text-brand-700 hover:bg-brand-200'
                    }`}
                  >
                    {n}
                  </button>
                  {active && (
                    <span
                      ref={popRef}
                      style={{ transform: `translateX(${shift}px)` }}
                      className="absolute left-0 top-full z-30 mt-1 block w-64 max-w-[calc(100vw-2rem)] rounded-xl border border-neutral-200 bg-white p-3 text-xs leading-relaxed text-neutral-700 shadow-lg ring-1 ring-black/5"
                    >
                      <span className="mr-1 font-bold text-brand-600">{n}</span>
                      {defs.get(n) ?? '—'}
                    </span>
                  )}
                </span>
              )
            })}
          </p>
        ))}
      </div>
      {defs.size > 0 && (
        <details className="mt-3 border-t border-neutral-100 pt-2">
          <summary className="flex cursor-pointer select-none items-center gap-1.5 text-xs font-medium text-neutral-400 hover:text-brand-600">
            <BookOpen className="h-3.5 w-3.5" /> Word list
          </summary>
          <div className="mt-2 grid gap-x-6 gap-y-1 text-xs leading-relaxed text-neutral-600 @xl:grid-cols-2">
            {[...defs].map(([n, d]) => (
              <div key={n}>
                <span className="font-bold text-brand-600">{n}</span> {d}
              </div>
            ))}
          </div>
        </details>
      )}
    </div>
  )
}
