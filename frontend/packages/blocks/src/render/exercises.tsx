import { useContext, useState } from 'react'
import { Check, RotateCcw } from 'lucide-react'
import { parseExercises, parseTheory, type Exercise, type Node } from '../parse.ts'
import type { CheckFn } from '../check.ts'
import { useProgress } from '../state.ts'
import { Inline, PROSE_CLS } from '../markdown.tsx'
import { BADGE, BlockCtx } from './context.ts'
import { RenderNodes } from './nodes.tsx'

export function ExercisesList({ exercises, checkAnswer, onTheory }: { exercises: string; checkAnswer: CheckFn; onTheory: (s: string) => void }) {
  const items = parseExercises(exercises)
  return (
    <div className="@container space-y-4">
      {items.map((ex) => (
        <BlockCtx.Provider
          key={ex.id}
          value={{ section: ex.section, dense: false, check: checkAnswer, onTheory, keyBase: ex.id, instruction: ex.instruction }}
        >
          <ExerciseCard ex={ex} />
        </BlockCtx.Provider>
      ))}
    </div>
  )
}

export function ExerciseCard({ ex }: { ex: Exercise }) {
  const progress = useProgress()
  const [nonce, setNonce] = useState(0)
  // Short ids ("1", "1.2", "A") sit inline; long titles from chat must stack or they crush the instruction.
  const isBar = !/^\d+(\.\d+)?[a-z]?$/i.test(ex.id) && ex.id.length > 3

  const own = (k: string) => k === ex.id || k.startsWith(`${ex.id}.`) || k.startsWith(`${ex.id}:`)
  const keys = progress.enabled ? progress.keys().filter(own) : []
  const done = keys.length > 0 && keys.every((k) => progress.get(k)?.correct)
  const touched = keys.some((k) => progress.get(k))

  const reset = () => {
    progress.remove(keys)
    setNonce((n) => n + 1)
  }

  const resetBtn = progress.enabled && (done || touched) && (
    <button
      type="button"
      onClick={reset}
      title="Reset this exercise"
      className="mt-0.5 inline-flex shrink-0 items-center gap-1 rounded-md px-1.5 py-1 text-xs font-medium text-neutral-400 transition-colors hover:bg-neutral-200/60 hover:text-neutral-700"
    >
      <RotateCcw className="h-3.5 w-3.5" />
    </button>
  )

  return (
    <article className={`min-w-0 rounded-2xl border p-4 shadow-sm sm:p-5 ${done ? 'border-emerald-200 bg-emerald-50/40' : 'border-neutral-200/90 bg-neutral-50/50'}`}>
      <header className={`mb-3.5 ${isBar ? 'space-y-2' : 'flex items-start gap-3'}`}>
        <div className={isBar ? 'flex items-start gap-2' : 'contents'}>
          <span
            className={`shrink-0 rounded-lg px-1.5 text-xs font-bold text-white shadow-sm ${done ? 'bg-emerald-500' : BADGE} ${
              isBar
                ? 'inline-flex max-w-full break-words px-2.5 py-1 leading-snug'
                : 'mt-0.5 grid h-6 min-w-6 place-items-center'
            }`}
          >
            {done ? <Check className="h-3.5 w-3.5" /> : ex.id}
          </span>
          {isBar && resetBtn}
        </div>
        {ex.instruction && (
          <p className="min-w-0 flex-1 text-[15px] font-semibold leading-snug text-neutral-800 break-words [overflow-wrap:anywhere]">
            <Inline text={ex.instruction} />
          </p>
        )}
        {!isBar && resetBtn}
      </header>
      <div key={nonce} className={`min-w-0 space-y-3 [display:flow-root] ${PROSE_CLS}`}>
        <RenderNodes nodes={ex.nodes} />
      </div>
    </article>
  )
}

export function Theory({ markdown }: { markdown: string }) {
  const sections = parseTheory(markdown)
  return (
    <div className="@container space-y-4">
      {sections.map((sec, i) => {
        const ordered = [...sec.nodes.filter((n) => n.t === 'image'), ...sec.nodes.filter((n) => n.t !== 'image')]
        const isBar = sec.letter.length > 2
        return (
          <section key={sec.letter || i} className={i > 0 ? 'border-t border-neutral-200/80 pt-4' : ''}>
            <div id={`theory-${sec.letter}`} className={`scroll-mt-6 ${isBar ? 'space-y-2' : 'flex gap-3'}`}>
              {sec.letter && (
                <div
                  className={`inline-flex h-7 shrink-0 items-center justify-center rounded-lg text-xs font-bold text-white shadow-sm ${BADGE} ${
                    isBar ? 'px-2.5' : 'mt-0.5 w-7'
                  }`}
                >
                  {sec.letter}
                </div>
              )}
              <div className="min-w-0 flex-1 space-y-3 [display:flow-root]">
                {sec.title && <h3 className="text-base font-bold leading-snug text-neutral-900">{sec.title}</h3>}
                <TheoryNodes nodes={ordered} />
              </div>
            </div>
          </section>
        )
      })}
    </div>
  )
}

function TheoryNodes({ nodes }: { nodes: Node[] }) {
  const parent = useContext(BlockCtx)
  return (
    <BlockCtx.Provider value={{ ...parent, dense: false }}>
      <RenderNodes nodes={nodes} variant="theory" />
    </BlockCtx.Provider>
  )
}
