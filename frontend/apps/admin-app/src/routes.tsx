import { Route, Routes } from 'react-router-dom'
import { AdminAppPage } from './AdminApp.tsx'

export function AdminAppRoutes() {
  return (
    <Routes>
      <Route path="*" element={<AdminAppPage />} />
    </Routes>
  )
}
