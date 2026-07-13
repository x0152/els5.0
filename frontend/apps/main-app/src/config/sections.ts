import type { AppIcon } from './appIcons'

export type SectionGroup = 'personal' | 'practice' | 'books' | 'media' | 'admin'

export interface Section {
  id: string
  label: string
  to: string
  icon: AppIcon
  group?: SectionGroup
  disabled?: boolean
}

export const groupOrder: SectionGroup[] = ['personal', 'practice', 'books', 'media', 'admin']

export const groupNames: Record<SectionGroup, string> = {
  personal: 'PERSONAL',
  practice: 'PRACTICE',
  books: 'BOOKS',
  media: 'MEDIA',
  admin: 'ADMIN',
}

export function isSectionGroup(value: string | undefined): value is SectionGroup {
  return (
    value === 'personal' ||
    value === 'practice' ||
    value === 'books' ||
    value === 'media' ||
    value === 'admin'
  )
}
