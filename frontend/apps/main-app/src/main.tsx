import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { queryClient } from './lib/queryClient'
import App from './App'
import './index.css'

const rootEl = document.getElementById('root')
if (!rootEl) throw new Error('#root not found')

createRoot(rootEl).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
)

const splashWindow = window as typeof window & {
  __elsSplashStart?: number
  __elsHangTimer?: number
  __elsKillTimer?: number
}
const loader = document.getElementById('app-loader')
if (loader) {
  const MIN_VISIBLE_MS = 600
  const elapsed = Date.now() - (splashWindow.__elsSplashStart ?? Date.now())
  const delay = Math.max(0, MIN_VISIBLE_MS - elapsed)
  window.setTimeout(() => {
    if (splashWindow.__elsHangTimer) clearTimeout(splashWindow.__elsHangTimer)
    if (splashWindow.__elsKillTimer) clearTimeout(splashWindow.__elsKillTimer)
    loader.classList.add('fade-out')
    window.setTimeout(() => loader.remove(), 1200)
  }, delay)
}
