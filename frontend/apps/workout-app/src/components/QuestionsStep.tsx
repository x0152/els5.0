import { useState } from 'react'
import { cn } from '@els/ui'
import { Check, X } from 'lucide-react'
import type { Question } from '../lib/types.ts'
import { ContinueButton, StepShell } from './StepShell.tsx'

export function QuestionsStep({ questions, onDone }: { questions: Question[]; onDone: (score: number) => void }) {
  const [picked, setPicked] = useState<Record<number, number>>({})
  const [checked, setChecked] = useState(false)

  const answered = Object.keys(picked).length === questions.length
  const correct = questions.filter((q, i) => picked[i] === q.answer).length
  const score = Math.round((correct / Math.max(questions.length, 1)) * 100)

  return (
    <StepShell>
      <ol className="flex flex-col gap-5">
        {questions.map((q, qi) => (
          <li key={qi}>
            <p className="mb-2 font-medium text-neutral-800">
              {qi + 1}. {q.text}
            </p>
            <div className="grid gap-1.5 sm:grid-cols-2">
              {q.options.map((opt, oi) => {
                const isPicked = picked[qi] === oi
                const isAnswer = q.answer === oi
                return (
                  <button
                    key={oi}
                    type="button"
                    disabled={checked}
                    onClick={() => setPicked((p) => ({ ...p, [qi]: oi }))}
                    className={cn(
                      'flex items-center gap-2 rounded-xl border px-3 py-2 text-left text-sm transition-colors',
                      !checked && (isPicked ? 'border-brand-500 bg-brand-50 text-brand-900' : 'border-neutral-200 bg-white hover:bg-neutral-50'),
                      checked && isAnswer && 'border-emerald-400 bg-emerald-50 text-emerald-900',
                      checked && isPicked && !isAnswer && 'border-rose-300 bg-rose-50 text-rose-900',
                      checked && !isPicked && !isAnswer && 'border-neutral-200 bg-white text-neutral-400',
                    )}
                  >
                    {checked && isAnswer && <Check className="h-4 w-4 shrink-0 text-emerald-600" />}
                    {checked && isPicked && !isAnswer && <X className="h-4 w-4 shrink-0 text-rose-500" />}
                    {opt}
                  </button>
                )
              })}
            </div>
          </li>
        ))}
      </ol>

      {!checked ? (
        <ContinueButton onClick={() => setChecked(true)} label="Check answers" disabled={!answered} />
      ) : (
        <>
          <p className={cn('text-sm font-semibold', score >= 80 ? 'text-emerald-600' : 'text-amber-600')}>
            {correct} of {questions.length} correct
          </p>
          <ContinueButton onClick={() => onDone(score)} />
        </>
      )}
    </StepShell>
  )
}
