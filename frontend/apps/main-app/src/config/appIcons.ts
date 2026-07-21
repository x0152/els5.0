import type { ComponentType } from 'react'
import {
  Bot,
  BookA,
  BookMarked,
  BookOpen,
  BookOpenText,
  BookText,
  Dumbbell,
  Film,
  FileText,
  GraduationCap,
  Headphones,
  LayoutGrid,
  Library,
  Mic,
  NotebookPen,
  PencilRuler,
  PenLine,
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
  diary: PenLine,
  docs: FileText,
  films: Film,
  journal: NotebookPen,
  grammarbook: BookOpen,
  listening: Headphones,
  reading: BookOpenText,
  profile: UserCircle,
  quest: Swords,
  reader: Library,
  settings: Settings,
  speaking: Mic,
  vocab: BookMarked,
  wordbook: BookA,
  workout: Dumbbell,
  writing: PencilRuler,
  phrasebook: BookText,
  essentialbook: GraduationCap,
}

export const defaultAppIcon: AppIcon = LayoutGrid

export function getAppIcon(id: string): AppIcon {
  return appIconById[id] ?? defaultAppIcon
}
