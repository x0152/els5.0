import { useEffect, useState } from 'react'
import { Check, Loader2, Pencil, X } from 'lucide-react'
import { cn, Input, Textarea } from '@els/ui'
import { Widget } from './Widget'
import { useMe, useUpdateProfile, type MeProfile } from '../store/me'

export function ProfileDetails() {
  const meQ = useMe()
  if (!meQ.data) return null
  return <DetailsForm me={meQ.data} />
}

function DetailsForm({ me }: { me: MeProfile }) {
  const update = useUpdateProfile()
  const [editing, setEditing] = useState(false)
  const [firstName, setFirstName] = useState(me.firstName)
  const [lastName, setLastName] = useState(me.lastName)
  const [englishLevel, setEnglishLevel] = useState(me.englishLevel)
  const [aboutMe, setAboutMe] = useState(me.aboutMe)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    if (!editing) {
      setFirstName(me.firstName)
      setLastName(me.lastName)
      setEnglishLevel(me.englishLevel)
      setAboutMe(me.aboutMe)
    }
  }, [me, editing])

  const onSave = async () => {
    setErr(null)
    if (!firstName.trim() || !lastName.trim()) {
      setErr('First and last name are required')
      return
    }
    try {
      await update.mutateAsync({ firstName, lastName, englishLevel, aboutMe })
      setEditing(false)
    } catch (x) {
      setErr(x instanceof Error ? x.message : 'Failed to save')
    }
  }

  return (
    <Widget
      title="Personal details"
      className="sm:col-span-3"
      action={
        editing ? (
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => setEditing(false)}
              disabled={update.isPending}
              className="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-xs font-medium text-neutral-600 hover:bg-neutral-100 disabled:opacity-60"
            >
              <X size={14} /> Cancel
            </button>
            <button
              type="button"
              onClick={onSave}
              disabled={update.isPending}
              className="inline-flex items-center gap-1 rounded-lg bg-brand-600 px-2.5 py-1.5 text-xs font-semibold text-white hover:bg-brand-700 disabled:opacity-60"
            >
              {update.isPending ? <Loader2 size={14} className="animate-spin" /> : <Check size={14} />}
              Save
            </button>
          </div>
        ) : (
          <button
            type="button"
            onClick={() => setEditing(true)}
            className="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-xs font-medium text-brand-700 hover:bg-brand-50"
          >
            <Pencil size={14} /> Edit
          </button>
        )
      }
    >
      <div className="grid grid-cols-1 gap-4 p-5 sm:grid-cols-2">
        <Field label="First name">
          <Input value={firstName} onChange={(e) => setFirstName(e.target.value)} disabled={!editing} />
        </Field>
        <Field label="Last name">
          <Input value={lastName} onChange={(e) => setLastName(e.target.value)} disabled={!editing} />
        </Field>
        <Field label="English level" className="sm:col-span-2">
          <Input
            value={englishLevel}
            onChange={(e) => setEnglishLevel(e.target.value)}
            disabled={!editing}
            placeholder="e.g. B2 (upper-intermediate)"
          />
        </Field>
        <Field label="About me" className="sm:col-span-2">
          <Textarea
            value={aboutMe}
            onChange={(e) => setAboutMe(e.target.value)}
            disabled={!editing}
            rows={4}
            placeholder="A few words about yourself"
          />
        </Field>
        {err && <div className="text-xs text-rose-700 sm:col-span-2">{err}</div>}
      </div>
    </Widget>
  )
}

function Field({
  label,
  className,
  children,
}: {
  label: string
  className?: string
  children: React.ReactNode
}) {
  return (
    <label className={cn('block', className)}>
      <span className="mb-1 block text-[11px] font-semibold uppercase tracking-wider text-neutral-500">
        {label}
      </span>
      {children}
    </label>
  )
}