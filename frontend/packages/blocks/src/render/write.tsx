import { useContext, useEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Check, Lightbulb, Loader2, TriangleAlert } from 'lucide-react'
import { useProgress, usePracticeMeta, useCheckFree, type PracticeAnswer } from '../state.ts'
import { Inline } from '../markdown.tsx'
import { ACCENT, BlockCtx } from './context.ts'
import { RevealCorrection } from './gaps.tsx'

function WriteSample({ sample }: { sample?: string }) {
  const [show, setShow] = useState(false)
  const meta = usePracticeMeta()
  const checkFree = useCheckFree()
  return (
    <div>
      <div className="flex items-center justify-between gap-2">
        {meta || checkFree ? <span /> : <span className="text-[11px] italic text-neutral-400">Free answer</span>}
        {sample && (
          <button
            type="button"
            onClick={() => setShow((s) => !s)}
            className={`inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium transition-colors ${
              show ? 'bg-brand-100 text-brand-700' : 'text-neutral-400 hover:bg-brand-50 hover:text-brand-600'
            }`}
          >
            <Lightbulb className="h-3.5 w-3.5" /> Example
          </button>
        )}
      </div>
      {show && sample && (
        <div
          className={`mt-1.5 select-none rounded-xl px-3 py-2 text-sm leading-relaxed text-neutral-700 ${ACCENT}`}
          onCopy={(e) => e.preventDefault()}
          onCut={(e) => e.preventDefault()}
          onContextMenu={(e) => e.preventDefault()}
        >
          <Inline text={sample} />
        </div>
      )}
    </div>
  )
}

function FreeAnswer({ writeKey, n, placeholder, rows = 2 }: { writeKey?: string; n?: number; placeholder?: string; rows?: number }) {
  const { instruction } = useContext(BlockCtx)
  const meta = usePracticeMeta()
  const checkFree = useCheckFree()
  const canCheck = !!meta || !!checkFree
  const progress = useProgress()
  // Without a checker the write can never be marked correct — don't count it towards completion.
  const persist = !!writeKey && progress.enabled && canCheck
  const saved = persist ? progress.get(writeKey!) : undefined
  const [value, setValue] = useState(saved?.answer ?? '')
  const [result, setResult] = useState<PracticeAnswer | undefined>(saved)
  const ref = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (persist) progress.register(writeKey!)
  }, [persist, writeKey, progress])

  const grow = () => {
    const el = ref.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${el.scrollHeight}px`
  }
  useEffect(grow, [])

  const check = useMutation({
    mutationFn: () =>
      checkFree
        ? checkFree({ instruction: instruction ?? '', answer: value.trim() })
        : meta!.api.checkFree({ kind: meta!.kind, number: meta!.number, instruction: instruction ?? '', answer: value.trim() }),
    onSuccess: (r) => {
      const a: PracticeAnswer = { answer: value.trim(), correct: r.correct, correction: r.correction, explanation: r.explanation }
      setResult(a)
      if (persist) progress.set(writeKey!, a)
    },
  })

  const ok = result?.correct
  const stateCls = ok
    ? 'border-emerald-500 focus:border-emerald-500'
    : result
      ? 'border-rose-400 focus:border-rose-400'
      : 'border-neutral-300 focus:border-brand-500'

  return (
    <div className="space-y-1.5">
      <div className="flex items-start gap-2">
        {n != null && <span className="mt-2 w-6 shrink-0 text-right text-sm tabular-nums text-neutral-400">{n}.</span>}
        <textarea
          ref={ref}
          rows={rows}
          value={value}
          onInput={grow}
          onChange={(e) => {
            setValue(e.target.value)
            if (result) setResult(undefined)
          }}
          placeholder={placeholder || 'Write your answer…'}
          className={`min-w-0 flex-1 resize-none overflow-hidden rounded-xl border bg-white px-3 py-2 text-sm leading-relaxed text-neutral-800 outline-none transition focus:ring-2 focus:ring-brand-100 placeholder:text-neutral-400 ${stateCls}`}
        />
        {canCheck && (
          <button
            type="button"
            disabled={!value.trim() || check.isPending}
            onClick={() => check.mutate()}
            title="Check answer"
            className="mt-1 inline-flex h-8 shrink-0 items-center gap-1 rounded-full bg-brand-600 px-3 text-xs font-semibold text-white transition-colors hover:bg-brand-700 disabled:opacity-50"
          >
            {check.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : ok ? <Check className="h-3.5 w-3.5" /> : 'Check'}
          </button>
        )}
      </div>
      {result && !ok && (
        <div className="rounded-xl border border-rose-200 bg-white p-3 text-xs shadow-sm">
          <span className="flex items-center gap-1.5 font-semibold text-rose-600">
            <TriangleAlert className="h-3.5 w-3.5" /> Not quite
          </span>
          {result.correction && <RevealCorrection correction={result.correction} label="Better" />}
          {result.explanation && <span className="mt-1 block leading-relaxed text-neutral-600">{result.explanation}</span>}
        </div>
      )}
    </div>
  )
}

export function WriteArea({ prompt, sample, rows }: { prompt: string; sample?: string; rows: number }) {
  const { keyBase } = useContext(BlockCtx)
  return (
    <div className="space-y-1.5">
      {prompt && (
        <p className="text-sm leading-relaxed text-neutral-700">
          <Inline text={prompt} />
        </p>
      )}
      <FreeAnswer writeKey={keyBase !== undefined ? `${keyBase}:area` : undefined} rows={rows} />
      <WriteSample sample={sample} />
    </div>
  )
}

export function WriteLines({ prompts, sample, lines }: { prompts: string[]; sample?: string; lines: number }) {
  const { keyBase } = useContext(BlockCtx)
  const items = prompts.length ? prompts : Array.from({ length: lines }, () => '')
  return (
    <div className="space-y-1.5">
      <div className="space-y-2">
        {items.map((p, i) => (
          <FreeAnswer key={i} n={i + 1} placeholder={p} rows={1} writeKey={keyBase !== undefined ? `${keyBase}:line:${i}` : undefined} />
        ))}
      </div>
      <WriteSample sample={sample} />
    </div>
  )
}
