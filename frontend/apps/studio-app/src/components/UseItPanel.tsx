import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Check, Loader2, PenLine, RefreshCw, Sparkles } from 'lucide-react'
import { cn } from '@els/ui'
import { api } from '../lib/api.ts'
import type { Item } from '../lib/types.ts'

export function UseItPanel({ item, onDone }: { item: Item; onDone: () => void }) {
  const queryClient = useQueryClient()
  const [reply, setReply] = useState('')

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['studio'] })

  const taskM = useMutation({
    mutationFn: () => api.studio.studioRegenTask({ params: { path: { id: item.id } } }),
    onSuccess: () => {
      checkM.reset()
      invalidate()
    },
  })

  const checkM = useMutation({
    mutationFn: () =>
      api.studio.studioCheckReply({ params: { path: { id: item.id } }, body: { reply: reply.trim() } }),
    onSuccess: (data) => {
      if (data?.ok) {
        invalidate()
        onDone()
      }
    },
  })

  return (
    <div className="flex flex-1 flex-col rounded-2xl border border-brand-300 bg-white shadow-sm ring-1 ring-brand-200">
      <div className="flex items-center justify-between border-b border-neutral-100 px-4 py-3">
        <span className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-brand-50 text-brand-600">
            <PenLine className="h-4 w-4" />
          </span>
          Use it
        </span>
        {item.written && (
          <span className="flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-0.5 text-xs font-semibold text-emerald-700 ring-1 ring-emerald-200">
            <Check className="h-3 w-3" /> done
          </span>
        )}
      </div>
      <div className="flex flex-col gap-2.5 p-4 lg:flex-row">
        <div className="flex min-w-0 flex-col gap-2.5 lg:w-1/2">
          {item.task ? (
            <div className="flex flex-1 items-start gap-1.5 rounded-xl bg-neutral-50 px-3 py-2">
              <p className="flex-1 text-sm italic leading-relaxed text-neutral-600">{item.task}</p>
              <button
                onClick={() => taskM.mutate()}
                disabled={taskM.isPending}
                title="New situation"
                className="shrink-0 rounded-md p-1 text-neutral-400 hover:bg-neutral-100 hover:text-brand-600 disabled:opacity-50"
              >
                <RefreshCw className={cn('h-3.5 w-3.5', taskM.isPending && 'animate-spin')} />
              </button>
            </div>
          ) : (
            <button
              onClick={() => taskM.mutate()}
              disabled={taskM.isPending}
              className="inline-flex h-9 items-center justify-center gap-2 self-start rounded-lg bg-white px-4 text-sm font-semibold text-neutral-800 shadow-sm ring-1 ring-inset ring-neutral-200 hover:bg-neutral-50 disabled:opacity-50"
            >
              {taskM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4 text-brand-600" />}
              Generate a situation
            </button>
          )}
          {checkM.data && (
            <p
              className={cn(
                'rounded-xl px-3 py-2 text-xs leading-relaxed ring-1',
                checkM.data.ok
                  ? 'bg-emerald-50 text-emerald-800 ring-emerald-200'
                  : 'bg-amber-50 text-amber-800 ring-amber-200',
              )}
            >
              {checkM.data.comment}
            </p>
          )}
        </div>
        <div className="flex min-w-0 flex-1 flex-col gap-2.5">
          <textarea
            value={reply}
            onChange={(e) => setReply(e.target.value)}
            placeholder={item.task ? 'Your reply to the situation…' : 'Generate a situation first, then reply here…'}
            className="min-h-20 w-full flex-1 resize-none rounded-lg border border-neutral-200 px-3 py-2 text-sm placeholder:text-neutral-400 focus:border-brand-400 focus:outline-none focus:ring-2 focus:ring-brand-100"
          />
          <button
            onClick={() => checkM.mutate()}
            disabled={!item.task || !reply.trim() || checkM.isPending}
            className="inline-flex h-9 items-center justify-center gap-2 self-end rounded-lg bg-brand-600 px-4 text-sm font-semibold text-white shadow-sm shadow-brand-600/25 hover:bg-brand-700 disabled:opacity-50"
          >
            {checkM.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Check
          </button>
        </div>
      </div>
    </div>
  )
}
