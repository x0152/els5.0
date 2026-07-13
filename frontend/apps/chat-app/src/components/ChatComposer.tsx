import { useEffect, useImperativeHandle, useRef, useState, type FormEvent, type KeyboardEvent, type Ref } from 'react'
import { ArrowUp, Square } from 'lucide-react'
import { cn } from '@els/ui'

const MAX_HEIGHT = 200

export type ComposerHandle = { focus: () => void }

export function ChatComposer({
  streaming,
  onSend,
  onStop,
  variant = 'panel',
  ref,
}: {
  streaming: boolean
  onSend: (text: string) => void
  onStop: () => void
  variant?: 'panel' | 'page'
  ref?: Ref<ComposerHandle>
}) {
  const [text, setText] = useState('')
  const taRef = useRef<HTMLTextAreaElement | null>(null)
  const page = variant === 'page'

  useImperativeHandle(ref, () => ({ focus: () => taRef.current?.focus() }), [])

  useEffect(() => {
    const el = taRef.current
    if (!el) return
    el.style.height = 'auto'
    el.style.height = `${Math.min(MAX_HEIGHT, el.scrollHeight)}px`
  }, [text])

  const submit = () => {
    if (!text.trim()) return
    onSend(text)
    setText('')
  }

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (streaming) {
      onStop()
      return
    }
    submit()
  }

  const onKey = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      if (streaming) return
      submit()
    }
  }

  const canSend = !streaming && text.trim().length > 0

  return (
    <form onSubmit={handleSubmit} className="border-t border-neutral-200 bg-white p-3">
      <div className={cn('mx-auto', page && 'max-w-3xl')}>
        <div className="flex items-end gap-2 rounded-2xl border border-neutral-200 bg-white px-3 py-2 transition-colors focus-within:border-brand-400 focus-within:ring-2 focus-within:ring-brand-100">
          <textarea
            ref={taRef}
            rows={1}
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={onKey}
            placeholder="Message the assistant…"
            className="flex-1 resize-none bg-transparent py-1 text-sm leading-relaxed text-neutral-900 placeholder:text-neutral-400 focus:outline-none"
            style={{ maxHeight: MAX_HEIGHT }}
          />
          <button
            type="submit"
            title={streaming ? 'Stop' : 'Send (Enter)'}
            disabled={!streaming && !canSend}
            className={cn(
              'inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors',
              streaming
                ? 'bg-neutral-900 text-white hover:bg-neutral-700'
                : canSend
                  ? 'bg-brand-600 text-white hover:bg-brand-700'
                  : 'cursor-not-allowed bg-neutral-200 text-neutral-400',
            )}
          >
            {streaming ? <Square size={13} /> : <ArrowUp size={15} />}
          </button>
        </div>
        <div className="mt-1.5 flex justify-end text-[10.5px] text-neutral-400">Enter — send · Shift+Enter — newline</div>
      </div>
    </form>
  )
}
