import type { ReactNode } from 'react'
import { Button } from '@els/ui'
import { ArrowRight } from 'lucide-react'

export function StepShell({ children }: { children: ReactNode }) {
  return <section className="flex flex-1 flex-col gap-5 rounded-3xl border border-neutral-200 bg-white p-6 shadow-sm sm:p-8">{children}</section>
}

export function ContinueButton({ onClick, label = 'Continue', disabled }: { onClick: () => void; label?: string; disabled?: boolean }) {
  return (
    <div className="mt-auto flex justify-end border-t border-neutral-100 pt-4">
      <Button variant="brand" onClick={onClick} disabled={disabled}>
        {label} <ArrowRight className="h-4 w-4" />
      </Button>
    </div>
  )
}
