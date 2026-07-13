import { useEffect, useLayoutEffect, useRef, useState, type PointerEvent as ReactPointerEvent } from 'react'
import type { Node } from '../parse.ts'
import { useBlockProgress } from './blockProgress.ts'

type P = { x: number; y: number }
type MatchBlock = Extract<Node, { t: 'match' }>

function curve(a: P, b: P): string {
  const mx = a.x + (b.x - a.x) / 2
  return `M ${a.x} ${a.y} C ${mx} ${a.y}, ${mx} ${b.y}, ${b.x} ${b.y}`
}

export function MatchConnect({ block }: { block: MatchBlock }) {
  const wrap = useRef<HTMLDivElement>(null)
  const leftDots = useRef<Record<string, HTMLElement | null>>({})
  const rightDots = useRef<Record<string, HTMLElement | null>>({})
  const [pts, setPts] = useState<{ l: Record<string, P>; r: Record<string, P> }>({ l: {}, r: {} })
  const bp = useBlockProgress<Record<string, string>>('match')
  const [conns, setConns] = useState<Record<string, string>>(() => bp.saved ?? {})
  const [drag, setDrag] = useState<({ n: string } & P) | null>(null)

  const measure = () => {
    const c = wrap.current?.getBoundingClientRect()
    if (!c) return
    const center = (el: HTMLElement | null | undefined): P | null => {
      if (!el) return null
      const b = el.getBoundingClientRect()
      return { x: b.left - c.left + b.width / 2, y: b.top - c.top + b.height / 2 }
    }
    const l: Record<string, P> = {}
    const r: Record<string, P> = {}
    for (const it of block.left) {
      const p = center(leftDots.current[it.n])
      if (p) l[it.n] = p
    }
    for (const it of block.right) {
      const p = center(rightDots.current[it.l])
      if (p) r[it.l] = p
    }
    setPts({ l, r })
  }

  useLayoutEffect(() => {
    measure()
  }, [])
  useEffect(() => {
    window.addEventListener('resize', measure)
    return () => window.removeEventListener('resize', measure)
  })

  const rel = (e: ReactPointerEvent): P => {
    const c = wrap.current!.getBoundingClientRect()
    return { x: e.clientX - c.left, y: e.clientY - c.top }
  }

  function onUp(e: ReactPointerEvent) {
    if (!drag) return
    const p = rel(e)
    let best: string | null = null
    let bestD = 40
    for (const it of block.right) {
      const rp = pts.r[it.l]
      if (!rp) continue
      const d = Math.hypot(rp.x - p.x, rp.y - p.y)
      if (d < bestD) {
        bestD = d
        best = it.l
      }
    }
    if (best) {
      const next = { ...conns, [drag.n]: best }
      setConns(next)
      const done = block.left.every((it) => next[it.n] === it.answer)
      bp.save(next, done)
    }
    setDrag(null)
  }

  return (
    <div
      ref={wrap}
      className="relative select-none"
      style={{ touchAction: drag ? 'none' : undefined }}
      onPointerMove={(e) => drag && setDrag({ n: drag.n, ...rel(e) })}
      onPointerUp={onUp}
      onPointerLeave={() => setDrag(null)}
    >
      <svg className="pointer-events-none absolute inset-0 h-full w-full overflow-visible">
        {Object.entries(conns).map(([n, l]) => {
          const a = pts.l[n]
          const b = pts.r[l]
          if (!a || !b) return null
          const correct = block.left.find((x) => x.n === n)?.answer === l
          return <path key={n} d={curve(a, b)} fill="none" strokeWidth={2.5} className={correct ? 'stroke-emerald-500' : 'stroke-rose-400'} />
        })}
        {drag && pts.l[drag.n] && (
          <path d={curve(pts.l[drag.n]!, drag)} fill="none" strokeWidth={2.5} strokeDasharray="5 4" className="stroke-brand-400" />
        )}
      </svg>

      <div className="grid min-w-0 grid-cols-2 items-stretch gap-x-4 sm:gap-x-10">
        <div className="flex min-w-0 flex-col justify-between gap-2.5">
          {block.left.map((it) => {
            const connected = conns[it.n]
            const correct = connected && it.answer === connected
            return (
              <div key={it.n} className="flex min-w-0 items-center justify-between gap-2 rounded-xl border border-neutral-200/90 bg-white px-3 py-2.5 text-sm shadow-sm">
                <span className="flex min-w-0 gap-1.5">
                  <span className="shrink-0 font-semibold text-neutral-400">{it.n}</span>
                  <span className="min-w-0 break-words">{it.text}</span>
                </span>
                <span
                  ref={(el) => {
                    leftDots.current[it.n] = el
                  }}
                  onPointerDown={(e) => {
                    e.preventDefault()
                    measure()
                    setDrag({ n: it.n, ...rel(e) })
                  }}
                  className={`relative h-4 w-4 shrink-0 cursor-grab touch-none rounded-full border-2 transition-colors before:absolute before:-inset-2.5 before:content-[''] ${
                    connected
                      ? correct
                        ? 'border-emerald-500 bg-emerald-500'
                        : 'border-rose-400 bg-rose-400'
                      : 'border-brand-400 bg-white hover:bg-brand-100'
                  }`}
                />
              </div>
            )
          })}
        </div>
        <div className="flex min-w-0 flex-col justify-between gap-2.5">
          {block.right.map((it) => (
            <div key={it.l} className="flex min-w-0 items-center gap-2 rounded-xl border border-neutral-200/90 bg-white px-3 py-2.5 text-sm shadow-sm">
              <span
                ref={(el) => {
                  rightDots.current[it.l] = el
                }}
                className="h-4 w-4 shrink-0 rounded-full border-2 border-neutral-400 bg-white"
              />
              <span className="flex min-w-0 gap-1.5">
                <span className="shrink-0 font-semibold text-neutral-400">{it.l}</span>
                <span className="min-w-0 break-words">{it.text}</span>
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
