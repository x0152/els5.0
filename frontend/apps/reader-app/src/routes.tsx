import { Route, Routes } from 'react-router-dom'
import { Library } from './pages/Library.tsx'
import { Collection } from './pages/Collection.tsx'
import { Read } from './pages/Read.tsx'

export function ReaderAppRoutes() {
  return (
    <Routes>
      <Route index element={<Library />} />
      <Route path="collection/:key" element={<Collection />} />
      <Route path=":id" element={<Read />} />
      <Route path="*" element={<Library />} />
    </Routes>
  )
}
