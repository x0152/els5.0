const TOKEN_KEY = 'els.auth.token'
const ORIGINAL_TOKEN_KEY = 'els.auth.original_token'
const ORIGINAL_LABEL_KEY = 'els.auth.original_label'

export function readToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function writeToken(token: string): void {
  try {
    localStorage.setItem(TOKEN_KEY, token)
  } catch {
    // ignore storage write failures (private mode etc.)
  }
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
  writeToken(args.newToken)
}

export function isImpersonating(): boolean {
  try {
    return sessionStorage.getItem(ORIGINAL_TOKEN_KEY) !== null
  } catch {
    return false
  }
}
