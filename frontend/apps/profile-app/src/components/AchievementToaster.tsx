import { useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { X } from 'lucide-react'
import { achievementMeta } from '../achievements/meta'
import { useAckItems, useOnboardingProgress, type ProgressItem } from '../store/onboarding'

export function AchievementToaster() {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const progressQ = useOnboardingProgress()
  const ack = useAckItems()
  const [hidden, setHidden] = useState<string[]>([])

  const refetch = progressQ.refetch
  useEffect(() => {
    void refetch()
  }, [pathname, refetch])

  const onAchievementsPage = pathname.includes('/v1/profile/achievements')
  const next = onAchievementsPage
    ? undefined
    : (progressQ.data ?? []).find((i) => i.done && !i.acked && !hidden.includes(i.id))

  if (!next) return null

  const dismiss = () => {
    setHidden((h) => [...h, next.id])
    ack.mutate([next.id])
  }

  return (
    <Toast
      key={next.id}
      item={next}
      onDismiss={dismiss}
      onOpen={() => {
        dismiss()
        navigate('/v1/profile/achievements')
      }}
    />
  )
}

function Toast({
  item,
  onDismiss,
  onOpen,
}: {
  item: ProgressItem
  onDismiss: () => void
  onOpen: () => void
}) {
  const [entered, setEntered] = useState(false)

  useEffect(() => {
    const enter = requestAnimationFrame(() => setEntered(true))
    const timer = setTimeout(onDismiss, 7000)
    return () => {
      cancelAnimationFrame(enter)
      clearTimeout(timer)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const meta = achievementMeta(item)
  const Icon = meta.icon

  return (
    <div
      className={`fixed bottom-[calc(4.5rem+env(safe-area-inset-bottom,0px))] right-4 z-50 transition-all duration-300 md:bottom-4 ${
        entered ? 'translate-y-0 opacity-100' : 'translate-y-3 opacity-0'
      }`}
    >
      <div className="flex w-80 items-center gap-3 rounded-2xl bg-white p-4 shadow-xl ring-1 ring-neutral-200">
        <button type="button" onClick={onOpen} className="flex flex-1 items-center gap-3 text-left">
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-full bg-gradient-to-br from-amber-300 to-amber-500 text-white ring-1 ring-amber-400">
            <Icon size={24} />
          </span>
          <span className="min-w-0">
            <span className="block text-[11px] font-semibold uppercase tracking-wider text-amber-600">
              Achievement unlocked
            </span>
            <span className="block truncate text-sm font-medium text-neutral-900">
              {meta.title}
            </span>
          </span>
        </button>
        <button
          type="button"
          onClick={onDismiss}
          className="shrink-0 text-neutral-300 hover:text-neutral-500"
        >
          <X size={16} />
        </button>
      </div>
    </div>
  )
}
