import { useMemo, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowRight, BookMarked, Dumbbell, Layers, Loader2, Plus, Search } from 'lucide-react'
import { api } from '../lib/api.ts'
import { Button, cn, ConfirmDialog, EmptyState, Input, LoadingState, useAgentView } from '@els/ui'
import { AddWordModal } from '../components/AddWordModal.tsx'
import { WordCard } from '../components/WordCard.tsx'
import { WordDetailModal } from '../components/WordDetailModal.tsx'
import { useDeleteUnit, useUnits } from '../store/units.ts'
import { STATUS_LABELS } from '../lib/types.ts'
import type { Unit, UnitStatus } from '../lib/types.ts'

const FILTERS: { id: '' | UnitStatus; label: string }[] = [
  { id: '', label: 'All' },
  { id: 'new', label: STATUS_LABELS.new },
  { id: 'learning', label: STATUS_LABELS.learning },
  { id: 'learned', label: STATUS_LABELS.learned },
]

export function VocabCatalog() {
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [query, setQuery] = useState('')
  const [status, setStatus] = useState<'' | UnitStatus>('')
  const [showAdd, setShowAdd] = useState(false)
  const [active, setActive] = useState<Unit | null>(null)
  const [deleting, setDeleting] = useState<Unit | null>(null)

  const unitsQ = useUnits(query, status)
  const deleteM = useDeleteUnit()
  const dueQ = useQuery({
    queryKey: ['vocab', 'cards', 'due'],
    queryFn: () => api.vocab.dueVocabCards({}),
    staleTime: 60 * 1000,
  })
  const due = dueQ.data?.count ?? 0

  const items = useMemo(() => unitsQ.data?.pages.flatMap((p) => p?.items ?? []) ?? [], [unitsQ.data])
  const total = unitsQ.data?.pages[0]?.total ?? 0

  useAgentView(
    active
      ? {
          app: 'vocab',
          screen: 'word',
          title: active.text,
          info: 'The user opened a word from their collection. Full collection — list_vocab_words; add a word — add_vocab_word.',
          ids: { unitId: active.id },
          state: { kind: active.kind, status: active.status, translation: active.translation ?? '' },
        }
      : {
          app: 'vocab',
          screen: 'catalog',
          info: 'The user is in their personal word collection. Word list — list_vocab_words; add — add_vocab_word.',
          state: { search: query, filter: status || 'all', total },
        },
  )

  function submitSearch(e: FormEvent) {
    e.preventDefault()
    setQuery(search.trim())
  }

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-7xl space-y-6 p-6">
        <header className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
              <BookMarked className="h-6 w-6 text-brand-600" />
              My Vocabulary
            </h1>
            <p className="mt-1 text-sm text-neutral-500">
              {total > 0 ? `${total} item${total === 1 ? '' : 's'} to memorize` : 'Build your personal word collection'}
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="secondary" onClick={() => navigate('cards')}>
              <Layers className="h-4 w-4" />
              Cards
            </Button>
            <Button variant="secondary" onClick={() => navigate('practice')}>
              <Dumbbell className="h-4 w-4" />
              Practice
            </Button>
            <Button variant="brand" onClick={() => setShowAdd(true)}>
              <Plus className="h-4 w-4" />
              Add word
            </Button>
          </div>
        </header>

        {due > 0 && (
          <button
            type="button"
            onClick={() => navigate('cards')}
            className="flex w-full items-center justify-between rounded-2xl bg-brand-600 px-5 py-4 text-left text-white transition hover:bg-brand-700"
          >
            <span className="flex items-center gap-3">
              <Layers className="h-5 w-5" />
              <span>
                <span className="font-semibold">
                  {due} word{due === 1 ? '' : 's'} ready to review
                </span>
                <span className="ml-2 hidden text-sm text-white/70 sm:inline">A quick round of cards moves them forward.</span>
              </span>
            </span>
            <ArrowRight className="h-5 w-5 shrink-0" />
          </button>
        )}

        <div className="flex flex-wrap items-center gap-3">
          <form onSubmit={submitSearch} className="relative flex-1 min-w-[220px]">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-neutral-400" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search words and translations…"
              className="rounded-xl py-2.5 pl-9"
            />
          </form>
          <div className="flex flex-wrap gap-2">
            {FILTERS.map((f) => (
              <button
                key={f.id || 'all'}
                type="button"
                onClick={() => setStatus(f.id)}
                className={cn(
                  'rounded-full px-3 py-1.5 text-sm font-medium ring-1 transition',
                  status === f.id
                    ? 'bg-brand-600 text-white ring-brand-600'
                    : 'bg-white text-neutral-700 ring-neutral-200 hover:bg-neutral-50',
                )}
              >
                {f.label}
              </button>
            ))}
          </div>
        </div>

        {unitsQ.isLoading ? (
          <LoadingState className="py-24 text-neutral-400" />
        ) : items.length === 0 ? (
          <CatalogEmpty searching={!!query || !!status} onAdd={() => setShowAdd(true)} />
        ) : (
          <>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
              {items.map((u) => (
                <WordCard key={u.id} unit={u} onOpen={setActive} onDelete={setDeleting} />
              ))}
            </div>
            {unitsQ.hasNextPage && (
              <div className="flex justify-center pt-2">
                <Button variant="secondary" onClick={() => unitsQ.fetchNextPage()} disabled={unitsQ.isFetchingNextPage}>
                  {unitsQ.isFetchingNextPage ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Load more'}
                </Button>
              </div>
            )}
          </>
        )}
      </div>

      {showAdd && <AddWordModal onClose={() => setShowAdd(false)} />}
      {active && <WordDetailModal unit={active} onClose={() => setActive(null)} />}
      {deleting && (
        <ConfirmDialog
          title="Delete word"
          description={`Remove "${deleting.text}" from your collection?`}
          onConfirm={() => {
            deleteM.mutate(deleting.id)
            setDeleting(null)
          }}
          onClose={() => setDeleting(null)}
        />
      )}
    </div>
  )
}

function CatalogEmpty({ searching, onAdd }: { searching: boolean; onAdd: () => void }) {
  return (
    <EmptyState
      icon={<BookMarked className="h-8 w-8" />}
      title={searching ? 'Nothing found' : 'No words yet'}
      description={searching ? 'Try a different search or filter.' : 'Add your first word and let the assistant describe it.'}
      action={
        !searching && (
          <Button variant="brand" onClick={onAdd}>
            <Plus className="h-4 w-4" />
            Add word
          </Button>
        )
      }
    />
  )
}
