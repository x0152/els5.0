import { Route, Routes } from 'react-router-dom'
import { StudioAppPage } from './StudioAppPage.tsx'

export function StudioAppRoutes() {
  return (
    <Routes>
      <Route index element={<StudioAppPage />} />
      <Route path="*" element={<StudioAppPage />} />
    </Routes>
  )
}
