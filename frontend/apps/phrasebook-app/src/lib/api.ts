import { createApi } from '@els/api-client'

const TOKEN_KEY = 'els.auth.token'

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
