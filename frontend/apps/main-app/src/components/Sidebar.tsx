import { useEffect, useRef, useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { LogOut, MessageCircle } from 'lucide-react'
import { Avatar, cn } from '@els/ui'
import { groupNames, groupOrder, type Section } from '../config/sections'
import { groupSections, useApps } from '../hooks/useApps'
import { Logo } from './Logo'
import { useAuth } from '../auth/AuthContext'

/**
 * Primary app navigation.
 *
 * - `md+`: classic vertical rail on the left (28 wide).
 * - `<md`: slim top strip (logo + profile menu) and a horizontally scrollable
 *   bottom tab bar — thumb-friendly on phones.
 *
 * The three variants are rendered into a single component so consumers
 * (`AppShell`) only need one mount point; Tailwind responsive classes hide
 * the ones that don't apply to the current breakpoint.
 */
export function Sidebar() {
  return (
    <>
      <DesktopSidebar />
      <MobileTopBar />
      <MobileBottomBar />
    </>
  )
}

function DesktopSidebar() {
  const { data: sections, isLoading } = useApps()
  const grouped = groupSections(sections ?? [])

  return (
    <aside className="hidden md:block fixed left-0 top-0 bottom-0 w-28 z-40 bg-white border-r border-neutral-200">
      <div className="h-full flex flex-col">
        <div className="shrink-0 flex items-center justify-center p-3">
          <NavLink
            to="/v1/profile"
            title="My profile"
            className="hover:scale-105 transition-transform"
          >
            <Logo />
          </NavLink>
        </div>

        <div className="mx-3 border-t border-neutral-100" />

        <div className="flex-1 overflow-y-auto px-2 py-4 scrollbar-hide">
          {isLoading && (
            <div className="text-center text-[10px] text-neutral-400 py-4">loading…</div>
          )}

          {groupOrder.map((group) => {
            const items = grouped[group]
            if (!items || items.length === 0) return null
            return (
              <div key={group} className="mb-6">
                <div className="text-center mb-2">
                  <span className="text-[10px] font-bold tracking-wider text-neutral-400">
                    {groupNames[group]}
                  </span>
                </div>
                <div className="space-y-1">
                  {items.map((s) => (
                    <SectionButton key={s.id} section={s} />
                  ))}
                </div>
              </div>
            )
          })}
        </div>

        <div className="mx-4 border-t border-neutral-100" />

        <div className="shrink-0 p-3">
          <ProfileMenu />
        </div>
      </div>
    </aside>
  )
}

function MobileTopBar() {
  const onChatPage = useLocation().pathname.startsWith('/v1/chat')
  return (
    <header className="md:hidden fixed left-0 right-0 top-0 z-40 h-14 bg-white/95 backdrop-blur border-b border-neutral-200">
      <div
        className="h-full flex items-center justify-between px-3"
        style={{ paddingTop: 'env(safe-area-inset-top, 0px)' }}
      >
        <NavLink
          to="/v1/profile"
          title="My profile"
          className="flex items-center gap-2 active:scale-95 transition-transform"
        >
          <Logo />
        </NavLink>
        <div className="flex items-center gap-2">
          {!onChatPage && (
            <button
              type="button"
              aria-label="Open assistant"
              onClick={() => document.dispatchEvent(new Event('els:ask'))}
              className="grid h-9 w-9 place-items-center rounded-full bg-brand-600 text-white shadow-sm active:scale-95 transition-transform"
            >
              <MessageCircle size={18} />
            </button>
          )}
          <MobileProfileMenu />
        </div>
      </div>
    </header>
  )
}

function MobileBottomBar() {
  const { data: sections, isLoading } = useApps()
  const flat = flattenSections(sections ?? [])

  return (
    <nav
      className="md:hidden fixed left-0 right-0 bottom-0 z-40 bg-white/95 backdrop-blur border-t border-neutral-200"
      style={{ paddingBottom: 'env(safe-area-inset-bottom, 0px)' }}
    >
      <div className="overflow-x-auto scrollbar-hide">
        <div className="flex items-stretch h-16 px-1 min-w-full w-max">
          {isLoading ? (
            <div className="flex-1 text-center text-[10px] text-neutral-400 py-4 px-6">
              loading…
            </div>
          ) : (
            flat.map((s) => <MobileTab key={s.id} section={s} />)
          )}
        </div>
      </div>
    </nav>
  )
}

/**
 * Flattens grouped sections preserving group order — on mobile we don't
 * surface group labels, just a single row of tabs.
 */
function flattenSections(sections: Section[]): Section[] {
  const grouped = groupSections(sections)
  const out: Section[] = []
  for (const g of groupOrder) {
    for (const s of grouped[g] ?? []) out.push(s)
  }
  return out
}

function MobileTab({ section }: { section: Section }) {
  const Icon = section.icon

  if (section.disabled) {
    return (
      <div
        title={`${section.label} — coming soon`}
        className="shrink-0 min-w-[64px] flex flex-col items-center justify-center gap-0.5 px-2 text-neutral-400 select-none cursor-not-allowed"
      >
        <Icon size={20} />
        <span className="text-[10px] font-medium leading-none">
          {section.label}
        </span>
      </div>
    )
  }

  return (
    <NavLink
      to={section.to}
      title={section.label}
      className={({ isActive }) =>
        cn(
          'relative shrink-0 min-w-[64px] flex flex-col items-center justify-center gap-0.5 px-2 transition-colors',
          isActive
            ? 'text-brand-700'
            : 'text-neutral-600 active:text-neutral-900',
        )
      }
    >
      {({ isActive }) => (
        <>
          {isActive && (
            <span className="absolute top-0 left-1/2 -translate-x-1/2 h-0.5 w-8 rounded-b-full bg-brand-500" />
          )}
          <Icon
            size={20}
            className={cn(isActive ? 'text-brand-600' : 'text-neutral-500')}
          />
          <span className="text-[10px] font-medium leading-none">
            {section.label}
          </span>
        </>
      )}
    </NavLink>
  )
}

function ProfileMenu() {
  const { user, signOut } = useAuth()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => {
      if (!ref.current?.contains(e.target as Node)) setOpen(false)
    }
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false)
    window.addEventListener('mousedown', onDown)
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('mousedown', onDown)
      window.removeEventListener('keydown', onKey)
    }
  }, [open])

  const initials = user?.initials ?? '?'
  const displayName = user?.displayName ?? 'Profile'
  const pictureUrl = user?.pictureUrl

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        title={displayName}
        className="w-full flex flex-col items-center gap-1.5 group"
      >
        <Avatar
          src={pictureUrl}
          name={displayName}
          initials={initials}
          className="w-11 h-11 text-sm transition-all group-hover:ring-brand-300"
        />
        <span className="text-[11px] font-medium text-neutral-700 leading-tight truncate max-w-full px-1">
          {user ? user.firstName || user.email.split('@')[0] : 'Profile'}
        </span>
      </button>

      {open && user ? (
        <div className="absolute left-full ml-3 bottom-0 min-w-[240px] bg-white border border-neutral-200 rounded-lg shadow-xl overflow-hidden">
          <ProfileDetails />
          <ProfileSignOut onClick={() => { setOpen(false); void signOut() }} />
        </div>
      ) : null}
    </div>
  )
}

