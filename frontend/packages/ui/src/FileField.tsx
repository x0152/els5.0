import { useRef, useState, type ReactNode } from 'react'
import { UploadCloud, X } from 'lucide-react'
import { cn } from './cn.ts'

export interface FileFieldProps {
  value: File | null
  onChange: (file: File | null) => void
  accept?: string
  placeholder?: string
  icon?: ReactNode
  className?: string
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

export function FileField({ value, onChange, accept, placeholder = 'Click or drop a file', icon, className }: FileFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [drag, setDrag] = useState(false)

  return (
    <div
      onClick={() => inputRef.current?.click()}
      onDragOver={(e) => {
        e.preventDefault()
        setDrag(true)
      }}
      onDragLeave={() => setDrag(false)}
      onDrop={(e) => {
        e.preventDefault()
        setDrag(false)
        const f = e.dataTransfer.files?.[0]
        if (f) onChange(f)
      }}
      className={cn(
        'flex cursor-pointer items-center gap-3 rounded-xl border-2 border-dashed px-4 py-3 text-sm transition-colors',
        drag
          ? 'border-brand-400 bg-brand-50'
          : value
            ? 'border-brand-200 bg-brand-50/40'
            : 'border-neutral-300 hover:border-brand-300 hover:bg-neutral-50',
        className,
      )}
    >
      <span className="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-white text-brand-600 ring-1 ring-neutral-200">
        {icon ?? <UploadCloud className="h-4 w-4" />}
      </span>
      {value ? (
        <div className="min-w-0 flex-1">
          <p className="truncate font-medium text-neutral-800">{value.name}</p>
          <p className="text-xs text-neutral-400">{formatSize(value.size)}</p>
        </div>
      ) : (
        <span className="flex-1 text-neutral-500">{placeholder}</span>
      )}
      {value && (
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation()
            onChange(null)
            if (inputRef.current) inputRef.current.value = ''
          }}
          className="grid h-7 w-7 shrink-0 place-items-center rounded-full text-neutral-400 hover:bg-neutral-200 hover:text-neutral-700"
        >
          <X className="h-4 w-4" />
        </button>
      )}
      <input ref={inputRef} type="file" accept={accept} className="hidden" onChange={(e) => onChange(e.target.files?.[0] ?? null)} />
    </div>
  )
}
