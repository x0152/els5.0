export function watchProgress(positionMs: number, durationMs: number): { percent: number; done: boolean } {
  if (!durationMs) return { percent: 0, done: false }
  const percent = Math.min(100, Math.round((positionMs / durationMs) * 100))
  return { percent, done: percent >= 90 }
}
