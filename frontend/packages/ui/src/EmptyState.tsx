import type { ReactNode } from 'react'
import { TriangleAlert } from 'lucide-react'
import { cn } from './cn.ts'

export interface EmptyStateProps {
  icon?: ReactNode
  iconClassName?: string
  title: ReactNode
  description?: ReactNode
  action?: ReactNode
  className?: string
}

export function EmptyState({ icon, iconClassName, title, description, action, className }: EmptyStateProps) {
  return (
    <div className={cn('grid place-items-center rounded-3xl bg-white py-20 ring-1 ring-neutral-200', className)}>
      <div className="px-6 text-center">
        {icon && (
          <div
            className={cn(
              'mx-auto mb-4 grid h-16 w-16 place-items-center rounded-2xl bg-brand-50 text-brand-600',
              iconClassName,
            )}
          >
            {icon}
          </div>
        )}
        <h3 className="text-lg font-semibold text-neutral-900">{title}</h3>
        {description && <p className="mt-1 text-sm text-neutral-500">{description}</p>}
        {action && <div className="mt-5 flex justify-center">{action}</div>}
      </div>
    </div>
  )
}

export interface ErrorStateProps {
  title?: ReactNode
  description?: ReactNode
  action?: ReactNode
  className?: string
}

export function ErrorState({ title = 'Something went wrong', description, action, className }: ErrorStateProps) {
  return (
    <EmptyState
      icon={<TriangleAlert className="h-8 w-8" />}
      iconClassName="bg-red-50 text-red-600"
      title={title}
      description={description}
      action={action}
      className={className}
    />
  )
}
