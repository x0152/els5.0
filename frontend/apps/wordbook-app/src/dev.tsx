import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { DevShell } from '@els/dev-harness'
import { WordbookAppRoutes } from './routes.tsx'
import './index.css'

const rootEl = document.getElementById('root')
if (!rootEl) throw new Error('#root not found')

createRoot(rootEl).render(
  <StrictMode>
    <DevShell title="wordbook-app" initialPath="/">
      <WordbookAppRoutes />
    </DevShell>
  </StrictMode>,
)
