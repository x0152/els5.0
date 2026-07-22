import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'

export interface ProgressItem {
  id: string
  kind: 'checklist' | 'achievement'
  metric: string
  threshold: number
  value: number
  done: boolean
  acked: boolean
}

export const onboardingProgressKey = ['profile-app', 'onboarding-progress'] as const

export function useOnboardingProgress() {
  return useQuery({
    queryKey: onboardingProgressKey,
    queryFn: async (): Promise<ProgressItem[]> => {
      const res = await api.onboarding.onboardingProgress()
      return (res?.items ?? []) as ProgressItem[]
    },
    staleTime: 30_000,
  })
}

export function useAckItems() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (ids: string[]) => api.onboarding.onboardingAck({ body: { ids } }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: onboardingProgressKey })
    },
  })
}
