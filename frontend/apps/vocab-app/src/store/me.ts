import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api.ts'

export function useMe() {
  return useQuery({
    queryKey: ['vocab', 'me'],
    queryFn: () => api.account.accountMe(),
    staleTime: 10_000,
  })
}

export function useShowTranslations(): boolean {
  return useMe().data?.show_translations ?? true
}
