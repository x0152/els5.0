import { useState } from 'react'
import { LayoutGrid, MessageCircleQuestion, Sparkles } from 'lucide-react'
import { Button, cn, Modal } from '@els/ui'
import type { AppIcon } from '../config/appIcons'
import { TourMedia } from './AppTour'
import { SYSTEM_TOUR } from './tours'

const PAGE_ICONS: Record<string, AppIcon> = {
  system: LayoutGrid,
  analyze: Sparkles,
  assistant: MessageCircleQuestion,
}

export function SystemTour({ onClose }: { onClose: () => void }) {
  const [page, setPage] = useState(0)
  const p = SYSTEM_TOUR[page]!
  const last = page === SYSTEM_TOUR.length - 1

  return (
    <Modal onClose={onClose} className="max-w-xl p-0">
      <TourMedia key={p.id} appId={p.id} icon={PAGE_ICONS[p.id] ?? LayoutGrid} />
      <div className="p-6">
        <h2 className="text-lg font-semibold text-neutral-900">{p.title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-neutral-600">{p.description}</p>
        <ol className="mt-4 space-y-2">
          {p.steps.map((s, i) => (
            <li key={s} className="flex items-start gap-2.5 text-sm text-neutral-700">
              <span className="mt-0.5 grid h-4.5 w-4.5 shrink-0 place-items-center rounded-full bg-brand-50 text-[10px] font-semibold text-brand-600">
                {i + 1}
              </span>
              {s}
            </li>
          ))}
        </ol>
        <div className="mt-6 flex items-center justify-between">
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
