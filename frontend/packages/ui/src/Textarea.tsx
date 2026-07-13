import { forwardRef, type TextareaHTMLAttributes } from 'react'
import { cn } from './cn.ts'
import { inputClass } from './Input.tsx'

export type TextareaProps = TextareaHTMLAttributes<HTMLTextAreaElement>

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  function Textarea({ className, ...props }, ref) {
    return <textarea ref={ref} className={cn(inputClass, 'resize-none', className)} {...props} />
  },
)
