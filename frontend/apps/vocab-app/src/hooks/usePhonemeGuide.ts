import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { canonicalPhoneme, type PhonemeGuideInfo } from '@els/ui'
import { api } from '../lib/api.ts'

export function usePhonemeGuide() {
  const q = useQuery({
    queryKey: ['speech', 'phonemes'],
    queryFn: () => api.speech.listSpeechPhonemes(),
    staleTime: Infinity,
  })
  return useMemo(() => {
    const map = new Map<string, PhonemeGuideInfo>((q.data?.items ?? []).map((p) => [p.symbol, p]))
    return (symbol: string) => map.get(canonicalPhoneme(symbol))
  }, [q.data])
}
