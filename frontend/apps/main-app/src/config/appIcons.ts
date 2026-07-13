import type { ComponentType } from 'react'
import {
  Bot,
  BookA,
  BookMarked,
  BookOpen,
  BookText,
  Film,
  FileText,
  GraduationCap,
  LayoutGrid,
  Library,
  NotebookPen,
  Settings,
  Shield,
  Swords,
  UserCircle,
  type LucideProps,
} from 'lucide-react'

export type AppIcon = ComponentType<LucideProps>

export const appIconById: Record<string, AppIcon> = {
  account: UserCircle,
  admin: Shield,
  ai: Bot,
  chat: Bot,
  docs: FileText,
  films: Film,
  journal: NotebookPen,
  grammarbook: BookOpen,
  profile: UserCircle,
  quest: Swords,
  reader: Library,
  settings: Settings,
  vocab: BookMarked,
  wordbook: BookA,
  phrasebook: BookText,
  essentialbook: GraduationCap,
}

export const defaultAppIcon: AppIcon = LayoutGrid

export function getAppIcon(id: string): AppIcon {
  return appIconById[id] ?? defaultAppIcon
}
