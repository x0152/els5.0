import { type ReactNode, type PropsWithChildren } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'
import { MemoryRouter, BrowserRouter } from 'react-router-dom'
import { DevErrorBoundary } from './ErrorBoundary.tsx'
import { TokenPanel } from './TokenPanel.tsx'

export interface DevShellProps {
  /** First URL the in-memory router lands on. Ignored for `routerMode='browser'`. */
  initialPath?: string
  /** `'browser'` (default) preserves URL across reloads; `'memory'` is fully isolated. */
  routerMode?: 'browser' | 'memory'
  /** Label shown in the dev banner. Defaults to `document.title`. */
  title?: string
  /** Override the bundled QueryClient if a feature wants its own cache. */
  queryClient?: QueryClient
  /** Hide the dev banner (token panel + title). Default: shown. */
  showBanner?: boolean
  /** Mount React Query Devtools (bottom-right). Default: shown. */
  showQueryDevtools?: boolean
  children: ReactNode
}

const defaultClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, retry: false },
  },
})

export function DevShell({
  initialPath = '/',
  routerMode = 'browser',
  title,
  queryClient = defaultClient,
  showBanner = true,
  showQueryDevtools = true,
  children,
}: DevShellProps) {
  const resolvedTitle = title ?? (typeof document !== 'undefined' ? document.title : '')

  return (
    <QueryClientProvider client={queryClient}>
      <Router mode={routerMode} initialPath={initialPath}>
        <div className="h-dvh flex flex-col bg-neutral-50 text-neutral-900 overflow-hidden">
          {showBanner && (
            <div className="border-b border-neutral-200 bg-yellow-50 px-4 py-1.5 text-[11px] font-mono text-neutral-700 shrink-0 flex items-center gap-3">
              <span>🧪 dev-harness · {resolvedTitle}</span>
              <span className="text-neutral-300">|</span>
              <TokenPanel />
              <span className="ml-auto text-neutral-400">
                {routerMode === 'memory' ? 'memory router' : 'browser router'}
              </span>
            </div>
          )}
          <main className="flex-1 min-h-0 flex flex-col">
            <DevErrorBoundary>{children}</DevErrorBoundary>
          </main>
        </div>
      </Router>
      {showQueryDevtools && (
        <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-right" />
      )}
    </QueryClientProvider>
  )
}

function Router({
  mode,
  initialPath,
  children,
}: PropsWithChildren<{ mode: 'browser' | 'memory'; initialPath: string }>) {
  if (mode === 'memory') {
    return <MemoryRouter initialEntries={[initialPath]}>{children}</MemoryRouter>
  }
  return <BrowserRouter>{children}</BrowserRouter>
}
