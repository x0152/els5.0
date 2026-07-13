import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { Button, EmptyState, LoadingState, useAgentView } from '@els/ui'
import { BookSpread } from '@els/blocks'
import { useActiveBook, useLesson } from '../lib/lessons.ts'
import { imageApi } from '../lib/images.ts'
import { practiceApi } from '../lib/practice.ts'
import { produce } from '../lib/events.ts'

export function LessonPlay() {
  const { num = '' } = useParams()
  const navigate = useNavigate()
  const book = useActiveBook()
  const { data: lesson, isLoading } = useLesson(book, Number(num))

  useAgentView(
    lesson
      ? {
          app: 'wordbook',
          screen: 'lesson',
          title: `Lesson ${lesson.number}: ${lesson.title}`,
          info: 'The user is taking a Vocabulary in Use lesson. Lesson text — read_book_unit with book=wordbook and this number; lesson list — list_book_units book=wordbook.',
          ids: { number: String(lesson.number) },
        }
      : { app: 'wordbook', screen: 'lesson' },
  )

  if (isLoading) {
    return <LoadingState className="h-full items-center py-0 text-neutral-400" />
  }

  if (!lesson) {
    return (
      <div className="grid h-full place-items-center p-6">
        <EmptyState
          className="w-full max-w-md"
          title="Lesson not found"
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
      heading={`Lesson ${lesson.number} · ${lesson.title}`}
      backLabel="Lessons"
      onBack={() => navigate('..')}
      theory={lesson.theory}
      exercises={lesson.exercises}
      page={lesson.page}
      footer={lesson.footer}
      exercisesTitle="Practice"
      adapters={{ images: imageApi, produce }}
      practiceApi={practiceApi}
      practiceKey={{ kind: book, number: lesson.number }}
    />
  )
}