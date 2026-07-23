import { api } from '../lib/api'

export const ONBOARDING_RESET_EVENT = 'els:onboarding:reset'
export const TOUR_OPEN_EVENT = 'els:tour:open'
export const SYSTEM_TOUR_OPEN_EVENT = 'els:systemtour:open'

export const WIZARD_TOUR = 'wizard'
export const SYSTEM_TOUR = 'system'

let done = new Set<string>()
let loaded = false

export async function loadTours(): Promise<void> {
  const res = await api.onboarding.onboardingTours()
  done = new Set(res?.ids ?? [])
  loaded = true
}

export function isTourDone(id: string): boolean {
  return !loaded || done.has(id)
}

export function markTourDone(id: string): void {
  done.add(id)
  void api.onboarding.onboardingMarkTour({ body: { id } }).catch(() => {})
}

export function clearTours(): void {
  done = new Set()
}
