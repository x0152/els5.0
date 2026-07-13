import { test, expect } from '../../fixtures/test'
import { HomePage } from '../../pages/HomePage'

/**
 * Smoke test: frontend starts, `/` renders, backend responds to /api/v1/account/me
 * (even if 401 for an anonymous user — what matters is the request arrived and the UI rendered
 * an error state, not a white screen).
 *
 * Goal — verify the build-router-request-render pipeline. Business scenarios
 * (registration, login, timesheets) will go into separate specs when the
 * matching screens appear.
 */

test.describe('smoke: home page', () => {
  test('opens and shows the /account/me card', async ({ page }) => {
    const home = new HomePage(page)
    await home.goto()

    await expect(home.heading).toBeVisible()
    await expect(home.meCard).toBeVisible()
    await expect(home.refetchButton).toBeVisible()
  })

  test('unauthenticated user sees the /account/me error state', async ({ page }) => {
    const home = new HomePage(page)
    await home.goto()

    const errorHint = page.getByText(/401|unauthor/i)
    const loadingHint = page.getByText('loading…')

    await expect(loadingHint.or(errorHint).first()).toBeVisible()
    await expect(errorHint).toBeVisible()
  })
})
