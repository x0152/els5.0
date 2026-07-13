import { useEffect, useState } from 'react'
import { api } from '../lib/api'

export function EventProcessingCard() {
  const [enabled, setEnabled] = useState<boolean | null>(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.settings
      .getEventProcessing()
      .then((res) => alive && setEnabled(!!res?.enabled))
      .catch((e) => alive && setError(e instanceof Error ? e.message : 'Failed to load'))
    return () => {
      alive = false
    }
  }, [])

  const toggle = async () => {
    if (enabled === null || saving) return
    const next = !enabled
    setSaving(true)
    setError(null)
    try {
      await api.settings.setEventProcessing({ body: { enabled: next } })
      setEnabled(next)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="rounded-2xl border border-neutral-200 bg-white p-5 shadow-sm">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h2 className="text-base font-semibold text-neutral-900">Event processing</h2>
          <p className="mt-1 text-sm text-neutral-500">
            When disabled, incoming events stay pending and are not processed by workers.
          </p>
        </div>
        <button
          type="button"
          onClick={toggle}
          disabled={enabled === null || saving}
          className={`relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors disabled:opacity-50 ${
            enabled ? 'bg-emerald-500' : 'bg-neutral-300'
          }`}
          aria-pressed={!!enabled}
        >
          <span
            className={`inline-block h-5 w-5 transform rounded-full bg-white shadow transition-transform ${
              enabled ? 'translate-x-5' : 'translate-x-1'
            }`}
          />
        </button>
      </div>
      {error && <div className="mt-3 text-sm text-red-600">{error}</div>}
    </div>
  )
}
