import { DeferredBlocks, ImageApiCtx, type IllustrationStatus, type ImageApi } from '@els/blocks'
import { emitTargetedEvents } from '@els/core-events'
import { api } from '../lib/api.ts'
import type { Grammar, StepOutcome } from '../lib/types.ts'
import { StepShell } from './StepShell.tsx'

const imageApi: ImageApi = async (prompt, trigger, aspect) =>
  (await api.learn.ensureIllustration({ body: { prompt, trigger, aspect } })) as IllustrationStatus

export function GrammarStep({ grammar, onDone }: { grammar: Grammar; onDone: (outcome: StepOutcome) => void }) {
  return (
    <StepShell>
      <ImageApiCtx.Provider value={imageApi}>
        <DeferredBlocks
          md={grammar.exercises}
          onContinue={(firstAttempt) => {
            const wrong = firstAttempt.filter((r) => !r.correct)
            emitTargetedEvents(
              api,
              'writing',
              wrong.map((r) => ({ target: grammar.topic || 'grammar', outcome: 'fail' as const, context: r.prompt })),
              { app: 'workout' },
            )
            const score = firstAttempt.length ? Math.round(((firstAttempt.length - wrong.length) / firstAttempt.length) * 100) : 100
            onDone({ score })
          }}
        />
      </ImageApiCtx.Provider>
    </StepShell>
  )
}
