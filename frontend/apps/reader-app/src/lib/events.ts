import { api } from './api.ts'

export function emitReading(texts: string[], source: Record<string, unknown>) {
  const events = texts
    .map((t) => t.replace(/\s+/g, ' ').trim())
    .filter((t) => t.length >= 20)
    .map((text) => ({ skill: 'reading' as const, text, source }))
  if (events.length) void api.core.ingestCoreEvents({ body: { events } }).catch(() => {})
}
