import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, BookOpen } from 'lucide-react'
import { LoadingState, PhonemePopover, anchorOf, phonemeImage, useAgentView, type PhonemeAnchor, type PhonemeGuideInfo } from '@els/ui'
import { api } from '../lib/api.ts'

const KINDS = [
  { id: 'vowel', label: 'Vowels' },
  { id: 'diphthong', label: 'Diphthongs' },
  { id: 'consonant', label: 'Consonants' },
] as const

export function SoundsPage() {
  const navigate = useNavigate()
  const [selected, setSelected] = useState<{ info: PhonemeGuideInfo; anchor: PhonemeAnchor } | null>(null)

  const guideQ = useQuery({
    queryKey: ['speech', 'phonemes'],
    queryFn: () => api.speech.listSpeechPhonemes(),
    staleTime: Infinity,
  })
  const items = guideQ.data?.items ?? []

  useAgentView({
    app: 'speaking',
    screen: 'sounds',
    info: 'The user is browsing the reference guide of English sounds with articulation diagrams.',
  })

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-4xl space-y-6 p-6">
        <button
          onClick={() => navigate('..')}
          className="inline-flex items-center gap-1.5 text-sm font-medium text-neutral-500 hover:text-neutral-800"
        >
          <ArrowLeft className="h-4 w-4" /> Back to practice
        </button>
        <header>
          <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
            <BookOpen className="h-6 w-6 text-brand-600" />
            Sounds of English
          </h1>
          <p className="mt-1 text-sm text-neutral-500">
            Every sound with its mouth position and a short how-to. Tap a sound for details.
          </p>
        </header>

        {guideQ.isLoading ? (
          <LoadingState className="py-24 text-neutral-400" />
        ) : (
          KINDS.map((kind) => {
            const sounds = items.filter((p) => p.kind === kind.id)
            if (sounds.length === 0) return null
            return (
              <section key={kind.id}>
                <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">{kind.label}</h2>
                <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4">
                  {sounds.map((p) => {
                    const image = phonemeImage(p.symbol)
                    return (
                      <button
                        key={p.symbol}
                        type="button"
                        onClick={(e) => setSelected({ info: p, anchor: anchorOf(e.currentTarget) })}
                        className="rounded-2xl bg-white p-4 text-left ring-1 ring-neutral-200 transition hover:ring-brand-400"
                      >
                        <div className="flex items-center justify-between gap-2">
                          <span className="font-mono text-2xl font-bold text-neutral-900">/{p.symbol}/</span>
                          {image && <img src={image} alt="" className="h-14 w-auto" />}
                        </div>
                        <p className="mt-2 truncate text-xs text-neutral-500 italic">{p.examples}</p>
                      </button>
                    )
                  })}
                </div>
              </section>
            )
          })
        )}
      </div>

      {selected && (
        <PhonemePopover
          symbol={selected.info.symbol}
          info={selected.info}
          anchor={selected.anchor}
          onClose={() => setSelected(null)}
        />
      )}
    </div>
  )
}
