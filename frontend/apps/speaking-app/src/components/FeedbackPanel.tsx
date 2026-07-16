import { Sparkles } from 'lucide-react'
import type { Feedback } from '../lib/types.ts'

export function FeedbackPanel({ feedback }: { feedback: Feedback }) {
  return (
    <div className="rounded-2xl border border-brand-200 bg-brand-50/50 p-5">
      <h3 className="mb-2 flex items-center gap-2 font-semibold text-neutral-900">
        <Sparkles className="h-4 w-4 text-brand-600" />
        Coach feedback
      </h3>
      <p className="text-sm text-neutral-700">{feedback.summary}</p>
      {(feedback.tips ?? []).length > 0 && (
        <ul className="mt-4 space-y-3">
          {(feedback.tips ?? []).map((tip, i) => (
            <li key={i} className="rounded-xl bg-white p-3.5 ring-1 ring-neutral-200">
              <span className="inline-block max-w-full break-words rounded-lg bg-brand-100 px-2.5 py-1 font-mono text-sm font-semibold text-brand-700">
                {tip.sound}
              </span>
              <p className="mt-2 text-sm leading-relaxed text-neutral-700">{tip.advice}</p>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
