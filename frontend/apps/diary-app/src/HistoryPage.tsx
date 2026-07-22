import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { Badge, Button, ConfirmDialog, EmptyState, ErrorState, LoadingState } from '@els/ui'
import { ArrowLeft, BookOpen, ChevronDown, MessageCircleHeart, Trash2 } from 'lucide-react'
import { api } from './lib/api'
import { formatDay, type Entry } from './lib/types'
import { NotesSection } from './components/NotesSection'

function EntryCard({ entry }: { entry: Entry }) {
  const [open, setOpen] = useState(false)
  const corrections = entry.corrections ?? []
  return (
    <div className="rounded-2xl border border-neutral-200 bg-white shadow-sm transition-shadow hover:shadow-md">
      <button onClick={() => setOpen(!open)} className="flex w-full items-center justify-between gap-3 rounded-2xl px-5 py-4 text-left">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-brand-700">{formatDay(entry.date)}</p>
          <p className="mt-1 truncate text-neutral-800">{entry.text}</p>
        </div>
        <span className="flex shrink-0 items-center gap-2">
          {corrections.length > 0 && <Badge tone="brand">{corrections.length}</Badge>}
          <ChevronDown className={`h-4 w-4 text-neutral-400 transition-transform ${open ? 'rotate-180' : ''}`} />
        </span>
      </button>
      {open && (
        <div className="flex flex-col gap-3 border-t border-neutral-100 px-5 py-4">
          {entry.question && (
            <p className="rounded-xl bg-brand-50 px-3.5 py-2.5 text-sm font-medium text-brand-800">{entry.question}</p>
          )}
          <p className="whitespace-pre-wrap leading-relaxed text-neutral-800">{entry.text}</p>
          {entry.reply && (
          <div className="rounded-xl bg-neutral-50 p-4">
            <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-rose-50 text-rose-500">
                <MessageCircleHeart className="h-3.5 w-3.5" />
              </span>
              Friend's reply
            </div>
            <p className="mt-2 whitespace-pre-wrap text-sm leading-relaxed text-neutral-700">{entry.reply}</p>
          </div>
          )}
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
        <header className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <Link
              to=".."
              title="Back to Diary"
              className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-800"
            >
              <ArrowLeft className="h-5 w-5" />
            </Link>
            <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
              <BookOpen className="h-6 w-6" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-neutral-900">History</h1>
              <p className="text-sm text-neutral-500">
                {items.length === 0 ? 'No entries yet' : `Entries: ${entries.data.total}`}
              </p>
            </div>
          </div>
          {items.length > 0 && (
            <Button variant="ghost" size="sm" className="text-red-600 hover:bg-red-50 hover:text-red-700" onClick={() => setConfirming(true)} disabled={reset.isPending}>
              <Trash2 className="h-4 w-4" /> Reset
            </Button>
          )}
        </header>
        {items.length === 0 ? (
          <EmptyState
            icon={<BookOpen className="h-8 w-8" />}
            title="The diary is empty"
            description="Write your first entry — it will appear here."
          />
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
