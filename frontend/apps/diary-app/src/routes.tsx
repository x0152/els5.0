import { Route, Routes } from 'react-router-dom'
import { DiaryAppPage } from './DiaryAppPage.tsx'
import { HistoryPage } from './HistoryPage.tsx'

export function DiaryAppRoutes() {
  return (
    <Routes>
      <Route index element={<DiaryAppPage />} />
      <Route path="history" element={<HistoryPage />} />
      <Route path="*" element={<DiaryAppPage />} />
    </Routes>
  )
}
