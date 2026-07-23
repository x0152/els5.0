import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api.ts'
import type { AddUnitResult, UnitStatus } from '../lib/types.ts'

const PAGE_SIZE = 60
const listRoot = ['vocab', 'units'] as const

type StatusFilter = '' | UnitStatus

export function useUnits(search: string, status: StatusFilter) {
  return useInfiniteQuery({
    queryKey: [...listRoot, search, status],
    initialPageParam: 0,
    queryFn: ({ pageParam }) =>
      api.vocab.listVocabUnits({
        params: { query: { q: search || undefined, status: status || undefined, limit: PAGE_SIZE, offset: pageParam } },
      }),
    getNextPageParam: (last, pages) => {
      const loaded = pages.reduce((n, p) => n + (p?.items?.length ?? 0), 0)
      return loaded < (last?.total ?? 0) ? loaded : undefined
    },
  })
}

export function usePendingAdds(): string[] {
  const { data } = useQuery<string[]>({
    queryKey: ['vocab', 'pending-adds'],
    queryFn: () => [],
    staleTime: Infinity,
    gcTime: Infinity,
  })
  return data ?? []
}

export function useUnitOccurrences(text: string) {
  return useQuery({
    queryKey: ['vocab', 'occurrences', text],
    queryFn: () => api.vocab.vocabOccurrences({ params: { query: { text } } }),
    enabled: !!text,
    staleTime: 5 * 60 * 1000,
  })
}

export function useAddUnit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (text: string): Promise<AddUnitResult> => {
      const res = await api.vocab.addVocabUnit({ body: { text } })
      return res as AddUnitResult
    },
    onSuccess: (res) => {
      if (res?.correct) qc.invalidateQueries({ queryKey: listRoot })
    },
  })
}

export function useUpdateStatus() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { id: string; status: UnitStatus }) =>
      api.vocab.updateVocabUnitStatus({ params: { path: { id: vars.id } }, body: { status: vars.status } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: listRoot }),
  })
}

export function useDeleteUnit() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.vocab.deleteVocabUnit({ params: { path: { id } } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: listRoot }),
  })
}
