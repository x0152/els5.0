import { ApiError, type ApiErrorBody, type ApiMeta } from '@els/api-client'

const TOKEN_KEY = 'els.auth.token'

function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

async function request<T>(method: 'GET' | 'POST', path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = { accept: 'application/json' }
  if (body !== undefined) headers['content-type'] = 'application/json'
  const token = getToken()
  if (token) headers.authorization = `Bearer ${token}`

  let res: Response
  try {
    res = await fetch(path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
    })
  } catch (cause) {
    throw new ApiError(0, {
      code: 'network_error',
      message: cause instanceof Error ? cause.message : 'Network request failed',
    })
  }

  let payload: unknown
  try {
    payload = await res.json()
  } catch {
    throw new ApiError(res.status, {
      code: 'invalid_response',
      message: `Invalid JSON from ${path}`,
    })
  }

  if (isSuccess(payload)) return payload.data as T
  if (isFailure(payload)) throw new ApiError(res.status, payload.error, payload.meta)
  if (!res.ok) {
    throw new ApiError(res.status, {
      code: 'unknown_error',
      message: `Request failed with status ${res.status}`,
    })
  }
  return payload as T
}

function isSuccess(v: unknown): v is { ok: true; data: unknown; meta?: ApiMeta } {
  return typeof v === 'object' && v !== null && (v as { ok?: unknown }).ok === true && 'data' in v
}

function isFailure(
  v: unknown,
): v is { ok: false; error: ApiErrorBody; meta?: ApiMeta } {
  return (
    typeof v === 'object' && v !== null && (v as { ok?: unknown }).ok === false && 'error' in v
  )
}

/* ----------------------------- DTOs: describe ----------------------------- */

export type GridColumnType =
  | 'text'
  | 'email'
  | 'int'
  | 'float'
  | 'bool'
  | 'date'
  | 'datetime'
  | 'enum'
  | 'ref'

export interface GridEnumOption {
  value: string
  label: string
}

export interface GridRefSpec {
  source: string
  key_field: string
  label_field: string
  multi?: boolean
}

export interface GridConstraints {
  min_length?: number
  max_length?: number
  min?: number
  max?: number
  pattern?: string
  unique?: boolean
}

export interface GridColumn {
  id: string
  title: string
  type: GridColumnType
  required?: boolean
  readonly?: boolean
  enum?: GridEnumOption[]
  ref?: GridRefSpec
  constraints?: GridConstraints
}

export interface GridRow {
  id: string
  base_version: number
  cells: Record<string, unknown>
}

export type GridRefsHydrated = Record<string, Record<string, string>>

export interface DescribeGridResponse {
  schema_version: string
  columns: GridColumn[]
  sources: string[]
  rows: GridRow[]
  total: number
  limit: number
  offset: number
  refs_hydrated: GridRefsHydrated
  generated_at: string
}

/* ------------------------------ DTOs: apply ------------------------------- */

export type GridOpKind = 'create' | 'update' | 'delete'

export interface GridOp {
  kind: GridOpKind
  temp_id?: string
  id?: string
  base_version?: number
  data?: Record<string, unknown>
}

export interface ApplyGridRequest {
  schema_version: string
  operations: GridOp[]
}

export interface GridOpResult {
  index: number
  kind: GridOpKind
  temp_id?: string
  id: string
  base_version: number
}

export interface GridOpError {
  index: number
  temp_id?: string
  id?: string
  code: string
  field?: string
  message: string
}

export interface ApplyGridResponse {
  schema_version: string
  applied: GridOpResult[]
  failed: GridOpError[]
}

/* ------------------------------ DTOs: lookup ------------------------------ */

export interface GridLookupQuery {
  source: string
  values?: string[]
  q?: string
  limit?: number
  cursor?: string
}

export interface GridLookupItem {
  key: string
  label: string
}

export interface GridLookupResolution {
  input: string
  key?: string
  label?: string
  matched_by?: string
  resolved: boolean
}

export interface GridLookupQueryResult {
  source: string
  resolutions?: GridLookupResolution[]
  items?: GridLookupItem[]
  next_cursor?: string
}

export interface LookupGridResponse {
  queries: GridLookupQueryResult[]
}

/* -------------------------------- methods --------------------------------- */

export function describeGrid(
  basePath: string,
  params: { limit?: number; offset?: number } = {},
): Promise<DescribeGridResponse> {
  const qs = new URLSearchParams()
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.offset !== undefined) qs.set('offset', String(params.offset))
  const suffix = qs.toString()
  return request<DescribeGridResponse>(
    'GET',
    `${basePath}/grid${suffix ? `?${suffix}` : ''}`,
  )
}

export function applyGrid(
  basePath: string,
  body: ApplyGridRequest,
): Promise<ApplyGridResponse> {
  return request<ApplyGridResponse>('POST', `${basePath}/grid`, body)
}

export function lookupGrid(
  basePath: string,
  queries: GridLookupQuery[],
): Promise<LookupGridResponse> {
  return request<LookupGridResponse>('POST', `${basePath}/grid/lookup`, { queries })
}
