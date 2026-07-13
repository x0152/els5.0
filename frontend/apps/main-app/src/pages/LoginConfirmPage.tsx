import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { isApiError } from '@els/api-client'
import { SplashScreen } from '@els/ui'
import { api } from '../lib/api'
import { useAuth } from '../auth/AuthContext'
import { AuthLayout } from '../auth/AuthLayout'
import { Alert, PrimaryButton } from '../auth/form'

const TRANSITION_MS = 900

export default function LoginConfirmPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const { signInWithToken } = useAuth()

  const token = params.get('token') || ''
  const next = params.get('next') || '/'

  const [error, setError] = useState<string | null>(null)
  const [status, setStatus] = useState<'loading' | 'done' | 'error'>('loading')
  const once = useRef(false)

  useEffect(() => {
    if (once.current) return
    once.current = true
    if (!token) {
      setStatus('error')
      setError('The link is missing a token.')
      return
    }
    ;(async () => {
      try {
        const session = await api.auth.loginConfirm({ body: { token } })
        if (!session?.token) throw new Error('The server did not return a session token.')
        await signInWithToken(session.token)
        setStatus('done')
        window.setTimeout(() => navigate(next, { replace: true }), TRANSITION_MS)
      } catch (err) {
        setStatus('error')
        const message = isApiError(err)
          ? err.status === 0
            ? 'The server is unavailable.'
            : err.message
          : 'Failed to confirm sign-in.'
        setError(message)
      }
    })()
  }, [token, next, signInWithToken, navigate])

  if (status === 'loading') {
    return <SplashScreen fullscreen subtitle="Confirming sign-in…" />
  }

  if (status === 'done') {
    return <SplashScreen fullscreen subtitle="Entering the platform…" />
  }

  return (
    <AuthLayout
      title="Sign-in failed"
      subtitle="The link may have expired or already been used. Request a new one."
      footer={
        <Link
          to="/login"
          className="text-neutral-900 underline underline-offset-2 hover:text-red-700"
        >
          ← Back to sign-in page
        </Link>
      }
    >
      {error ? (
        <div className="space-y-5">
          <Alert tone="error">{error}</Alert>
          <Link to="/login" className="block">
            <PrimaryButton type="button">Request a new link</PrimaryButton>
          </Link>
        </div>
      ) : null}
    </AuthLayout>
  )
}
