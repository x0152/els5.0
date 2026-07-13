import type { SettingsComponents } from '@els/api-client'

export type Provider = SettingsComponents['schemas']['ProviderOutput']

export type Feature = 'main' | 'analysis' | 'vision' | 'image'

export const FEATURES: { id: Feature; title: string; description: string }[] = [
  { id: 'main', title: 'Main LLM', description: 'Chat agent and generation — everything not delegated to a dedicated provider.' },
  { id: 'analysis', title: 'Analysis', description: 'Word analysis in the dictionary and grammar checking in quests.' },
  { id: 'vision', title: 'Image reading', description: 'Recognizing film frames for the agent.' },
  { id: 'image', title: 'Image generation', description: 'Illustrations and images for quests.' },
]
