import { Route, Routes } from 'react-router-dom'
import { TrainerPage } from './TrainerPage.tsx'

export function WritingAppRoutes() {
  return (
    <Routes>
      <Route index element={<TrainerPage />} />
      <Route path="*" element={<TrainerPage />} />
    </Routes>
  )
}
