import { useState } from 'react'
import { Check, Eye, ImageIcon, KeyRound, ScanText, Sparkles, type LucideIcon } from 'lucide-react'
import { Button, Card, CardContent, CardDescription, CardHeader, CardTitle, Field, Input } from '@els/ui'
import { api } from '../lib/api'
import type { Feature, Provider } from '../lib/types'
import { ModelCombobox } from './ModelCombobox'

const ICONS: Record<Feature, LucideIcon> = {
  main: Sparkles,
  analysis: ScanText,
  vision: Eye,
  image: ImageIcon,
}

type Props = {
  feature: Feature
  title: string
  description: string
  provider: Provider | undefined
}

export function ProviderCard({ feature, title, description, provider }: Props) {
  const [baseUrl, setBaseUrl] = useState(provider?.base_url ?? '')
  const [model, setModel] = useState(provider?.model ?? '')
  const [apiKey, setApiKey] = useState('')
  const [keyTouched, setKeyTouched] = useState(false)
  const [hasKey, setHasKey] = useState(provider?.has_key ?? false)

  const [models, setModels] = useState<string[]>([])
  const [modelsLoading, setModelsLoading] = useState(false)
  const [modelsError, setModelsError] = useState<string | null>(null)

  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  const loadModels = async () => {
    setModelsLoading(true)
    setModelsError(null)
    try {
      const res = await api.settings.listAIProviderModels({
        params: {
          path: { feature },
          query: {
            ...(baseUrl.trim() ? { base_url: baseUrl.trim() } : {}),
            ...(keyTouched && apiKey ? { api_key: apiKey } : {}),
          },
        },
      })
      setModels(res?.items ?? [])
    } catch (e) {
      setModelsError(e instanceof Error ? e.message : 'Failed to load models')
    } finally {
      setModelsLoading(false)
    }
  }

  const save = async () => {
    setSaving(true)
    setSaved(false)
    setSaveError(null)
    try {
      const res = await api.settings.updateAIProvider({
        params: { path: { feature } },
        body: {
          base_url: baseUrl.trim(),
          model: model.trim(),
          ...(keyTouched ? { api_key: apiKey } : {}),
        },
      })
      setHasKey(res?.provider?.has_key ?? hasKey)
      setApiKey('')
      setKeyTouched(false)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } catch (e) {
      setSaveError(e instanceof Error ? e.message : 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const Icon = ICONS[feature]

  return (
    <Card className="rounded-2xl transition-shadow hover:shadow-md">
      <CardHeader>
        <div className="flex items-start gap-3">
          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-brand-50 text-brand-600 ring-1 ring-brand-100">
            <Icon size={18} />
          </span>
          <div className="min-w-0">
            <CardTitle>{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <Field label="Base URL">
          <Input
            value={baseUrl}
            onChange={(e) => setBaseUrl(e.target.value)}
            placeholder="https://api.openai.com/v1"
          />
        </Field>

        <Field label="API token">
          <div className="relative">
            <KeyRound size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
            <Input
              type="password"
              value={apiKey}
              onChange={(e) => {
                setApiKey(e.target.value)
                setKeyTouched(true)
              }}
              placeholder={hasKey ? '•••••••••• (saved, leave empty)' : 'Enter token'}
              className="pl-9"
            />
          </div>
        </Field>

        <Field label="Model">
          <ModelCombobox
            value={model}
            onChange={setModel}
            models={models}
            loading={modelsLoading}
            error={modelsError}
            onLoad={loadModels}
          />
        </Field>

        <div className="flex items-center gap-3 pt-1">
          <Button onClick={save} disabled={saving} variant="brand">
            {saving ? 'Saving...' : 'Save'}
          </Button>
          {saved && (
            <span className="flex items-center gap-1 text-sm text-green-600">
              <Check size={16} /> Saved
            </span>
          )}
          {saveError && <span className="text-sm text-red-500">{saveError}</span>}
        </div>
      </CardContent>
    </Card>
  )
}