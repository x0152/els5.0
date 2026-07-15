import { Route, Routes } from 'react-router-dom'
import { VocabCatalog } from './pages/VocabCatalog.tsx'
import { Practice } from './pages/Practice.tsx'
import { Cards } from './pages/Cards.tsx'

export function VocabAppRoutes() {
  return (
    <Routes>
      <Route index element={<VocabCatalog />} />
      <Route path="practice" element={<Practice />} />
      <Route path="cards" element={<Cards />} />
      <Route path="*" element={<VocabCatalog />} />
    </Routes>
  )
}
