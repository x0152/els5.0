import { useState } from 'react'
import { AlertTriangle, RefreshCw, RotateCcw } from 'lucide-react'
import { Button, LogoWordmark } from '@els/ui'

export interface ErrorPageProps {
  title?: string
  description?: string
  details?: string
  onRetry?: () => void | Promise<unknown>
}

export function ErrorPage({
  title = 'Service temporarily unavailable',
  description = 'Failed to load the list of applications. Please try again or refresh the page later.',
  details,
  onRetry,
}: ErrorPageProps) {
  const [retrying, setRetrying] = useState(false)

  const handleRetry = async () => {
    if (!onRetry || retrying) return
    try {
      setRetrying(true)
      await onRetry()
    } finally {
      setRetrying(false)
    }
  }

  return (
    <div className="min-h-dvh flex flex-col items-center justify-center bg-neutral-50 text-neutral-900 px-6">
      <div className="absolute top-8 left-1/2 -translate-x-1/2">
        <LogoWordmark wordmark="ELS" subtitle="English Learning Studio" markClassName="text-3xl" />
      </div>

      <div className="max-w-md w-full flex flex-col items-center text-center">
        <div className="w-20 h-20 rounded-full bg-brand-50 text-brand-600 flex items-center justify-center mb-6 ring-1 ring-brand-100">
          <AlertTriangle size={36} strokeWidth={1.75} />
        </div>

        <h1 className="text-2xl font-semibold tracking-tight mb-2">{title}</h1>
        <p className="text-sm text-neutral-500 leading-relaxed mb-6">{description}</p>

        {details && (
          <pre className="w-full text-left text-[11px] text-neutral-500 bg-white border border-neutral-200 rounded-lg p-3 mb-6 overflow-auto max-h-32">
            {details}
          </pre>
        )}

        <div className="flex items-center gap-3">
          {onRetry && (
            <Button
              variant="primary"
              size="md"
              onClick={handleRetry}
              disabled={retrying}
            >
              <RotateCcw size={16} className={retrying ? 'animate-spin' : undefined} />
              {retrying ? 'Retrying…' : 'Try again'}
            </Button>
          )}
          <Button
            variant="secondary"
            size="md"
            onClick={() => window.location.reload()}
          >
            <RefreshCw size={16} />
            Refresh page
          </Button>
        </div>
      </div>
    </div>
  )
}
