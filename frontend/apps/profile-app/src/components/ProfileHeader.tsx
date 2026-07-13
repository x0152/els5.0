import { useState } from 'react'
import { Mail, ShieldCheck, UserCircle2 } from 'lucide-react'
import { Avatar, Badge } from '@els/ui'
import { useMe, useUploadMyPicture } from '../store/me'

const ROLE_LABEL: Record<string, string> = {
  admin: 'Administrator',
  expert: 'Specialist',
  customer: 'Customer',
}

const STATUS_LABEL: Record<string, { label: string; tone: 'success' | 'warning' | 'danger' | 'neutral' }> = {
  active: { label: 'Active', tone: 'success' },
  invited: { label: 'Invited', tone: 'warning' },
  blocked: { label: 'Blocked', tone: 'danger' },
}

export function ProfileHeader() {
  const meQ = useMe()
  const upload = useUploadMyPicture()
  const [err, setErr] = useState<string | null>(null)

  const me = meQ.data

  const onPick = async (file: File) => {
    setErr(null)
    if (!/^image\//.test(file.type)) {
      setErr('An image must be uploaded')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      setErr('File is larger than 5 MB')
      return
    }
    try {
      await upload.mutateAsync(file)
    } catch (x) {
      setErr(x instanceof Error ? x.message : 'Failed to upload')
    }
  }

  if (meQ.isLoading || !me) {
    return (
      <div className="rounded-2xl bg-white ring-1 ring-neutral-200 p-6 animate-pulse">
        <div className="flex items-center gap-5">
          <div className="h-20 w-20 rounded-full bg-neutral-100" />
          <div className="flex-1 space-y-2">
            <div className="h-5 w-52 rounded bg-neutral-100" />
            <div className="h-4 w-40 rounded bg-neutral-100" />
          </div>
        </div>
      </div>
    )
  }

  const roleLabel = ROLE_LABEL[me.role] ?? me.role
  const st = STATUS_LABEL[me.status] ?? { label: me.status, tone: 'neutral' as const }

  return (
    <div className="rounded-2xl bg-white ring-1 ring-neutral-200 p-6">
      <div className="flex items-start gap-5">
        <Avatar
          src={me.pictureUrl}
          name={me.displayName}
          initials={me.initials}
          className="h-20 w-20 text-xl font-bold ring-2"
          onUpload={(file) => void onPick(file)}
          uploading={upload.isPending}
        />

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-2xl font-semibold text-neutral-900 tracking-tight">
              {me.displayName}
            </h1>
            <Badge tone={st.tone} className="text-[11px] font-semibold uppercase tracking-wider">
              {st.label}
            </Badge>
            {me.isGlobalAdmin && (
              <Badge tone="brand" className="text-[11px] font-semibold uppercase tracking-wider">
                global admin
              </Badge>
            )}
          </div>

          <div className="mt-2 flex items-center gap-5 text-sm text-neutral-600 flex-wrap">
            <span className="inline-flex items-center gap-1.5">
              <UserCircle2 size={14} className="text-neutral-400" />
              {roleLabel}
            </span>
            <span className="inline-flex items-center gap-1.5">
              <Mail size={14} className="text-neutral-400" />
              {me.email}
            </span>
            {me.isGlobalAdmin && (
              <span className="inline-flex items-center gap-1.5 text-brand-700">
                <ShieldCheck size={14} />
                full access
              </span>
            )}
          </div>

          {err && (
            <div className="mt-2 text-xs text-rose-700">{err}</div>
          )}
        </div>
      </div>
    </div>
  )
}
