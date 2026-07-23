import { useSyncExternalStore } from 'react'
import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './api.ts'
import { practiceApi } from './practice.ts'
import type { Unit } from './types.ts'

export const SERIES = 'grammarbook'
export const DEFAULT_BOOK = 'grammarbook'

export type BookInfo = { slug: string; series: string; level?: string; title: string; description?: string }

const BOOK_KEY = `${SERIES}.activeBook`
let bookListeners: (() => void)[] = []

export function setActiveBook(slug: string) {
  localStorage.setItem(BOOK_KEY, slug)
  bookListeners.forEach((l) => l())
}

export function useActiveBook() {
  return useSyncExternalStore(
    (cb) => {
      bookListeners.push(cb)
      return () => {
        bookListeners = bookListeners.filter((l) => l !== cb)
      }
    },
    () => localStorage.getItem(BOOK_KEY) ?? DEFAULT_BOOK,
  )
}

export function useBooks() {
  return useQuery({
    queryKey: ['books', SERIES],
    queryFn: async (): Promise<BookInfo[]> => {
      const res = await api.learn.listLearnBooks()
      return (res?.items ?? []).filter((b) => b.series === SERIES)
    },
    staleTime: 60_000,
  })
}

export function useMainCompletion(book: string, numbers: number[]) {
  const results = useQueries({
    queries: numbers.map((n) => ({
      queryKey: ['practice', 'progress', book, n, 'main'],
      queryFn: () => practiceApi.getProgress(book, n, 'main'),
      staleTime: 15_000,
    })),
  })
  const state: Record<number, 'done' | 'started'> = {}
  numbers.forEach((n, i) => {
    const d = results[i]?.data
    if (d?.completed) state[n] = 'done'
    else if (d && Object.keys(d.answers ?? {}).length > 0) state[n] = 'started'
  })
  return state
}

export function useUnits(book: string) {
  return useQuery({
    queryKey: [book, 'units'],
    queryFn: async (): Promise<Unit[]> => {
      const res = await api.learn.listChapters({ params: { path: { book } } })
      return res?.items ?? []
    },
    refetchInterval: (q) => (q.state.data?.some((u) => u.status === 'generating') ? 2000 : false),
  })
}

export function useUnit(book: string, number: number) {
  return useQuery({
    queryKey: [book, 'unit', number],
    enabled: Number.isFinite(number) && number > 0,
    queryFn: () => api.learn.getChapter({ params: { path: { book, number } } }),
  })
}

export function useGenerateUnit(book: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (topic: string) =>
      api.learn.generateChapter({ params: { path: { book } }, body: { topic } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: [book, 'units'] }),
  })
}

export function useDeleteUnit(book: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (number: number) =>
      api.learn.deleteChapter({ params: { path: { book, number } } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: [book, 'units'] }),
  })
}
