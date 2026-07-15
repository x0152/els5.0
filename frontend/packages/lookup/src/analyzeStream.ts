const TOKEN_KEY = 'els.auth.token'

function authToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export interface AnalyzeStreamItem {
  text: string
  kind: string
  description: string
  translation?: string
  frequency: number
  cefr: string
  common: boolean
  existing: boolean
  total: number
  media_count: number
  media: Array<{
    media_id: string
    media_type: string
    title: string
    kind?: string
    series_title?: string
    season?: number
    episode?: number
    author?: string
    count: number
    spots?: Array<{ ref: number; example?: string }>
  }>
}

export interface AnalyzeStreamHandlers {
  onItem: (item: AnalyzeStreamItem) => void
  onError: (message: string) => void
  onDone: () => void
}

export async function streamAnalyze(text: string, context: string, h: AnalyzeStreamHandlers, signal?: AbortSignal) {
  const token = authToken()
  const res = await fetch('/api/v1/vocab/analyze/stream', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ text, context }),
    signal,
  })
  if (!res.ok || !res.body) {
    h.onError(`HTTP ${res.status}`)
    return
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buf = ''
  let doneCalled = false
  const handlers: AnalyzeStreamHandlers = {
    ...h,
    onDone: () => {
      if (doneCalled) return
      doneCalled = true
      h.onDone()
    },
  }
  try {
    for (;;) {
      const { value, done } = await reader.read()
      if (done) break
      buf += decoder.decode(value, { stream: true })
      let idx: number
      while ((idx = buf.indexOf('\n\n')) !== -1) {
        handleEvent(buf.slice(0, idx), handlers)
        buf = buf.slice(idx + 2)
      }
    }
    buf += decoder.decode()
    if (buf.trim()) handleEvent(buf, handlers)
    handlers.onDone()
  } catch (err) {
    if (err instanceof Error && err.name === 'AbortError') return
    h.onError(err instanceof Error ? err.message : 'stream error')
  }
}

function handleEvent(chunk: string, h: AnalyzeStreamHandlers) {
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
  if (event === 'item') h.onItem(p as unknown as AnalyzeStreamItem)
  else if (event === 'error') h.onError(String(p.message ?? 'error'))
  else if (event === 'done') h.onDone()
}
