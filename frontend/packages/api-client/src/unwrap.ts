import { ApiError, type ApiErrorBody, type ApiMeta } from './errors.ts'

type FetchResult = {
  data?: unknown
  error?: unknown
  response: Response
}

type SuccessEnvelope = {
  ok: true
  data: unknown
  meta?: ApiMeta
}

type ErrorEnvelope = {
  ok: false
  error: ApiErrorBody
  meta?: ApiMeta
}

type UnwrapData<R> = R extends Promise<infer U>
  ? U extends { data?: infer D }
    ? D extends { ok: boolean; data: infer Inner }
      ? Inner
      : D
    : never
  : never

export async function unwrap<R extends Promise<FetchResult>>(
  promise: R,
): Promise<UnwrapData<R>> {
  let result: FetchResult
  try {
    result = await promise
  } catch (cause) {
    throw new ApiError(0, {
      code: 'network_error',
      message: cause instanceof Error ? cause.message : 'Network request failed',
    })
  }

  const { data, error, response } = result

  if (isErrorEnvelope(error)) {
    throw new ApiError(response.status, error.error, error.meta)
  }
  if (error !== undefined && error !== null) {
    const body = typeof error === 'string' ? error.trim().slice(0, 200) : ''
    throw new ApiError(response.status, {
      code: 'unknown_error',
      message: body || `Request failed with status ${response.status}`,
    })
  }

  if (isErrorEnvelope(data)) {
    throw new ApiError(response.status, data.error, data.meta)
  }
  if (isSuccessEnvelope(data)) {
    return data.data as UnwrapData<R>
  }

  if (!response.ok) {
    throw new ApiError(response.status, {
      code: 'unknown_error',
      message: `Request failed with status ${response.status}`,
    })
  }

  return data as UnwrapData<R>
}

function isSuccessEnvelope(v: unknown): v is SuccessEnvelope {
  return (
    typeof v === 'object' &&
    v !== null &&
    'ok' in v &&
    (v as { ok: unknown }).ok === true &&
    'data' in v
  )
}

function isErrorEnvelope(v: unknown): v is ErrorEnvelope {
  return (
    typeof v === 'object' &&
    v !== null &&
    'ok' in v &&
    (v as { ok: unknown }).ok === false &&
    'error' in v
  )
}
