import { useState } from 'react'
import { LayoutGrid, MessageCircleQuestion, Sparkles } from 'lucide-react'
import { Button, cn, Mascot, Modal } from '@els/ui'
import type { AppIcon } from '../config/appIcons'
import { TourMedia } from './AppTour'
import type { SystemTourPage } from './tours'
import { SYSTEM_TOUR } from './tours'

const PAGE_ICONS: Record<string, AppIcon> = {
  analyze: Sparkles,
  assistant: MessageCircleQuestion,
}

const HERO_KEYFRAMES = `
@keyframes els-shimmer { from { background-position: 200% 0 } to { background-position: -200% 0 } }
@keyframes els-float { 0%, 100% { transform: translateY(0) } 50% { transform: translateY(-10px) } }
`

const MARK_GRADIENT =
  'linear-gradient(110deg,#047857 25%,#10b981 45%,#6ee7b7 50%,#10b981 55%,#047857 75%)'

function WelcomeHero() {
  return (
    <div className="relative flex aspect-video items-center justify-center overflow-hidden rounded-t-3xl bg-gradient-to-b from-brand-50 to-white">
      <style>{HERO_KEYFRAMES}</style>
      <Mascot className="absolute inset-0 h-full w-full opacity-60" />
      <div className="absolute h-48 w-80 rounded-full bg-white/80 blur-2xl" />
      <div className="absolute bottom-4 text-[11px] font-semibold uppercase tracking-[0.3em] text-emerald-700/70">
        English Learning Studio
      </div>
      <div className="relative flex flex-col items-center">
        <div className="flex text-7xl font-extrabold tracking-tight">
          {'ELS'.split('').map((ch, i) => (
            <span
              key={i}
              style={{
                display: 'inline-block',
                backgroundImage: MARK_GRADIENT,
                backgroundSize: '200% 100%',
                WebkitBackgroundClip: 'text',
                backgroundClip: 'text',
                color: 'transparent',
                animation: `els-shimmer 2.4s linear ${i * 0.2}s infinite, els-float 2.8s ease-in-out ${i * 0.15}s infinite`,
              }}
            >
              {ch}
            </span>
          ))}
        </div>
      </div>
    </div>
  )
}

export function SystemTourPageView({ page }: { page: SystemTourPage }) {
  return (
    <>
      {page.id === 'system' ? (
        <WelcomeHero />
      ) : (
        <TourMedia key={page.id} appId={page.id} icon={PAGE_ICONS[page.id] ?? LayoutGrid} />
      )}
      <div className="p-6 pb-0">
        <h2 className="text-lg font-semibold text-neutral-900">{page.title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-neutral-600">{page.description}</p>
        <ol className="mt-4 space-y-2">
          {page.steps.map((s, i) => (
            <li key={s} className="flex items-start gap-2.5 text-sm text-neutral-700">
              <span className="mt-0.5 grid h-4.5 w-4.5 shrink-0 place-items-center rounded-full bg-brand-50 text-[10px] font-semibold text-brand-600">
                {i + 1}
              </span>
              {s}
            </li>
          ))}
        </ol>
      </div>
    </>
  )
}

export function SystemTour({ onClose }: { onClose: () => void }) {
  const [page, setPage] = useState(0)
  const p = SYSTEM_TOUR[page]!
  const last = page === SYSTEM_TOUR.length - 1

  return (
    <Modal onClose={onClose} className="max-w-xl p-0">
      <SystemTourPageView page={p} />
      <div className="p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            {SYSTEM_TOUR.map((tp, i) => (
              <button
                key={tp.id}
                type="button"
                aria-label={tp.title}
                onClick={() => setPage(i)}
                className={cn(
                  'h-2 w-2 rounded-full transition-colors',
                  i === page ? 'bg-brand-600' : 'bg-neutral-300 hover:bg-neutral-400',
                )}
              />
            ))}
          </div>
          <div className="flex items-center gap-2">
            {page > 0 && (
              <Button variant="secondary" onClick={() => setPage(page - 1)}>
                Back
              </Button>
            )}
            {last ? (
              <Button variant="brand" onClick={onClose}>
                Got it
              </Button>
            ) : (
              <Button variant="brand" onClick={() => setPage(page + 1)}>
                Next
              </Button>
            )}
          </div>
        </div>
      </div>
    </Modal>
  )
}
