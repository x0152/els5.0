import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './api.ts'
import { emitReading } from './events.ts'
import type { Book, BookSummary } from './types.ts'

export function useBooks() {
  return useQuery({
    queryKey: ['reader', 'list'],
    queryFn: async (): Promise<BookSummary[]> => {
      const res = await api.reader.listBooks()
      return (res?.items ?? []) as BookSummary[]
    },
    refetchInterval: (q) => {
      const list = q.state.data as BookSummary[] | undefined
      return list?.some((b) => b.status === 'processing') ? 2000 : false
    },
  })
}

export function useBook(id: string) {
  return useQuery({
    queryKey: ['reader', 'book', id],
    enabled: !!id,
    queryFn: async (): Promise<Book> => {
      const res = await api.reader.getBook({ params: { path: { id } } })
      return res as Book
    },
    refetchInterval: (q) => ((q.state.data as Book | undefined)?.status === 'processing' ? 2000 : false),
  })
}

export function useBookContent(url: string | undefined) {
  return useQuery({
    queryKey: ['reader', 'content', url],
    enabled: !!url,
    staleTime: Infinity,
    queryFn: async (): Promise<string> => {
      const res = await fetch(url as string)
      if (!res.ok) throw new Error('failed to load book')
      return res.text()
    },
  })
}

export function useUploadBook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (args: {
      file: File
      title?: string
      author?: string
      description?: string
      cover?: File
      kind: 'book' | 'article'
      groupTitle?: string
    }) => {
      const form = new FormData()
      form.append('file', args.file)
      form.append('kind', args.kind)
      if (args.kind === 'article' && args.groupTitle) form.append('group_title', args.groupTitle)
      if (args.title) form.append('title', args.title)
      if (args.author) form.append('author', args.author)
      if (args.description) form.append('description', args.description)
      if (args.cover) form.append('cover', args.cover)
      return api.reader.uploadBook({ body: form as unknown as never })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reader', 'list'] }),
  })
}

export function useImportArticle() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (args: { url: string; groupTitle?: string }) =>
      api.reader.importArticle({ body: { url: args.url, group_title: args.groupTitle } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reader', 'list'] }),
  })
}

export function useUpdateBook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (args: { id: string; title: string; author: string; description: string; cover?: File }) => {
      const form = new FormData()
      form.append('title', args.title)
      form.append('author', args.author)
      form.append('description', args.description)
      if (args.cover) form.append('cover', args.cover)
      return api.reader.updateBook({ params: { path: { id: args.id } }, body: form as unknown as never })
    },
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['reader', 'list'] })
      qc.invalidateQueries({ queryKey: ['reader', 'book', vars.id] })
    },
  })
}

export function useMarkRead() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (book: BookSummary) => {
      const detail = (await api.reader.getBook({ params: { path: { id: book.id } } })) as Book
      if (detail.content_url) {
        const res = await fetch(detail.content_url)
        if (res.ok) {
          const doc = new DOMParser().parseFromString(await res.text(), 'text/html')
          const texts = Array.from(doc.querySelectorAll('p')).map((p) => p.textContent ?? '')
          emitReading(texts, { app: 'reader', book_id: book.id })
        }
      }
      await api.reader.saveBookProgress({ params: { path: { id: book.id } }, body: { position: book.text_length } })
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reader', 'list'] }),
  })
}

export function useDeleteBook() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.reader.deleteBook({ params: { path: { id } } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['reader', 'list'] }),
  })
}

export function useSaveProgress() {
  return useMutation({
    mutationFn: (args: { id: string; position: number }) =>
      api.reader.saveBookProgress({ params: { path: { id: args.id } }, body: { position: args.position } }),
  })
}

export interface CollectionMeta {
  title: string
  description?: string
  cover_url?: string
}

export function useCollections() {
  return useQuery({
    queryKey: ['reader', 'collections'],
    queryFn: async (): Promise<CollectionMeta[]> => {
      const res = await api.reader.listCollections()
      return res?.items ?? []
    },
  })
}

export function useUpdateCollection() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (args: { title: string; newTitle?: string; description: string; cover?: File }) => {
      const form = new FormData()
      form.append('title', args.title)
      if (args.newTitle && args.newTitle !== args.title) form.append('new_title', args.newTitle)
      form.append('description', args.description)
      if (args.cover) form.append('cover', args.cover)
      return api.reader.updateCollection({ body: form as unknown as never })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['reader', 'collections'] })
      qc.invalidateQueries({ queryKey: ['reader', 'list'] })
    },
  })
}

export function useDeleteCollection() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (ids: string[]) => {
      for (const id of ids) await api.reader.deleteBook({ params: { path: { id } } })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['reader', 'collections'] })
      qc.invalidateQueries({ queryKey: ['reader', 'list'] })
    },
  })
}
