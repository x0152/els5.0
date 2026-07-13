import { Route, Routes } from 'react-router-dom'
import { SettingsAppPage } from './SettingsAppPage.tsx'

export function SettingsAppRoutes() {
  return (
    <Routes>
      <Route index element={<SettingsAppPage />} />
      <Route path="*" element={<SettingsAppPage />} />
    </Routes>
  )
}
