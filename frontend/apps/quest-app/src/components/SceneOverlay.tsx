import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { cn } from '@els/ui'

const OVERLAY_PHRASES = ['The story moves on…', 'A new scene is taking shape…', 'Meanwhile…']

interface Props {
  sceneNumber: number
  summary?: string
  backdropImage?: string
  leaving?: boolean
  onClose: () => void
}

// SceneTransitionDialog — cinematic scene transition as a
// dialog: background blurs, centered card with scene number and the
// previous outcome. Closes on click — a normal loading banner stays in chat.
export function SceneTransitionDialog({ sceneNumber, summary, backdropImage, leaving, onClose }: Props) {
  const [phrase, setPhrase] = useState(0)
  useEffect(() => {
    const timer = setInterval(() => setPhrase((n) => (n + 1) % OVERLAY_PHRASES.length), 2600)
    return () => clearInterval(timer)
  }, [])

  return (
    <div
      onClick={onClose}
      className={cn(
        'absolute inset-0 z-20 flex items-center justify-center bg-neutral-950/40 p-3 backdrop-blur-sm transition-opacity duration-700 sm:p-6',
        leaving ? 'pointer-events-none opacity-0' : 'animate-word-in',
      )}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="animate-rise-in relative flex min-h-[60%] w-full max-w-3xl flex-col items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-b from-neutral-900 via-brand-900 to-neutral-900 px-6 py-12 text-center shadow-2xl ring-1 ring-white/10 sm:px-12 sm:py-16"
      >
        {backdropImage && (
          <>
            <img
              src={backdropImage}
              alt=""
              className="pointer-events-none absolute inset-0 h-full w-full scale-110 object-cover opacity-25 blur-[3px]"
            />
            <span className="absolute left-4 top-4 rounded-full bg-black/50 px-2.5 py-1 text-[10px] font-medium uppercase tracking-wider text-white/60 backdrop-blur-sm">
              Previous scene
            </span>
          </>
        )}

        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="absolute right-3 top-3 grid h-7 w-7 place-items-center rounded-full bg-black/40 text-white/60 backdrop-blur-sm transition-colors hover:text-white"
        >
          <X className="h-4 w-4" />
        </button>

        <div className="relative space-y-5 sm:space-y-6">
          <div key={phrase} className="animate-word-in text-xs uppercase tracking-[0.25em] text-brand-300/80 sm:text-sm">
            {OVERLAY_PHRASES[phrase]}
          </div>

          <div className="text-5xl font-bold tracking-tight text-white sm:text-7xl">Scene {sceneNumber}</div>

          {summary && (
            <p className="mx-auto max-w-xl text-sm italic leading-relaxed text-neutral-300 sm:text-lg">{summary}</p>
          )}

          <div className="flex items-center justify-center gap-1.5 pt-1 sm:gap-2">
            {[0, 160, 320].map((delay) => (
              <span
                key={delay}
                className="h-1.5 w-1.5 animate-bounce rounded-full bg-brand-400/60 sm:h-2 sm:w-2"
                style={{ animationDelay: `${delay}ms` }}
              />
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
