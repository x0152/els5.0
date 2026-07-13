import { setToken } from './token'

const ORIGINAL_TOKEN_KEY = 'els.auth.original_token'
const ORIGINAL_LABEL_KEY = 'els.auth.original_label'

export interface ImpersonationState {
  originalToken: string
  originalLabel: string
}

export function getImpersonation(): ImpersonationState | null {
  try {
    const originalToken = sessionStorage.getItem(ORIGINAL_TOKEN_KEY)
    const originalLabel = sessionStorage.getItem(ORIGINAL_LABEL_KEY) ?? ''
    if (!originalToken) return null
    return { originalToken, originalLabel }
  } catch {
    return null
  }
}

export function isImpersonating(): boolean {
  return getImpersonation() !== null
}

export function startImpersonation(args: {
  originalToken: string
  originalLabel: string
  newToken: string
}): void {
  try {
    sessionStorage.setItem(ORIGINAL_TOKEN_KEY, args.originalToken)
    sessionStorage.setItem(ORIGINAL_LABEL_KEY, args.originalLabel)
  } catch {
    // ignore storage write failures (private mode etc.)
  }
  setToken(args.newToken)
}

export function stopImpersonation(): string | null {
  const state = getImpersonation()
  if (!state) return null
  try {
    sessionStorage.removeItem(ORIGINAL_TOKEN_KEY)
    sessionStorage.removeItem(ORIGINAL_LABEL_KEY)
  } catch {
    // ignore storage write failures (private mode etc.)
  }
  setToken(state.originalToken)
  return state.originalToken
}

export function clearImpersonation(): void {
  try {
    sessionStorage.removeItem(ORIGINAL_TOKEN_KEY)
    sessionStorage.removeItem(ORIGINAL_LABEL_KEY)
  } catch {
    // ignore storage write failures (private mode etc.)
  }
}
