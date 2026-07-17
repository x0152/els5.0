import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { Badge, Button, ConfirmDialog, EmptyState, ErrorState, LoadingState } from '@els/ui'
import { BookOpen, ChevronDown, MessageCircleHeart, Trash2 } from 'lucide-react'
import { api } from './lib/api'
import { formatDay, type Entry } from './lib/types'
import { NotesSection } from './components/NotesSection'

function EntryCard({ entry }: { entry: Entry }) {
  const [open, setOpen] = useState(false)
  const corrections = entry.corrections ?? []
  return (
    <div className="rounded-lg border border-neutral-200 bg-white">
      <button onClick={() => setOpen(!open)} className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left">
        <div className="min-w-0">
          <p className="text-sm font-medium capitalize text-neutral-500">{formatDay(entry.date)}</p>
          <p className="mt-0.5 truncate text-neutral-800">{entry.text}</p>
        </div>
        <span className="flex shrink-0 items-center gap-2">
          {corrections.length > 0 && <Badge>{corrections.length}</Badge>}
          <ChevronDown className={`h-4 w-4 text-neutral-400 transition-transform ${open ? 'rotate-180' : ''}`} />
        </span>
      </button>
      {open && (
        <div className="flex flex-col gap-3 border-t border-neutral-100 px-4 py-4">
          {entry.question && <p className="text-sm italic text-neutral-500">{entry.question}</p>}
          <p className="whitespace-pre-wrap leading-relaxed text-neutral-800">{entry.text}</p>
          <div className="rounded-md bg-neutral-50 p-3">
            <div className="flex items-center gap-2 text-sm font-medium text-neutral-500">
              <MessageCircleHeart className="h-4 w-4 text-rose-500" /> Friend's reply
            </div>
            <p className="mt-1 whitespace-pre-wrap text-sm leading-relaxed text-neutral-700">{entry.reply}</p>
          </div>
          <NotesSection notes={corrections} nativeSample={entry.native_sample} />
        </div>
      )}
    </div>
  )
}

export function HistoryPage() {
  const [confirming, setConfirming] = useState(false)
  const qc = useQueryClient()
  const entries = useQuery({
    queryKey: ['diary', 'entries'],
    queryFn: () => api.diary.diaryListEntries({ params: { query: { limit: 100 } } }),
  })

  const reset = useMutation({
    mutationFn: () => api.diary.diaryResetHistory(),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['diary'] }),
  })

  if (entries.isError) {
    return (
      <ErrorState
        title="Failed to load the history"
        description={isApiError(entries.error) ? entries.error.message : String(entries.error)}
        action={<Button variant="secondary" onClick={() => entries.refetch()}>Retry</Button>}
      />
    )
  }
  if (entries.isPending || !entries.data) return <LoadingState className="py-24" />

  const items = entries.data.items ?? []

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-8">
        <header className="flex items-start justify-between">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
              <BookOpen className="h-6 w-6 text-brand-600" /> History
            </h1>
            <p className="text-sm text-neutral-500">
              {items.length === 0 ? 'No entries yet' : `Entries: ${entries.data.total}`}
            </p>
          </div>
          {items.length > 0 && (
            <Button variant="ghost" size="sm" className="text-red-600 hover:bg-red-50 hover:text-red-700" onClick={() => setConfirming(true)} disabled={reset.isPending}>
              <Trash2 className="h-4 w-4" /> Reset
            </Button>
          )}
        </header>
        {items.length === 0 ? (
          <EmptyState title="The diary is empty" description="Write your first entry — it will appear here." />
        ) : (
          <div className="flex flex-col gap-3">
            {items.map((e) => (
              <EntryCard key={e.id} entry={e} />
            ))}
          </div>
        )}
      </div>
      {confirming && (
        <ConfirmDialog
          title="Reset history"
          description="Delete all diary entries? This cannot be undone."
          confirmLabel="Delete"
          onConfirm={() => {
            reset.mutate()
            setConfirming(false)
          }}
          onClose={() => setConfirming(false)}
        />
      )}
    </div>
  )
}
