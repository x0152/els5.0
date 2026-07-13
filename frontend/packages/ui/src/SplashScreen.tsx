import { cn } from './cn'

export interface SplashScreenProps {
  title?: string
  subtitle?: string
  fullscreen?: boolean
  className?: string
}

const SPLASH_KEYFRAMES = `
@keyframes els-shimmer { from { background-position: 200% 0 } to { background-position: -200% 0 } }
@keyframes els-float { 0%, 100% { transform: translateY(0) } 50% { transform: translateY(-14px) } }
@keyframes els-bar { 0% { left: -45% } 100% { left: 100% } }
`

const MARK_GRADIENT =
  'linear-gradient(110deg,#047857 25%,#10b981 45%,#6ee7b7 50%,#10b981 55%,#047857 75%)'

export function SplashScreen({
  title = 'ELS',
  subtitle = 'Loading the application…',
  fullscreen = true,
  className,
}: SplashScreenProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center',
        fullscreen
          ? 'fixed inset-0 z-[99999] bg-[radial-gradient(1200px_800px_at_50%_120%,#d1fae5_0%,transparent_60%),linear-gradient(to_bottom,#ffffff_0%,#ecfdf5_100%)]'
          : 'w-full py-12',
        className,
      )}
      role="status"
      aria-live="polite"
    >
      <style>{SPLASH_KEYFRAMES}</style>
      <div className={cn('flex font-extrabold tracking-tight', fullscreen ? 'text-7xl' : 'text-4xl')}>
        {title.split('').map((ch, i) => (
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
      <div className="mt-3 text-[10px] font-semibold uppercase tracking-[0.24em] text-emerald-700/70">
        English Learning Studio
      </div>
      <div className="relative mt-9 h-1 w-40 overflow-hidden rounded-full bg-emerald-100">
        <span
          className="absolute top-0 h-full w-2/5 rounded-full"
          style={{
            backgroundImage: 'linear-gradient(90deg,transparent,#10b981,transparent)',
            animation: 'els-bar 1.3s ease-in-out infinite',
          }}
        />
      </div>
      <div className={cn('text-neutral-500 mt-5', fullscreen ? 'text-base' : 'text-sm')}>
        {subtitle}
      </div>
    </div>
  )
}
