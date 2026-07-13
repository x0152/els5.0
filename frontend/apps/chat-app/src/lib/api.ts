import { createApi } from '@els/api-client'
import { getToken } from './token'

export const api = createApi({
  baseUrl: '',
  getToken,
})
