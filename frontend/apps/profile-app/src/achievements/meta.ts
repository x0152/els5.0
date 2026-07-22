import type { IconType } from 'react-icons'
import {
  GiBookCover,
  GiConversation,
  GiCrossedSwords,
  GiFilmProjector,
  GiQuillInk,
  GiRunningShoe,
  GiScrollUnfurled,
  GiSpellBook,
  GiTalk,
  GiWeightLiftingUp,
} from 'react-icons/gi'

export interface AchievementMeta {
  icon: IconType
  title: string
}

const METRIC_ICONS: Record<string, IconType> = {
  quests_completed: GiCrossedSwords,
  workouts_completed: GiWeightLiftingUp,
  vocab_words: GiSpellBook,
  diary_entries: GiQuillInk,
  chat_messages: GiConversation,
}

const METRIC_TITLES: Record<string, (n: number) => string> = {
  quests_completed: (n) => (n === 1 ? 'First quest completed' : `${n} quests completed`),
  workouts_completed: (n) => `${n} workouts completed`,
  vocab_words: (n) => `${n} words collected`,
  diary_entries: (n) => `${n} diary entries`,
  chat_messages: (n) => `${n} messages to the assistant`,
}

export const GROUP_LABELS: Record<string, string> = {
  quests_completed: 'Quests',
  workouts_completed: 'Workouts',
  vocab_words: 'Vocabulary',
  diary_entries: 'Diary',
  chat_messages: 'Assistant',
}

const CHECKLIST_META: Record<string, AchievementMeta> = {
  first_film: { icon: GiFilmProjector, title: 'First film watched with subtitles' },
  first_quest: { icon: GiCrossedSwords, title: 'First own quest generated and completed' },
  first_article: { icon: GiScrollUnfurled, title: 'First page added by URL and read' },
  first_workout: { icon: GiRunningShoe, title: 'First workout completed' },
  first_chat: { icon: GiTalk, title: 'First task from the assistant' },
  first_words: { icon: GiBookCover, title: 'First 5 words in vocabulary' },
  first_chapter: { icon: GiSpellBook, title: 'First book chapter completed' },
}

export function achievementMeta(item: { id: string; metric: string; threshold: number }): AchievementMeta {
  const checklist = CHECKLIST_META[item.id]
  if (checklist) return checklist
  return {
    icon: METRIC_ICONS[item.metric] ?? GiScrollUnfurled,
    title: METRIC_TITLES[item.metric]?.(item.threshold) ?? item.id,
  }
}
