import { forwardRef, type HTMLAttributes } from 'react'
import { cn } from './cn.ts'

export const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  function Card({ className, ...props }, ref) {
    return (
      <div
        ref={ref}
        className={cn(
          'rounded-lg border border-neutral-200 bg-white shadow-sm',
          className,
        )}
        {...props}
      />
    )
  },
)

export const CardHeader = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  function CardHeader({ className, ...props }, ref) {
    return (
      <div
        ref={ref}
        className={cn('flex flex-col gap-1 p-4 border-b border-neutral-100', className)}
        {...props}
      />
    )
  },
)

export const CardTitle = forwardRef<HTMLHeadingElement, HTMLAttributes<HTMLHeadingElement>>(
  function CardTitle({ className, ...props }, ref) {
    return (
      <h3
        ref={ref}
        className={cn('text-base font-semibold text-neutral-900', className)}
        {...props}
      />
    )
  },
)

export const CardDescription = forwardRef<
  HTMLParagraphElement,
  HTMLAttributes<HTMLParagraphElement>
>(function CardDescription({ className, ...props }, ref) {
  return (
    <p
      ref={ref}
      className={cn('text-sm text-neutral-500', className)}
      {...props}
    />
  )
})

export const CardContent = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  function CardContent({ className, ...props }, ref) {
    return <div ref={ref} className={cn('p-4', className)} {...props} />
  },
)

export const CardFooter = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  function CardFooter({ className, ...props }, ref) {
    return (
      <div
        ref={ref}
        className={cn('flex items-center gap-2 p-4 border-t border-neutral-100', className)}
        {...props}
      />
    )
  },
)
