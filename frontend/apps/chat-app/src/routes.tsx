import { Route, Routes } from 'react-router-dom'
import { ChatPage } from './components/ChatPage'

export function ChatAppRoutes() {
  return (
    <Routes>
      <Route index element={<ChatPage />} />
      <Route path="*" element={<ChatPage />} />
    </Routes>
  )
}
