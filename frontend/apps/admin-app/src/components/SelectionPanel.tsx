import { useRef, useState } from 'react'
import {
  Ban,
  Building2,
  Camera,
  CheckCircle2,
  Hash,
  KeyRound,
  Loader2,
  LogIn,
  Mail,
  Shield,
  Trash2,
  UserRound,
  X,
} from 'lucide-react'
import { Badge, cn, ConfirmDialog } from '@els/ui'
import { isApiError } from '@els/api-client'
import type { GridColumn } from '../lib/grid-client.ts'
import { applyGrid } from '../lib/grid-client.ts'
import { api } from '../lib/api.ts'
import { readToken, startImpersonation } from '../lib/impersonate.ts'
import { useMe, useUploadAccountPicture } from '../store/grid.ts'
import { useToast } from './Toasts.tsx'

export type SelectionPanelRow = {
  id: string
  base_version?: number
} & Record<string, unknown>

export interface ColumnMaps {
  byLabel: Map<string, string>
  byKey: Map<string, string>
}

interface SelectionPanelProps {
  basePath: string
  schemaVersion: string
  columns: GridColumn[]
  maps: Record<string, ColumnMaps>
  row: SelectionPanelRow | null
  onClose: () => void
  onApplied: () => void
}

type StatusKey = 'active' | 'blocked' | 'pending_password' | 'no_auth' | null

const ACCOUNT_STATUS_COL = 'status'
const ACCOUNT_EMAIL_COL = 'email'
const ACCOUNT_FIRST_NAME_COL = 'first_name'
const ACCOUNT_LAST_NAME_COL = 'last_name'
const ACCOUNT_PICTURE_URL_COL = 'picture_url'
const ACCOUNT_ID_COL = 'account_id'
const CLIENT_NAME_COL = 'name'
const CLIENT_REF_COL = 'client'

