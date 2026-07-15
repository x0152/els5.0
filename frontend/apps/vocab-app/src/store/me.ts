import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api.ts'

export function useShowTranslations(): boolean {
  const q = useQuery({
    queryKey: ['vocab', 'me'],
    queryFn: () => api.account.accountMe(),
    staleTime: 10_000,
  })
  return q.data?.show_translations ?? true
}
