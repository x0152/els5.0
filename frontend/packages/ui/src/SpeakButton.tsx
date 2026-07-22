import { useRef, useState, type ButtonHTMLAttributes, type MouseEvent, type ReactNode } from 'react'
import { AudioLines, Loader2, Pause, Play, Square, Volume2 } from 'lucide-react'
import { cn } from './cn.ts'
import { speak } from './speech.ts'

const VARIANTS = {
  icon: 'rounded-md p-1 text-neutral-400 transition-colors hover:bg-neutral-100 hover:text-brand-600 disabled:opacity-50',
  button:
    'inline-flex h-9 items-center justify-center gap-2 rounded-lg bg-white px-4 text-sm font-semibold text-neutral-800 shadow-sm ring-1 ring-inset ring-neutral-200 transition-colors hover:bg-neutral-50 disabled:opacity-50',
  ghost:
    'inline-flex h-9 items-center justify-center gap-2 rounded-lg px-3 text-sm font-semibold text-neutral-700 transition-colors hover:bg-neutral-100 disabled:opacity-50',
  pill: 'inline-flex items-center gap-1 rounded-full bg-brand-50 px-2.5 py-1 text-xs font-medium text-brand-700 ring-1 ring-brand-100 transition-colors hover:bg-brand-100 disabled:opacity-50',
  round:
    'flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-lg shadow-brand-600/30 transition hover:brightness-110 active:scale-95 disabled:opacity-50',
}

export interface SpeakButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  text?: string
  rate?: number
  voice?: string
  onPlay?: () => Promise<HTMLAudioElement | null | void> | void
  variant?: keyof typeof VARIANTS
  icon?: ReactNode
  iconClassName?: string
  pendingText?: ReactNode
}

const fmt = (s: number) => `${Math.floor(s / 60)}:${String(Math.floor(s % 60)).padStart(2, '0')}`

export function SpeakButton({
  text,
  rate,
  voice,
  onPlay,
  variant = 'icon',
  icon,
  iconClassName,
  pendingText,
  className,
  children,
  disabled,
  onClick,
  ...props
}: SpeakButtonProps) {
  const [pending, setPending] = useState(false)
  const [audio, setAudio] = useState<HTMLAudioElement | null>(null)
  const [paused, setPaused] = useState(false)
  const [time, setTime] = useState(0)
  const userPause = useRef(false)

  const reset = () => {
    setAudio(null)
    setPaused(false)
    setTime(0)
  }

  const track = (a: HTMLAudioElement) => {
    setAudio(a)
    setPaused(false)
    setTime(0)
    a.addEventListener('timeupdate', () => setTime(a.currentTime))
    a.addEventListener('ended', reset)
    a.addEventListener('pause', () => {
      if (a.ended) return
      if (userPause.current) {
        userPause.current = false
        setPaused(true)
      } else {
        reset()
      }
    })
  }

  const handle = (e: MouseEvent<HTMLButtonElement>) => {
    onClick?.(e)
    const run = onPlay
      ? onPlay()
      : speak(text ?? '', { ...(rate ? { rate } : {}), ...(voice ? { voice } : {}) })
    if (run instanceof Promise) {
      setPending(true)
      void run
        .then((a) => {
          if (a instanceof HTMLAudioElement) track(a)
        })
        .finally(() => setPending(false))
    }
  }

  const stopAudio = () => {
    if (!audio) return
    audio.pause()
    reset()
  }

  const togglePause = () => {
    if (!audio) return
    if (paused) {
      setPaused(false)
      void audio.play()
    } else {
      userPause.current = true
      audio.pause()
    }
  }

  const iconCls = cn(variant === 'round' ? 'h-7 w-7' : 'h-4 w-4', iconClassName)

  if (audio) {
    const duration = Number.isFinite(audio.duration) ? audio.duration : 0
    const status = paused ? (
      <Pause className={cn(iconCls, 'shrink-0')} fill="currentColor" />
    ) : (
      <AudioLines className={cn(iconCls, 'shrink-0 animate-pulse')} />
    )
    return (
      <span
        title={paused ? 'Paused' : 'Playing'}
        className={cn(
          VARIANTS[variant],
          'group relative cursor-default',
          variant === 'icon' && 'inline-flex items-center align-middle',
          className,
          '!opacity-100',
        )}
        {...(props as object)}
      >
        {status}
        {children}
        <span className="absolute left-0 top-1/2 z-10 origin-left -translate-y-1/2 scale-x-0 opacity-0 transition-all duration-200 group-hover:scale-x-100 group-hover:opacity-100">
          <span
            className={cn(
              VARIANTS[variant],
              className,
              'flex w-max items-center gap-1.5 whitespace-nowrap',
              variant === 'icon' && 'bg-white',
              variant === 'round' && 'w-max px-5',
            )}
          >
            {status}
            {children}
            <span className="text-[11px] font-medium tabular-nums">
              {fmt(time)}{duration ? ` / ${fmt(duration)}` : ''}
            </span>
            <button
              type="button"
              onClick={togglePause}
              title={paused ? 'Resume' : 'Pause'}
              className="rounded p-0.5 transition-opacity hover:opacity-70"
            >
              {paused ? (
                <Play className="h-3.5 w-3.5" fill="currentColor" />
              ) : (
                <Pause className="h-3.5 w-3.5" fill="currentColor" />
              )}
            </button>
            <button
              type="button"
              onClick={stopAudio}
              title="Stop"
              className="rounded p-0.5 transition-opacity hover:opacity-70"
            >
              <Square className="h-3 w-3" fill="currentColor" />
            </button>
          </span>
        </span>
      </span>
    )
  }

  return (
    <button
      type="button"
      onClick={handle}
      disabled={disabled || pending}
      className={cn(VARIANTS[variant], className, pending && '!opacity-100')}
      {...props}
    >
      {pending ? (
        <Loader2 className={cn(iconCls, 'animate-spin')} />
      ) : (
        (icon ?? (variant === 'round' ? (
          <Play className={cn(iconCls, 'translate-x-0.5')} fill="currentColor" />
        ) : (
          <Volume2 className={iconCls} />
        )))
      )}
      {pending ? (pendingText ?? children) : children}
    </button>
  )
}
