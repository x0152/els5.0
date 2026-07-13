import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Swords } from 'lucide-react'
import { Button, ConfirmDialog, EmptyState, LoadingState, useAgentView } from '@els/ui'
import { CreateMissionModal } from '../components/CreateMissionModal.tsx'
import { MissionCard } from '../components/MissionCard.tsx'
import { useCreateMission, useDeleteMission, useMissions } from '../store/missions.ts'
import type { MissionSummary } from '../lib/types.ts'

export function QuestHub() {
  const navigate = useNavigate()
  const missionsQ = useMissions()
  const createM = useCreateMission()
  const deleteM = useDeleteMission()
  const [showCreate, setShowCreate] = useState(false)
  const [deleting, setDeleting] = useState<MissionSummary | null>(null)

  const missions = missionsQ.data ?? []
  const pending = missions.filter((m) => m.generationStatus !== 'ready')
  const fresh = missions.filter((m) => m.generationStatus === 'ready' && !m.started)
  const active = missions.filter((m) => m.generationStatus === 'ready' && m.started && !m.isComplete)
  const completed = missions.filter((m) => m.generationStatus === 'ready' && m.started && m.isComplete)

  useAgentView({
    app: 'quest',
    screen: 'hub',
    info: 'The user is on the quests list. List — list_quests; quest details — read_quest (part=info or dialogue).',
    state: { active: active.length, completed: completed.length },
  })

  function open(id: string) {
    navigate(id)
  }

  function create(payload: { prompt: string; genre: string; practiceGoals: string }) {
    createM.mutate({ ...payload, language: 'English' }, { onSuccess: () => setShowCreate(false) })
  }

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-7xl space-y-8 p-6">
        <header className="flex items-center justify-between">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
              <Swords className="h-6 w-6 text-brand-600" />
              Quests
            </h1>
            <p className="mt-1 text-sm text-neutral-500">Interactive missions for language practice</p>
          </div>
          <Button variant="brand" onClick={() => setShowCreate(true)}>
            <Plus className="h-4 w-4" />
            New adventure
          </Button>
        </header>

        {missionsQ.isLoading ? (
          <LoadingState className="py-24 text-neutral-400" />
        ) : missions.length === 0 ? (
          <EmptyState
            icon={<Swords className="h-8 w-8" />}
            title="No adventures yet"
            description="Create your first mission and start playing."
            action={
              <Button variant="brand" onClick={() => setShowCreate(true)}>
                <Plus className="h-4 w-4" />
                New adventure
              </Button>
            }
          />
        ) : (
          <>
            <Section title="Generating" items={pending} onOpen={open} onDelete={setDeleting} />
            <Section title="Continue" items={active} onOpen={open} onDelete={setDeleting} />
            <Section title="Discover" items={fresh} onOpen={open} onDelete={setDeleting} />
            <Section title="Completed" items={completed} onOpen={open} onDelete={setDeleting} />
          </>
        )}
      </div>

      {showCreate && (
        <CreateMissionModal
          submitting={createM.isPending}
          error={createM.isError ? (createM.error as Error)?.message || 'Failed to create the mission' : null}
          onClose={() => setShowCreate(false)}
          onCreate={create}
        />
      )}

      {deleting && (
        <ConfirmDialog
          title="Delete adventure"
          description={`Delete "${deleting.title || 'this adventure'}"? All progress will be lost.`}
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

function Section({
  title,
  items,
  onOpen,
  onDelete,
}: {
  title: string
  items: MissionSummary[]
  onOpen: (id: string) => void
  onDelete: (mission: MissionSummary) => void
}) {
  if (items.length === 0) return null
  return (
    <section>
      <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">{title}</h2>
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {items.map((m) => (
          <MissionCard key={m.id} mission={m} onOpen={onOpen} onDelete={onDelete} />
        ))}
      </div>
    </section>
  )
}
