import { useEffect, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { cn } from './cn.ts'
import { Button } from './Button.tsx'

export interface ModalProps {
  onClose: () => void
  title?: ReactNode
  className?: string
  children: ReactNode
}

export function Modal({ onClose, title, className, children }: ModalProps) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', onKey)
      document.body.style.overflow = prev
    }
  }, [onClose])

  return createPortal(
    <div
      onClick={onClose}
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-0 backdrop-blur-sm sm:items-center sm:p-6"
    >
      <div
        role="dialog"
        aria-modal="true"
        onClick={(e) => e.stopPropagation()}
        className={cn(
          'max-h-[90dvh] w-full max-w-lg overflow-y-auto rounded-t-3xl bg-white p-6 shadow-2xl sm:rounded-3xl',
          className,
        )}
      >
        {title !== undefined && (
          <div className="mb-5 flex items-center justify-between">
            <h2 className="flex items-center gap-2 text-lg font-semibold text-neutral-900">{title}</h2>
            <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close">
              <X className="h-4 w-4" />
            </Button>
          </div>
        )}
        {children}
      </div>
    </div>,
    document.body,
  )
}
