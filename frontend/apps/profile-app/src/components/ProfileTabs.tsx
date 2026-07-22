import { useLocation, useNavigate } from 'react-router-dom'
import { Tabs } from '@els/ui'

export function ProfileTabs() {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const value = pathname.includes('/achievements') ? 'achievements' : 'overview'
  return (
    <Tabs
      value={value}
      onChange={(v) => navigate(v === 'achievements' ? '/v1/profile/achievements' : '/v1/profile')}
      options={[
        { value: 'overview', label: 'Overview' },
        { value: 'achievements', label: 'Achievements' },
      ]}
    />
  )
}
