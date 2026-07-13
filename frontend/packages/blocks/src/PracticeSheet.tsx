import { mockCheckAnswer } from './check.ts'
import { PROSE_CLS } from './markdown.tsx'
import { BlocksProvider, type BlocksAdapters } from './Blocks.tsx'
import { ExercisesList } from './render/exercises.tsx'

export type PracticeSheetProps = {
  exercises: string
  adapters?: BlocksAdapters
}

// Renders a full exercise set (`## N` headings) as numbered cards.
export function PracticeSheet({ exercises, adapters = {} }: PracticeSheetProps) {
  return (
    <BlocksProvider adapters={adapters}>
      <div className={PROSE_CLS}>
        <ExercisesList exercises={exercises} checkAnswer={adapters.check ?? mockCheckAnswer} onTheory={() => {}} />
      </div>
    </BlocksProvider>
  )
}

export function PracticeSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="space-y-4">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="animate-pulse rounded-2xl border border-neutral-200/80 bg-neutral-50/50 p-4 sm:p-5">
          <div className="mb-3.5 flex items-center gap-3">
            <div className="h-6 w-6 shrink-0 rounded-lg bg-neutral-300" />
            <div className="h-4 w-2/3 rounded bg-neutral-200" />
          </div>
          <div className="space-y-2 pl-9">
            <div className="h-3 w-full rounded bg-neutral-200" />
            <div className="h-3 w-5/6 rounded bg-neutral-200" />
            <div className="h-3 w-3/4 rounded bg-neutral-200" />
          </div>
        </div>
      ))}
    </div>
  )
}
