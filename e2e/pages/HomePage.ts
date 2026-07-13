import type { Locator, Page } from '@playwright/test'

/**
 * Page Object for the main-app home page (`/`).
 *
 * UI is minimal for now: sidebar + “GET /api/v1/account/me” card. As
 * new screens appear — selectors and navigation methods grow here too.
 *
 * Selectors intentionally rely on role/accessible name, not classes:
 * such tests do not break when styles change or layout is refactored.
 */
export class HomePage {
  readonly page: Page
  readonly heading: Locator
  readonly refetchButton: Locator
  readonly meCard: Locator

  constructor(page: Page) {
    this.page = page
    this.heading = page.getByRole('heading', { level: 1, name: 'Home' })
    this.refetchButton = page.getByRole('button', { name: 'Refetch' })
    this.meCard = page.getByText('GET /api/v1/account/me', { exact: true })
  }

  async goto(): Promise<void> {
    await this.page.goto('/')
  }
}
