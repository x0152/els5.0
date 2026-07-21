import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { isApiError } from '@els/api-client'
import { emitTextEvents } from '@els/core-events'
import { Button, Spinner, Textarea, cn, useAgentView } from '@els/ui'
import { CheckCircle2, MessageSquareText, Pencil, PencilRuler, RotateCcw, Sparkles } from 'lucide-react'
import { api } from './lib/api'
import type { TrainerIssue, TrainerVerdict } from './lib/types'

const SOURCE = { app: 'writing' }

const LEVELS = [
  { id: 1, title: 'No mistakes', desc: 'only genuine grammar errors' },
  { id: 2, title: 'Natural', desc: '+ words and collocations that sound off' },
  { id: 3, title: 'Like a native', desc: "+ anything a native would phrase differently" },
] as const

type Level = (typeof LEVELS)[number]

const SEVERITY_STYLE: Record<string, { mark: string; badge: string; label: string }> = {
  grammar: { mark: 'bg-red-100 decoration-red-500', badge: 'bg-red-100 text-red-700', label: 'grammar' },
  style: { mark: 'bg-amber-100 decoration-amber-500', badge: 'bg-amber-100 text-amber-700', label: 'sounds unnatural' },
  native: { mark: 'bg-violet-100 decoration-violet-500', badge: 'bg-violet-100 text-violet-700', label: 'a native would say it differently' },
}

type Span = { text: string; issueIdx: number | null }

function toSpans(text: string, issues: TrainerIssue[]): Span[] {
  const marks: { start: number; end: number; issueIdx: number }[] = []
  const lower = text.toLowerCase()
  issues.forEach((issue, issueIdx) => {
    let from = 0
    while (true) {
      const i = lower.indexOf(issue.fragment.toLowerCase(), from)
      if (i < 0) break
      const end = i + issue.fragment.length
      if (!marks.some((m) => i < m.end && end > m.start)) {
        marks.push({ start: i, end, issueIdx })
        break
      }
      from = end
    }
  })
  marks.sort((a, b) => a.start - b.start)
  const spans: Span[] = []
  let pos = 0
  for (const m of marks) {
    if (m.start > pos) spans.push({ text: text.slice(pos, m.start), issueIdx: null })
    spans.push({ text: text.slice(m.start, m.end), issueIdx: m.issueIdx })
    pos = m.end
  }
  if (pos < text.length) spans.push({ text: text.slice(pos), issueIdx: null })
  return spans
}

function HighlightedDraft({ text, issues }: { text: string; issues: TrainerIssue[] }) {
  return (
    <p className="leading-loose text-neutral-800">
      {toSpans(text, issues).map((s, i) =>
        s.issueIdx === null ? (
          <span key={i}>{s.text}</span>
        ) : (
          <span
            key={i}
            className={`rounded px-0.5 underline decoration-wavy decoration-2 underline-offset-4 ${SEVERITY_STYLE[issues[s.issueIdx]!.severity]?.mark ?? ''}`}
          >
            {s.text}
            <sup className="ml-0.5 font-semibold text-neutral-500">{s.issueIdx + 1}</sup>
          </span>
        ),
      )}
    </p>
  )
}

type Attempt = { draft: string; verdict: TrainerVerdict }

