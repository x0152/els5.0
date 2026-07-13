import { test as base, expect } from '@playwright/test'
import { ApiClient, uniqueEmail, type SessionOutput } from '../support/api-client'

/**
 * Base `test` with mixed-in fixtures.
 *
 * Extension policy:
 * - `api`      — ready APIRequestContext to the backend (for seeding).
 * - `newUser`  — freshly registered account for a specific test.
 *
 * Roles/personas (expertUser, clientUser, adminUser) will be added when the backend
 * has the matching roles and seed helpers. For now the codebase only has
 * “plain registration”, so fixtures are deliberately minimal.
 */

type Fixtures = {
  api: ApiClient
  newUser: {
    email: string
    password: string
    session: SessionOutput
  }
}

export const test = base.extend<Fixtures>({
  api: async ({}, use) => {
    const client = await ApiClient.create()
    await use(client)
    await client.dispose()
  },

  newUser: async ({ api }, use) => {
    const email = uniqueEmail('user')
    const password = 'Test1234!'
    const session = await api.register({ email, password })
    await use({ email, password, session })
  },
})

export { expect }
