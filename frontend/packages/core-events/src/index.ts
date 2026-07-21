import type { createApi } from '@els/api-client'

type Api = ReturnType<typeof createApi>
export type Skill = 'reading' | 'writing' | 'speaking' | 'listening'

/** Publish free language samples the user consumed or produced (best effort). */
export function emitTextEvents(api: Api, skill: Skill, texts: string[], source: Record<string, unknown>) {
  const events = texts
    .map((t) => t.replace(/\s+/g, ' ').trim())
    .filter((t) => t.length >= 20)
    .map((text) => ({ skill, text, source }))
  if (events.length) void api.core.ingestCoreEvents({ body: { events } }).catch(() => {})
}

/** Publish targeted outcomes, e.g. words the user failed or nailed (best effort). */
export function emitTargetedEvents(
  api: Api,
  skill: Skill,
  items: { target: string; outcome: 'ok' | 'fail'; context?: string }[],
  source: Record<string, unknown>,
) {
  const events = items
    .filter((i) => i.target.trim())
    .map((i) => ({ skill, target: i.target.trim(), outcome: i.outcome, context: i.context, source }))
  if (events.length) void api.core.ingestCoreEvents({ body: { events } }).catch(() => {})
}
