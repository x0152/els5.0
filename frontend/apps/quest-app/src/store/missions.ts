import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api.ts'
import { missionBusy, summaryBusy } from '../lib/helpers.ts'
import type { ActiveReply, Mission, MissionSummary } from '../lib/types.ts'

const listKey = ['quest', 'missions'] as const
const missionKey = (id: string) => ['quest', 'mission', id] as const

export function useMissions() {
  return useQuery({
    queryKey: listKey,
    queryFn: async (): Promise<MissionSummary[]> => {
      const res = await api.quest.listQuestMissions()
      return res?.missions ?? []
    },
    refetchInterval: (q) => {
      const list = q.state.data as MissionSummary[] | undefined
      return list && list.some(summaryBusy) ? 2500 : false
    },
  })
}

export function useMission(id: string) {
  return useQuery({
    queryKey: missionKey(id),
    enabled: !!id,
    queryFn: () => api.quest.getQuestMission({ params: { path: { id } } }),
    refetchInterval: (q) => {
      const data = q.state.data
      if (!data?.mission) return 1500
      if (data.activeReply?.status === 'running') return 700
      return missionBusy(data.mission, data.activeReply) ? 1500 : false
    },
  })
}

export function useCreateMission() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { prompt: string; genre: string; language: string; practiceGoals: string }) =>
      api.quest.createQuestMission({ body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: listKey }),
  })
}

export function useRespond(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { text: string; strict: boolean }) =>
      api.quest.respondQuestMission({
        params: { path: { id } },
        body: { text: vars.text, strict: vars.strict },
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: missionKey(id) }),
  })
}

export function useSuggestNativeReply(id: string) {
  return useMutation({
    mutationFn: async (text: string): Promise<string[]> => {
      const res = await api.quest.suggestQuestNativeReply({
        params: { path: { id } },
        body: { text },
      })
      return res?.variants ?? []
    },
  })
}

export function useResetMission(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.quest.resetQuestMission({ params: { path: { id } } }),
    onSuccess: (res) => {
      if (res?.mission) {
        qc.setQueryData(missionKey(id), { mission: res.mission as Mission, activeReply: undefined })
      }
      qc.invalidateQueries({ queryKey: listKey })
    },
  })
}

export function useRegenerateImage(id: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { kind: 'cover' | 'scene' | 'avatar'; key?: string }) =>
      api.quest.regenerateQuestMissionImages({
        params: { path: { id } },
        body: { kind: vars.kind, key: vars.key },
      }),
    onSuccess: (res) => {
      const mission = res?.mission as Mission | undefined
      if (mission) {
        qc.setQueryData(missionKey(id), (prev: { mission: Mission; activeReply?: ActiveReply } | undefined) => ({
          mission,
          activeReply: prev?.activeReply,
        }))
      }
      qc.invalidateQueries({ queryKey: missionKey(id) })
    },
  })
}

export function useDeleteMission() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.quest.deleteQuestMission({ params: { path: { id } } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: listKey }),
  })
}
