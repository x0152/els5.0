import type { api } from './api.ts'

export type Area = NonNullable<NonNullable<Awaited<ReturnType<typeof api.studio.studioListAreas>>>['items']>[number]
export type Item = NonNullable<NonNullable<Awaited<ReturnType<typeof api.studio.studioListItems>>>['items']>[number]

export const isDue = (i: Item) => !!i.next_review_at && new Date(i.next_review_at) <= new Date()
