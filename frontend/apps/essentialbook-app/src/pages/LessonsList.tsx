import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { BookOpen, CheckCircle2, GraduationCap, ListChecks, Loader2, Play, Sparkles, Trash2, TriangleAlert } from 'lucide-react'
import { parseExercises } from '@els/blocks'
import { Badge, Button, Input, useAgentView } from '@els/ui'
import { setActiveBook, useActiveBook, useBooks, useDeleteLesson, useGenerateLesson, useLessons, useMainCompletion } from '../lib/lessons.ts'

export function LessonsList() {
  const navigate = useNavigate()
  const book = useActiveBook()
  const { data: books = [] } = useBooks()
  const current = books.find((b) => b.slug === book)
  const { data: lessons = [] } = useLessons(book)
  const generate = useGenerateLesson(book)
  const remove = useDeleteLesson(book)
  const completion = useMainCompletion(book, lessons.map((l) => l.number))
  const [topic, setTopic] = useState('')

  useAgentView({
    app: 'essentialbook',
    screen: 'lessons',
    info: 'The user is on the Essential Words lesson list. List — list_book_units book=essentialbook; lesson text — read_book_unit book=essentialbook.',
    state: { lessons: lessons.length, book },
  })

  const ready = lessons.filter((l) => l.status !== 'generating' && l.status !== 'error')
  const doneCount = ready.filter((l) => completion[l.number] === 'done').length
  const pct = ready.length ? Math.round((doneCount / ready.length) * 100) : 0
  const totalWords = ready.reduce((sum, l) => sum + l.words.length, 0)
  const totalExercises = ready.reduce((sum, l) => sum + parseExercises(l.exercises).length, 0)
  const next = ready.find((l) => completion[l.number] !== 'done')

  const onGenerate = () => {
    const value = topic.trim()
    if (!value || generate.isPending) return
    generate.mutate(value, { onSuccess: () => setTopic('') })
  }
  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-4xl space-y-6 p-6">
        <header className="relative overflow-hidden rounded-3xl bg-gradient-to-br from-brand-600 via-brand-700 to-brand-900 p-6 text-white shadow-lg">
          <GraduationCap className="absolute -right-6 -top-6 h-40 w-40 text-white/10" />
          <div className="relative">
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-white/60">Vocabulary</span>
              {current?.level && (
                <span className="rounded-full bg-white/15 px-2.5 py-0.5 text-xs font-semibold">{current.level}</span>
              )}
            </div>
            <h1 className="mt-1 text-2xl font-bold">{current?.title ?? 'Essential Words'}</h1>
            {current?.description && <p className="mt-2 max-w-2xl text-sm leading-relaxed text-white/75">{current.description}</p>}
            {books.length > 1 && (
              <div className="mt-4 flex flex-wrap gap-2">
                {books.map((b) => (
                  <button
                    key={b.slug}
                    onClick={() => setActiveBook(b.slug)}
                    className={`rounded-full px-3.5 py-1.5 text-xs font-semibold transition ${
                      b.slug === book ? 'bg-white text-brand-700' : 'bg-white/10 text-white/80 hover:bg-white/20'
                    }`}
                  >
                    {b.level || b.title}
                  </button>
                ))}
              </div>
            )}
            <div className="mt-5 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm">
              <span className="flex items-center gap-1.5 text-white/90">
                <BookOpen className="h-4 w-4 text-white/60" /> {ready.length} lessons
              </span>
              <span className="flex items-center gap-1.5 text-white/90">
                <Sparkles className="h-4 w-4 text-white/60" /> {totalWords} words
              </span>
              <span className="flex items-center gap-1.5 text-white/90">
                <ListChecks className="h-4 w-4 text-white/60" /> {totalExercises} exercises
              </span>
              <span className="flex items-center gap-1.5 text-white/90">
                <CheckCircle2 className="h-4 w-4 text-white/60" /> {doneCount} of {ready.length} completed
              </span>
            </div>
            <div className="mt-4 flex items-center gap-4">
              <div className="h-2 flex-1 overflow-hidden rounded-full bg-white/20">
                <div className="h-full rounded-full bg-white transition-all" style={{ width: `${pct}%` }} />
              </div>
              <span className="text-sm font-semibold">{pct}%</span>
              {next && (
                <Button variant="secondary" className="shrink-0 rounded-xl bg-white text-brand-700 hover:bg-white/90" onClick={() => navigate(String(next.number))}>
                  <Play className="h-4 w-4" /> Continue
                </Button>
              )}
            </div>
          </div>
        </header>

        <div className="grid gap-4 sm:grid-cols-2">
          {lessons.map((l) => {
            if (l.status === 'generating') {
              return (
                <div key={l.number} className="flex items-center justify-between gap-3 rounded-2xl bg-white p-5 ring-1 ring-neutral-200">
                  <div className="flex items-center gap-3">
                    <Loader2 className="h-5 w-5 animate-spin text-brand-600" />
                    <div>
                      <div className="text-sm font-semibold text-neutral-900">{l.title}</div>
                      <div className="text-xs text-neutral-500">Selecting words and exercises…</div>
                    </div>
                  </div>
                  <button onClick={() => remove.mutate(l.number)} className="rounded-lg p-2 text-neutral-400 hover:bg-red-50 hover:text-red-600">
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              )
            }
            if (l.status === 'error') {
              return (
                <div key={l.number} className="flex items-center justify-between gap-3 rounded-2xl bg-red-50 p-5 ring-1 ring-red-200">
                  <div className="flex items-center gap-3">
                    <TriangleAlert className="h-5 w-5 text-red-600" />
                    <div>
                      <div className="text-sm font-semibold text-neutral-900">{l.title}</div>
                      <div className="text-xs text-red-600">Generation failed</div>
                    </div>
                  </div>
                  <button onClick={() => remove.mutate(l.number)} className="rounded-lg p-2 text-neutral-400 hover:bg-red-100 hover:text-red-600">
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              )
            }
            const state = completion[l.number]
            return (
              <button
                key={l.number}
                onClick={() => navigate(String(l.number))}
                className={`group flex flex-col gap-3 rounded-2xl p-5 text-left ring-1 transition hover:shadow-md ${
                  state === 'done' ? 'bg-emerald-50/60 ring-emerald-200 hover:ring-emerald-300' : 'bg-white ring-neutral-200 hover:ring-brand-300'
                }`}
              >
                <div className="flex items-center gap-3">
                  <span
                    className={`grid h-10 w-10 shrink-0 place-items-center rounded-xl text-lg font-bold shadow-sm ${
                      state === 'done' ? 'bg-emerald-600 text-white' : 'bg-gradient-to-br from-brand-500 to-brand-700 text-white'
                    }`}
                  >
                    {state === 'done' ? <CheckCircle2 className="h-5 w-5" /> : l.number}
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-semibold text-neutral-900">Lesson {l.number}</div>
                    <div className="text-xs text-neutral-500">{l.words.length} words</div>
                  </div>
                  {state === 'done' && (
                    <Badge tone="success" className="shrink-0 font-semibold">Completed</Badge>
                  )}
                  {state === 'started' && (
                    <Badge tone="warning" className="shrink-0 font-semibold">In progress</Badge>
                  )}
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {l.words.map((w) => (
                    <Badge key={w} tone="brand">
                      {w}
                    </Badge>
                  ))}
                </div>
              </button>
            )
          })}
        </div>

        <div className="rounded-2xl bg-white p-5 ring-1 ring-neutral-200">
          <div className="flex items-center gap-2 text-sm font-semibold text-neutral-900">
            <Sparkles className="h-4 w-4 text-brand-600" />
            Generate a lesson
          </div>
          <p className="mt-1 text-xs text-neutral-500">
            Describe a topic — the LLM will pick words and build theory and exercises. Images stay as prompts; generate them with a click inside the lesson.
          </p>
          <div className="mt-3 flex gap-2">
            <Input
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && onGenerate()}
              placeholder="e.g. Travel and airports, Feelings…"
              disabled={generate.isPending}
              className="min-w-0 flex-1 rounded-xl px-3.5"
            />
            <Button variant="brand" className="shrink-0 rounded-xl" onClick={onGenerate} disabled={generate.isPending || !topic.trim()}>
              {generate.isPending ? 'Generating…' : 'Create'}
            </Button>
          </div>
          {generate.isError && (
            <p className="mt-2 text-xs text-red-600">Generation failed. Please try again.</p>
          )}
        </div>
      </div>
    </div>
  )
}
