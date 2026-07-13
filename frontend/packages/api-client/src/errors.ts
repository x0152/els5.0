export interface ApiErrorDetail {
  field?: string
  message: string
}

export interface ApiErrorBody {
  code: string
  message: string
  details?: ApiErrorDetail[] | null
}

export interface ApiMeta {
  request_id?: string
  pagination?: {
    limit: number
    offset: number
    total: number
    has_more: boolean
  }
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly details: ApiErrorDetail[]
  readonly meta?: ApiMeta
  readonly requestId?: string

  constructor(status: number, body: ApiErrorBody, meta?: ApiMeta) {
    super(body.message)
    this.name = 'ApiError'
    this.status = status
    this.code = body.code
    this.details = body.details ?? []
    this.meta = meta
    this.requestId = meta?.request_id
  }
}

export function isApiError(err: unknown): err is ApiError {
  return err instanceof ApiError
}
