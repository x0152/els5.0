import { Route, Routes } from 'react-router-dom'
import { WorkoutAppPage } from './WorkoutAppPage.tsx'
import { LessonPage } from './LessonPage.tsx'

export function WorkoutAppRoutes() {
  return (
    <Routes>
      <Route index element={<WorkoutAppPage />} />
      <Route path="lesson/:id" element={<LessonPage />} />
      <Route path="*" element={<WorkoutAppPage />} />
    </Routes>
  )
}
