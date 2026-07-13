import { useEffect, useState } from 'react'
import { SlidersHorizontal } from 'lucide-react'
import { LoadingState } from '@els/ui'
import { api } from './lib/api'
import { FEATURES, type Provider } from './lib/types'
import { ProviderCard } from './components/ProviderCard'
import { EventProcessingCard } from './components/EventProcessingCard'

export function SettingsAppPage() {
  const [providers, setProviders] = useState<Provider[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.settings
      .listAIProviders()
      .then((res) => alive && setProviders(res?.items ?? []))
      .catch((e) => alive && setError(e instanceof Error ? e.message : 'Failed to load settings'))
    return () => {
      alive = false
    }
  }, [])

  return (
    <div className="h-full w-full overflow-auto bg-neutral-50">
      <div className="mx-auto max-w-6xl space-y-8 p-6 lg:p-8">
        <header className="flex items-center gap-3">
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-brand-600 text-white shadow-sm">
            <SlidersHorizontal className="h-5 w-5" />
          </span>
          <div>
            <h1 className="text-2xl font-bold text-neutral-900">Platform settings</h1>
            <p className="mt-0.5 text-sm text-neutral-500">Manage event processing and the platform AI providers.</p>
          </div>
        </header>

        <EventProcessingCard />

        <section className="space-y-4">
          <div>
            <h2 className="text-lg font-semibold text-neutral-900">AI providers</h2>
            <p className="mt-1 text-sm text-neutral-500">
              For each feature, set the base URL, token and a model from the <code className="rounded bg-neutral-100 px-1 py-0.5 text-xs text-neutral-700">/models</code> list.
            </p>
          </div>

          {error && <div className="rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}

          {!providers && !error && <LoadingState className="py-10" />}

          {providers && (
            <div className="grid gap-5 lg:grid-cols-2">
              {FEATURES.map((f) => {
                const provider = providers.find((p) => p.feature === f.id)
                return (
                  <ProviderCard
                    key={provider ? `${f.id}:${provider.base_url}:${provider.model}` : f.id}
                    feature={f.id}
                    title={f.title}
                    description={f.description}
                    provider={provider}
                  />
                )
              })}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
