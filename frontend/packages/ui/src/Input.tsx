import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from './cn.ts'

export const inputClass =
  'w-full rounded-lg border border-neutral-200 bg-white px-3 py-2 text-sm text-neutral-900 outline-none transition-colors placeholder:text-neutral-400 focus:border-brand-400 focus:ring-2 focus:ring-brand-100 disabled:bg-neutral-50 disabled:text-neutral-500'

export type InputProps = InputHTMLAttributes<HTMLInputElement>

export const Input = forwardRef<HTMLInputElement, InputProps>(
  function Input({ className, ...props }, ref) {
    return <input ref={ref} className={cn(inputClass, className)} {...props} />
  },
)
