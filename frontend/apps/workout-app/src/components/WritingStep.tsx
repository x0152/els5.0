import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { emitTextEvents } from '@els/core-events'
import { Button, Spinner, Textarea, cn } from '@els/ui'
import { Send } from 'lucide-react'
import { api } from '../lib/api.ts'
import type { StepOutcome, Writing } from '../lib/types.ts'
import { ContinueButton, StepShell } from './StepShell.tsx'

type Issue = { fragment: string; severity: 'grammar' | 'style' | 'native'; hint: string }

const SEVERITY_CLS: Record<Issue['severity'], string> = {
  grammar: 'bg-rose-50 text-rose-800 ring-rose-200',
  style: 'bg-amber-50 text-amber-800 ring-amber-200',
  native: 'bg-sky-50 text-sky-800 ring-sky-200',
}

export function WritingStep({ writing, onDone }: { writing: Writing; onDone: (outcome: StepOutcome) => void }) {
  const [draft, setDraft] = useState('')
  const [attempts, setAttempts] = useState(0)
  const [verdict, setVerdict] = useState<{ pass: boolean; comment: string; issues: Issue[] } | null>(null)

  const check = useMutation({
    mutationFn: () => api.writing.writingTrainerCheck({ body: { dialogue: writing.dialogue, draft: draft.trim(), level: 2 } }),
    onSuccess: (data) => {
      setAttempts((n) => n + 1)
      setVerdict(data ? { pass: data.pass, comment: data.comment, issues: data.issues ?? [] } : null)
      if (data?.pass) emitTextEvents(api, 'writing', [draft.trim()], { app: 'workout' })
    },
  })

  const finish = () => onDone({ score: Math.max(100 - (attempts - 1) * 15, 40) })

  return (
    <StepShell>
      <div className="flex flex-col gap-2 rounded-xl bg-neutral-50 p-4">
        {writing.dialogue.split('\n').map((line, i) => (
          <p key={i} className="text-[15px] leading-6 text-neutral-800">
            {line}
          </p>
        ))}
      </div>

      <Textarea
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        rows={4}
        placeholder="Write your reply in English…"
        disabled={verdict?.pass}
      />

      {verdict && (
        <div className="flex flex-col gap-2">
          <p className={cn('text-sm font-medium', verdict.pass ? 'text-emerald-600' : 'text-neutral-700')}>{verdict.comment}</p>
          {verdict.issues.map((issue, i) => (
            <div key={i} className={cn('rounded-lg px-3 py-2 text-sm ring-1', SEVERITY_CLS[issue.severity])}>
              <span className="font-semibold">“{issue.fragment}”</span> — {issue.hint}
            </div>
          ))}
        </div>
      )}

      {verdict?.pass ? (
        <ContinueButton onClick={finish} />
      ) : (
        <div className="mt-auto flex items-center justify-end gap-2 border-t border-neutral-100 pt-4">
          {verdict && (
            <Button variant="ghost" onClick={finish}>
              Skip
            </Button>
          )}
          <Button variant="brand" onClick={() => check.mutate()} disabled={!draft.trim() || check.isPending}>
            {check.isPending ? <Spinner className="h-4 w-4" /> : <Send className="h-4 w-4" />}
            {verdict ? 'Check again' : 'Check'}
          </Button>
        </div>
      )}
    </StepShell>
  )
}
