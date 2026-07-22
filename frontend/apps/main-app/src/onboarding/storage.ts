const PREFIX = 'els.onboarding.'
const WIZARD_KEY = `${PREFIX}wizard`

export const ONBOARDING_RESET_EVENT = 'els:onboarding:reset'
export const TOUR_OPEN_EVENT = 'els:tour:open'

export function isWizardDone(): boolean {
  return localStorage.getItem(WIZARD_KEY) === 'done'
}

export function markWizardDone(): void {
  localStorage.setItem(WIZARD_KEY, 'done')
}

export function isTourDone(appId: string): boolean {
  return localStorage.getItem(`${PREFIX}tour.${appId}`) === 'done'
}

export function markTourDone(appId: string): void {
  localStorage.setItem(`${PREFIX}tour.${appId}`, 'done')
}
