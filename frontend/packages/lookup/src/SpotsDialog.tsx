import { Fragment, useEffect, type ReactNode } from 'react'
import { BookOpen, Film, X } from 'lucide-react'
import ReactMarkdown from 'react-markdown'

export interface Spot {
  ref: number
  example: string
}

interface SpotsDialogProps {
  title: string
  mediaType: string
  kind?: string
  seriesTitle?: string
  season?: number
  episode?: number
  author?: string
  spots: Spot[]
  hrefFor: (ref: number) => string
  onClose: () => void
}

function formatTime(ms: number): string {
  const total = Math.max(0, Math.floor(ms / 1000))
  const s = `${total % 60}`.padStart(2, '0')
  const m = Math.floor(total / 60) % 60
  const h = Math.floor(total / 3600)
  return h > 0 ? `${h}:${`${m}`.padStart(2, '0')}:${s}` : `${m}:${s}`
}

type Active = { i: boolean; b: boolean; u: boolean }

function renderSubtitle(text: string): ReactNode {
  const cleaned = text.replace(/\{[^}]*\}/g, '').replace(/<\/?font[^>]*>/gi, '')
  return cleaned.split(/\\N|\n/).map((line, li) => {
    const active: Active = { i: false, b: false, u: false }
    const parts: ReactNode[] = []
    line.split(/(<\/?[ibu]>)/i).forEach((part, idx) => {
      const tag = /^<(\/?)([ibu])>$/i.exec(part)
      if (tag) {
        active[tag[2]!.toLowerCase() as keyof Active] = !tag[1]
        return
      }
      if (!part) return
      let node: ReactNode = part
      if (active.i) node = <em>{node}</em>
      if (active.b) node = <strong>{node}</strong>
      if (active.u) node = <u>{node}</u>
      parts.push(<Fragment key={idx}>{node}</Fragment>)
    })
    return (
      <Fragment key={li}>
        {li > 0 && <br />}
        {parts}
      </Fragment>
    )
  })
}

export function SpotsDialog({ title, mediaType, kind, seriesTitle, season, episode, author, spots, hrefFor, onClose }: SpotsDialogProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  const isSeries = mediaType === 'film' && kind === 'series'
  const sxe = season || episode ? `S${season ?? 0}E${episode ?? 0}` : ''
  const headTitle = isSeries ? seriesTitle || title : title
  const subtitle = mediaType === 'film' ? (isSeries ? [sxe, title].filter(Boolean).join(' · ') : '') : author || ''

  return (
    <div
      className="fixed inset-0 z-[2147483647] flex items-center justify-center bg-black/40 p-4"
      onClick={onClose}
    >
      <div
        className="flex max-h-[80vh] w-full max-w-lg flex-col overflow-hidden rounded-2xl bg-white shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3 border-b border-neutral-100 px-5 py-3">
          <div className="flex min-w-0 items-start gap-2">
            {mediaType === 'film' ? <Film size={16} className="mt-0.5 shrink-0 text-neutral-400" /> : <BookOpen size={16} className="mt-0.5 shrink-0 text-neutral-400" />}
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <p className="truncate text-sm font-semibold text-neutral-800">{headTitle || 'Untitled'}</p>
                <span className="shrink-0 text-xs text-neutral-400">{spots.length}</span>
              </div>
              {subtitle && <p className="truncate text-xs text-neutral-500">{subtitle}</p>}
            </div>
          </div>
          <button type="button" onClick={onClose} className="rounded-lg p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-700">
            <X size={18} />
          </button>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto p-2">
          {spots.map((s, i) => (
            <a
              key={i}
              href={hrefFor(s.ref)}
              className="flex gap-2 rounded-xl px-3 py-2 text-sm text-neutral-700 transition-colors hover:bg-neutral-50"
            >
              {mediaType === 'film' && (
                <span className="shrink-0 pt-0.5 text-xs font-medium tabular-nums text-brand-600">{formatTime(s.ref)}</span>
              )}
              {mediaType === 'film' ? (
                <span className="leading-snug">{renderSubtitle(s.example) || `Spot ${i + 1}`}</span>
              ) : (
                <div className="leading-relaxed [&_p]:m-0">
                  <ReactMarkdown>{s.example || `Spot ${i + 1}`}</ReactMarkdown>
                </div>
              )}
            </a>
          ))}
        </div>
      </div>
    </div>
  )
}
