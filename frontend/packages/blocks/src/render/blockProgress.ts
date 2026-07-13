import { useContext, useEffect } from 'react'
import { useProgress } from '../state.ts'
import { BlockCtx } from './context.ts'

// Persists a whole-block state (sort/highlight/match) as one progress entry,
// so these exercises survive reloads and count towards completion.
export function useBlockProgress<T>(suffix: string): {
  saved: T | undefined
  save: (state: T, correct: boolean) => void
} {
  const { keyBase } = useContext(BlockCtx)
  const progress = useProgress()
  const key = keyBase !== undefined && progress.enabled ? `${keyBase}:${suffix}` : undefined

  useEffect(() => {
    if (key) progress.register(key)
  }, [key, progress])

  let saved: T | undefined
  const raw = key ? progress.get(key)?.answer : undefined
  if (raw) {
    try {
      saved = JSON.parse(raw) as T
    } catch {
      saved = undefined
    }
  }
  return {
    saved,
    save: (state, correct) => {
      if (key) progress.set(key, { answer: JSON.stringify(state), correct })
    },
  }
}
