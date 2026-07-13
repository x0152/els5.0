import { Route, Routes } from 'react-router-dom'
import { UnitsList } from './pages/UnitsList.tsx'
import { UnitPlay } from './pages/UnitPlay.tsx'

export function GrammarbookAppRoutes() {
  return (
    <Routes>
      <Route index element={<UnitsList />} />
      <Route path=":num" element={<UnitPlay />} />
      <Route path="*" element={<UnitsList />} />
    </Routes>
  )
}
