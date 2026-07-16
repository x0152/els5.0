import { Route, Routes } from 'react-router-dom'
import { SpeakingPage } from './pages/SpeakingPage.tsx'
import { SoundsPage } from './pages/SoundsPage.tsx'

export function SpeakingAppRoutes() {
  return (
    <Routes>
      <Route index element={<SpeakingPage />} />
      <Route path="sounds" element={<SoundsPage />} />
      <Route path="*" element={<SpeakingPage />} />
    </Routes>
  )
}
