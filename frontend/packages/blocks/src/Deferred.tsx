import { createContext, useMemo, useRef, useState } from 'react'
import { CheckCircle2, ListChecks, RotateCcw } from 'lucide-react'
import { parseBlocks } from './parse.ts'
import { mockCheckAnswer } from './check.ts'
import { BlocksProvider } from './Blocks.tsx'
import { useProgress } from './state.ts'
import { PROSE_CLS } from './markdown.tsx'
import { BlockCtx } from './render/context.ts'
import { RenderNodes } from './render/nodes.tsx'

export type DeferredResult = { prompt: string; answer: string; expected: string; correct: boolean }

// Deferred check mode: gaps collect answers silently and only verify them when the
// learner presses Check. First-attempt results are what gets reported (retries are
// for learning, not for the stats).
export const DeferredCtx = createContext<{ round: number; report: (key: string, r: DeferredResult) => void } | null>(null)

export function DeferredBlocks({
  md,
  onContinue,
  continueLabel = 'Continue',
}: {
  md: string
  onContinue: (firstAttempt: DeferredResult[]) => void
  continueLabel?: string
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

  return (
    <BlocksProvider>
      <DeferredCtx.Provider value={ctxValue}>
        <BlockCtx.Provider value={{ dense: false, check: mockCheckAnswer, onTheory: () => {}, keyBase: 'b' }}>
          <div className={`space-y-3 [display:flow-root] ${PROSE_CLS}`}>
            <RenderNodes nodes={parseBlocks(md)} />
          </div>
        </BlockCtx.Provider>
        <DeferredControls
          round={round}
          onCheck={() => setRound((r) => r + 1)}
          wrongCount={() => Object.values(latest.current).filter((r) => !r.correct).length}
          onContinue={() => onContinue(Object.values(first.current))}
          continueLabel={continueLabel}
        />
      </DeferredCtx.Provider>
    </BlocksProvider>
  )
}

function DeferredControls({
  round,
  onCheck,
  wrongCount,
  onContinue,
  continueLabel,
}: {
  round: number
  onCheck: () => void
  wrongCount: () => number
  onContinue: () => void
  continueLabel: string
}) {
  const progress = useProgress()
  const total = progress.keys().length
  const wrong = wrongCount()
  const checked = round > 0

  return (
    <div className="mt-4 flex flex-wrap items-center gap-2 border-t border-neutral-100 pt-3.5">
      {!checked ? (
        <button
          type="button"
          onClick={onCheck}
          disabled={total === 0}
          className="inline-flex items-center gap-1.5 rounded-xl bg-brand-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-700 disabled:opacity-50"
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
