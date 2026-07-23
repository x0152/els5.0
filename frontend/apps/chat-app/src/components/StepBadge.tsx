import { useState } from 'react'
import {
  AlertTriangle,
  BookOpen,
  Captions,
  CheckCircle2,
  Clock,
  Film,
  Library,
  List,
  Loader2,
  Swords,
  Wrench,
} from 'lucide-react'
import type { ChatStep } from '../lib/chat'

function StepIcon({ icon, running }: { icon?: string; running: boolean }) {
  if (running) return <Loader2 size={12} className="shrink-0 animate-spin" />
  if (icon === 'alert-triangle') return <AlertTriangle size={12} className="shrink-0" />
  if (icon === 'clock') return <Clock size={12} className="shrink-0" />
  if (icon === 'list') return <List size={12} className="shrink-0" />
  if (icon === 'film') return <Film size={12} className="shrink-0" />
  if (icon === 'captions') return <Captions size={12} className="shrink-0" />
  if (icon === 'library') return <Library size={12} className="shrink-0" />
  if (icon === 'book-open') return <BookOpen size={12} className="shrink-0" />
  if (icon === 'swords') return <Swords size={12} className="shrink-0" />
  return <Wrench size={12} className="shrink-0" />
}

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

function argsSummary(raw?: string): string {
  if (!raw) return ''
  try {
    const obj = JSON.parse(raw) as Record<string, unknown>
    for (const [key, value] of Object.entries(obj)) {
      if (/(^|_)id$/.test(key)) continue
      const s = typeof value === 'string' ? value : JSON.stringify(value)
      if (!s || UUID_RE.test(s)) continue
      return s.length > 48 ? s.slice(0, 45) + '…' : s
    }
    return ''
  } catch {
    return ''
  }
}

export function StepBadge({ step }: { step: ChatStep }) {
  const [open, setOpen] = useState(false)
  const running = !step.done
  const summary = argsSummary(step.args)
  const canExpand = !running && (!!step.result || !!step.args)

  return (
    <div className="w-full">
      <button
        type="button"
        onClick={() => canExpand && setOpen((v) => !v)}
        className={`inline-flex max-w-full items-center gap-1.5 rounded-md border px-2 py-1 text-[12px] font-mono leading-snug transition-colors ${
          running
            ? 'border-amber-300 bg-amber-50 text-amber-800'
            : 'border-neutral-300 bg-white text-neutral-700 hover:border-neutral-400 hover:bg-neutral-50'
        }`}
      >
        <StepIcon icon={step.icon} running={running} />
        <span className="min-w-0 max-w-[40ch] truncate">
          <span className="font-medium">{step.label}</span>
          {summary && <span className="ml-1.5 text-neutral-500">{summary}</span>}
        </span>
        {!running && <CheckCircle2 size={11} className="shrink-0 text-brand-500" />}
      </button>
      {open && (
        <pre className="mt-1 max-h-48 overflow-auto whitespace-pre-wrap break-words rounded-md border border-neutral-200 bg-neutral-50 p-2 text-[11.5px] font-mono text-neutral-600">
          {step.result || step.args}
        </pre>
      )}
    </div>
  )
}
