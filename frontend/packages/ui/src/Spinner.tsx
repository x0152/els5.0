import { Loader2 } from 'lucide-react'
import { cn } from './cn.ts'

export function Spinner({ className }: { className?: string }) {
  return <Loader2 className={cn('h-6 w-6 animate-spin', className)} />
}

export function LoadingState({ className }: { className?: string }) {
  return (
    <div className={cn('flex justify-center py-16 text-brand-500', className)}>
      <Spinner />
    </div>
  )
}
