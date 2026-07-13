import { request, type APIRequestContext } from '@playwright/test'
import { env } from './env'

/**
 * Thin wrapper over APIRequestContext for quick data seeding.
 *
 * Later we will replace it with a typed client from `frontend/packages/api-client`
 * (codegen from backend/docs/openapi/*.yaml), so we do not keep hand-written types.
 * For now — exactly what is needed for registration/login.
 */

type ApiEnvelope<T> = {
  ok: boolean
  data?: T
  error?: { code: string; message: string }
  meta?: Record<string, unknown>
}

export type SessionOutput = {
  account_id: string
  email: string
  token: string
  expires_at?: string
}

export class ApiClient {
  constructor(private readonly ctx: APIRequestContext) {}

  static async create(): Promise<ApiClient> {
    const ctx = await request.newContext({ baseURL: env.apiUrl })
    return new ApiClient(ctx)
  }

  async dispose(): Promise<void> {
    await this.ctx.dispose()
  }

  async register(input: { email: string; password: string }): Promise<SessionOutput> {
    const res = await this.ctx.post('/api/v1/auth/register', { data: input })
    return unwrap<SessionOutput>(await res.json(), res.status())
  }

  async login(input: { email: string; password: string }): Promise<SessionOutput> {
    const res = await this.ctx.post('/api/v1/auth/login', { data: input })
    return unwrap<SessionOutput>(await res.json(), res.status())
  }

  async me(token: string): Promise<unknown> {
    const res = await this.ctx.get('/api/v1/account/me', {
      headers: { Authorization: `Bearer ${token}` },
    })
    return unwrap(await res.json(), res.status())
  }
}

function unwrap<T>(body: ApiEnvelope<T>, status: number): T {
  if (!body.ok || !body.data) {
    const code = body.error?.code ?? 'unknown'
    const msg = body.error?.message ?? 'no message'
    throw new Error(`API error [${status} ${code}]: ${msg}`)
  }
  return body.data
}

/** Unique email for a specific test — so runs do not conflict. */
export function uniqueEmail(prefix = 'e2e'): string {
  return `${prefix}+${Date.now()}-${Math.floor(Math.random() * 1e6)}@example.test`
}
