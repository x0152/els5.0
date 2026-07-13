import { useState, useEffect, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { isApiError } from '@els/api-client'
import { api } from '../lib/api'
import { useAuth } from '../auth/AuthContext'
import { AuthLayout } from '../auth/AuthLayout'
import { Alert, Field, PrimaryButton } from '../auth/form'

export default function LoginPage() {
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const next = params.get('next') || '/'
  const initialEmail = params.get('email') || ''

  const { isAuthenticated, signInWithToken } = useAuth()

  const [email, setEmail] = useState(initialEmail)
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isAuthenticated) navigate(next, { replace: true })
  }, [isAuthenticated, navigate, next])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (loading) return
    setError(null)
    setLoading(true)
    try {
      const session = await api.auth.loginStart({ body: { email, password } })
      if (!session?.token) throw new Error('The server did not return a session token.')
      await signInWithToken(session.token)
      navigate(next, { replace: true })
    } catch (err) {
      setError(errorMessage(err, 'Invalid email or password.'))
    } finally {
      setLoading(false)
    }
  }

  const forgotPasswordHref = email
    ? `/forgot-password?email=${encodeURIComponent(email)}`
    : '/forgot-password'

  return (
    <AuthLayout title="Sign in" subtitle="Enter your email and password.">
      <form onSubmit={handleSubmit} className="space-y-5">
        {error ? <Alert tone="error">{error}</Alert> : null}
        <Field
          label="Email"
          type="email"
          autoComplete="email"
          autoFocus={!initialEmail}
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="you@company.com"
        />
        <Field
          label="Password"
          type="password"
          autoComplete="current-password"
          autoFocus={!!initialEmail}
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="••••••••"
        />
        <div className="flex justify-end -mt-2">
          <Link
            to={forgotPasswordHref}
            className="text-sm text-neutral-600 underline underline-offset-2 hover:text-red-700"
          >
            Forgot password?
          </Link>
        </div>
        <PrimaryButton type="submit" loading={loading}>
          Sign in
        </PrimaryButton>
      </form>
    </AuthLayout>
  )
}

function errorMessage(err: unknown, fallback: string): string {
  if (isApiError(err)) {
    if (err.status === 0) return 'The server is unavailable. Check that the backend is running.'
    if (err.status === 401) return 'Invalid email or password.'
    return err.message || fallback
  }
  return fallback
}
