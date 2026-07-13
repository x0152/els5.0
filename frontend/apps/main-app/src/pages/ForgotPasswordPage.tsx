import { useState, type FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { isApiError } from '@els/api-client'
import { api } from '../lib/api'
import { AuthLayout } from '../auth/AuthLayout'
import { Alert, Field, PrimaryButton } from '../auth/form'

type Stage =
  | { kind: 'form' }
  | { kind: 'sent'; sentTo: string }

export default function ForgotPasswordPage() {
  const [params] = useSearchParams()
  const initialEmail = params.get('email') || ''

  const [email, setEmail] = useState(initialEmail)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [stage, setStage] = useState<Stage>({ kind: 'form' })

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (loading) return
    setError(null)
    setLoading(true)
    try {
      const res = await api.auth.forgotPassword({ body: { email } })
      const sentTo =
        (res as { Body?: { sent_to?: string }; sent_to?: string } | undefined)?.Body?.sent_to ||
        (res as { sent_to?: string } | undefined)?.sent_to ||
        maskEmail(email)
      setStage({ kind: 'sent', sentTo })
    } catch (err) {
      setError(errorMessage(err, 'Failed to send the email. Please try again.'))
    } finally {
      setLoading(false)
    }
  }

  if (stage.kind === 'sent') {
    return (
      <AuthLayout
        title="Check your email"
        subtitle={
          <>
            If an account with the address <b>{stage.sentTo}</b> exists, we have sent it an
            email with a password reset link. Open the email in this browser and follow the
            link.
          </>
        }
        footer={
          <Link
            to="/login"
            className="text-neutral-900 underline underline-offset-2 hover:text-red-700"
          >
            ← Back to sign-in page
          </Link>
        }
      >
        <Alert tone="info">
          Email did not arrive? Check your “Spam” folder or try sending the link again in a
          minute.
        </Alert>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout
      title="Password recovery"
      subtitle="Enter your corporate email. If the account exists, we will send an email with a one-time password reset link."
      footer={
        <Link
          to="/login"
          className="text-neutral-900 underline underline-offset-2 hover:text-red-700"
        >
          ← Back to sign-in page
        </Link>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        {error ? <Alert tone="error">{error}</Alert> : null}
        <Field
          label="Email"
          type="email"
          autoComplete="email"
          autoFocus
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@company.com"
        />
        <PrimaryButton type="submit" loading={loading}>
          Send reset link
        </PrimaryButton>
      </form>
    </AuthLayout>
  )
}

function errorMessage(err: unknown, fallback: string): string {
  if (isApiError(err)) {
    if (err.status === 0) return 'The server is unavailable. Check that the backend is running.'
    return err.message || fallback
  }
  return fallback
}

function maskEmail(raw: string): string {
  const at = raw.indexOf('@')
  if (at <= 1) return raw
  const head = raw.slice(0, 1)
  const tail = raw.slice(at)
  return `${head}***${tail}`
}
