import { Route, Routes } from 'react-router-dom'
import { ListeningPage } from './ListeningPage.tsx'

export function ListeningAppRoutes() {
  return (
    <Routes>
      <Route index element={<ListeningPage />} />
      <Route path="*" element={<ListeningPage />} />
    </Routes>
  )
}
