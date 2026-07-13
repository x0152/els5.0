import { useEffect, useRef, useState } from 'react'
import { Check, RefreshCw, Search, ChevronDown } from 'lucide-react'
import { cn, Input } from '@els/ui'

type Props = {
  value: string
  onChange: (value: string) => void
  models: string[]
  loading: boolean
  error: string | null
  onLoad: () => void
}

const DROPDOWN_MAX_PX = 360

export function ModelCombobox({ value, onChange, models, loading, error, onLoad }: Props) {
  const [open, setOpen] = useState(false)
  const [dropUp, setDropUp] = useState(false)
  const [query, setQuery] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const onClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onClick)
    return () => document.removeEventListener('mousedown', onClick)
  }, [])

  const filtered = query.trim()
    ? models.filter((m) => m.toLowerCase().includes(query.trim().toLowerCase()))
    : models

  const openAndLoad = () => {
    const rect = ref.current?.getBoundingClientRect()
    if (rect) {
      const spaceBelow = window.innerHeight - rect.bottom
      setDropUp(spaceBelow < DROPDOWN_MAX_PX && rect.top > spaceBelow)
    }
    setOpen(true)
    if (models.length === 0 && !loading) onLoad()
  }

  return (
    <div ref={ref} className="relative">
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Input
            value={value}
            onChange={(e) => onChange(e.target.value)}
            onFocus={openAndLoad}
            placeholder="Model (e.g. gpt-4o-mini)"
            className="pr-9"
          />
          <button
            type="button"
            onClick={() => (open ? setOpen(false) : openAndLoad())}
            className="absolute right-2 top-1/2 -translate-y-1/2 text-neutral-400 hover:text-neutral-600"
          >
            <ChevronDown size={16} className={cn('transition-transform', open && 'rotate-180')} />
          </button>
        </div>
        <button
          type="button"
          onClick={onLoad}
          title="Refresh model list"
          className="rounded-lg border border-neutral-300 p-2 text-neutral-500 transition-colors hover:border-brand-300 hover:bg-brand-50 hover:text-brand-600"
        >
          <RefreshCw size={16} className={cn(loading && 'animate-spin')} />
        </button>
      </div>

      {open && (
        <div
          className={cn(
            'absolute z-20 w-full overflow-hidden rounded-xl bg-white shadow-xl ring-1 ring-neutral-200',
            dropUp ? 'bottom-full mb-1.5' : 'mt-1.5',
          )}
        >
          <div className="flex items-center gap-2 border-b border-neutral-100 px-3 py-2.5">
            <Search size={15} className="text-neutral-400" />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search models…"
              className="w-full bg-transparent text-sm outline-none placeholder:text-neutral-400"
            />
            {!loading && !error && models.length > 0 && (
              <span className="shrink-0 text-xs text-neutral-400">{filtered.length}</span>
            )}
          </div>
          <div className="max-h-72 overflow-auto p-1">
            {loading && <div className="px-3 py-6 text-center text-sm text-neutral-400">Loading…</div>}
            {!loading && error && <div className="px-3 py-6 text-center text-sm text-red-500">{error}</div>}
            {!loading && !error && filtered.length === 0 && (
              <div className="px-3 py-6 text-center text-sm text-neutral-400">Nothing found</div>
            )}
            {!loading &&
              !error &&
              filtered.map((m) => {
                const selected = m === value
                return (
                  <button
                    key={m}
                    type="button"
                    onClick={() => {
                      onChange(m)
                      setOpen(false)
                      setQuery('')
                    }}
                    className={cn(
                      'flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-brand-50',
                      selected ? 'bg-brand-50 font-medium text-brand-700' : 'text-neutral-700',
                    )}
                  >
                    <span className="truncate">{m}</span>
                    {selected && <Check size={15} className="ml-auto shrink-0 text-brand-600" />}
                  </button>
                )
              })}
          </div>
        </div>
      )}
    </div>
  )
}
