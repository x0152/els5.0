import type { ProduceEvent } from '@els/blocks'
import { api } from './api.ts'

export function reviewed(target: string, outcome: 'ok' | 'fail', skill: 'reading' | 'writing') {
  void api.core
    .ingestCoreEvents({ body: { events: [{ target, outcome, skill, source: { app: 'vocab', feature: 'cards' } }] } })
    .catch(() => {})
}

export function pronounced(target: string, outcome: 'ok' | 'fail') {
  void api.core
    .ingestCoreEvents({
      body: { events: [{ target, outcome, skill: 'speaking', source: { app: 'vocab', feature: 'pronunciation' } }] },
    })
    .catch(() => {})
}

export const produce: ProduceEvent = ({ skill, text, context }) => {
  const t = text.trim()
  if (t.length < 2) return
  void api.core
    .ingestCoreEvents({ body: { events: [{ skill, text: t, context, source: { app: 'vocab' } }] } })
    .catch(() => {})
}
