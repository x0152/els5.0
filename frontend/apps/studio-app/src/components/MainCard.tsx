import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CheckCircle2, Clock, Loader2, Quote, Shuffle, Turtle } from 'lucide-react'
import { SpeakButton, cn } from '@els/ui'
import { api } from '../lib/api.ts'
import { isDue, type Item } from '../lib/types.ts'

export function MainCard({ item, hidden }: { item: Item; hidden: boolean }) {
  const queryClient = useQueryClient()
  const done = [item.listened, item.spoken, item.written, item.recalled].filter(Boolean).length

  const meQ = useQuery({
    queryKey: ['studio', 'me'],
    queryFn: () => api.account.accountMe(),
    staleTime: 10_000,
  })
  const showTranslations = meQ.data?.show_translations ?? true

  const exampleM = useMutation({
    mutationFn: () => api.studio.studioRegenExample({ params: { path: { id: item.id } } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['studio', 'items'] }),
  })

  return (
    <div className="relative flex flex-[3] flex-col overflow-hidden rounded-2xl border border-brand-200 bg-gradient-to-br from-brand-50 to-white p-5 shadow-sm lg:p-7">
      <Quote className="absolute -right-2 -top-2 h-24 w-24 -scale-x-100 text-brand-100" />
      <div className="flex items-center gap-2">
        <p className="text-xs font-semibold uppercase tracking-wide text-brand-700">Now studying</p>
        {isDue(item) && (
          <span className="flex items-center gap-1 rounded-full bg-amber-50 px-2 py-0.5 text-xs font-semibold text-amber-700 ring-1 ring-amber-200">
            <Clock className="h-3 w-3" /> review due
          </span>
        )}
      </div>
      <div
        className={cn(
          'flex min-h-0 flex-1 flex-col justify-center overflow-y-auto py-4 lg:py-0',
          hidden && 'pointer-events-none select-none blur-md',
        )}
      >
        <p className="relative text-2xl font-bold leading-tight text-neutral-900 lg:text-3xl">{item.text}</p>
        {item.transcription && (
          <p className="mt-2.5 font-mono text-sm text-neutral-500">/{item.transcription}/</p>
        )}
        {showTranslations && item.translation && (
          <p className="mt-1 text-sm text-neutral-600">{item.translation}</p>
        )}
        {item.explanation && (
          <div className="mt-3 max-w-lg text-sm leading-relaxed">
            <p className="text-neutral-700">{item.explanation}</p>
            {showTranslations && item.explanation_native && (
              <p className="mt-0.5 text-neutral-400">{item.explanation_native}</p>
            )}
          </div>
        )}
        {item.example && (
          <div className="mt-4 flex max-w-lg items-start gap-2 rounded-xl bg-white/70 px-4 py-2.5 ring-1 ring-brand-100">
            <p className="flex-1 text-sm italic leading-relaxed text-neutral-600">&ldquo;{item.example}&rdquo;</p>
            <SpeakButton
              title="Listen to the example"
              className="shrink-0 hover:bg-brand-50"
              text={item.example}
            />
          </div>
        )}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <SpeakButton variant="button" text={item.text}>
          Listen
        </SpeakButton>
        <SpeakButton
          variant="ghost"
          className="hover:bg-white/70"
          icon={<Turtle className="h-4 w-4" />}
          text={item.text}
          rate={0.7}
        >
          Slow
        </SpeakButton>
        <button
          onClick={() => exampleM.mutate()}
          disabled={exampleM.isPending}
          className="inline-flex h-9 items-center gap-2 rounded-lg px-3 text-sm font-semibold text-neutral-700 hover:bg-white/70 disabled:opacity-50"
        >
          {exampleM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Shuffle className="h-4 w-4" />}
          Another example
        </button>
        <span className="ml-auto flex items-center gap-1.5 text-xs text-neutral-400">
          <CheckCircle2 className={done === 4 ? 'h-4 w-4 text-emerald-500' : 'h-4 w-4 text-neutral-300'} />
          {done} of 4 done
        </span>
      </div>
    </div>
  )
}
