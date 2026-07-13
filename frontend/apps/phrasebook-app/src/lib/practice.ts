import type { PracticeApi } from '@els/blocks'
import { api } from './api.ts'

export const practiceApi: PracticeApi = {
  listVariants: async (kind, number) =>
    (await api.learn.listPracticeVariants({ params: { path: { kind, number } } }))?.items ?? [],
  generateVariant: async (kind, number) => (await api.learn.generatePracticeVariant({ params: { path: { kind, number } } }))!,
  deleteVariant: async (id) => {
    await api.learn.deletePracticeVariant({ params: { path: { id } } })
  },
  getProgress: async (kind, number, variant) =>
    (await api.learn.getPracticeProgress({ params: { path: { kind, number }, query: { variant } } }))!,
  saveProgress: async (kind, number, variant, progress) => {
    await api.learn.savePracticeProgress({
      params: { path: { kind, number } },
      body: { variant, answers: progress.answers, completed: progress.completed },
    })
  },
  resetProgress: async (kind, number, variant) => {
    await api.learn.resetPracticeProgress({ params: { path: { kind, number }, query: { variant } } })
  },
  checkFree: async (input) => (await api.learn.checkPracticeAnswer({ body: input }))!,
}
