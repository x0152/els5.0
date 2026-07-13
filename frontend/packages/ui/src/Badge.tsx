import type { HTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from './cn.ts'

const badgeVariants = cva(
  'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium ring-1',
  {
    variants: {
      tone: {
        neutral: 'bg-neutral-100 text-neutral-700 ring-neutral-200',
        brand: 'bg-brand-50 text-brand-700 ring-brand-100',
        success: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
        warning: 'bg-amber-50 text-amber-700 ring-amber-100',
        danger: 'bg-red-50 text-red-700 ring-red-100',
      },
    },
    defaultVariants: {
      tone: 'neutral',
    },
  },
)

export interface BadgeProps
  extends HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ tone, className, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ tone }), className)} {...props} />
}
