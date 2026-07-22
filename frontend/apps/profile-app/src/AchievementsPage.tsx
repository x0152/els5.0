import { useEffect } from 'react'
import { cn, ErrorState, LoadingState } from '@els/ui'
import { ProfileHeader } from './components/ProfileHeader'
import { ProfileTabs } from './components/ProfileTabs'
import { achievementMeta, GROUP_LABELS } from './achievements/meta'
import { useAckItems, useOnboardingProgress, type ProgressItem } from './store/onboarding'

export function AchievementsPage() {
  const progressQ = useOnboardingProgress()
  const ack = useAckItems()
  const items = progressQ.data ?? []

  useEffect(() => {
    const ids = items.filter((i) => i.done && !i.acked).map((i) => i.id)
    if (ids.length > 0 && !ack.isPending) ack.mutate(ids)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [progressQ.data])

  const firstSteps = items.filter((i) => i.kind === 'checklist')
  const groups = Object.entries(GROUP_LABELS)
    .map(([metric, label]) => ({
      metric,
      label,
      items: items.filter((i) => i.kind === 'achievement' && i.metric === metric),
    }))
    .filter((g) => g.items.length > 0)

  const unlocked = items.filter((i) => i.done).length

  return (
    <div className="h-full min-h-0 w-full flex flex-col bg-neutral-50">
      <div className="flex-1 min-h-0 overflow-y-auto">
        <div className="mx-auto max-w-7xl p-6 space-y-6">
          <ProfileHeader />
          <ProfileTabs />

          {progressQ.isLoading ? (
            <LoadingState className="rounded-xl bg-white ring-1 ring-neutral-200" />
          ) : progressQ.error ? (
            <ErrorState
              title="Loading error"
              description={progressQ.error instanceof Error ? progressQ.error.message : 'Error'}
            />
          ) : (
            <>
              <div className="text-sm text-neutral-500">
                Unlocked <span className="font-semibold text-neutral-800">{unlocked}</span> of{' '}
                {items.length}
              </div>
              {firstSteps.length > 0 && <Group label="First steps" items={firstSteps} />}
              {groups.map((g) => (
                <Group key={g.metric} label={g.label} items={g.items} />
              ))}
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function Group({ label, items }: { label: string; items: ProgressItem[] }) {
  return (
    <section className="rounded-2xl bg-white p-5 ring-1 ring-neutral-200">
      <h3 className="mb-4 text-sm font-semibold text-neutral-800">{label}</h3>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
        {items.map((item) => (
          <AchievementBadge key={item.id} item={item} />
        ))}
      </div>
    </section>
  )
}

function AchievementBadge({ item }: { item: ProgressItem }) {
  const meta = achievementMeta(item)
  const Icon = meta.icon
  const pct = Math.min(100, Math.round((item.value / item.threshold) * 100))
  return (
    <div
      className={cn(
        'flex flex-col items-center gap-2 rounded-2xl p-4 text-center ring-1',
        item.done ? 'bg-amber-50 ring-amber-200' : 'bg-neutral-50 ring-neutral-200',
      )}
    >
      <span
        className={cn(
          'grid h-14 w-14 place-items-center rounded-full ring-1',
          item.done
            ? 'bg-gradient-to-br from-amber-300 to-amber-500 text-white ring-amber-400 shadow-sm'
            : 'bg-white text-neutral-300 ring-neutral-200',
        )}
      >
        <Icon size={30} />
      </span>
      <div
        className={cn(
          'text-xs font-medium leading-tight',
          item.done ? 'text-neutral-800' : 'text-neutral-500',
        )}
      >
        {meta.title}
      </div>
      {!item.done && (
        <div className="w-full">
          <div className="h-1 w-full overflow-hidden rounded-full bg-neutral-200">
            <div className="h-full bg-brand-500" style={{ width: `${pct}%` }} />
          </div>
          <div className="mt-1 text-[10px] text-neutral-400">
            {Math.min(item.value, item.threshold)}/{item.threshold}
          </div>
        </div>
      )}
    </div>
  )
}
