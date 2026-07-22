import { useContext, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { BookOpen, Check, Eye, Lightbulb, Loader2, TriangleAlert, X } from 'lucide-react'
import { parseGap, gapPrompt } from '../parse.ts'
import { checkLocal, type CheckResult } from '../check.ts'
import { useProgress, useProduce, useFill } from '../state.ts'
import { DeferredCtx } from '../Deferred.tsx'
import { Inline } from '../markdown.tsx'
import { BlockCtx } from './context.ts'

export function RevealCorrection({ correction, label = 'Correct' }: { correction: string; label?: string }) {
  const [show, setShow] = useState(false)
  useEffect(() => setShow(false), [correction])
  if (!show) {
    return (
      <button
        type="button"
        onClick={() => setShow(true)}
        className="mt-1.5 inline-flex items-center gap-1 rounded-md bg-brand-50 px-2 py-1 font-medium text-brand-700 transition-colors hover:bg-brand-100"
      >
        <Eye className="h-3.5 w-3.5" /> Show answer
      </button>
    )
  }
  return (
    <span className="mt-1.5 block select-none" onCopy={(e) => e.preventDefault()} onCut={(e) => e.preventDefault()} onContextMenu={(e) => e.preventDefault()}>
      <span className="text-neutral-500">{label}: </span>
      <span className="font-semibold text-brand-700">{correction}</span>
    </span>
  )
}

export function GapLines({ md }: { md: string }) {
  const lines = md.split('\n').map((l) => l.trimEnd())
  const numbered = lines.some((l) => /^\d+\.\s/.test(l.trim()))
  return (
    <div className="space-y-2.5">
      {lines.map((raw, i) => {
        const ln = raw.trim()
        if (!ln) return null
        const heading = /^(#{1,4})\s+(.+)$/.exec(ln)
        if (heading) {
          const level = heading[1]!.length
          const label = heading[2]!
          const cls =
            level <= 2
              ? 'mt-4 text-base font-bold text-neutral-900 first:mt-0'
              : 'mt-3 text-sm font-semibold text-neutral-800 first:mt-0'
          return (
            <div key={i} className={cls}>
              <Inline text={label} />
            </div>
          )
        }
        if (!/\{\{/.test(ln) && !/^\d+\.\s/.test(ln)) {
          return (
            <p key={i} className="text-sm leading-relaxed text-neutral-600">
              <Inline text={ln} />
            </p>
          )
        }
        const m = /^(\d+)\.\s+(.*)$/.exec(ln)
        return (
          <div key={i} className="flex gap-2.5 text-sm leading-[1.75] text-neutral-800">
            {numbered && (
              <span className="w-5 shrink-0 select-none text-right text-sm font-semibold tabular-nums text-brand-500/80">{m?.[1]}</span>
            )}
            <div className="min-w-0 flex-1">
              <ItemText text={m ? m[2] ?? '' : ln} lineIdx={i} />
            </div>
          </div>
        )
      })}
    </div>
  )
}

export function ItemText({ text, lineIdx }: { text: string; lineIdx?: number }) {
  const { keyBase } = useContext(BlockCtx)
  // Drop bold/italic markers wrapped around a gap (**{{...}}**) — they would render as literal `**`.
  const clean = text.replace(/(\*\*|__|\*|_)(\{\{[^}]*\}\})\1/g, '$2')
  const prompt = gapPrompt(clean)
  const parts = clean.split(/(\{\{[^}]*\}\})/g)
  return (
    <>
      {parts.map((p, i) => {
        const m = /^\{\{([^}]*)\}\}$/.exec(p)
        if (m) {
          const gapKey = keyBase !== undefined && lineIdx !== undefined ? `${keyBase}:${lineIdx}:${i}` : undefined
          return <Gap key={i} gap={parseGap(m[1] ?? '')} prompt={prompt} gapKey={gapKey} />
        }
        return <Inline key={i} text={p} />
      })}
    </>
  )
}

function Gap({ gap, prompt, narrow, gapKey }: { gap: ReturnType<typeof parseGap>; prompt: string; narrow?: boolean; gapKey?: string }) {
  const { section, dense, check: checkFn, onTheory } = useContext(BlockCtx)
  const deferred = useContext(DeferredCtx)
  const progress = useProgress()
  const produce = useProduce()
  const onFill = useFill()
  const [filled, setFilled] = useState(!!gap.fill)
  const produced = useRef(false)
  const persist = !!gapKey && progress.enabled
  const saved = persist ? progress.get(gapKey!) : undefined
  const [answer, setAnswer] = useState(saved?.answer ?? gap.fill ?? '')
  const [hints, setHints] = useState(false)
  const [dismissed, setDismissed] = useState(!!saved)
  const rootRef = useRef<HTMLSpanElement>(null)

  useEffect(() => {
    if (!hints) return
    const onDown = (e: PointerEvent) => {
      if (!rootRef.current?.contains(e.target as Element)) setHints(false)
    }
    document.addEventListener('pointerdown', onDown)
    return () => document.removeEventListener('pointerdown', onDown)
  }, [hints])
  const [result, setResult] = useState<CheckResult | undefined>(
    saved ? { correct: saved.correct, correction: saved.correction ?? '', explanation: saved.explanation ?? '' } : undefined,
  )
  const submitted = useRef(saved?.answer ?? '')
  const check = useMutation({
    mutationFn: checkFn,
    onSuccess: (r) => {
      setResult(r)
      setDismissed(false)
      if (persist) progress.set(gapKey!, { answer: submitted.current, correct: r.correct, correction: r.correction, explanation: r.explanation })
      if (produce && !produced.current && prompt.split('___').length === 2) {
        produced.current = true
        produce({ skill: 'writing', text: prompt.replace('___', submitted.current), context: prompt })
      }
    },
  })

  useEffect(() => {
    if (persist) progress.register(gapKey!)
  }, [persist, gapKey, progress])

  // Deferred mode: verify the collected answer locally when the sheet-level Check fires.
  const checkedRound = useRef(0)
  const answerRef = useRef(answer)
  useEffect(() => {
    answerRef.current = answer
  }, [answer])
  const alreadyCorrect = !!result?.correct
  useEffect(() => {
    if (!deferred || deferred.round === 0 || deferred.round === checkedRound.current) return
    checkedRound.current = deferred.round
    const key = gapKey ?? prompt
    if (alreadyCorrect) {
      deferred.report(key, { prompt, answer: answerRef.current, expected: gap.answers[0] ?? '', correct: true })
      return
    }
    const text = answerRef.current.trim()
    const ok = text !== '' && checkLocal(gap.answers, text)
    submitted.current = text
    setResult({ correct: ok, correction: ok ? '' : gap.answers[0] ?? '', explanation: '' })
    setDismissed(false)
    if (persist) progress.set(gapKey!, { answer: text, correct: ok, correction: ok ? '' : gap.answers[0] ?? '' })
    deferred.report(key, { prompt, answer: text, expected: gap.answers[0] ?? '', correct: ok })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [deferred?.round])

  const pending = check.isPending
  const done = !pending && !!result
  const ok = done && result!.correct

  // Deferred batch-check opens many feedbacks at once — keep them in flow, not absolute popovers.
  const inline = dense || !!deferred
  const popRef = useRef<HTMLSpanElement>(null)
  const [shift, setShift] = useState(0)
  const popupVisible = (gap.type === 'choice' && !ok && hints) || (done && !ok && !dismissed)
  useLayoutEffect(() => {
    if (!popupVisible || inline) {
      setShift(0)
      return
    }
    const r = popRef.current?.getBoundingClientRect()
    if (!r) return
    const over = r.right - window.innerWidth + 12
    if (over > 0) setShift(-Math.min(over, Math.max(r.left - 12, 0)))
  }, [popupVisible, inline])

  function submit(value: string) {
    const text = value.trim()
    if (!text || pending || deferred) return
    if (onFill) {
      if (gap.ordinal !== undefined && text !== gap.fill) onFill({ ordinal: gap.ordinal, answer: text })
      setFilled(true)
      return
    }
    submitted.current = text
    check.mutate({ prompt, answers: gap.answers, answer: text })
  }

  const len = Math.max(answer.length, narrow ? 2 : 8)
  const width = narrow ? '2.5rem' : `${Math.min(len + 2, 40)}ch`
  const center = !inline

  const stateCls = onFill
    ? filled
      ? 'border-brand-500 text-brand-700'
      : 'border-neutral-400 text-brand-700 focus:border-brand-500'
    : ok
      ? 'border-emerald-500 text-emerald-600'
      : done
        ? 'border-rose-400 text-rose-500'
        : 'border-neutral-400 text-brand-700 focus:border-brand-500'

  const inputRow = (
    <span className="inline-flex items-center gap-1">
        <input
          value={answer}
          disabled={ok || pending}
          onChange={(e) => {
            setAnswer(e.target.value)
            if (onFill) setFilled(false)
            if (deferred && done && !ok) setResult(undefined)
          }}
        onBlur={() => {
          if (onFill && !filled) submit(answer)
        }}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            submit(answer)
          }
        }}
        style={dense ? undefined : { width }}
        className={`max-w-full border-0 border-b-2 bg-transparent px-1 pb-0.5 text-sm font-medium outline-none transition-colors disabled:opacity-100 ${
          dense ? 'w-full text-left' : center ? 'text-center' : 'text-left'
        } ${stateCls}`}
      />
      {pending && <Loader2 className="h-4 w-4 shrink-0 animate-spin text-neutral-400" />}
      {ok && <Check className="h-4 w-4 shrink-0 text-emerald-500" />}
      {onFill && filled && <Check className="h-4 w-4 shrink-0 text-brand-500" />}
      {gap.type === 'choice' && !ok && !pending && (
        <button
          type="button"
          title="Hint"
          onClick={() => setHints((v) => !v)}
          className={`shrink-0 rounded-full p-0.5 transition-colors ${hints ? 'text-brand-600' : 'text-neutral-300 hover:text-brand-500'}`}
        >
          <Lightbulb className="h-4 w-4" />
        </button>
      )}
    </span>
  )

  const chips = gap.type === 'choice' && !ok && hints && (
    <span className="flex w-full max-w-sm flex-wrap gap-1.5 rounded-xl border border-neutral-200 bg-white p-2.5 shadow-lg ring-1 ring-black/5">
      {gap.choices.map((c) => (
        <button
          key={c}
          onClick={() => {
            setAnswer(c)
            setHints(false)
            if (deferred) {
              if (done && !ok) setResult(undefined)
              return
            }
            submit(c)
          }}
          className="rounded-full border border-neutral-200 bg-neutral-50 px-2 py-0.5 text-xs text-neutral-600 transition-colors hover:border-brand-300 hover:bg-brand-50 hover:text-brand-700"
        >
          {c}
        </button>
      ))}
    </span>
  )

  const feedback = done && !ok && !dismissed && (
    deferred ? (
      <span className="block text-xs">
        {result!.correction && <RevealCorrection correction={result!.correction} />}
      </span>
    ) : (
      <span className="block w-full max-w-sm rounded-xl border border-rose-200 bg-white p-3 text-xs shadow-lg ring-1 ring-black/5">
        <span className="flex items-center gap-1.5 font-semibold text-rose-600">
          <TriangleAlert className="h-3.5 w-3.5" /> Not quite
          <button
            type="button"
            onClick={() => setDismissed(true)}
            className="ml-auto grid h-5 w-5 place-items-center rounded-full text-neutral-300 transition-colors hover:bg-neutral-100 hover:text-neutral-600"
          >
            <X className="h-3.5 w-3.5" />
          </button>
        </span>
        {result!.correction && <RevealCorrection correction={result!.correction} />}
        {result!.explanation && <span className="mt-1 block leading-relaxed text-neutral-600">{result!.explanation}</span>}
        {section && (
          <button
            onClick={() => onTheory(section)}
            className="mt-2 inline-flex items-center gap-1 rounded-md bg-brand-50 px-2 py-1 font-medium text-brand-700 transition-colors hover:bg-brand-100"
          >
            <BookOpen className="h-3.5 w-3.5" /> Review section {section}
          </button>
        )}
      </span>
    )
  )

  const popupOpen = !!(chips || feedback)
  return (
    <span
      ref={rootRef}
      className={`relative align-baseline ${popupOpen ? 'z-40' : 'z-10'} ${
        inline ? 'inline-flex max-w-full flex-col items-stretch gap-1' : 'inline-flex max-w-full flex-wrap items-baseline gap-x-1'
      }`}
    >
      {inputRow}
      {popupOpen && (
        <span
          ref={popRef}
          style={inline ? undefined : { transform: `translateX(${shift}px)` }}
          className={`${
            inline ? 'relative mt-0.5 w-full max-w-sm' : 'absolute left-0 top-full z-30 mt-1.5 w-72 max-w-[calc(100vw-1.5rem)]'
          } flex flex-col items-start gap-1.5`}
        >
          {chips}
          {feedback}
        </span>
      )}
      {check.isError && <span className="absolute left-0 top-full mt-1 text-xs text-rose-600">Check failed.</span>}
    </span>
  )
}
