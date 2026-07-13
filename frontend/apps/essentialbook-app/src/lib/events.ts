import type { ProduceEvent } from '@els/blocks'
import { api } from './api.ts'

export const produce: ProduceEvent = ({ skill, text, context }) => {
  const t = text.trim()
  if (t.length < 2) return
  void api.core
    .ingestCoreEvents({ body: { events: [{ skill, text: t, context, source: { app: 'essentialbook' } }] } })
    .catch(() => {})
}
