import { createApi } from '@els/api-client'
import { getToken } from '../auth/token'

export const api = createApi({
  baseUrl: '',
  getToken,
})
