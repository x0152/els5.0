import { useRef, type ReactNode } from 'react'
import { Camera, Loader2 } from 'lucide-react'
import { cn } from './cn.ts'

export function nameInitials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w.charAt(0).toUpperCase())
    .join('')
}

export interface AvatarProps {
  src?: string | null
  name?: string
  initials?: string
  icon?: ReactNode
  className?: string
  onUpload?: (file: File) => void
  uploading?: boolean
}

export function Avatar({ src, name = '', initials, icon, className, onUpload, uploading }: AvatarProps) {
  const fileRef = useRef<HTMLInputElement | null>(null)
  const fallback = initials ?? nameInitials(name)

  const core = src ? (
    <img src={src} alt={name} className={cn('h-10 w-10 rounded-full object-cover ring-1 ring-brand-100', className)} />
  ) : (
    <div
      title={name || undefined}
      className={cn(
        'flex h-10 w-10 select-none items-center justify-center rounded-full bg-brand-50 font-semibold text-brand-700 ring-1 ring-brand-100',
        className,
      )}
    >
      {fallback || icon}
    </div>
  )

  if (!onUpload) return core

  return (
    <div className="relative inline-block shrink-0">
      {core}
      <button
        type="button"
        onClick={() => fileRef.current?.click()}
        disabled={uploading}
        title="Upload photo"
        aria-label="Upload photo"
        className="absolute -bottom-1 -right-1 flex h-7 w-7 items-center justify-center rounded-full bg-neutral-900 text-white shadow transition-colors hover:bg-neutral-800 disabled:opacity-60"
      >
        {uploading ? <Loader2 size={13} className="animate-spin" /> : <Camera size={13} />}
      </button>
      <input
        ref={fileRef}
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0]
          e.target.value = ''
          if (file) onUpload(file)
        }}
      />
    </div>
  )
}
