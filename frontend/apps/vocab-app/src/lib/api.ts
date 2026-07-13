import { createApi } from '@els/api-client'

const TOKEN_KEY = 'els.auth.token'

/**
 * Module-level api-client singleton used by this feature.
 *
 * Same instance is used in isolated dev (`pnpm dev:<feature>`,
 * proxied through Vite to the backend) and in production
 * (main-app loads the feature via lazy-import; the client makes
 * relative `/api/...` requests against whatever host the SPA was
 * served from).
 *
 * Token is read from localStorage; the dev-harness banner has a
 * paste-token panel that writes there for isolated dev.
 */
export const api = createApi({
  baseUrl: '',
  getToken: () => {
    try {
      return localStorage.getItem(TOKEN_KEY)
    } catch {
      return null
    }
  },
})