export function SelectionPanel({
  basePath,
  schemaVersion,
  columns,
  maps,
  row,
  onClose,
  onApplied,
}: SelectionPanelProps) {
  const toast = useToast()
  const [busy, setBusy] = useState<
    null | 'block' | 'unblock' | 'reset' | 'resend' | 'impersonate' | 'delete'
  >(null)
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const uploadPicture = useUploadAccountPicture(basePath)
  // Server-driven flag: only accounts with `impersonation_enabled = true`
  // get a session token from POST /auth/impersonate. Without this gate
  // any global admin in dev sees the button but gets 403 in prod.
  const meQ = useMe()
  const impersonationEnabled = meQ.data?.impersonation_enabled === true

  if (!row) return null

  const isAccountSection = columns.some((c) => c.id === ACCOUNT_STATUS_COL)
  const summary = summarizeRow(row, columns, maps)
  const statusKey = resolveStatusKey(columns, maps, row)
  const pictureUrl = asString(row[ACCOUNT_PICTURE_URL_COL]) || null
  const accountId = asString(row[ACCOUNT_ID_COL]) || null

  async function onPickPicture(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    if (!accountId) {
      toast.error('Failed to determine the account')
      return
    }
    if (!/^image\//.test(file.type)) {
      toast.error('The file must be an image')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      toast.error('File size — up to 5 MB')
      return
    }
    try {
      await uploadPicture.mutateAsync({ accountId, file })
      toast.success('Avatar updated')
      onApplied()
    } catch (err) {
      toast.error(errorText(err, 'Failed to upload avatar'))
    }
  }

  async function runStatusChange(next: 'active' | 'blocked') {
    if (!row) return
    if (statusKey === next) return
    const action = next === 'blocked' ? 'block' : 'unblock'
    setBusy(action)
    try {
      const res = await applyGrid(basePath, {
        schema_version: schemaVersion,
        operations: [
          {
            kind: 'update',
            id: row.id,
            base_version: row.base_version,
            data: { [ACCOUNT_STATUS_COL]: next },
          },
        ],
      })
      if (res.failed.length > 0) {
        toast.error(res.failed[0]?.message ?? 'Failed to change status')
        return
      }
      toast.success(next === 'blocked' ? 'Account blocked' : 'Account unblocked')
      onApplied()
    } catch (e) {
      toast.error(errorText(e, 'Failed to change status'))
    } finally {
      setBusy(null)
    }
  }

  async function runResetPassword() {
    if (!summary.email) {
      toast.error('The account has no email')
      return
    }
    setBusy('reset')
    try {
      await api.auth.forgotPassword({ body: { email: summary.email } })
      toast.success(`Password reset link sent to ${summary.email}`)
    } catch (e) {
      toast.error(errorText(e, 'Failed to send the email'))
    } finally {
      setBusy(null)
    }
  }

  async function runResendInvite() {
    if (!summary.email) {
      toast.error('The account has no email')
      return
    }
    setBusy('resend')
    try {
      await api.auth.resendInvite({ body: { email: summary.email } })
      toast.success(`Invitation sent to ${summary.email}`)
    } catch (e) {
      toast.error(errorText(e, 'Failed to send the invitation'))
    } finally {
      setBusy(null)
    }
  }

  const deleteTarget = isAccountSection
    ? summary.title || summary.email || 'this account'
    : summary.title || 'this customer'

  async function runDelete() {
    if (!row) return
    setConfirmingDelete(false)
    setBusy('delete')
    try {
      const res = await applyGrid(basePath, {
        schema_version: schemaVersion,
        operations: [
          {
            kind: 'delete',
            id: row.id,
            base_version: row.base_version,
          },
        ],
      })
      if (res.failed.length > 0) {
        toast.error(res.failed[0]?.message ?? 'Failed to delete')
        return
      }
      toast.success('Deleted')
      onApplied()
      onClose()
    } catch (e) {
      toast.error(errorText(e, 'Failed to delete'))
    } finally {
      setBusy(null)
    }
  }

  async function runImpersonate() {
    if (!accountId) {
      toast.error('Failed to determine the account')
      return
    }
    const originalToken = readToken()
    if (!originalToken) {
      toast.error('No active session')
      return
    }
    setBusy('impersonate')
    try {
      const me = await api.account.accountMe()
      const originalLabel =
        ([me?.first_name, me?.last_name].filter(Boolean).join(' ').trim() ||
          me?.email ||
          'global admin') as string
      const res = await api.auth.impersonate({ body: { account_id: accountId } })
      if (!res?.token) {
        toast.error('The server did not return a token')
        return
      }
      startImpersonation({ originalToken, originalLabel, newToken: res.token })
      window.location.assign('/')
    } catch (e) {
      toast.error(errorText(e, 'Failed to switch to the user'))
      setBusy(null)
    }
  }

  const canBlock = isAccountSection && (statusKey === 'active' || statusKey === 'pending_password')
  const canUnblock = isAccountSection && statusKey === 'blocked'
  const canResetPassword = isAccountSection && statusKey === 'active'
  const canResendInvite = isAccountSection && statusKey === 'pending_password'
  const canImpersonate = isAccountSection && impersonationEnabled && !!accountId && statusKey === 'active'
  const canDelete = !!row.base_version

  return (
    <aside className="h-full w-[360px] shrink-0 bg-white border-l border-neutral-200 flex flex-col min-h-0">
      <header className="shrink-0 flex items-center justify-between px-5 pt-4 pb-3 border-b border-neutral-100">
        <div className="text-[11px] font-semibold tracking-[0.1em] uppercase text-neutral-500">
          {isAccountSection ? 'Member' : 'Customer'}
        </div>
        <button
          type="button"
          onClick={onClose}
          className="p-1.5 rounded-md text-neutral-400 hover:text-neutral-900 hover:bg-neutral-100"
          aria-label="Close"
          title="Clear selection"
        >
          <X size={16} />
        </button>
      </header>

      <div className="flex-1 overflow-y-auto">
        <div className="px-5 pt-5 pb-4">
          <div className="flex items-start gap-4">
            <Avatar
              name={summary.title}
              kind={isAccountSection ? 'user' : 'client'}
              pictureUrl={pictureUrl}
              editable={isAccountSection && !!accountId}
              uploading={uploadPicture.isPending}
              onEdit={() => fileInputRef.current?.click()}
            />
            {isAccountSection ? (
              <input
                ref={fileInputRef}
                type="file"
                accept="image/png,image/jpeg,image/webp,image/gif"
                className="hidden"
                onChange={onPickPicture}
              />
            ) : null}
            <div className="min-w-0 pt-0.5">
              <h2 className="text-[15px] font-semibold text-neutral-900 truncate">
                {summary.title || <span className="text-neutral-400 italic">no name</span>}
              </h2>
              {summary.email ? (
                <a
                  href={`mailto:${summary.email}`}
                  className="inline-flex items-center gap-1 text-[13px] text-neutral-600 hover:text-brand-700 mt-1 truncate max-w-full"
                  title={summary.email}
                >
                  <Mail size={12} className="text-neutral-400 shrink-0" />
                  <span className="truncate">{summary.email}</span>
                </a>
              ) : null}
              {isAccountSection ? (
                <div className="mt-2">
                  <StatusPill statusKey={statusKey} />
                </div>
              ) : null}
            </div>
          </div>

          {summary.clientLabel ? (
            <div className="mt-4 p-3 rounded-lg bg-neutral-50 ring-1 ring-neutral-200/80 flex items-center gap-2">
              <Building2 size={14} className="text-neutral-400 shrink-0" />
              <div className="min-w-0 flex-1">
                <div className="text-[10px] font-semibold tracking-wider uppercase text-neutral-400">
                  Customer
                </div>
                <div className="text-sm text-neutral-900 truncate">{summary.clientLabel}</div>
              </div>
            </div>
          ) : null}

          <dl className="mt-4 text-[13px] space-y-2.5">
            <MetaRow icon={<Hash size={12} />} label="ID">
              <code className="text-[11px] text-neutral-700 font-mono truncate" title={row.id}>
                {row.id}
              </code>
            </MetaRow>
          </dl>
        </div>

        {isAccountSection ? (
          <div className="px-5 pb-5">
            <div className="text-[10px] font-semibold tracking-[0.1em] uppercase text-neutral-400 mb-2">
              Actions
            </div>
            <div className="flex flex-col gap-1.5">
              {canBlock ? (
                <ActionButton
                  icon={<Ban size={14} />}
                  label="Block"
                  description="Deny access to the system"
                  tone="danger"
                  loading={busy === 'block'}
                  disabled={busy !== null}
                  onClick={() => runStatusChange('blocked')}
                />
              ) : null}
              {canUnblock ? (
                <ActionButton
                  icon={<CheckCircle2 size={14} />}
                  label="Unblock"
                  description="Remove the account block"
                  tone="success"
                  loading={busy === 'unblock'}
                  disabled={busy !== null}
                  onClick={() => runStatusChange('active')}
                />
              ) : null}
              {canResetPassword ? (
                <ActionButton
                  icon={<KeyRound size={14} />}
                  label="Reset password"
                  description="Send an email with a reset link"
                  tone="neutral"
                  loading={busy === 'reset'}
                  disabled={busy !== null}
                  onClick={runResetPassword}
                />
              ) : null}
              {canResendInvite ? (
                <ActionButton
                  icon={<Mail size={14} />}
                  label="Resend invitation"
                  description="Resend the password setup email"
                  tone="brand"
                  loading={busy === 'resend'}
                  disabled={busy !== null}
                  onClick={runResendInvite}
                />
              ) : null}
              {canImpersonate ? (
                <ActionButton
                  icon={<LogIn size={14} />}
                  label="Sign in as this user"
                  description="Open the system on behalf of the account"
                  tone="neutral"
                  loading={busy === 'impersonate'}
                  disabled={busy !== null}
                  onClick={runImpersonate}
                />
              ) : null}
              {canDelete ? (
                <ActionButton
                  icon={<Trash2 size={14} />}
                  label="Delete account"
                  description="Permanently delete the record"
                  tone="danger"
                  loading={busy === 'delete'}
                  disabled={busy !== null}
                  onClick={() => setConfirmingDelete(true)}
                />
              ) : null}

              {!canBlock && !canUnblock && !canResetPassword && !canResendInvite && !canImpersonate && !canDelete ? (
                <div className="text-[12px] text-neutral-500 italic px-3 py-2 rounded-md bg-neutral-50 ring-1 ring-neutral-200">
                  No actions available for this status.
                </div>
              ) : null}
            </div>
          </div>
        ) : (
          <div className="px-5 pb-5">
            <div className="text-[10px] font-semibold tracking-[0.1em] uppercase text-neutral-400 mb-2">
              Actions
            </div>
            <div className="flex flex-col gap-1.5">
              {canDelete ? (
                <ActionButton
                  icon={<Trash2 size={14} />}
                  label="Delete customer"
                  description="Permanently delete the record"
                  tone="danger"
                  loading={busy === 'delete'}
                  disabled={busy !== null}
                  onClick={() => setConfirmingDelete(true)}
                />
              ) : (
                <div className="text-[12px] text-neutral-500 italic px-3 py-2 rounded-md bg-neutral-50 ring-1 ring-neutral-200">
                  Customer management is done through the table fields.
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {confirmingDelete && (
        <ConfirmDialog
          title={isAccountSection ? 'Delete account' : 'Delete customer'}
          description={`Delete “${deleteTarget}”? This action is irreversible.`}
          onConfirm={() => void runDelete()}
          onClose={() => setConfirmingDelete(false)}
        />
      )}
    </aside>
  )
}

/* ------------------------------- primitives ------------------------------- */

interface AvatarProps {
  name: string
  kind: 'user' | 'client'
  pictureUrl?: string | null
  editable?: boolean
  uploading?: boolean
  onEdit?: () => void
}

function Avatar({ name, kind, pictureUrl, editable, uploading, onEdit }: AvatarProps) {
  const initialsText = initials(name)
  const isUser = kind === 'user'
  const title = name || (isUser ? 'User' : 'Customer')
  return (
    <div className="relative shrink-0">
      {pictureUrl ? (
        <img
          src={pictureUrl}
          alt={title}
          className="w-14 h-14 rounded-2xl object-cover ring-1 ring-neutral-200 select-none"
        />
      ) : (
        <div
          className={cn(
            'w-14 h-14 rounded-2xl flex items-center justify-center ring-1 font-semibold text-[15px] select-none',
            isUser
              ? 'bg-gradient-to-br from-brand-50 to-brand-100 text-brand-700 ring-brand-200'
              : 'bg-gradient-to-br from-indigo-50 to-indigo-100 text-indigo-700 ring-indigo-200',
          )}
          title={title}
        >
          {initialsText || (isUser ? <UserRound size={20} /> : <Building2 size={20} />)}
        </div>
      )}

      {!pictureUrl ? (
        <span
          className={cn(
            'absolute -bottom-1 -right-1 w-5 h-5 rounded-full bg-white flex items-center justify-center ring-1 pointer-events-none',
            isUser ? 'ring-brand-200 text-brand-600' : 'ring-indigo-200 text-indigo-600',
          )}
        >
          {isUser ? <Shield size={11} /> : <Building2 size={11} />}
        </span>
      ) : null}

      {editable ? (
        <button
          type="button"
          onClick={onEdit}
          disabled={uploading}
          className={cn(
            'absolute inset-0 rounded-2xl flex items-center justify-center',
            'bg-black/0 hover:bg-black/40 text-white opacity-0 hover:opacity-100 transition',
            'disabled:cursor-not-allowed',
            uploading && 'bg-black/40 opacity-100',
          )}
          title="Upload avatar"
          aria-label="Upload avatar"
        >
          {uploading ? (
            <Loader2 size={18} className="animate-spin" />
          ) : (
            <Camera size={18} />
          )}
        </button>
      ) : null}
    </div>
  )
}

function StatusPill({ statusKey }: { statusKey: StatusKey }) {
  if (!statusKey) return null
  const theme: Record<NonNullable<StatusKey>, { tone: 'success' | 'danger' | 'warning' | 'neutral'; dot: string; label: string }> = {
    active: { tone: 'success', dot: 'bg-emerald-500', label: 'Active' },
    blocked: { tone: 'danger', dot: 'bg-red-500', label: 'Blocked' },
    pending_password: { tone: 'warning', dot: 'bg-amber-400', label: 'Pending password' },
    no_auth: { tone: 'neutral', dot: 'bg-neutral-400', label: 'No authentication' },
  }
  const t = theme[statusKey]
  return (
    <Badge tone={t.tone} className="gap-1.5 text-[11px] font-semibold">
      <span className={cn('w-1.5 h-1.5 rounded-full', t.dot)} />
      {t.label}
    </Badge>
  )
}

function MetaRow({
  icon,
  label,
  children,
}: {
  icon: React.ReactNode
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="flex items-start gap-2">
      <span className="mt-0.5 text-neutral-400 shrink-0">{icon}</span>
      <div className="min-w-0 flex-1">
        <dt className="text-[10px] font-medium tracking-wider uppercase text-neutral-400">
          {label}
        </dt>
        <dd className="text-sm text-neutral-800 truncate">{children}</dd>
      </div>
    </div>
  )
}

type ActionTone = 'neutral' | 'brand' | 'danger' | 'success'

interface ActionButtonProps {
  icon: React.ReactNode
  label: string
  description?: string
  tone: ActionTone
  loading?: boolean
  disabled?: boolean
  onClick: () => void
}

function ActionButton({ icon, label, description, tone, loading, disabled, onClick }: ActionButtonProps) {
  const tones: Record<ActionTone, string> = {
    neutral: 'hover:bg-neutral-50 text-neutral-800 ring-neutral-200 [&>span.icon]:text-neutral-500',
    brand: 'hover:bg-brand-50 text-brand-800 ring-brand-200 [&>span.icon]:text-brand-600',
    danger: 'hover:bg-red-50 text-red-800 ring-red-200 [&>span.icon]:text-red-600',
    success: 'hover:bg-emerald-50 text-emerald-800 ring-emerald-200 [&>span.icon]:text-emerald-600',
  }
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className={cn(
        'group w-full flex items-start gap-3 px-3 py-2.5 rounded-lg ring-1 bg-white text-left transition-all',
        'disabled:opacity-60 disabled:cursor-not-allowed disabled:hover:bg-white',
        tones[tone],
      )}
    >
      <span className="icon mt-0.5 shrink-0 w-7 h-7 flex items-center justify-center rounded-md bg-neutral-50 ring-1 ring-inherit">
        {loading ? <Loader2 size={14} className="animate-spin" /> : icon}
      </span>
      <span className="min-w-0 flex-1">
        <span className="block text-[13px] font-medium leading-tight">{label}</span>
        {description ? (
          <span className="block text-[11px] text-neutral-500 mt-0.5 leading-snug">{description}</span>
        ) : null}
      </span>
    </button>
  )
}

/* -------------------------------- helpers --------------------------------- */

interface RowSummary {
  title: string
  email: string | null
  clientLabel: string | null
}

function summarizeRow(
  row: SelectionPanelRow,
  columns: GridColumn[],
  maps: Record<string, ColumnMaps>,
): RowSummary {
  const hasAccount = columns.some((c) => c.id === ACCOUNT_STATUS_COL)
  const first = asString(row[ACCOUNT_FIRST_NAME_COL])
  const last = asString(row[ACCOUNT_LAST_NAME_COL])
  const title = hasAccount
    ? [first, last].filter(Boolean).join(' ').trim() ||
      asString(row[ACCOUNT_EMAIL_COL]) ||
      'No name'
    : asString(row[CLIENT_NAME_COL]) || 'No title'
  const email = asString(row[ACCOUNT_EMAIL_COL]) || null
  const clientLabel = resolveClientLabel(row, maps)
  return { title, email, clientLabel }
}

function resolveClientLabel(
  row: SelectionPanelRow,
  maps: Record<string, ColumnMaps>,
): string | null {
  const raw = row[CLIENT_REF_COL]
  if (raw === null || raw === undefined || raw === '') return null
  const display = String(raw)
  const map = maps[CLIENT_REF_COL]
  if (map?.byKey.has(display)) return map.byKey.get(display) ?? null
  return display
}

function resolveStatusKey(
  columns: GridColumn[],
  maps: Record<string, ColumnMaps>,
  row: SelectionPanelRow,
): StatusKey {
  const col = columns.find((c) => c.id === ACCOUNT_STATUS_COL)
  if (!col) return null
  const raw = row[ACCOUNT_STATUS_COL]
  if (!raw) return null
  const display = String(raw)
  const keyFromLabel = maps[ACCOUNT_STATUS_COL]?.byLabel.get(display)
  const key = keyFromLabel ?? display
  if (key === 'active' || key === 'blocked' || key === 'pending_password' || key === 'no_auth') {
    return key
  }
  return null
}

function asString(v: unknown): string {
  if (v === null || v === undefined) return ''
  return String(v)
}

function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? '')
    .join('')
}

function errorText(e: unknown, fallback: string): string {
  if (isApiError(e)) {
    const detail = e.details?.[0]?.message
    return detail || e.message || fallback
  }
  if (e instanceof Error) return e.message || fallback
  return fallback
}
