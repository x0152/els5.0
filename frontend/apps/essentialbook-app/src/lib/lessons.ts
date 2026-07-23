import { useSyncExternalStore } from 'react'
import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './api.ts'
import { practiceApi } from './practice.ts'
import type { Lesson } from './types.ts'

export const SERIES = 'essentialbook'
export const DEFAULT_BOOK = 'essentialbook'

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

export type BookInfo = { slug: string; series: string; level?: string; title: string; description?: string }

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

export function useLessons(book: string) {
  return useQuery({
    queryKey: [book, 'lessons'],
    queryFn: async (): Promise<Lesson[]> => {
      const res = await api.learn.listChapters({ params: { path: { book } } })
      return (res?.items ?? []).map((l) => ({ ...l, words: l.words ?? [] }))
    },
    refetchInterval: (q) => (q.state.data?.some((l) => l.status === 'generating') ? 2000 : false),
  })
}

export function useLesson(book: string, number: number) {
  return useQuery({
    queryKey: [book, 'lesson', number],
    enabled: Number.isFinite(number) && number > 0,
    queryFn: () => api.learn.getChapter({ params: { path: { book, number } } }),
  })
}

export function useGenerateLesson(book: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (topic: string) =>
      api.learn.generateChapter({ params: { path: { book } }, body: { topic } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: [book, 'lessons'] }),
  })
}

export function useDeleteLesson(book: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (number: number) =>
      api.learn.deleteChapter({ params: { path: { book, number } } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: [book, 'lessons'] }),
  })
}
