export type BookStatus = 'processing' | 'ready' | 'failed'

export type BookKind = 'book' | 'article'

export interface BookSummary {
  id: string
  title: string
  author?: string
  description?: string
  cover_url?: string
  status: BookStatus
  kind: BookKind
  group_title?: string
  text_length: number
  position: number
  percent: number
  created_at: string
}

export interface Book extends BookSummary {
  error?: string
  content_url?: string
}
