import { useCallback, useEffect, useRef, useState } from 'react'
import { api } from '../lib/api'
import { buildItems, streamChat, type ChatItem, type ChatSegment, type ChatStep } from '../lib/chat'

let seq = 0
const nid = () => `m${Date.now()}_${seq++}`

type Assistant = Extract<ChatItem, { kind: 'assistant' }>

const appendText = (m: Assistant, delta: string): Assistant => {
  const segs = [...m.segments]
  const last = segs[segs.length - 1]
  if (!last || last.steps.length > 0) segs.push({ id: nid(), text: delta, steps: [] })
  else segs[segs.length - 1] = { ...last, text: last.text + delta }
  return { ...m, segments: segs }
}

const addStep = (m: Assistant, step: ChatStep): Assistant => {
  const segs: ChatSegment[] = [...m.segments]
  const last = segs[segs.length - 1]
  if (last) segs[segs.length - 1] = { ...last, steps: [...last.steps, step] }
  else segs.push({ id: nid(), text: '', steps: [step] })
  return { ...m, segments: segs }
}

const endStep = (m: Assistant, id: string, result: string): Assistant => ({
  ...m,
  segments: m.segments.map((seg) => ({
    ...seg,
    steps: seg.steps.map((st) => (st.id === id ? { ...st, result, done: true } : st)),
  })),
})

export function useChat(active: boolean) {
  const [items, setItems] = useState<ChatItem[]>([])
  const [model, setModel] = useState('')
  const [models, setModels] = useState<string[]>([])
  const [streaming, setStreaming] = useState(false)
  const [initializing, setInitializing] = useState(false)
  const abortRef = useRef<AbortController | null>(null)
  const itemsRef = useRef<ChatItem[]>([])
  useEffect(() => {
    itemsRef.current = items
  }, [items])

  useEffect(() => {
    if (!active) return
    setInitializing(true)
    let alive = true
    const pendingBubble: ChatItem = { kind: 'assistant', id: 'pending', segments: [], pending: true }
    const pollWhileGenerating = async () => {
      setStreaming(true)
      try {
        for (;;) {
          await new Promise((r) => setTimeout(r, 1500))
          if (!alive) return
          const h = await api.ai.aiHistory()
          if (!alive) return
          const built = buildItems(h?.messages ?? [])
          if (h?.generating) {
            setItems([...built, pendingBubble])
          } else {
            setItems(built)
            break
          }
        }
      } catch {
        /* ignore */
      } finally {
        if (alive) setStreaming(false)
      }
    }
    void (async () => {
      try {
        const h = await api.ai.aiHistory()
        if (!alive) return
        const built = buildItems(h?.messages ?? [])
        setItems(h?.generating ? [...built, pendingBubble] : built)
        setModel(h?.model ?? '')
        if (h?.generating) void pollWhileGenerating()
        else setStreaming(false)
      } catch {
        setStreaming(false)
      }
      if (alive) setInitializing(false)
      try {
        const r = await api.ai.aiModels()
        if (!alive) return
        setModels((r?.models ?? []).map((m) => m.id))
        setModel((prev) => prev || r?.selected || r?.default || '')
      } catch {
        /* ignore */
      }
    })()
    return () => {
      alive = false
    }
  }, [active])

  const runStream = useCallback(async (content: string, regenerate: boolean) => {
    const assistantId = nid()
    setItems((prev) => [...prev, { kind: 'assistant', id: assistantId, segments: [], pending: true }])
    setStreaming(true)
    const ac = new AbortController()
    abortRef.current = ac
    const patch = (fn: (m: Assistant) => Assistant) =>
      setItems((prev) => prev.map((m) => (m.id === assistantId && m.kind === 'assistant' ? fn(m) : m)))
    try {
      await streamChat(
        content,
        {
          onText: (d) => patch((m) => appendText(m, d)),
          onToolStart: (s) =>
            patch((m) => addStep(m, { id: s.id, tool: s.tool, label: s.label, icon: s.icon, args: s.args, done: false })),
          onToolEnd: (s) => patch((m) => endStep(m, s.id, s.result)),
          onError: (msg) => patch((m) => appendText(m, `\n\n⚠️ ${msg}`)),
          onDone: () => patch((m) => ({ ...m, pending: false })),
        },
        ac.signal,
        regenerate,
      )
    } catch {
      /* aborted */
    }
    patch((m) => ({ ...m, pending: false }))
    setStreaming(false)
  }, [])

  const send = useCallback(
    async (text: string) => {
      const content = text.trim()
      if (!content || streaming) return
      setItems((prev) => [...prev, { kind: 'user', id: nid(), content }])
      await runStream(content, false)
    },
    [streaming, runStream],
  )

  const askRef = useRef<string | null>(null)
  const flushAsk = useCallback(() => {
    const text = askRef.current
    if (!text || !active || initializing || streaming) return
    askRef.current = null
    void send(`Explain this and how it's used: «${text}»`)
  }, [active, initializing, streaming, send])

  useEffect(() => {
    const onAsk = (e: Event) => {
      const text = ((e as CustomEvent<string>).detail ?? '').trim()
      if (text.length < 2) return
      askRef.current = text
      flushAsk()
    }
    document.addEventListener('els:ask', onAsk)
    return () => document.removeEventListener('els:ask', onAsk)
  }, [flushAsk])

  useEffect(() => {
    flushAsk()
  }, [flushAsk])

  const regenerate = useCallback(async () => {
    if (streaming) return
    const i = itemsRef.current.map((m) => m.kind).lastIndexOf('assistant')
    if (i === -1) return
    setItems((prev) => prev.slice(0, i))
    await runStream('', true)
  }, [streaming, runStream])

  const stop = useCallback(() => {
    if (!abortRef.current) return
    abortRef.current.abort()
    abortRef.current = null
    setStreaming(false)
  }, [])

  const reset = useCallback(async () => {
    await api.ai.aiResetContext()
    setItems((prev) => [...prev, { kind: 'separator', id: nid() }])
  }, [])

  const clear = useCallback(async () => {
    await api.ai.aiClearChat()
    setItems([])
  }, [])

  const selectModel = useCallback(
    async (m: string) => {
      const prev = model
      setModel(m)
      try {
        await api.ai.aiSetModel({ body: { model: m } })
      } catch {
        setModel(prev)
      }
    },
    [model],
  )

  return { items, model, models, streaming, send, stop, reset, clear, selectModel, regenerate }
}
