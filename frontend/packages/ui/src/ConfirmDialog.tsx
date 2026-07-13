import type { ReactNode } from 'react'
import { Button } from './Button.tsx'
import { Modal } from './Modal.tsx'

export interface ConfirmDialogProps {
  title: ReactNode
  description?: ReactNode
  confirmLabel?: ReactNode
  pending?: boolean
  onConfirm: () => void
  onClose: () => void
}

export function ConfirmDialog({
  title,
  description,
  confirmLabel = 'Delete',
  pending,
  onConfirm,
  onClose,
}: ConfirmDialogProps) {
  return (
    <Modal onClose={() => !pending && onClose()} title={title} className="max-w-sm">
      {description && <p className="text-sm text-neutral-600">{description}</p>}
      <div className="mt-6 flex justify-end gap-3">
        <Button variant="secondary" onClick={onClose} disabled={pending}>
          Cancel
        </Button>
        <Button variant="danger" onClick={onConfirm} disabled={pending}>
          {confirmLabel}
        </Button>
      </div>
    </Modal>
  )
}
