import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { isApiError } from '@els/api-client'
import { api } from '../lib/api'
import { clearImpersonation } from './impersonation'
import { clearToken, getToken, setToken } from './token'

export interface CurrentUser {
  accountId: string
  entityId: string
  email: string
  firstName: string
  lastName: string
  role: string
  status: string
  isGlobalAdmin: boolean
  /** Backend allows this account to issue impersonation tokens. */
  impersonationEnabled: boolean
  displayName: string
  initials: string
  pictureUrl?: string
}

interface AuthContextValue {
  user: CurrentUser | null
  isLoading: boolean
  isAuthenticated: boolean
  signInWithToken: (token: string) => Promise<void>
  signOut: () => Promise<void>
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside <AuthProvider>')
  return ctx
}

function toUser(raw: {
  account_id: string
  entity_id: string
  email: string
  first_name: string
  last_name: string
  role: string
  status: string
  is_global_admin: boolean
  impersonation_enabled: boolean
  picture_url?: string
}): CurrentUser {
  const displayName = [raw.first_name, raw.last_name].filter(Boolean).join(' ') || raw.email
  const initials =
    `${raw.first_name?.[0] ?? ''}${raw.last_name?.[0] ?? ''}`.toUpperCase() ||
    raw.email[0]?.toUpperCase() ||
    '?'
  return {
    accountId: raw.account_id,
    entityId: raw.entity_id,
    email: raw.email,
    firstName: raw.first_name,
    lastName: raw.last_name,
    role: raw.role,
    status: raw.status,
    isGlobalAdmin: raw.is_global_admin,
    impersonationEnabled: raw.impersonation_enabled,
    displayName,
    initials,
    pictureUrl: raw.picture_url || undefined,
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [isLoading, setLoading] = useState<boolean>(() => !!getToken())

  const loadMe = useCallback(async (): Promise<void> => {
    const token = getToken()
    if (!token) {
      setUser(null)
      setLoading(false)
      return
    }
    try {
      const me = await api.account.accountMe()
      if (!me) throw new Error('account/me returned empty payload')
      setUser(toUser(me))
    } catch (err) {
      if (isApiError(err) && (err.status === 401 || err.status === 403)) {
        clearToken()
      }
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadMe()
  }, [loadMe])

  const signInWithToken = useCallback(
    async (token: string): Promise<void> => {
      setToken(token)
      setLoading(true)
      await loadMe()
    },
    [loadMe],
  )

  const signOut = useCallback(async (): Promise<void> => {
    try {
      await api.auth.logout()
    } catch {
    }
    clearImpersonation()
    clearToken()
    setUser(null)
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isLoading,
      isAuthenticated: !!user,
      signInWithToken,
      signOut,
      refresh: loadMe,
    }),
    [user, isLoading, signInWithToken, signOut, loadMe],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
