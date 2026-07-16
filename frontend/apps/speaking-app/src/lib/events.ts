import { api } from './api.ts'

export function spoken(text: string, score: number) {
  void api.core
    .ingestCoreEvents({
      body: { events: [{ skill: 'speaking', text, meta: { score }, source: { app: 'speaking', feature: 'practice' } }] },
    })
    .catch(() => {})
}
