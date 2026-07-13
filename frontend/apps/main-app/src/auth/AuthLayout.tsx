import type { ReactNode } from 'react'
import { LogoWordmark } from '@els/ui'

const BRAND = '#059669'

export function AuthLayout({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string
  subtitle?: ReactNode
  children: ReactNode
  footer?: ReactNode
}) {
  return (
    <div className="min-h-dvh w-full flex bg-neutral-50 text-neutral-900">
      <div
        className="hidden lg:flex flex-col justify-between w-[42%] max-w-[560px] p-12 text-white relative overflow-hidden"
        style={{
          background: `linear-gradient(160deg, ${BRAND} 0%, #064e3b 100%)`,
        }}
      >
        <div className="relative z-10 flex justify-center">
          <LogoWordmark wordmark="ELS" color="#ffffff" markClassName="text-3xl" />
        </div>

        <div className="relative z-10 space-y-6 max-w-sm">
          <div className="text-4xl font-semibold leading-tight tracking-tight">
            English Learning Studio
          </div>
          <div className="text-base text-white/85 leading-relaxed">
            Learn English with real content — books, films, and podcasts. ELS remembers
            every word and grammar point you encounter and builds a personal review
            program from them.
          </div>
        </div>

        <div className="relative z-10 text-xs text-white/60 text-center">
          © {new Date().getFullYear()} ELS
        </div>

        <div
          aria-hidden
          className="absolute inset-0 pointer-events-none opacity-30"
          style={{
            backgroundImage:
              'radial-gradient(circle at 20% 20%, rgba(255,255,255,0.35) 0%, transparent 42%), radial-gradient(circle at 80% 70%, rgba(255,255,255,0.2) 0%, transparent 55%)',
          }}
        />
        <div
          aria-hidden
          className="absolute -bottom-24 -right-24 w-[420px] h-[420px] rounded-full pointer-events-none"
          style={{
            background: 'radial-gradient(circle, rgba(255,255,255,0.15) 0%, transparent 70%)',
          }}
        />
      </div>

      <div className="flex-1 flex items-center justify-center p-6 sm:p-10">
        <div className="w-full max-w-[420px]">
          <div className="lg:hidden mb-8 flex justify-center">
            <LogoWordmark wordmark="ELS" color={BRAND} markClassName="text-3xl" />
          </div>
          <h1 className="text-2xl font-semibold tracking-tight text-neutral-900">{title}</h1>
          {subtitle ? (
            <p className="mt-2 text-sm text-neutral-600 leading-relaxed">{subtitle}</p>
          ) : null}
          <div className="mt-8">{children}</div>
          {footer ? <div className="mt-8 text-sm text-neutral-500">{footer}</div> : null}
        </div>
      </div>
    </div>
  )
}

export const AUTH_BRAND = BRAND
