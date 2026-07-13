import { Mail, ShieldCheck, UserCircle2 } from 'lucide-react'
import { ErrorState, LoadingState } from '@els/ui'
import { ProfileHeader } from './components/ProfileHeader'
import { ProfileDetails } from './components/ProfileDetails'
import { StatCard } from './components/Widget'
import { useMe } from './store/me'

export function ProfilePage() {
  const meQ = useMe()

  return (
    <div className="h-full min-h-0 w-full flex flex-col bg-neutral-50">
      <div className="flex-1 min-h-0 overflow-y-auto">
        <div className="mx-auto max-w-7xl p-6 space-y-6">
          <ProfileHeader />

          {meQ.isLoading ? (
            <LoadingState className="rounded-xl bg-white ring-1 ring-neutral-200" />
          ) : meQ.error ? (
            <ErrorState
              title="Loading error"
              description={meQ.error instanceof Error ? meQ.error.message : 'Error'}
            />
          ) : meQ.data ? (
            <AccountOverview
              email={meQ.data.email}
              role={meQ.data.role}
              status={meQ.data.status}
              isGlobalAdmin={meQ.data.isGlobalAdmin}
            />
          ) : null}
        </div>
      </div>
    </div>
  )
}

function AccountOverview({
  email,
  role,
  status,
  isGlobalAdmin,
}: {
  email: string
  role: string
  status: string
  isGlobalAdmin: boolean
}) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <StatCard label="Email" value={email} icon={<Mail size={18} />} />
      <StatCard label="Role" value={role || '—'} icon={<UserCircle2 size={18} />} tone="brand" />
      <StatCard
        label="Status"
        value={status || '—'}
        icon={<ShieldCheck size={18} />}
        tone={isGlobalAdmin ? 'emerald' : 'neutral'}
        hint={isGlobalAdmin ? 'global admin' : undefined}
      />
      <ProfileDetails />
    </div>
  )
}