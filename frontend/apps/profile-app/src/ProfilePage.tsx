import { ErrorState, LoadingState } from '@els/ui'
import { ProfileHeader } from './components/ProfileHeader'
import { ProfileDetails } from './components/ProfileDetails'
import { ProfileTabs } from './components/ProfileTabs'
import { GettingStarted } from './components/GettingStarted'
import { useMe } from './store/me'

export function ProfilePage() {
  const meQ = useMe()

  return (
    <div className="h-full min-h-0 w-full flex flex-col bg-neutral-50">
      <div className="flex-1 min-h-0 overflow-y-auto">
        <div className="mx-auto max-w-7xl p-6 space-y-6">
          <ProfileHeader />
          <ProfileTabs />

          {meQ.isLoading ? (
            <LoadingState className="rounded-xl bg-white ring-1 ring-neutral-200" />
          ) : meQ.error ? (
            <ErrorState
              title="Loading error"
              description={meQ.error instanceof Error ? meQ.error.message : 'Error'}
            />
          ) : (
            <div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-3">
              <div className="lg:col-span-2">
                <ProfileDetails />
              </div>
              <GettingStarted />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
