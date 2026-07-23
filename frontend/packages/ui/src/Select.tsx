import { useEffect, useRef, useState, type CSSProperties } from 'react'
import { createPortal } from 'react-dom'
import { Check, ChevronDown } from 'lucide-react'
import { cn } from './cn.ts'
import { inputClass } from './Input.tsx'

export interface SelectOption {
  value: string
  label: string
}

export interface SelectProps {
  value: string
  onChange: (value: string) => void
  options: SelectOption[]
  placeholder?: string
  disabled?: boolean
  dark?: boolean
  className?: string
  title?: string
  up?: boolean
}

export function Select({
  value,
  onChange,
  options,
  placeholder = 'Select…',
  disabled,
  dark,
  className,
  title,
  up: forceUp,
}: SelectProps) {
  const btnRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const [open, setOpen] = useState(false)
  const [hi, setHi] = useState(-1)
  const [style, setStyle] = useState<CSSProperties>({})

  const selected = options.find((o) => o.value === value)

  const openMenu = () => {
    const r = btnRef.current!.getBoundingClientRect()
    const estimate = Math.min(280, options.length * 34 + 10)
    const below = window.innerHeight - r.bottom
    const up = forceUp ?? (below < estimate && r.top > below)
    const alignRight = r.left + r.width / 2 > window.innerWidth / 2
    setStyle({
      minWidth: r.width,
      maxWidth: 'min(20rem, calc(100vw - 16px))',
      maxHeight: Math.max(120, Math.min(280, (up ? r.top : below) - 12)),
      ...(alignRight ? { right: window.innerWidth - r.right } : { left: r.left }),
      ...(up ? { bottom: window.innerHeight - r.top + 4 } : { top: r.bottom + 4 }),
    })
    setHi(options.findIndex((o) => o.value === value))
    setOpen(true)
  }

  const pick = (v: string) => {
    setOpen(false)
    onChange(v)
  }

  useEffect(() => {
    if (!open) return
    const hide = () => setOpen(false)
    const onDown = (e: PointerEvent) => {
      const t = e.target as Node
      if (!btnRef.current?.contains(t) && !menuRef.current?.contains(t)) setOpen(false)
    }
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' || e.key === 'Tab') setOpen(false)
      else if (e.key === 'ArrowDown') {
        e.preventDefault()
        setHi((h) => Math.min(h + 1, options.length - 1))
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        setHi((h) => Math.max(h - 1, 0))
      } else if (e.key === 'Enter') {
        e.preventDefault()
        const o = options[hi]
        if (o) pick(o.value)
      }
    }
    const onScroll = (e: Event) => {
      if (!menuRef.current?.contains(e.target as Node)) setOpen(false)
    }
    window.addEventListener('pointerdown', onDown)
    window.addEventListener('keydown', onKey)
    window.addEventListener('scroll', onScroll, true)
    window.addEventListener('resize', hide)
    return () => {
      window.removeEventListener('pointerdown', onDown)
      window.removeEventListener('keydown', onKey)
      window.removeEventListener('scroll', onScroll, true)
      window.removeEventListener('resize', hide)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, options, hi])

  useEffect(() => {
    if (open && hi >= 0) {
      menuRef.current?.children[hi]?.scrollIntoView({ block: 'nearest' })
    }
  }, [open, hi])

  return (
    <>
      <button
        ref={btnRef}
        type="button"
        disabled={disabled}
        title={title}
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={() => (open ? setOpen(false) : openMenu())}
        onKeyDown={(e) => {
          if (!open && (e.key === 'ArrowDown' || e.key === 'ArrowUp')) {
            e.preventDefault()
            openMenu()
          }
        }}
        className={cn(
          inputClass,
          'flex cursor-pointer items-center justify-between gap-2 text-left disabled:cursor-default',
          dark && 'border-white/20 bg-white/10 text-white focus:border-white/40 focus:ring-white/20',
          className,
        )}
      >
        <span className={cn('truncate', !selected && !value && (dark ? 'text-white/50' : 'text-neutral-400'))}>
          {selected?.label ?? (value || placeholder)}
        </span>
        <ChevronDown size={14} className={cn('shrink-0 opacity-50 transition-transform', open && 'rotate-180')} />
      </button>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            role="listbox"
            style={style}
            className={cn(
              'fixed z-[70] overflow-y-auto rounded-xl border p-1 shadow-xl',
              dark ? 'border-white/10 bg-neutral-900/95 text-white backdrop-blur' : 'border-neutral-200 bg-white',
            )}
          >
            {options.map((o, i) => {
              const isSelected = o.value === value
              return (
                <button
                  key={o.value}
                  type="button"
                  role="option"
                  aria-selected={isSelected}
                  onClick={() => pick(o.value)}
                  onMouseEnter={() => setHi(i)}
                  className={cn(
                    'flex w-full items-center justify-between gap-2 rounded-lg px-2.5 py-1.5 text-left text-sm transition-colors',
                    i === hi && (dark ? 'bg-white/10' : 'bg-neutral-100'),
                    isSelected && 'font-medium',
                  )}
                >
                  <span className="truncate">{o.label}</span>
                  {isSelected && <Check size={14} className={cn('shrink-0', dark ? 'text-brand-400' : 'text-brand-600')} />}
                </button>
              )
            })}
          </div>,
          (document.fullscreenElement ?? document.body) as HTMLElement,
        )}
    </>
  )
}
