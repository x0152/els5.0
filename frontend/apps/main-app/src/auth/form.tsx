import type { ButtonHTMLAttributes, InputHTMLAttributes, ReactNode } from 'react'
import { forwardRef } from 'react'
import { AUTH_BRAND } from './AuthLayout'

type FieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string
  hint?: ReactNode
  error?: string | null
}

export const Field = forwardRef<HTMLInputElement, FieldProps>(function Field(
  { label, hint, error, className, ...rest },
  ref,
) {
  const hasError = !!error
  return (
    <label className="block">
      <span className="block text-xs font-medium uppercase tracking-wide text-neutral-500 mb-2">
        {label}
      </span>
      <input
        ref={ref}
        {...rest}
        className={[
          'w-full h-11 px-3.5 rounded-lg border bg-white text-[15px] text-neutral-900',
          'placeholder:text-neutral-400',
          'outline-none transition-colors',
          hasError
            ? 'border-red-500 focus:border-red-600 focus:ring-2 focus:ring-red-500/20'
            : 'border-neutral-300 focus:border-neutral-900 focus:ring-2 focus:ring-neutral-900/10',
          className ?? '',
        ].join(' ')}
      />
      {hasError ? (
        <span className="mt-1.5 block text-xs text-red-600">{error}</span>
      ) : hint ? (
        <span className="mt-1.5 block text-xs text-neutral-500">{hint}</span>
      ) : null}
    </label>
  )
})

type PrimaryButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  loading?: boolean
}

export function PrimaryButton({
  loading,
  disabled,
  children,
  className,
  ...rest
}: PrimaryButtonProps) {
  return (
    <button
      {...rest}
      disabled={disabled || loading}
      style={{ backgroundColor: AUTH_BRAND }}
      className={[
        'relative w-full h-11 rounded-lg text-white font-medium text-[15px]',
        'flex items-center justify-center gap-2',
        'transition-all duration-150',
        'hover:brightness-110 active:brightness-95',
        'disabled:opacity-60 disabled:cursor-not-allowed',
        'focus:outline-none focus:ring-4 focus:ring-[color-mix(in_srgb,currentColor_30%,transparent)]',
        className ?? '',
      ].join(' ')}
    >
      {loading ? <Spinner /> : null}
      <span>{children}</span>
    </button>
  )
}

function Spinner() {
  return (
    <svg className="animate-spin h-4 w-4 text-white" viewBox="0 0 24 24" fill="none" aria-hidden>
      <circle cx="12" cy="12" r="10" stroke="currentColor" strokeOpacity="0.3" strokeWidth="4" />
      <path
        d="M22 12a10 10 0 0 1-10 10"
        stroke="currentColor"
        strokeWidth="4"
        strokeLinecap="round"
      />
    </svg>
  )
}

export function Alert({
  tone = 'error',
  children,
}: {
  tone?: 'error' | 'success' | 'info'
  children: ReactNode
}) {
  const styles =
    tone === 'error'
      ? 'bg-red-50 text-red-800 border-red-200'
      : tone === 'success'
        ? 'bg-emerald-50 text-emerald-800 border-emerald-200'
        : 'bg-blue-50 text-blue-800 border-blue-200'
  return (
    <div
      role={tone === 'error' ? 'alert' : 'status'}
      className={[
        'rounded-lg border px-3.5 py-3 text-sm leading-relaxed',
        styles,
      ].join(' ')}
    >
      {children}
    </div>
  )
}
