import { Route, Routes } from 'react-router-dom'
import { ProfilePage } from './ProfilePage.tsx'
import { AchievementsPage } from './AchievementsPage.tsx'

export function ProfileAppRoutes() {
  return (
    <Routes>
      <Route index element={<ProfilePage />} />
      <Route path="achievements" element={<AchievementsPage />} />
      <Route path="*" element={<ProfilePage />} />
    </Routes>
  )
}
