import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { SplashScreen } from '@els/ui'
import { useApps } from '../hooks/useApps'

/**
 * Paths that must stay reachable for any signed-in user even when the
 * backend `accountApps()` catalog omits them (historically `/v1/profile`
 * was only linked from the shell, not listed as an app — that made
 * `allowed === false` here while the index route keeps sending users to
 * `/v1/profile`, causing an infinite `history.replaceState` loop).
 */
const IMPLICIT_ALLOWED_PREFIXES = ['/v1/profile'] as const

const DEFAULT_FALLBACK = '/v1/profile'

/**
 * Route-level permission gate driven by `accountApps()`.
 *
 * Rationale: backend already filters the apps list per-user permissions,
 * so the sidebar shows only what's allowed. But a user can still navigate
 * to a feature URL directly (typed, bookmarked, deep-linked) and would
 * land on a feature whose API calls are guaranteed to 403. This wrapper
 * matches the current path against the per-user app list and redirects
 * to home before any feature module mounts.
 *
 * Disabled apps (`AppOutput.disabled = true`) are also blocked — sidebar
 * already greys them out, this just keeps URL access in sync.
 */
export function RequireApp() {
  const { data: apps, isLoading, error } = useApps()
  const location = useLocation()

  if (isLoading) {
    return <SplashScreen fullscreen subtitle="Checking section access…" />
  }

  // If apps list itself failed to load (network/backend down), fall through
  // to the feature — its own loading/error UI will surface the problem
  // rather than us redirecting to a home that's just as broken.
  if (error) return <Outlet />

  const pathname = location.pathname
  const allowed =
    isImplicitlyAllowed(pathname) ||
    (apps ?? []).some(
      (a) => !a.disabled && pathMatchesApp(pathname, normalizeAppPath(a.to)),
    )
  if (!allowed) {
    // Never send to `/`: `<Route index>` immediately redirects to
    // `/v1/profile`, which can be denied again → replaceState storm.
    const target = firstAllowedDestination(apps)
    return <Navigate to={target} replace />
  }

  return <Outlet />
}

function isImplicitlyAllowed(pathname: string): boolean {
  return IMPLICIT_ALLOWED_PREFIXES.some((p) => pathMatchesApp(pathname, p))
}

function firstAllowedDestination(
  apps: { disabled?: boolean; to: string }[] | undefined,
): string {
  const first = (apps ?? []).find((a) => !a.disabled)?.to
  if (first) return normalizeAppPath(first)
  return DEFAULT_FALLBACK
}

function normalizeAppPath(uri: string): string {
  const t = uri.trim()
  if (!t) return t
  return t.startsWith('/') ? t : `/${t}`
}

/**
 * `app.to` looks like `/v1/profile`; we accept it and any nested path
 * (`/v1/profile/edit`) but reject lookalike prefixes (`/v1/profile-xyz`).
 */
function pathMatchesApp(pathname: string, appPath: string): boolean {
  const p = normalizeAppPath(appPath)
  if (!p) return false
  if (pathname === p) return true
  return pathname.startsWith(p.endsWith('/') ? p : `${p}/`)
}
