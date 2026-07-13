import { Route, Routes } from 'react-router-dom'
import { ProfilePage } from './ProfilePage.tsx'

export function ProfileAppRoutes() {
  return (
    <Routes>
      <Route index element={<ProfilePage />} />
      <Route path="*" element={<ProfilePage />} />
    </Routes>
  )
}
