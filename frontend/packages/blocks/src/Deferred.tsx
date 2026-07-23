import { createContext, useMemo, useRef, useState, type ReactNode } from 'react'
import { CheckCircle2, ListChecks, RotateCcw } from 'lucide-react'
import { parseBlocks } from './parse.ts'
import { mockCheckAnswer } from './check.ts'
import { BlocksProvider, type BlocksAdapters } from './Blocks.tsx'
import { PROSE_CLS } from './markdown.tsx'
import { BlockCtx } from './render/context.ts'
import { RenderNodes } from './render/nodes.tsx'

export type DeferredResult = { prompt: string; answer: string; expected: string; correct: boolean }

// Deferred check mode: gaps collect answers silently and only verify them when the
// learner presses Check. First-attempt results are what gets reported (retries are
// for learning, not for the stats).
export const DeferredCtx = createContext<{ round: number; report: (key: string, r: DeferredResult) => void } | null>(null)

// Provides the deferred context and hands the Check/Continue controls to the caller,
// so any layout (plain list, book spread) can place them where they belong.
export function DeferredProvider({
  onContinue,
  continueLabel = 'Continue',
  controlsClassName,
  children,
}: {
  onContinue: (firstAttempt: DeferredResult[]) => void
  continueLabel?: string
  controlsClassName?: string
  children: (controls: ReactNode) => ReactNode
}) {
  const [round, setRound] = useState(0)
  const latest = useRef<Record<string, DeferredResult>>({})
  const first = useRef<Record<string, DeferredResult>>({})
  const [, force] = useState(0)

  const ctxValue = useMemo(
    () => ({
      round,
      report: (key: string, r: DeferredResult) => {
        latest.current[key] = r
        if (!(key in first.current)) first.current[key] = r
        force((n) => n + 1)
      },
    }),
    [round],
  )

  const controls = (
    <DeferredControls
      round={round}
      onCheck={() => setRound((r) => r + 1)}
      wrongCount={() => Object.values(latest.current).filter((r) => !r.correct).length}
      onContinue={() => onContinue(Object.values(first.current))}
      continueLabel={continueLabel}
      className={controlsClassName}
    />
  )

  return <DeferredCtx.Provider value={ctxValue}>{children(controls)}</DeferredCtx.Provider>
}

export function DeferredBlocks({
  md,
  onContinue,
  continueLabel = 'Continue',
  adapters,
}: {
  md: string
  onContinue: (firstAttempt: DeferredResult[]) => void
  continueLabel?: string
  adapters?: BlocksAdapters
}) {
  return (
    <DeferredProvider onContinue={onContinue} continueLabel={continueLabel}>
      {(controls) => (
        <BlocksProvider adapters={adapters}>
          <BlockCtx.Provider value={{ dense: false, check: mockCheckAnswer, onTheory: () => {}, keyBase: 'b' }}>
            <div className={`space-y-3 [display:flow-root] ${PROSE_CLS}`}>
              <RenderNodes nodes={parseBlocks(md)} />
            </div>
          </BlockCtx.Provider>
          {controls}
        </BlocksProvider>
      )}
    </DeferredProvider>
  )
}

function DeferredControls({
  round,
  onCheck,
  wrongCount,
  onContinue,
  continueLabel,
  className = 'mt-4 flex flex-wrap items-center gap-2 border-t border-neutral-100 pt-3.5',
}: {
  round: number
  onCheck: () => void
  wrongCount: () => number
  onContinue: () => void
  continueLabel: string
  className?: string
}) {
  const wrong = wrongCount()
  const checked = round > 0

  return (
    <div className={className}>
      {!checked ? (
        <button
          type="button"
          onClick={onCheck}
          className="ml-auto inline-flex items-center gap-1.5 rounded-xl bg-brand-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-700"
        >
          <ListChecks className="h-4 w-4" /> Check
        </button>
      ) : (
        <>
          <span className={`inline-flex items-center gap-1.5 text-sm font-medium ${wrong === 0 ? 'text-emerald-600' : 'text-rose-500'}`}>
            <CheckCircle2 className="h-4 w-4" />
            {wrong === 0 ? 'All correct!' : `${wrong} to fix`}
          </span>
          {wrong > 0 && (
            <button
              type="button"
              onClick={onCheck}
              className="inline-flex items-center gap-1.5 rounded-xl border border-neutral-200 bg-white px-3.5 py-2 text-sm font-medium text-neutral-700 shadow-sm transition-colors hover:bg-neutral-50"
            >
              <RotateCcw className="h-3.5 w-3.5" /> Check again
            </button>
          )}
          <button
            type="button"
            onClick={onContinue}
            className="ml-auto inline-flex items-center gap-1.5 rounded-xl bg-brand-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-700"
          >
            {continueLabel}
          </button>
        </>
      )}
    </div>
  )
}
