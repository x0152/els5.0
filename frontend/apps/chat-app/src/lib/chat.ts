import { readAgentContext } from '@els/ui'
import { getToken } from './token'

export type ChatStep = {
  id: string
  tool: string
  label: string
  icon?: string
  args?: string
  result?: string
  done: boolean
}

export type ChatSegment = { id: string; text: string; steps: ChatStep[] }

export type ChatItem =
  | { kind: 'user'; id: string; content: string }
  | { kind: 'assistant'; id: string; segments: ChatSegment[]; model?: string; pending?: boolean }
  | { kind: 'separator'; id: string }

type RawToolCall = { id?: string; name?: string; arguments?: string }
type RawMessage = {
  id: string
  role: string
  content: string
  tool_calls?: RawToolCall[] | null
  tool_call_id?: string
  tool_name?: string
  model?: string
}

const ICON_BY_TOOL: Record<string, string> = {
  read_recent_errors: 'alert-triangle',
  current_time: 'clock',
  list_book_units: 'list',
  read_book_unit: 'book-open',
  list_films: 'film',
  read_film_subtitles: 'captions',
  list_books: 'library',
  read_book_text: 'book-open',
  list_quests: 'swords',
  read_quest: 'swords',
}

const LABEL_BY_TOOL: Record<string, string> = {
  read_recent_errors: 'Reading recent mistakes',
  current_time: 'Current time',
  list_book_units: 'Listing units',
  read_book_unit: 'Reading a unit',
  list_films: 'Listing films',
  read_film_subtitles: 'Reading subtitles',
  list_books: 'Listing books',
  read_book_text: 'Reading the book',
  list_quests: 'Listing quests',
  read_quest: 'Reading the quest',
}

export function buildItems(messages: RawMessage[]): ChatItem[] {
  const items: ChatItem[] = []
  let current: Extract<ChatItem, { kind: 'assistant' }> | null = null
  const flush = () => {
    if (current) {
      items.push(current)
      current = null
    }
  }

  for (const m of messages) {
    if (m.role === 'user') {
      flush()
      items.push({ kind: 'user', id: m.id, content: m.content })
    } else if (m.role === 'separator') {
      flush()
      items.push({ kind: 'separator', id: m.id })
    } else if (m.role === 'assistant') {
      if (!current) current = { kind: 'assistant', id: m.id, segments: [] }
      const steps: ChatStep[] = []
      for (const tc of m.tool_calls ?? []) {
        if (!tc.id) continue
        steps.push({
          id: tc.id,
          tool: tc.name ?? '',
          label: LABEL_BY_TOOL[tc.name ?? ''] ?? tc.name ?? '',
          icon: ICON_BY_TOOL[tc.name ?? ''],
          args: tc.arguments,
          done: true,
        })
      }
      current.segments.push({ id: m.id, text: m.content ?? '', steps })
      if (m.model) current.model = m.model
    } else if (m.role === 'tool') {
      for (const seg of current?.segments ?? []) {
        const step = seg.steps.find((s) => s.id === m.tool_call_id)
        if (step) {
          step.result = m.content
          step.done = true
        }
      }
    }
  }
  flush()
  return items
}

export type StreamHandlers = {
  onText: (delta: string) => void
  onToolStart: (s: { id: string; tool: string; label: string; icon?: string; args?: string }) => void
  onToolEnd: (s: { id: string; tool: string; result: string }) => void
  onError: (message: string) => void
  onDone: () => void
}

export async function streamChat(message: string, h: StreamHandlers, signal?: AbortSignal, regenerate = false) {
  const token = getToken()
  const { view } = readAgentContext()
  const res = await fetch('/api/v1/ai/stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ message, regenerate, view }),
    signal,
  })
  if (!res.ok || !res.body) {
    h.onError(`HTTP ${res.status}`)
    return
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  for (;;) {
    const { value, done } = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })
    let idx: number
    while ((idx = buf.indexOf('\n\n')) !== -1) {
      handleEvent(buf.slice(0, idx), h)
      buf = buf.slice(idx + 2)
    }
  }
}

function handleEvent(chunk: string, h: StreamHandlers) {
  let event = 'message'
  let data = ''
  for (const line of chunk.split('\n')) {
    if (line.startsWith('event:')) event = line.slice(6).trim()
    else if (line.startsWith('data:')) data += line.slice(5).trim()
  }
  if (!data) return
  let p: Record<string, unknown>
  try {
    p = JSON.parse(data)
  } catch {
    return
  }
  switch (event) {
    case 'text':
      h.onText(String(p.delta ?? ''))
      break
    case 'tool_start':
      h.onToolStart(p as never)
      break
    case 'tool_end':
      h.onToolEnd(p as never)
      break
    case 'error':
      h.onError(String(p.message ?? 'error'))
      break
    case 'done':
      h.onDone()
      break
  }
}
