import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button, EmptyState, LoadingState, useAgentView } from '@els/ui'
import { BookSpread } from '@els/blocks'
import { useActiveBook, useUnit } from '../lib/units.ts'
import { imageApi } from '../lib/images.ts'
import { practiceApi } from '../lib/practice.ts'
import { produce } from '../lib/events.ts'

export function UnitPlay() {
  const { num = '' } = useParams()
  const navigate = useNavigate()
  const book = useActiveBook()
  const { data: unit, isLoading } = useUnit(book, Number(num))

  useAgentView(
    unit
      ? {
          app: 'grammarbook',
          screen: 'unit',
          title: `Unit ${unit.number}: ${unit.title}`,
          info: 'The user is taking a grammar unit (Murphy). Unit text — read_book_unit with book=grammar and this number; unit list — list_book_units book=grammar.',
          ids: { number: String(unit.number) },
        }
      : { app: 'grammarbook', screen: 'unit' },
  )

  if (isLoading) {
    return <LoadingState className="h-full items-center py-0 text-neutral-400" />
  }

  if (!unit) {
    return (
      <div className="grid h-full place-items-center p-6">
        <EmptyState
          className="w-full max-w-md"
          title="Unit not found"
          action={
            <Button variant="secondary" onClick={() => navigate('..')}>
              <ArrowLeft className="h-4 w-4" /> Back to list
            </Button>
          }
        />
      </div>
    )
  }

  return (
    <BookSpread
      heading={`Unit ${unit.number} · ${unit.title}`}
      backLabel="Units"
      onBack={() => navigate('..')}
      theory={unit.theory}
      exercises={unit.exercises}
      page={unit.page}
      footer={unit.footer}
      adapters={{ images: imageApi, produce }}
      practiceApi={practiceApi}
      practiceKey={{ kind: book, number: unit.number }}
    />
  )
}
