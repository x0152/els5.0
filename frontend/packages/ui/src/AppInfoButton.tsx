import { Info } from 'lucide-react'
import { cn } from './cn.ts'

export function AppInfoButton({ className }: { className?: string }) {
  return (
    <button
      type="button"
      title="How this app works"
      onClick={() => window.dispatchEvent(new Event('els:tour:open'))}
      className={cn(
        'rounded-full p-1 text-neutral-300 transition-colors hover:bg-neutral-100 hover:text-brand-600',
        className,
      )}
    >
      <Info className="h-4.5 w-4.5" />
    </button>
  )
}
