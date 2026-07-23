import { useEffect, useReducer, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Check } from 'lucide-react'
import { Button, Modal } from '@els/ui'
import { getAppIcon, type AppIcon } from '../config/appIcons'
import { ONBOARDING_RESET_EVENT, TOUR_OPEN_EVENT, isTourDone, markTourDone } from './storage'
import { TOURS } from './tours'

export function AppTour({ suspended }: { suspended: boolean }) {
  const { pathname } = useLocation()
  const [, force] = useReducer((x: number) => x + 1, 0)
  const [forced, setForced] = useState(false)

  useEffect(() => {
    const onOpen = () => setForced(true)
    window.addEventListener(ONBOARDING_RESET_EVENT, force)
    window.addEventListener(TOUR_OPEN_EVENT, onOpen)
    return () => {
      window.removeEventListener(ONBOARDING_RESET_EVENT, force)
      window.removeEventListener(TOUR_OPEN_EVENT, onOpen)
    }
  }, [])

  useEffect(() => setForced(false), [pathname])

  const appId = pathname.match(/^\/v1\/([^/]+)/)?.[1] ?? ''
  const tour = TOURS[appId]
  if (suspended || !tour || (!forced && isTourDone(appId))) return null

  const close = () => {
    markTourDone(appId)
    setForced(false)
    force()
    if (appId === 'profile') window.dispatchEvent(new Event('els:getting-started:highlight'))
  }

  return (
    <Modal onClose={close} className="max-w-xl p-0">
      <TourMedia key={appId} appId={appId} icon={getAppIcon(appId)} />
      <div className="p-6">
        <h2 className="text-lg font-semibold text-neutral-900">{tour.title}</h2>
        <p className="mt-2 text-sm leading-relaxed text-neutral-600">{tour.description}</p>
        <ul className="mt-4 space-y-2">
          {tour.features.map((f) => (
            <li key={f} className="flex items-start gap-2.5 text-sm text-neutral-700">
              <span className="mt-0.5 grid h-4.5 w-4.5 shrink-0 place-items-center rounded-full bg-brand-50">
                <Check size={11} className="text-brand-600" />
              </span>
              {f}
            </li>
          ))}
        </ul>
        <div className="mt-5 flex justify-end">
          <Button variant="brand" onClick={close}>
            Got it
          </Button>
        </div>
      </div>
    </Modal>
  )
}

export function TourMedia({ appId, icon: Icon }: { appId: string; icon: AppIcon }) {
  const sources = [`/tours/${appId}.mp4`, `/tours/${appId}.gif`]
  const [idx, setIdx] = useState(0)
  const src = sources[idx]

  if (!src) {
    return (
      <div className="flex aspect-video items-center justify-center rounded-t-3xl bg-gradient-to-br from-brand-500 to-brand-700">
        <Icon className="h-16 w-16 text-white/90" />
      </div>
    )
  }

  return src.endsWith('.mp4') ? (
    <video
      key={src}
      src={src}
      autoPlay
      loop
      muted
      playsInline
      title="Click to watch full screen"
      onClick={(e) => void e.currentTarget.requestFullscreen?.()}
      onError={() => setIdx(idx + 1)}
      className="aspect-video w-full cursor-zoom-in rounded-t-3xl bg-neutral-100 object-cover"
    />
  ) : (
    <img
      key={src}
      src={src}
      alt=""
      title="Click to watch full screen"
      onClick={(e) => void e.currentTarget.requestFullscreen?.()}
      onError={() => setIdx(idx + 1)}
      className="aspect-video w-full cursor-zoom-in rounded-t-3xl bg-neutral-100 object-cover"
    />
  )
}
