import { useEffect, useRef, useState } from 'react'
import { ImagePlus, X } from 'lucide-react'
import { cn } from './cn.ts'

export interface ImageFieldProps {
  value: File | null
  onChange: (file: File | null) => void
  initialUrl?: string
  aspect?: string
  placeholder?: string
  className?: string
}

export function ImageField({
  value,
  onChange,
  initialUrl,
  aspect = 'aspect-[2/3]',
  placeholder = 'Click or drop an image',
  className,
}: ImageFieldProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [drag, setDrag] = useState(false)

  useEffect(() => {
    if (!value) {
      setPreview(null)
      return
    }
    const url = URL.createObjectURL(value)
    setPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [value])

  const shown = preview ?? initialUrl

  const clear = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange(null)
    if (inputRef.current) inputRef.current.value = ''
  }

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
        if (f?.type.startsWith('image/')) onChange(f)
      }}
      className={cn(
        'group relative flex cursor-pointer items-center justify-center overflow-hidden rounded-xl border-2 border-dashed transition-colors',
        drag ? 'border-brand-400 bg-brand-50' : 'border-neutral-300 hover:border-brand-300 hover:bg-neutral-50',
        aspect,
        className,
      )}
    >
      {shown ? (
        <img src={shown} alt="" className="h-full w-full object-cover" />
      ) : (
        <div className="flex flex-col items-center gap-1.5 px-3 text-center text-neutral-400">
          <ImagePlus className="h-6 w-6" />
          <span className="text-xs font-medium">{placeholder}</span>
        </div>
      )}
      {value && (
        <button
          type="button"
          onClick={clear}
          className="absolute right-2 top-2 grid h-7 w-7 place-items-center rounded-full bg-black/55 text-white opacity-0 transition-opacity hover:bg-black/70 group-hover:opacity-100"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      )}
      <input ref={inputRef} type="file" accept="image/*" className="hidden" onChange={(e) => onChange(e.target.files?.[0] ?? null)} />
    </div>
  )
}
