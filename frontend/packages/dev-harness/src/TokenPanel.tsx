import { useEffect, useRef, useState } from 'react'

const TOKEN_KEY = 'els.auth.token'

function readToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

function writeToken(token: string): void {
  try {
    localStorage.setItem(TOKEN_KEY, token)
  } catch {
    // localStorage may be blocked (private mode, sandboxed iframe) — ignore.
  }
}

function clearToken(): void {
  try {
    localStorage.removeItem(TOKEN_KEY)
  } catch {
    // see writeToken — same reasoning, nothing useful we can do.
  }
}

function maskToken(token: string): string {
  if (token.length <= 12) return token
  return `${token.slice(0, 6)}…${token.slice(-4)}`
}

/**
 * Small status pill in the dev banner.
 *
 * Shows whether `localStorage["els.auth.token"]` is set (which is
 * what every feature's `lib/api.ts` reads), and lets the developer
 * paste/replace/clear it without opening DevTools. Reload-on-change so
 * react-query refetches with the new token.
 */
export function TokenPanel() {
  const [token, setToken] = useState<string | null>(() => readToken())
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState('')
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => {
      if (!ref.current?.contains(e.target as Node)) setOpen(false)
    }
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false)
    window.addEventListener('mousedown', onDown)
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('mousedown', onDown)
      window.removeEventListener('keydown', onKey)
    }
  }, [open])

  const onSave = () => {
    const trimmed = draft.trim()
    if (!trimmed) return
    writeToken(trimmed)
    setToken(trimmed)
    setOpen(false)
    setDraft('')
    window.location.reload()
  }

  const onClear = () => {
    clearToken()
    setToken(null)
    setOpen(false)
    setDraft('')
    window.location.reload()
  }

  const hasToken = !!token

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={
          hasToken
            ? 'inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-emerald-100 text-emerald-800 hover:bg-emerald-200'
            : 'inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-amber-100 text-amber-900 hover:bg-amber-200'
        }
        title={hasToken ? `token: ${maskToken(token)}` : 'no token in localStorage'}
      >
        <span className="text-[10px]">{hasToken ? '🔑' : '⚠️'}</span>
        <span>{hasToken ? `token ${maskToken(token)}` : 'no token'}</span>
      </button>

      {open && (
        <div className="absolute left-0 top-full mt-1 w-[420px] bg-white border border-neutral-200 rounded shadow-lg p-3 z-[100]">
          <div className="text-[11px] text-neutral-500 mb-2 leading-snug">
            Paste a JWT — it goes into <code className="bg-neutral-100 px-1 rounded">localStorage["{TOKEN_KEY}"]</code>{' '}
            and the page reloads. Feature's <code className="bg-neutral-100 px-1 rounded">lib/api.ts</code>{' '}
            will pick it up on next request.
          </div>
          <textarea
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            placeholder="eyJhbGciOi..."
            rows={4}
            className="w-full font-mono text-[11px] border border-neutral-300 rounded p-2 resize-y focus:outline-none focus:ring-1 focus:ring-neutral-400"
          />
          <div className="mt-2 flex items-center justify-between gap-2">
            <button
              type="button"
              onClick={onClear}
              disabled={!hasToken}
              className="text-[11px] px-2 py-1 rounded text-neutral-600 hover:bg-neutral-100 disabled:opacity-40 disabled:cursor-not-allowed"
            >
              clear current
            </button>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => { setOpen(false); setDraft('') }}
                className="text-[11px] px-2 py-1 rounded text-neutral-700 hover:bg-neutral-100"
              >
                cancel
              </button>
              <button
                type="button"
                onClick={onSave}
                disabled={!draft.trim()}
                className="text-[11px] px-2 py-1 rounded bg-neutral-900 text-white hover:bg-neutral-700 disabled:opacity-40 disabled:cursor-not-allowed"
              >
                save & reload
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
