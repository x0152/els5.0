import type { ImageApi, IllustrationStatus } from '@els/blocks'
import { api } from './api.ts'

export const imageApi: ImageApi = async (prompt, trigger, aspect) =>
  (await api.learn.ensureIllustration({ body: { prompt, trigger, aspect } })) as IllustrationStatus

export function wordImagePrompt(text: string) {
  const t = text.trim()
  return `A clear educational illustration of "${t}": a simple memorable scene that shows the meaning.`
}
