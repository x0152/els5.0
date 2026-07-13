import { Route, Routes } from 'react-router-dom'
import { LessonsList } from './pages/LessonsList.tsx'
import { LessonPlay } from './pages/LessonPlay.tsx'

export function PhrasebookAppRoutes() {
  return (
    <Routes>
      <Route index element={<LessonsList />} />
      <Route path=":num" element={<LessonPlay />} />
      <Route path="*" element={<LessonsList />} />
    </Routes>
  )
}
