import { useState } from 'react'
import { Bot, ChevronRight, Maximize2, RotateCcw, Trash2 } from 'lucide-react'
import { cn, ConfirmDialog, Select } from '@els/ui'
import { useChat } from '../hooks/useChat'
import { ChatMessages } from './ChatMessages'
import { ChatComposer } from './ChatComposer'
import { ImageViewerProvider } from './ImageViewer'

type Chat = ReturnType<typeof useChat>

function ModelSelect({ chat }: { chat: Chat }) {
  return (
    <Select
      value={chat.model}
      onChange={(e) => void chat.selectModel(e.target.value)}
      className="max-w-[140px] truncate px-2 py-1 text-xs text-neutral-600 hover:border-neutral-300"
    >
      {chat.model && !chat.models.includes(chat.model) && <option value={chat.model}>{chat.model}</option>}
      {chat.models.map((m) => (
        <option key={m} value={m}>
          {m}
        </option>
      ))}
    </Select>
  )
}

function IconButton({
  onClick,
  title,
  danger,
  children,
}: {
  onClick: () => void
  title: string
  danger?: boolean
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      title={title}
      className={cn(
        'rounded-lg p-1.5 text-neutral-400 transition-colors',
        danger ? 'hover:bg-red-50 hover:text-red-600' : 'hover:bg-neutral-100 hover:text-neutral-700',
      )}
    >
      {children}
    </button>
  )
}

export function ChatConversation({
  variant,
  active = true,
  onClose,
  onExpand,
}: {
  variant: 'panel' | 'page'
  active?: boolean
  onClose?: () => void
  onExpand?: () => void
}) {
  const chat = useChat(active)
  const page = variant === 'page'
  const [confirmingClear, setConfirmingClear] = useState(false)

  return (
    <ImageViewerProvider>
      <div className="flex h-full min-h-0 flex-col bg-neutral-50">
        <header className="flex h-14 shrink-0 items-center gap-2 border-b border-neutral-200 bg-white px-3">
          <div className={cn('flex flex-1 items-center gap-2', page && 'mx-auto w-full max-w-3xl px-3')}>
            {onClose && (
              <IconButton onClick={onClose} title="Collapse">
                <ChevronRight className="h-5 w-5" />
              </IconButton>
            )}
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
              <Bot className="h-4 w-4" />
            </span>
            <div className="min-w-0 flex-1">
              <h2 className="truncate text-sm font-semibold leading-tight text-neutral-900">Assistant</h2>
              {page && <p className="text-xs text-neutral-400">Your personal English tutor</p>}
            </div>
            <ModelSelect chat={chat} />
            {onExpand && (
              <IconButton onClick={onExpand} title="Open full screen">
                <Maximize2 className="h-4 w-4" />
              </IconButton>
            )}
            <IconButton onClick={() => void chat.reset()} title="Reset context">
              <RotateCcw className="h-4 w-4" />
            </IconButton>
            <IconButton onClick={() => setConfirmingClear(true)} title="Clear chat" danger>
              <Trash2 className="h-4 w-4" />
            </IconButton>
          </div>
        </header>

        {confirmingClear && (
          <ConfirmDialog
            title="Clear chat"
            description="Delete the whole conversation history? This cannot be undone."
            confirmLabel="Clear"
            onConfirm={() => {
              void chat.clear()
              setConfirmingClear(false)
            }}
            onClose={() => setConfirmingClear(false)}
          />
        )}

        <ChatMessages
          items={chat.items}
          streaming={chat.streaming}
          onRegenerate={() => void chat.regenerate()}
          onPickSuggestion={chat.send}
          variant={variant}
        />
        <ChatComposer streaming={chat.streaming} onSend={chat.send} onStop={chat.stop} variant={variant} />
      </div>
    </ImageViewerProvider>
  )
}