export function TrainerPage() {
  const [dialogue, setDialogue] = useState('')
  const [scenario, setScenario] = useState('')
  const [draft, setDraft] = useState('')
  const [level, setLevel] = useState<Level>(LEVELS[0])
  const [attempts, setAttempts] = useState<Attempt[]>([])

  const suggest = useMutation({
    mutationFn: () => api.writing.writingGenerateSituation({ body: {} }),
    onSuccess: (res) => {
      if (!res) return
      setDialogue(res.dialogue)
      setScenario(res.scenario)
      setDraft('')
      setAttempts([])
    },
  })

  const check = useMutation({
    mutationFn: (vars: { dialogue: string; draft: string; level: number }) =>
      api.writing.writingTrainerCheck({ body: vars }),
    onSuccess: (verdict, vars) => {
      if (!verdict) return
      setAttempts((prev) => [...prev, { draft: vars.draft, verdict }])
      if (verdict.pass) emitTextEvents(api, 'writing', [vars.draft], SOURCE)
    },
  })

  const changeLevel = (l: Level) => {
    setLevel(l)
    setAttempts([])
  }

  const restart = () => setAttempts([])

  const result = attempts[attempts.length - 1]
  const passed = result?.verdict.pass ?? false
  const issues = result?.verdict.issues ?? []
  const nextLevel = LEVELS.find((l) => l.id === level.id + 1)
  const canCheck = !check.isPending && draft.trim().length >= 5

  useAgentView({
    app: 'writing',
    screen: 'trainer',
    info: 'The user drafts a reply to a dialogue and the trainer points at problems without giving corrections.',
    state: { level: level.id, attempts: attempts.length, phase: passed ? 'passed' : 'drafting' },
  })

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey) && canCheck) {
      e.preventDefault()
      check.mutate({ dialogue, draft, level: level.id })
    }
  }

  return (
    <div className="h-full w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto flex max-w-2xl flex-col gap-6 px-6 py-8">
        <header className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
            <PencilRuler className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-neutral-900">Phrase trainer</h1>
            <p className="text-sm text-neutral-500">It shows where the problem is — rewriting is up to you.</p>
          </div>
        </header>

        <section className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-500">Dialogue (context)</h2>
            <Button variant="ghost" size="sm" onClick={() => suggest.mutate()} disabled={suggest.isPending}>
              {suggest.isPending ? <Spinner className="h-4 w-4" /> : <MessageSquareText className="h-4 w-4" />}
              Suggest a situation
            </Button>
          </div>
          {scenario && <p className="text-xs italic text-neutral-400">{scenario}</p>}
          <Textarea
            value={dialogue}
            onChange={(e) => setDialogue(e.target.value)}
            rows={4}
            placeholder="Paste the dialogue you want to reply to, or let AI suggest one…"
          />
          {suggest.isError && (
            <p className="text-sm text-red-600">{isApiError(suggest.error) ? suggest.error.message : 'Generation failed'}</p>
          )}
        </section>

        <section className="flex flex-col gap-2">
          <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-500">Strictness level</h2>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
            {LEVELS.map((l) => (
              <button
                key={l.id}
                onClick={() => changeLevel(l)}
                className={cn(
                  'rounded-xl border p-3 text-left transition-colors',
                  l.id === level.id ? 'border-brand-500 bg-brand-50' : 'border-neutral-200 bg-white hover:border-neutral-300',
                )}
              >
                <p className="text-sm font-semibold text-neutral-900">{l.id}. {l.title}</p>
                <p className="mt-0.5 text-xs text-neutral-500">{l.desc}</p>
              </button>
            ))}
          </div>
        </section>

        <section className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-500">Your phrase</h2>
            {attempts.length > 0 && (
              <div className="flex items-center gap-1.5 text-xs text-neutral-400">
                {attempts.map((a, i) => (
                  <span
                    key={i}
                    className={cn(
                      'rounded-full px-2 py-0.5 font-medium',
                      a.verdict.pass ? 'bg-emerald-100 text-emerald-700' : 'bg-neutral-100 text-neutral-500',
                    )}
                  >
                    {a.verdict.pass ? '✓' : (a.verdict.issues?.length ?? 0)}
                  </span>
                ))}
                <span>attempt {attempts.length}</span>
              </div>
            )}
          </div>
          <Textarea
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={onKeyDown}
            rows={4}
            placeholder="How would you reply in this dialogue? (Cmd+Enter to check)"
          />
          {check.isError && (
            <p className="text-sm text-red-600">{isApiError(check.error) ? check.error.message : 'Check failed'}</p>
          )}
          <div className="flex items-center gap-2 self-end">
            {result && (
              <Button variant="ghost" size="sm" onClick={restart}>
                <RotateCcw className="h-4 w-4" /> Reset
              </Button>
            )}
            <Button variant="brand" onClick={() => check.mutate({ dialogue, draft, level: level.id })} disabled={!canCheck}>
              {check.isPending ? <Spinner className="h-4 w-4" /> : <Pencil className="h-4 w-4" />}
              {attempts.length > 0 ? 'Check again' : 'Check'}
            </Button>
          </div>
        </section>

        {result && !check.isPending && (
          <section className="flex flex-col gap-3">
            {passed ? (
              <div className="flex flex-col gap-3 rounded-2xl border border-emerald-200 bg-emerald-50 p-5">
                <p className="flex items-center gap-2 font-medium text-emerald-800">
                  <CheckCircle2 className="h-5 w-5" /> {result.verdict.comment}
                </p>
                {attempts.length > 1 && (
                  <p className="text-sm text-emerald-700">Nailed it in {attempts.length} attempts.</p>
                )}
                <div className="flex gap-2">
                  {nextLevel && (
                    <Button variant="brand" size="sm" onClick={() => changeLevel(nextLevel)}>
                      Raise the bar: level {nextLevel.id} — {nextLevel.title}
                    </Button>
                  )}
                  <Button variant="secondary" size="sm" onClick={() => suggest.mutate()} disabled={suggest.isPending}>
                    {suggest.isPending ? <Spinner className="h-4 w-4" /> : <Sparkles className="h-4 w-4" />}
                    New situation
                  </Button>
                </div>
              </div>
            ) : (
              <div className="rounded-2xl border border-neutral-200 bg-white p-5">
                <p className="text-sm text-neutral-500">{result.verdict.comment}</p>
                <div className="mt-3 rounded-lg bg-neutral-50 p-3">
                  <HighlightedDraft text={result.draft} issues={issues} />
                </div>
                <ol className="mt-4 flex flex-col gap-2">
                  {issues.map((issue, i) => (
                    <li key={i} className="flex items-baseline gap-2 text-sm">
                      <span className="font-semibold text-neutral-400">{i + 1}.</span>
                      <span className={`shrink-0 rounded-full px-2 py-0.5 text-xs font-medium ${SEVERITY_STYLE[issue.severity]?.badge ?? ''}`}>
                        {SEVERITY_STYLE[issue.severity]?.label ?? issue.severity}
                      </span>
                      <span className="text-neutral-700">{issue.hint}</span>
                    </li>
                  ))}
                </ol>
                <p className="mt-3 text-xs text-neutral-400">No ready-made answer here — fix the phrase above and check again.</p>
              </div>
            )}
          </section>
        )}
      </div>
    </div>
  )
}
