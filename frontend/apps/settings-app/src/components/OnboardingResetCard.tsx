import { useState } from 'react'
import { api } from '../lib/api'

export function OnboardingResetCard() {
  const [done, setDone] = useState(false)

  const reset = async () => {
    await api.onboarding.onboardingResetTours()
    window.dispatchEvent(new Event('els:onboarding:reset'))
    setDone(true)
    window.setTimeout(() => setDone(false), 2000)
  }

  return (
    <div className="rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h2 className="text-base font-semibold text-neutral-900">Onboarding</h2>
          <p className="mt-1 text-sm text-neutral-500">
            Reset the welcome wizard and app tours so they show again.
          </p>
        </div>
        <button
          type="button"
          onClick={reset}
          className="shrink-0 rounded-lg bg-neutral-900 px-3.5 py-2 text-sm font-medium text-white hover:bg-neutral-800"
        >
          {done ? 'Reset ✓' : 'Reset'}
        </button>
      </div>
    </div>
  )
}
