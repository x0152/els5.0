import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'

export interface MeProfile {
  accountId: string
  entityId: string
  email: string
  firstName: string
  lastName: string
  englishLevel: string
  aboutMe: string
  nativeLanguage: string
  showTranslations: boolean
  autoWordImages: boolean
  role: 'admin' | 'expert' | 'customer' | string
  status: string
  isGlobalAdmin: boolean
  /** Backend allows this account to issue impersonation tokens. */
  impersonationEnabled: boolean
  pictureUrl?: string
  displayName: string
  initials: string
}

export const meQueryKey = ['profile-app', 'me'] as const

export function useMe() {
  return useQuery({
    queryKey: meQueryKey,
    queryFn: async (): Promise<MeProfile> => {
      const res = await api.account.accountMe()
      if (!res) throw new Error('account/me returned empty payload')
      const displayName =
        [res.first_name, res.last_name].filter(Boolean).join(' ').trim() || res.email
      const initials =
        `${res.first_name?.[0] ?? ''}${res.last_name?.[0] ?? ''}`.toUpperCase() ||
        res.email[0]?.toUpperCase() ||
        '?'
      return {
        accountId: res.account_id,
        entityId: res.entity_id,
        email: res.email,
        firstName: res.first_name,
        lastName: res.last_name,
        englishLevel: res.english_level ?? '',
        aboutMe: res.about_me ?? '',
        nativeLanguage: res.native_language ?? '',
        showTranslations: res.show_translations ?? true,
        autoWordImages: res.auto_word_images ?? false,
        role: res.role,
        status: res.status,
        isGlobalAdmin: res.is_global_admin,
        impersonationEnabled: res.impersonation_enabled,
        pictureUrl: res.picture_url || undefined,
        displayName,
        initials,
      }
    },
    staleTime: 30_000,
  })
}

export interface UpdateProfileInput {
  firstName: string
  lastName: string
  englishLevel: string
  aboutMe: string
  nativeLanguage: string
  showTranslations: boolean
  autoWordImages: boolean
}

export function useUpdateProfile() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: UpdateProfileInput) =>
      api.account.accountUpdateProfile({
        body: {
          first_name: input.firstName,
          last_name: input.lastName,
          english_level: input.englishLevel,
          about_me: input.aboutMe,
          native_language: input.nativeLanguage,
          show_translations: input.showTranslations,
          auto_word_images: input.autoWordImages,
        },
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: meQueryKey })
    },
  })
}

export function useUploadMyPicture() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (file: File) => {
      const form = new FormData()
      form.append('file', file)
      return api.account.accountMeUploadPicture({
        body: form as unknown as never,
      })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: meQueryKey })
    },
  })
}
