import { Route, Routes } from 'react-router-dom'
import { Library } from './pages/Library.tsx'
import { Series } from './pages/Series.tsx'
import { Watch } from './pages/Watch.tsx'

export function FilmsAppRoutes() {
  return (
    <Routes>
      <Route index element={<Library />} />
      <Route path="series/:key" element={<Series />} />
      <Route path=":id" element={<Watch />} />
      <Route path="*" element={<Library />} />
    </Routes>
  )
}
