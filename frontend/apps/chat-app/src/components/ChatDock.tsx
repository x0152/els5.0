import { MessageCircle } from 'lucide-react'
import { useLocation, useNavigate } from 'react-router-dom'
import { cn } from '@els/ui'
import { ChatConversation } from './ChatConversation'

export function ChatDock({ open, onOpen, onClose }: { open: boolean; onOpen: () => void; onClose: () => void }) {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const onChatPage = pathname.startsWith('/v1/chat')

  const expand = () => {
    onClose()
    navigate('/v1/chat')
  }

  return (
    <>
      {!onChatPage && (
        <button
          onClick={onOpen}
          title="Assistant"
          aria-label="Open assistant"
          className={cn(
            'group fixed z-40 hidden items-center justify-center bg-brand-600 text-white transition-all duration-300 hover:bg-brand-700',
            'right-0 top-1/2 -translate-y-1/2 rounded-l-xl rounded-r-none px-1.5 py-4 shadow-md active:scale-95',
            'md:flex md:flex-col md:gap-2',
            open && 'pointer-events-none scale-90 opacity-0',
          )}
        >
          <MessageCircle className="h-5 w-5" />
          <span className="text-[10px] font-semibold uppercase tracking-widest [writing-mode:vertical-rl]">
            Assistant
          </span>
        </button>
      )}

      <div
        onClick={onClose}
        className={cn(
          'fixed inset-0 z-40 bg-neutral-900/20 backdrop-blur-[1px] transition-opacity duration-300 md:hidden',
          open ? 'opacity-100' : 'pointer-events-none opacity-0',
        )}
      />

      <div
        className={cn(
          'fixed right-0 top-0 bottom-0 z-50 flex w-full flex-col border-l border-neutral-200 shadow-2xl transition-transform duration-300 ease-out sm:w-[460px] sm:shadow-none',
          open ? 'translate-x-0' : 'pointer-events-none translate-x-full',
        )}
      >
        <ChatConversation variant="panel" active={open} onClose={onClose} onExpand={expand} />
      </div>
    </>
  )
}
