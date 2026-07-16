import { api } from './api.ts'

const stripTags = (s: string) => s.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim()

export function emitListening(text: string, source: Record<string, unknown>) {
  const t = stripTags(text)
  if (t.length < 2) return
  void api.core.ingestCoreEvents({ body: { events: [{ skill: 'listening', text: t, source }] } }).catch(() => {})
}

export function emitUnclear(text: string, source: Record<string, unknown>) {
  const t = stripTags(text)
  if (t.length < 2) return
  void api.core.markCoreEventUnclear({ body: { skill: 'listening', text: t, source } }).catch(() => {})
}

export function requestAnalyze(text: string) {
  const t = stripTags(text.replace(/\{[^}]*\}/g, '').replace(/\\N/g, ' '))
  if (t.length < 2) return
  document.dispatchEvent(new CustomEvent('els:analyze', { detail: t }))
}

export function requestAsk(text: string) {
  const t = stripTags(text.replace(/\{[^}]*\}/g, '').replace(/\\N/g, ' '))
  if (t.length < 2) return
  document.dispatchEvent(new CustomEvent('els:ask', { detail: t }))
}
