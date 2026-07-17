import type { DiaryComponents } from '@els/api-client'

export type Entry = DiaryComponents['schemas']['EntryOutput']
export type Correction = DiaryComponents['schemas']['CorrectionOutput']
export type Today = DiaryComponents['schemas']['TodayOutput']
export type TrainerIssue = DiaryComponents['schemas']['TrainerIssueOutput']
export type TrainerVerdict = DiaryComponents['schemas']['TrainerCheckOutput']

export function formatDay(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', { weekday: 'long', day: 'numeric', month: 'long' })
}
