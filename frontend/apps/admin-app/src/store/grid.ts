import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  applyGrid,
  describeGrid,
  lookupGrid,
  type ApplyGridRequest,
  type ApplyGridResponse,
  type DescribeGridResponse,
  type GridLookupItem,
} from '../lib/grid-client.ts'
import { api } from '../lib/api.ts'

/** Large limit so the admin UI can load “everything” in one request. */
const DEFAULT_LIMIT = 500

/**
 * Read the current account; admin-app needs the permission flags
 * (`is_global_admin`, `impersonation_enabled`) to decide which actions
 * to even render — backend would 403 otherwise.
 */
export function useMe() {
  return useQuery({
    queryKey: ['admin-app', 'me'] as const,
    queryFn: async () => {
      const res = await api.account.accountMe()
      if (!res) throw new Error('account/me returned empty payload')
      return res
    },
    staleTime: 30_000,
  })
}

export function gridDescribeKey(basePath: string) {
  return ['admin-app', 'grid', 'describe', basePath] as const
}

export function gridLookupKey(basePath: string, source: string) {
  return ['admin-app', 'grid', 'lookup', basePath, source] as const
}

export function useGridDescribe(basePath: string) {
  return useQuery({
    queryKey: gridDescribeKey(basePath),
    queryFn: (): Promise<DescribeGridResponse> =>
      describeGrid(basePath, { limit: DEFAULT_LIMIT, offset: 0 }),
    staleTime: 15_000,
  })
}

export function useGridApply(basePath: string) {
  return useMutation({
    mutationFn: (body: ApplyGridRequest): Promise<ApplyGridResponse> =>
      applyGrid(basePath, body),
  })
}

export function useInvalidateGridDescribe(basePath: string) {
  const qc = useQueryClient()
  return () => qc.invalidateQueries({ queryKey: gridDescribeKey(basePath) })
}

/**
 * Loads the full list of available values for a ref source (via `q=""`),
 * to use as dropdown options in the table.
 */
export function useUploadAccountPicture(basePath: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (args: { accountId: string; file: File }) => {
      const form = new FormData()
      form.append('file', args.file)
      await api.account.accountUploadPicture({
        params: { path: { account_id: args.accountId } },
        body: form as unknown as never,
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: gridDescribeKey(basePath) })
    },
  })
}

export function useGridLookupSource(basePath: string, source: string | null) {
  return useQuery({
    enabled: !!source,
    queryKey: source ? gridLookupKey(basePath, source) : ['admin-app', 'grid', 'lookup', 'disabled'],
    queryFn: async (): Promise<GridLookupItem[]> => {
      if (!source) return []
      const res = await lookupGrid(basePath, [{ source, q: '', limit: DEFAULT_LIMIT }])
      return res.queries[0]?.items ?? []
    },
    staleTime: 30_000,
  })
}
