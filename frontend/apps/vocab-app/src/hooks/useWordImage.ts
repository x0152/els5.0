import { useMutation, useQuery } from '@tanstack/react-query'
import { imageApi, wordImagePrompt } from '../lib/images.ts'

export function useWordImage(text: string, enabled = true) {
  const prompt = wordImagePrompt(text)
  const query = useQuery({
    queryKey: ['illustration', 'square', prompt],
    enabled: enabled && !!text.trim(),
    queryFn: () => imageApi(prompt, false, 'square'),
    refetchInterval: (q) => (q.state.data?.status === 'generating' ? 2500 : false),
    staleTime: Infinity,
  })
  const trigger = useMutation({
    mutationFn: () => imageApi(prompt, true, 'square'),
    onSuccess: () => {
      void query.refetch()
    },
  })
  const status = trigger.isPending ? 'generating' : (query.data?.status ?? 'pending')
  return {
    status,
    url: query.data?.url,
    generate: () => trigger.mutate(),
  }
}
