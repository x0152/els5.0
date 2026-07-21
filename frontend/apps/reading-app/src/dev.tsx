import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { DevShell } from '@els/dev-harness'
import { ReadingAppRoutes } from './routes.tsx'
import './index.css'

const rootEl = document.getElementById('root')
if (!rootEl) throw new Error('#root not found')

const queryClient = new QueryClient()

createRoot(rootEl).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <DevShell title="reading-app" initialPath="/">
        <ReadingAppRoutes />
      </DevShell>
    </QueryClientProvider>
  </StrictMode>,
)
