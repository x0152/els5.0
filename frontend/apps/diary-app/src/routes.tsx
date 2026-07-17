import { NavLink, Outlet, Route, Routes, useLocation } from 'react-router-dom'
import { DiaryAppPage } from './DiaryAppPage.tsx'
import { HistoryPage } from './HistoryPage.tsx'
import { TrainerPage } from './TrainerPage.tsx'

const PAGES = ['history', 'trainer']

function useBasePath(): string {
  const { pathname } = useLocation()
  const cuts = PAGES.map((p) => pathname.indexOf(`/${p}`)).filter((i) => i >= 0)
  const base = cuts.length ? pathname.slice(0, Math.min(...cuts)) : pathname
  return base.replace(/\/$/, '')
}

function Layout() {
  const base = useBasePath()
  const link = ({ isActive }: { isActive: boolean }) =>
    `rounded-md px-3 py-1.5 text-sm font-medium ${isActive ? 'bg-brand-600 text-white' : 'text-neutral-600 hover:bg-neutral-100'}`
  return (
    <div className="flex h-full flex-col">
      <nav className="flex shrink-0 gap-1 border-b border-neutral-200 bg-white px-6 py-2">
        <NavLink to={base || '/'} end className={link}>
          Diary
        </NavLink>
        <NavLink to={`${base}/history`} className={link}>
          History
        </NavLink>
        <NavLink to={`${base}/trainer`} className={link}>
          Trainer
        </NavLink>
      </nav>
      <div className="min-h-0 flex-1">
        <Outlet />
      </div>
    </div>
  )
}

export function DiaryAppRoutes() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<DiaryAppPage />} />
        <Route path="history" element={<HistoryPage />} />
        <Route path="trainer" element={<TrainerPage />} />
        <Route path="*" element={<DiaryAppPage />} />
      </Route>
    </Routes>
  )
}
