import { useEffect, useState } from 'react'
import { LogOut, ShieldAlert } from 'lucide-react'
import { useAuth } from '../auth/AuthContext'
import { getImpersonation, stopImpersonation } from '../auth/impersonation'

export function ImpersonationBanner() {
  const { user } = useAuth()
  const [busy, setBusy] = useState(false)
  const [tick, setTick] = useState(0)
  const state = getImpersonation()

  useEffect(() => {
    function onStorage() {
      setTick((n) => n + 1)
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  void tick

  if (!state) return null

  async function onExit() {
    setBusy(true)
    stopImpersonation()
    window.location.assign('/v1/admin')
  }

  const targetLabel = user?.displayName || user?.email || '—'
  const originalLabel = state.originalLabel || 'global admin'

  return (
    <div className="sticky top-0 z-40 bg-amber-100 text-amber-900 border-b border-amber-300">
      <div className="px-4 py-2 flex items-center gap-3 text-sm">
        <ShieldAlert size={16} className="shrink-0" />
        <div className="min-w-0 flex-1 truncate">
          <span className="font-semibold">Impersonation:</span>{' '}
          you are signed in as <span className="font-semibold">{targetLabel}</span>
          <span className="hidden sm:inline">
            {' '}— original account <span className="font-medium">{originalLabel}</span>
          </span>
        </div>
        <button
          type="button"
          onClick={onExit}
          disabled={busy}
          className="inline-flex items-center gap-1.5 px-2.5 h-7 rounded-md bg-amber-200 hover:bg-amber-300 disabled:opacity-60 text-amber-900 font-medium text-[12px] ring-1 ring-amber-300"
          title="Return to original account"
        >
          <LogOut size={12} />
          Exit
        </button>
      </div>
    </div>
  )
}
