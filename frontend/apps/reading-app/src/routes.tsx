import { Route, Routes } from 'react-router-dom'
import { ReadingPage } from './ReadingPage.tsx'

export function ReadingAppRoutes() {
  return (
    <Routes>
      <Route index element={<ReadingPage />} />
      <Route path="*" element={<ReadingPage />} />
    </Routes>
  )
}
