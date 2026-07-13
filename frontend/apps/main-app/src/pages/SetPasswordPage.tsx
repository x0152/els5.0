import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { isApiError } from '@els/api-client'
import { SplashScreen } from '@els/ui'
import { api } from '../lib/api'
import { AuthLayout } from '../auth/AuthLayout'
import { Alert, Field, PrimaryButton } from '../auth/form'

const REDIRECT_DELAY_MS = 1200

export type PasswordPageMode = 'set' | 'reset'

interface PasswordPageProps {
  mode?: PasswordPageMode
}

export default function SetPasswordPage({ mode = 'set' }: PasswordPageProps) {
  const [params] = useSearchParams()
  const navigate = useNavigate()

  const token = params.get('token') || ''
  const next = params.get('next') || '/'

  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [done, setDone] = useState(false)

  const isReset = mode === 'reset'
  const copy = isReset
    ? {
        title: 'Reset password',
        subtitle: 'Choose a new password. Minimum length is 8 characters.',
        button: 'Save new password',
        errorFallback: 'Failed to reset the password. The link may have expired or already been used.',
        doneSubtitle: 'Password updated. Redirecting to the sign-in page…',
        missingTitle: 'Open the link from the email',
        missingSubtitle:
          'Password reset is only available via the one-time link from the email. Request it on the “Forgot password?” page.',
      }
    : {
        title: 'Set password',
        subtitle:
          'Choose a password to sign in faster next time. Minimum length is 8 characters.',
        button: 'Save and continue to sign-in',
        errorFallback: 'Failed to set the password. The link may have expired or already been used.',
        doneSubtitle: 'Password set. Redirecting to the sign-in page…',
        missingTitle: 'Open the link from the email',
        missingSubtitle:
          'You can set a password only via the invitation link from the email. If the email did not arrive, contact your administrator.',
      }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (loading) return
    setError(null)
    if (password.length < 8) {
      setError('Password must be at least 8 characters long.')
      return
    }
    if (password !== confirm) {
      setError('Passwords do not match.')
      return
    }
    setLoading(true)
    try {
      const body = {
        token,
        password,
        password_confirm: confirm,
      }
      const result = isReset
        ? await api.auth.resetPassword({ body })
        : await api.auth.setPassword({ body })
      setDone(true)
      const email = result?.email
      const search = new URLSearchParams()
      if (email) search.set('email', email)
      if (next && next !== '/') search.set('next', next)
      const target = search.size > 0 ? `/login?${search.toString()}` : '/login'
      window.setTimeout(() => navigate(target, { replace: true }), REDIRECT_DELAY_MS)
    } catch (err) {
      if (isApiError(err)) {
        setError(err.status === 0 ? 'The server is unavailable.' : err.message)
      } else {
        setError(copy.errorFallback)
      }
      setLoading(false)
    }
  }

  if (done) {
    return <SplashScreen fullscreen subtitle={copy.doneSubtitle} />
  }

  if (!token) {
    return (
      <AuthLayout
        title={copy.missingTitle}
        subtitle={copy.missingSubtitle}
        footer={
          <Link
            to={isReset ? '/forgot-password' : '/login'}
            className="text-neutral-900 underline underline-offset-2 hover:text-red-700"
          >
            {isReset ? '← Request a reset link' : '← Back to sign-in page'}
          </Link>
        }
      >
        <Alert tone="info">
          The link is generated only by the server and is sent to your email. Open the email in
          this browser to continue.
        </Alert>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout
      title={copy.title}
      subtitle={copy.subtitle}
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
          label="New password"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="At least 8 characters"
          autoFocus
          required
        />
        <Field
          label="Repeat password"
          type="password"
          autoComplete="new-password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          placeholder="Once more"
          required
        />
        <PrimaryButton type="submit" loading={loading}>
          {copy.button}
        </PrimaryButton>
      </form>
    </AuthLayout>
  )
}