function MobileProfileMenu() {
  const { user, signOut } = useAuth()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDown = (e: MouseEvent) => {
      if (!ref.current?.contains(e.target as Node)) setOpen(false)
    }
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false)
    window.addEventListener('mousedown', onDown)
    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('mousedown', onDown)
      window.removeEventListener('keydown', onKey)
    }
  }, [open])

  const initials = user?.initials ?? '?'
  const displayName = user?.displayName ?? 'Profile'
  const pictureUrl = user?.pictureUrl

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        title={displayName}
        className="flex items-center active:scale-95 transition-transform"
      >
        <Avatar src={pictureUrl} name={displayName} initials={initials} className="w-9 h-9 text-xs" />
      </button>

      {open && user ? (
        <div className="absolute right-0 top-full mt-2 min-w-[240px] bg-white border border-neutral-200 rounded-lg shadow-xl overflow-hidden">
          <ProfileDetails />
          <ProfileSignOut onClick={() => { setOpen(false); void signOut() }} />
        </div>
      ) : null}
    </div>
  )
}

function ProfileDetails() {
  const { user } = useAuth()
  if (!user) return null
  return (
    <div className="px-4 py-3 border-b border-neutral-100">
      <div className="text-sm font-semibold text-neutral-900 truncate">
        {user.displayName}
      </div>
      <div className="text-xs text-neutral-500 truncate">{user.email}</div>
      {user.isGlobalAdmin ? (
        <div className="mt-2 inline-flex items-center text-[10px] font-semibold tracking-wider uppercase text-brand-700 bg-brand-50 px-2 py-0.5 rounded">
          global admin
        </div>
      ) : null}
    </div>
  )
}

function ProfileSignOut({ onClick }: { onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full flex items-center gap-2 px-4 py-2.5 text-sm text-neutral-700 hover:bg-neutral-50"
    >
      <LogOut size={16} className="text-neutral-500" />
      Sign out
    </button>
  )
}

function SectionButton({ section }: { section: Section }) {
  const Icon = section.icon

  if (section.disabled) {
    return (
      <div
        title={`${section.label} — coming soon`}
        className="relative w-full flex flex-col items-center gap-1 py-2 px-1 rounded-lg cursor-not-allowed select-none text-neutral-500"
      >
        <Icon size={20} className="text-neutral-400" />
        <span className="text-[11px] font-medium leading-tight text-center">
          {section.label}
        </span>
      </div>
    )
  }

  return (
    <NavLink
      to={section.to}
      title={section.label}
      className={({ isActive }) =>
        cn(
          'relative w-full flex flex-col items-center gap-1 py-2 px-1 rounded-lg transition-all',
          isActive
            ? 'bg-brand-50 text-brand-700'
            : 'text-neutral-700 hover:bg-neutral-100 hover:text-neutral-900',
        )
      }
    >
      {({ isActive }) => (
        <>
          {isActive && (
            <span className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-6 rounded-r-full bg-brand-500" />
          )}
          <Icon size={20} className={cn(isActive ? 'text-brand-600' : 'text-neutral-500')} />
          <span className="text-[11px] font-medium leading-tight text-center">
            {section.label}
          </span>
        </>
      )}
    </NavLink>
  )
}
