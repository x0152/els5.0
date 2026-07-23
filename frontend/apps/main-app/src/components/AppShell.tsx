import { useEffect, useState } from 'react'
import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { isApiError } from '@els/api-client'
import { cn, MiniPlayerProvider, useAgentView } from '@els/ui'
import { VocabLookupProvider } from '@els/lookup'
import { ChatDock } from '@els/chat-app'
import { AchievementToaster } from '@els/profile-app'
import { api } from '../lib/api'
import { useApps } from '../hooks/useApps'
import { AppTour } from '../onboarding/AppTour'
import { OnboardingWizard } from '../onboarding/OnboardingWizard'
import { SystemTour } from '../onboarding/SystemTour'
import {
  ONBOARDING_RESET_EVENT,
  SYSTEM_TOUR,
  SYSTEM_TOUR_OPEN_EVENT,
  WIZARD_TOUR,
  clearTours,
  isTourDone,
  loadTours,
  markTourDone,
} from '../onboarding/storage'
import { ErrorPage } from './ErrorPage'
import { ImpersonationBanner } from './ImpersonationBanner'
import { Sidebar } from './Sidebar'

export function AppShell() {
  const { isError, error, refetch } = useApps()
  const [chatOpen, setChatOpen] = useState(false)
  const [toursLoaded, setToursLoaded] = useState(false)
  const [wizardOpen, setWizardOpen] = useState(false)
  const [systemTourOpen, setSystemTourOpen] = useState(false)
  const navigate = useNavigate()
  const [, , app, ...rest] = useLocation().pathname.split('/')
  useAgentView(app ? { app, screen: rest.join('/') || 'home' } : null)

  useEffect(() => {
    loadTours()
      .then(() => {
        setWizardOpen(!isTourDone(WIZARD_TOUR))
        setSystemTourOpen(!isTourDone(SYSTEM_TOUR))
        setToursLoaded(true)
      })
      .catch(() => {})
    const onAsk = () => setChatOpen(true)
    const onReset = () => {
      clearTours()
      setWizardOpen(true)
      setSystemTourOpen(true)
    }
    const onSystemTour = () => setSystemTourOpen(true)
    document.addEventListener('els:ask', onAsk)
    window.addEventListener(ONBOARDING_RESET_EVENT, onReset)
    window.addEventListener(SYSTEM_TOUR_OPEN_EVENT, onSystemTour)
    return () => {
      document.removeEventListener('els:ask', onAsk)
      window.removeEventListener(ONBOARDING_RESET_EVENT, onReset)
      window.removeEventListener(SYSTEM_TOUR_OPEN_EVENT, onSystemTour)
    }
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
            chatOpen && app !== 'chat' && 'sm:pr-[460px]',
          )}
        >
          <ImpersonationBanner />
          <Outlet />
        </div>
        <ChatDock open={chatOpen} onOpen={() => setChatOpen(true)} onClose={() => setChatOpen(false)} />
        <VocabLookupProvider api={api} />
        {wizardOpen && (
          <OnboardingWizard
            onDone={() => {
              setWizardOpen(false)
              navigate('/v1/profile')
            }}
          />
        )}
        {!wizardOpen && systemTourOpen && (
          <SystemTour
            onClose={() => {
              markTourDone(SYSTEM_TOUR)
              setSystemTourOpen(false)
            }}
          />
        )}
        <AppTour suspended={!toursLoaded || wizardOpen || systemTourOpen} />
        {!wizardOpen && <AchievementToaster />}
      </div>
    </MiniPlayerProvider>
  )
}
