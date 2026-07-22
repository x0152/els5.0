import { Link } from 'react-router-dom'
import { Check, ChevronRight, Rocket } from 'lucide-react'
import { cn } from '@els/ui'
import { Widget } from './Widget'
import { useOnboardingProgress } from '../store/onboarding'

const STEPS: Record<string, { title: string; to?: string }> = {
  first_film: { title: 'Watch a film with subtitles', to: '/v1/films' },
  first_quest: { title: 'Generate your own quest and complete it', to: '/v1/quest' },
  first_article: { title: 'Add any page by URL and read it in Reader', to: '/v1/reader' },
  first_workout: { title: 'Complete your first workout', to: '/v1/workout' },
  first_chat: { title: 'Ask the assistant anything' },
  first_words: { title: 'Add 5 words to your vocabulary', to: '/v1/vocab' },
  first_chapter: { title: 'Finish a unit in Grammarbook', to: '/v1/grammarbook' },
}

export function GettingStarted() {
  const progressQ = useOnboardingProgress()

  const items = (progressQ.data ?? []).flatMap((i) => {
    const step = STEPS[i.id]
    return i.kind === 'checklist' && step ? [{ ...i, step }] : []
  })
  if (items.length === 0 || items.every((i) => i.done)) return null

  const doneCount = items.filter((i) => i.done).length

  return (
    <Widget
      title="Getting started"
      icon={<Rocket size={16} />}
      action={
        <span className="flex items-center gap-3">
          <span className="text-xs font-medium text-neutral-500">
            {doneCount}/{items.length}
          </span>
          <Link
            to="/v1/profile/achievements"
            className="text-xs font-medium text-brand-600 hover:text-brand-700"
          >
            All achievements →
          </Link>
        </span>
      }
    >
      <ul className="divide-y divide-neutral-100">
          {items.map((item) => {
            const step = item.step
            const row = (
              <>
                <span
                  className={cn(
                    'grid h-6 w-6 shrink-0 place-items-center rounded-full ring-1',
                    item.done
                      ? 'bg-emerald-50 text-emerald-600 ring-emerald-200'
                      : 'bg-white text-neutral-300 ring-neutral-200',
                  )}
                >
                  <Check size={14} />
                </span>
                <span
                  className={cn(
                    'flex-1 text-sm',
                    item.done ? 'text-neutral-400 line-through' : 'text-neutral-800',
                  )}
                >
                  {step.title}
                </span>
                {!item.done && item.threshold > 1 && (
                  <span className="text-xs text-neutral-400">
                    {Math.min(item.value, item.threshold)}/{item.threshold}
                  </span>
                )}
                {!item.done && <ChevronRight size={16} className="text-neutral-300" />}
              </>
            )
            const rowCls = 'flex w-full items-center gap-3 px-5 py-3 text-left hover:bg-neutral-50'
            return (
              <li key={item.id}>
                {item.done ? (
                  <div className="flex items-center gap-3 px-5 py-3">{row}</div>
                ) : step.to ? (
                  <Link to={step.to} className={rowCls}>
                    {row}
                  </Link>
                ) : (
                  <button
                    type="button"
                    className={rowCls}
                    onClick={() => document.dispatchEvent(new CustomEvent('els:ask'))}
                  >
                    {row}
                  </button>
                )}
              </li>
            )
          })}
      </ul>
    </Widget>
  )
}
