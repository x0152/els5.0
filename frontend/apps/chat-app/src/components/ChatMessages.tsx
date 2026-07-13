import { useEffect, useRef } from 'react'
import { Bot, RefreshCw, User } from 'lucide-react'
import { Avatar, cn } from '@els/ui'
import type { ChatItem } from '../lib/chat'
import { useMe } from '../hooks/useMe'
import { Markdown } from './Markdown'
import { StepBadge } from './StepBadge'

const SUGGESTIONS = [
  'What mistakes have I made recently?',
  'Suggest what to study next',
  'Quiz me on my weak words',
]

function UserAvatar() {
  const me = useMe()
  return (
    <Avatar
      src={me?.pictureUrl}
      initials={me?.initials}
      icon={<User size={14} />}
      className="mt-0.5 h-7 w-7 shrink-0 bg-brand-100 text-[11px] ring-brand-200"
    />
  )
}

function Separator() {
  return (
    <div className="flex items-center gap-3 py-1 text-[11px] text-neutral-400">
      <span className="h-px flex-1 bg-neutral-200" />
      <span>Context reset</span>
      <span className="h-px flex-1 bg-neutral-200" />
    </div>
  )
}

function TypingDots() {
  return (
    <span className="inline-flex items-center gap-1">
      {[0, 120, 240].map((d) => (
        <span
          key={d}
          className="inline-block h-1.5 w-1.5 animate-bounce rounded-full bg-neutral-400"
          style={{ animationDelay: `${d}ms` }}
        />
      ))}
    </span>
  )
}

function EmptyState({ onPick }: { onPick?: (text: string) => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 p-8 text-center">
      <span className="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-lg shadow-brand-600/20">
        <Bot className="h-7 w-7" />
      </span>
      <div className="space-y-1">
        <p className="text-base font-semibold text-neutral-800">How can I help?</p>
        <p className="max-w-sm text-sm text-neutral-400">Ask the assistant about your progress, mistakes, or what to study next.</p>
      </div>
      {onPick && (
        <div className="flex max-w-md flex-wrap justify-center gap-2 pt-1">
          {SUGGESTIONS.map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => onPick(s)}
              className="rounded-full border border-neutral-200 bg-white px-3 py-1.5 text-xs font-medium text-neutral-600 shadow-sm transition-colors hover:border-brand-300 hover:bg-brand-50 hover:text-brand-700"
            >
              {s}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

export function ChatMessages({
  items,
  streaming,
  onRegenerate,
  onPickSuggestion,
  variant = 'panel',
}: {
  items: ChatItem[]
  streaming: boolean
  onRegenerate: () => void
  onPickSuggestion?: (text: string) => void
  variant?: 'panel' | 'page'
}) {
  const endRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [items, streaming])

  const lastAssistantId = [...items].reverse().find((m) => m.kind === 'assistant')?.id
  const page = variant === 'page'

  if (items.length === 0) {
    return <EmptyState onPick={onPickSuggestion} />
  }

  return (
    <div className="flex-1 overflow-y-auto">
      <div className={cn('mx-auto flex flex-col gap-4 px-4 py-4', page && 'max-w-3xl px-6 py-8 gap-6')}>
        {items.map((item) => {
          if (item.kind === 'separator') return <Separator key={item.id} />
          if (item.kind === 'user') {
            return (
              <div key={item.id} className="flex flex-row-reverse gap-2.5">
                <UserAvatar />
                <div className="max-w-[calc(100%-2.75rem)] rounded-2xl rounded-tr-sm border border-brand-100 bg-brand-50 px-3.5 py-2 text-sm text-neutral-900 whitespace-pre-wrap break-words">
                  {item.content}
                </div>
              </div>
            )
          }
          const lastSeg = item.segments[item.segments.length - 1]
          const running = item.segments.some((seg) => seg.steps.some((st) => !st.done))
          const waiting =
            streaming && item.id === lastAssistantId && !running && (!lastSeg || lastSeg.steps.length > 0)
          const canRegenerate = item.id === lastAssistantId && !item.pending
          return (
            <div key={item.id} className="group flex gap-2.5">
              <div className="mt-0.5 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-white text-brand-600 ring-1 ring-neutral-200">
                <Bot size={14} />
              </div>
              <div className="flex min-w-0 max-w-[calc(100%-2.75rem)] flex-1 flex-col gap-1">
                <div className="flex flex-col gap-2 rounded-2xl rounded-tl-sm border border-neutral-200 bg-white px-3.5 py-2.5 shadow-sm">
                  {item.segments.map((seg) => (
                    <div key={seg.id} className="flex flex-col gap-1.5">
                      {seg.text && <Markdown text={seg.text} />}
                      {seg.steps.map((s) => (
                        <StepBadge key={s.id} step={s} />
                      ))}
                    </div>
                  ))}
                  {waiting && <TypingDots />}
                </div>
                {canRegenerate && (
                  <button
                    type="button"
                    onClick={onRegenerate}
                    disabled={streaming}
                    title="Regenerate response"
                    className="inline-flex w-fit items-center gap-1 rounded px-1.5 py-0.5 text-[11px] text-neutral-500 opacity-0 transition-opacity hover:bg-neutral-100 hover:text-neutral-800 group-hover:opacity-100 disabled:opacity-40"
                  >
                    <RefreshCw size={11} /> Regenerate
                  </button>
                )}
              </div>
            </div>
          )
        })}
        <div ref={endRef} />
      </div>
    </div>
  )
}
