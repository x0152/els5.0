import { api } from './api.ts'

export function emitWriting(text: string, source: Record<string, unknown>) {
  const t = text.trim()
  if (t.length < 2) return
  void api.core.ingestCoreEvents({ body: { events: [{ skill: 'writing', text: t, source }] } }).catch(() => {})
}
