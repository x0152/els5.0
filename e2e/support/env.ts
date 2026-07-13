import 'dotenv/config'

function required(name: string, fallback?: string): string {
  const value = process.env[name] ?? fallback
  if (!value) {
    throw new Error(`Missing required env var: ${name}`)
  }
  return value
}

export const env = {
  baseUrl: required('BASE_URL', 'http://localhost:5173'),
  apiUrl: required('API_URL', 'http://localhost:8000'),
  mailUrl: process.env.MAIL_URL ?? 'http://localhost:8025',
  isCI: Boolean(process.env.CI),
} as const

export type Env = typeof env
