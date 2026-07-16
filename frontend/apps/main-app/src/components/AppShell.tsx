import { useEffect, useState } from 'react'
import { Outlet, useNavigate } from 'react-router-dom'
import { isApiError } from '@els/api-client'
import { cn, MiniPlayerProvider } from '@els/ui'
import { VocabLookupProvider } from '@els/lookup'
import { ChatDock } from '@els/chat-app'
import { api } from '../lib/api'
import { useApps } from '../hooks/useApps'
import { ErrorPage } from './ErrorPage'
import { ImpersonationBanner } from './ImpersonationBanner'
import { Sidebar } from './Sidebar'

export function AppShell() {
  const { isError, error, refetch } = useApps()
  const [chatOpen, setChatOpen] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const onAsk = () => setChatOpen(true)
    document.addEventListener('els:ask', onAsk)
    return () => document.removeEventListener('els:ask', onAsk)
  }, [])

  if (isError) {
    const details = isApiError(error)
      ? `${error.status} ${error.code}: ${error.message}`
      : error instanceof Error
        ? error.message
        : undefined

    return (
      <ErrorPage
        title="Service temporarily unavailable"
        description="Failed to load the list of applications. Check your connection and try again."
        details={details}
        onRetry={refetch}
      />
    )
  }

  return (
    <MiniPlayerProvider onNavigate={navigate}>
      <div className="h-dvh bg-neutral-50 text-neutral-900 overflow-hidden">
        <Sidebar />
        <div
          className={cn(
            'md:pl-28 pt-14 md:pt-0 pb-[calc(4rem+env(safe-area-inset-bottom,0px))] md:pb-0 h-dvh overflow-y-auto transition-[padding] duration-300 ease-out',
            chatOpen && 'sm:pr-[460px]',
          )}
        >
          <ImpersonationBanner />
          <Outlet />
        </div>
        <ChatDock open={chatOpen} onOpen={() => setChatOpen(true)} onClose={() => setChatOpen(false)} />
        <VocabLookupProvider api={api} />
      </div>
    </MiniPlayerProvider>
  )
}
