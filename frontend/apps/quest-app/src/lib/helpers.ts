import type { ActiveReply, Mission, MissionSummary } from './types.ts'

export const GENRES: { id: string; label: string; emoji: string }[] = [
  { id: 'random', label: 'Random', emoji: '🎲' },
  { id: 'everyday', label: 'Everyday', emoji: '☕' },
  { id: 'detective', label: 'Detective', emoji: '🕵️' },
  { id: 'comedy', label: 'Comedy', emoji: '😄' },
  { id: 'fantasy', label: 'Fantasy', emoji: '🐉' },
  { id: 'sci-fi', label: 'Sci-Fi', emoji: '🚀' },
  { id: 'horror', label: 'Horror', emoji: '👻' },
]

export function genreEmoji(genre?: string): string {
  return GENRES.find((g) => g.id === genre)?.emoji ?? '✨'
}

export const OUTCOME_LABEL: Record<string, { label: string; tone: string }> = {
  perfect: { label: 'Perfect run', tone: 'text-amber-600' },
  good: { label: 'Mission accomplished', tone: 'text-brand-600' },
  partial: { label: 'Partial success', tone: 'text-amber-600' },
  failed: { label: 'Mission failed', tone: 'text-rose-600' },
  unexpected: { label: 'Unexpected ending', tone: 'text-indigo-600' },
  abandoned: { label: 'Mission abandoned', tone: 'text-neutral-500' },
}

export function avatarKey(name: string): string {
  return name.toLowerCase().trim().split(/\s+/).filter(Boolean).join(' ')
}

function hasGenerating(map?: Record<string, string>): boolean {
  return !!map && Object.values(map).some((s) => s === 'generating')
}

export function missionImagesBusy(m: Mission): boolean {
  return (
    m.coverImageStatus === 'generating' ||
    hasGenerating(m.sceneImageStatus) ||
    hasGenerating(m.characterAvatarStatus)
  )
}

export function sceneGenerating(m: Mission, active?: ActiveReply): boolean {
  const r = active?.result
  return (
    active?.status === 'done' &&
    !!r?.sceneAdvanced &&
    !r.isComplete &&
    m.currentStage < (r.currentStage ?? 0)
  )
}

export function missionBusy(m: Mission, active?: ActiveReply): boolean {
  return (
    active?.status === 'running' ||
    m.generationStatus === 'generating' ||
    missionImagesBusy(m) ||
    sceneGenerating(m, active)
  )
}

export function summaryBusy(m: MissionSummary): boolean {
  return m.generationStatus === 'generating' || m.coverImageStatus === 'generating'
}
