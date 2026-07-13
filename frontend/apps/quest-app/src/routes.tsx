import { Route, Routes } from 'react-router-dom'
import { QuestHub } from './pages/QuestHub.tsx'
import { MissionPlay } from './pages/MissionPlay.tsx'
import { AnimationsLab } from './pages/AnimationsLab.tsx'

export function QuestAppRoutes() {
  return (
    <Routes>
      <Route index element={<QuestHub />} />
      <Route path="dev/animations" element={<AnimationsLab />} />
      <Route path=":id" element={<MissionPlay />} />
      <Route path="*" element={<QuestHub />} />
    </Routes>
  )
}
