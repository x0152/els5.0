import { api } from './api.ts'

export function spoken(text: string, score: number) {
  void api.core
    .ingestCoreEvents({
      body: {
        events: [
          {
            skill: 'speaking',
            text,
            outcome: score >= 60 ? 'ok' : 'fail',
            meta: { score },
            source: { app: 'speaking', feature: 'practice' },
          },
        ],
      },
    })
    .catch(() => {})
}
