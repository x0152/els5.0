import type { HTMLAttributes, SVGProps } from 'react'
import { cn } from './cn.ts'

const LEFT_PAGE = 'M30 16 L11 19 Q7 32 11 45 L30 48 Z'
const RIGHT_PAGE = 'M34 16 L53 19 Q57 32 53 45 L34 48 Z'

export const LogoPath = `${LEFT_PAGE} ${RIGHT_PAGE}`

export interface LogoProps extends Omit<SVGProps<SVGSVGElement>, 'fill'> {
  color?: string
}

export function Logo({ color = '#059669', ...props }: LogoProps) {
  return (
    <svg
      viewBox="0 0 64 64"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
      {...props}
    >
      <mask id="elsBookMask">
        <rect width="64" height="64" fill="black" />
        <path d={LEFT_PAGE} fill="white" />
        <path d={RIGHT_PAGE} fill="white" />
        <g stroke="black" strokeWidth="2.4" strokeLinecap="round">
          <line x1="15" y1="26" x2="27" y2="24" />
          <line x1="15" y1="32" x2="27" y2="30" />
          <line x1="15" y1="38" x2="27" y2="36" />
          <line x1="37" y1="24" x2="49" y2="26" />
          <line x1="37" y1="30" x2="49" y2="32" />
          <line x1="37" y1="36" x2="49" y2="38" />
        </g>
      </mask>
      <rect width="64" height="64" fill={color} mask="url(#elsBookMask)" />
    </svg>
  )
}

export interface LogoWordmarkProps extends HTMLAttributes<HTMLDivElement> {
  wordmark?: string
  subtitle?: string
  color?: string
  markClassName?: string
  subtitleClassName?: string
}

export function LogoWordmark({
  wordmark = 'ELS',
  subtitle,
  color = '#059669',
  className,
  markClassName,
  subtitleClassName,
  ...props
}: LogoWordmarkProps) {
  return (
    <div
      className={cn('inline-flex flex-col items-center leading-none', className)}
      {...props}
    >
      <span
        className={cn('text-2xl font-extrabold tracking-tight', markClassName)}
        style={{ color }}
      >
        {wordmark}
      </span>
      {subtitle ? (
        <span
          className={cn(
            'mt-2 text-[10px] font-semibold uppercase tracking-[0.2em] opacity-70',
            subtitleClassName,
          )}
          style={{ color }}
        >
          {subtitle}
        </span>
      ) : null}
    </div>
  )
}
