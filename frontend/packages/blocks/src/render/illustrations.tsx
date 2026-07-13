import { useContext, useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { ImageIcon, Loader2, Maximize2, Sparkles, TriangleAlert, X } from 'lucide-react'
import { useImageApi, type ImageAspect } from '../images.ts'
import type { Size } from '../parse.ts'
import { BlockCtx } from './context.ts'

const IMG_W = { sm: 'w-28', md: 'w-44', lg: 'w-64', full: 'w-full' } as const

function useIllustration(prompt: string, aspect: ImageAspect) {
  const api = useImageApi()
  const query = useQuery({
    queryKey: ['illustration', aspect, prompt],
    enabled: !!api && !!prompt,
    queryFn: () => api!(prompt, false, aspect),
    refetchInterval: (q) => (q.state.data?.status === 'generating' ? 2500 : false),
    staleTime: Infinity,
  })
  const trigger = useMutation({
    mutationFn: () => api!(prompt, true, aspect),
    onSuccess: (data) => {
      query.refetch()
      return data
    },
  })
  const status = trigger.isPending ? 'generating' : query.data?.status ?? (api ? 'pending' : 'none')
  return {
    enabled: !!api,
    status,
    url: query.data?.url,
    error: query.data?.error,
    generate: () => trigger.mutate(),
  }
}

export function Illustration({
  prompt,
  index,
  className = '',
  floatCls = '',
  aspect = 'square',
  style,
}: {
  prompt: string
  index?: number
  className?: string
  floatCls?: string
  aspect?: ImageAspect
  style?: React.CSSProperties
}) {
  const { enabled, status, url, generate } = useIllustration(prompt, aspect)
  const [zoomed, setZoomed] = useState(false)
  const base = `relative grid place-items-center overflow-hidden bg-neutral-100 text-neutral-400 ring-1 ring-neutral-200 ${className} ${floatCls}`

  if (status === 'ready' && url) {
    return (
      <>
        <button type="button" onClick={() => setZoomed(true)} className={`group/illus ${base} cursor-zoom-in`} style={style} title={prompt}>
          <img src={url} alt={prompt} className="h-full w-full object-cover transition-transform duration-200 group-hover/illus:scale-105" />
          <span className="absolute right-1.5 top-1.5 grid h-6 w-6 place-items-center rounded-full bg-black/45 text-white opacity-0 backdrop-blur-sm transition-opacity group-hover/illus:opacity-100">
            <Maximize2 className="h-3.5 w-3.5" />
          </span>
          {index != null && (
            <span className="absolute left-1.5 top-1 grid h-5 w-5 place-items-center rounded-full bg-white/90 text-[11px] font-bold text-neutral-500 shadow-sm">
              {index}
            </span>
          )}
          <span className="pointer-events-none absolute inset-x-0 bottom-0 translate-y-full bg-gradient-to-t from-black/85 to-transparent p-2 text-left text-[10px] leading-tight text-white transition-transform duration-200 group-hover/illus:translate-y-0">
            {prompt}
          </span>
        </button>
        {zoomed && <Lightbox url={url} alt={prompt} prompt={prompt} onClose={() => setZoomed(false)} />}
      </>
    )
  }

  return (
    <div
      title={prompt}
      className={`relative flex min-h-32 flex-col items-center justify-center gap-2 rounded-xl bg-neutral-100 p-3 text-center text-neutral-500 ring-1 ring-neutral-200 ${className || 'w-40'} ${floatCls}`}
      style={style}
    >
      {index != null && (
        <span className="absolute left-1.5 top-1.5 grid h-5 w-5 place-items-center rounded-full bg-white text-[11px] font-bold text-neutral-500 shadow-sm">
          {index}
        </span>
      )}
      {status === 'generating' ? (
        <>
          <Loader2 className="h-5 w-5 animate-spin text-brand-500" />
          <span className="text-[11px] leading-tight">Generating…</span>
        </>
      ) : enabled ? (
        <button
          type="button"
          onClick={generate}
          className="inline-flex items-center gap-1 whitespace-nowrap rounded-full bg-brand-600 px-3 py-1 text-[11px] font-semibold text-white transition-colors hover:bg-brand-700"
        >
          {status === 'error' ? <TriangleAlert className="h-3 w-3" /> : <ImageIcon className="h-3 w-3" />}
          {status === 'error' ? 'Retry' : 'Generate'}
        </button>
      ) : (
        <Sparkles className="h-4 w-4 text-brand-400" />
      )}
    </div>
  )
}

function Lightbox({ url, alt, prompt, onClose }: { url: string; alt: string; prompt: string; onClose: () => void }) {
  const [show, setShow] = useState(false)
  useEffect(() => {
    setShow(true)
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  return createPortal(
    <div
      onClick={onClose}
      className={`fixed inset-0 z-[100] flex flex-col items-center justify-center gap-3 bg-black/80 p-4 backdrop-blur-sm transition-opacity duration-200 ${show ? 'opacity-100' : 'opacity-0'}`}
    >
      <button
        type="button"
        onClick={onClose}
        className="absolute right-4 top-4 grid h-10 w-10 place-items-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
      >
        <X className="h-5 w-5" />
      </button>
      <img
        src={url}
        alt={alt}
        onClick={(e) => e.stopPropagation()}
        className={`max-h-[80vh] max-w-full rounded-xl object-contain shadow-2xl transition-transform duration-200 ${show ? 'scale-100' : 'scale-95'}`}
      />
      <p onClick={(e) => e.stopPropagation()} className="max-w-2xl text-center text-sm leading-relaxed text-white/80">
        {prompt}
      </p>
    </div>,
    document.body,
  )
}

export function ImagePlaceholder({ prompt, align, size = 'md' }: { prompt: string; align?: 'left' | 'right'; size?: Size }) {
  const { dense } = useContext(BlockCtx)
  if (size === 'full' || (dense && !align)) {
    return (
      <Illustration
        prompt={prompt}
        aspect="landscape"
        className="mx-auto my-2 w-full max-w-full rounded-xl"
        style={{ aspectRatio: '16/9', maxHeight: size === 'full' ? '26rem' : '14rem' }}
      />
    )
  }
  const effectiveAlign = align ?? 'right'
  const floatCls = effectiveAlign === 'right' ? 'float-right ml-3 mb-2' : 'float-left mr-3 mb-2'
  return (
    <Illustration
      prompt={prompt}
      aspect="landscape"
      className={`rounded-lg ${IMG_W[size] ?? IMG_W.md}`}
      floatCls={floatCls}
      style={{ aspectRatio: '4/3', maxWidth: 'min(48%, 16rem)' }}
    />
  )
}
