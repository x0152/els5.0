import { useState } from 'react'
import { api } from '../lib/api'

export function WorkoutResetCard() {
  const [state, setState] = useState<'idle' | 'busy' | 'done' | 'error'>('idle')

  const reset = async () => {
    if (!window.confirm('Delete all workout progress and generated lessons? This cannot be undone.'))
      return
    setState('busy')
    try {
      await api.workout.workoutReset()
      setState('done')
      window.setTimeout(() => setState('idle'), 2000)
    } catch {
      setState('error')
      window.setTimeout(() => setState('idle'), 3000)
    }
  }

  return (
    <div className="rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h2 className="text-base font-semibold text-neutral-900">Workout</h2>
          <p className="mt-1 text-sm text-neutral-500">
            Delete all workout progress, streak and the generated lessons. The next visit to
            Workout generates a fresh lesson from scratch.
          </p>
        </div>
        <button
          type="button"
          onClick={() => void reset()}
          disabled={state === 'busy'}
          className="shrink-0 rounded-lg bg-rose-600 px-3.5 py-2 text-sm font-medium text-white hover:bg-rose-700 disabled:opacity-60"
        >
          {state === 'busy' ? 'Resetting…' : state === 'done' ? 'Reset ✓' : state === 'error' ? 'Failed' : 'Reset'}
        </button>
      </div>
    </div>
  )
}
