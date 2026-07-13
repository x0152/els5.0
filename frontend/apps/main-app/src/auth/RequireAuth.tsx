import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { SplashScreen } from '@els/ui'
import { useAuth } from './AuthContext'

export function RequireAuth() {
  const { isAuthenticated, isLoading } = useAuth()
  const location = useLocation()

  if (isLoading) {
    return <SplashScreen fullscreen subtitle="Checking session…" />
  }
  if (!isAuthenticated) {
    const next = location.pathname + location.search
    return <Navigate to={`/login?next=${encodeURIComponent(next)}`} replace />
  }
  return <Outlet />
}
