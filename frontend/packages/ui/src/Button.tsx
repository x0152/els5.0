import { forwardRef, type ButtonHTMLAttributes } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from './cn.ts'

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 rounded-lg text-sm font-semibold transition-all active:scale-[0.98] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        primary:
          'bg-neutral-900 text-white shadow-sm hover:bg-neutral-800 focus-visible:ring-neutral-900',
        brand:
          'bg-brand-600 text-white shadow-sm shadow-brand-600/25 hover:bg-brand-700 focus-visible:ring-brand-600',
        secondary:
          'bg-white text-neutral-800 shadow-sm ring-1 ring-inset ring-neutral-200 hover:bg-neutral-50 hover:ring-neutral-300 focus-visible:ring-neutral-400',
        ghost:
          'bg-transparent text-neutral-700 hover:bg-neutral-100 focus-visible:ring-neutral-400',
        danger:
          'bg-red-600 text-white shadow-sm hover:bg-red-700 focus-visible:ring-red-500',
      },
      size: {
        sm: 'h-8 px-3',
        md: 'h-9 px-4',
        lg: 'h-10 px-5',
        icon: 'h-9 w-9 p-0',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  },
)

export interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  function Button({ className, variant, size, type = 'button', ...props }, ref) {
    return (
      <button
        ref={ref}
        type={type}
        className={cn(buttonVariants({ variant, size }), className)}
        {...props}
      />
    )
  },
)
