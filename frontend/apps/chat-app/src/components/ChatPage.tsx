import { ChatConversation } from './ChatConversation'

export function ChatPage() {
  return (
    <div className="h-full w-full">
      <ChatConversation variant="page" active />
    </div>
  )
}
