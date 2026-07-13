import { Route, Routes } from 'react-router-dom'
import { JournalPage } from './JournalPage.tsx'

export function JournalAppRoutes() {
  return (
    <Routes>
      <Route index element={<JournalPage />} />
      <Route path="*" element={<JournalPage />} />
    </Routes>
  )
}
