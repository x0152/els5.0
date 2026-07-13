import { forwardRef, type SelectHTMLAttributes } from 'react'
import { cn } from './cn.ts'
import { inputClass } from './Input.tsx'

export type SelectProps = SelectHTMLAttributes<HTMLSelectElement>

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  function Select({ className, ...props }, ref) {
    return <select ref={ref} className={cn(inputClass, 'cursor-pointer', className)} {...props} />
  },
)
