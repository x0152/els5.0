import { BookSpread, DeferredBlocks, DeferredProvider, type BlocksAdapters, type DeferredResult, type IllustrationStatus, type ImageApi } from '@els/blocks'
import { emitTargetedEvents } from '@els/core-events'
import { api } from '../lib/api.ts'
import type { Grammar, StepOutcome } from '../lib/types.ts'
import { StepShell } from './StepShell.tsx'

const imageApi: ImageApi = async (prompt, trigger, aspect) =>
  (await api.learn.ensureIllustration({ body: { prompt, trigger, aspect } })) as IllustrationStatus

const adapters: BlocksAdapters = { images: imageApi }

export function GrammarStep({ grammar, onDone }: { grammar: Grammar; onDone: (outcome: StepOutcome) => void }) {
  const onContinue = (firstAttempt: DeferredResult[]) => {
    const wrong = firstAttempt.filter((r) => !r.correct)
    emitTargetedEvents(
      api,
      'writing',
      wrong.map((r) => ({ target: grammar.topic || 'grammar', outcome: 'fail' as const, context: r.prompt })),
      { app: 'workout' },
    )
    const score = firstAttempt.length ? Math.round(((firstAttempt.length - wrong.length) / firstAttempt.length) * 100) : 100
    onDone({ score })
  }

  if (!grammar.theory) {
    return (
      <StepShell>
        <DeferredBlocks md={grammar.exercises} adapters={adapters} onContinue={onContinue} />
      </StepShell>
    )
  }

  return (
    <DeferredProvider
      onContinue={onContinue}
      controlsClassName="flex flex-wrap items-center gap-2 rounded-2xl border border-neutral-200 bg-white p-4 shadow-sm"
    >
      {(controls) => (
        <>
          <BookSpread heading={grammar.topic} theory={grammar.theory!} exercises={grammar.exercises} page={1} spread={false} adapters={adapters} />
          {controls}
        </>
      )}
    </DeferredProvider>
  )
}
