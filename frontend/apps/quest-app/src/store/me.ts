import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api.ts'

export function useMe() {
  return useQuery({
    queryKey: ['quest', 'me'],
    queryFn: () => api.account.accountMe(),
    staleTime: 60_000,
  })
}
