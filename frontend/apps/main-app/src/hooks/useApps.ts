import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import { getAppIcon } from '../config/appIcons'
import { isSectionGroup, type Section, type SectionGroup } from '../config/sections'

export function useApps() {
  return useQuery({
    queryKey: ['account', 'apps'],
    queryFn: async (): Promise<Section[]> => {
      const res = await api.account.accountApps()
      return (res?.items ?? []).map(toSection)
    },
    staleTime: 60_000,
  })
}

export function groupSections(sections: Section[]): Record<SectionGroup, Section[]> {
  const acc: Record<SectionGroup, Section[]> = {
    personal: [],
    practice: [],
    books: [],
    media: [],
    admin: [],
  }
  for (const s of sections) if (s.group) acc[s.group].push(s)
  return acc
}

function toSection(app: {
  id: string
  name: string
  uri: string
  group?: string
  disabled: boolean
}): Section {
  return {
    id: app.id,
    label: app.name,
    to: app.uri,
    icon: getAppIcon(app.id),
    group: isSectionGroup(app.group) ? app.group : undefined,
    disabled: app.disabled,
  }
}
