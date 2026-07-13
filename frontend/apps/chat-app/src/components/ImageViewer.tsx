import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'
import { ImageViewerContext, type Viewer } from './imageViewerContext'

export function ImageViewerProvider({ children }: { children: ReactNode }) {
  const [img, setImg] = useState<{ src: string; alt?: string } | null>(null)
  const open = useCallback<Viewer>((src, alt) => setImg({ src, alt }), [])

  return (
    <ImageViewerContext.Provider value={open}>
      {children}
      {img && <Lightbox src={img.src} alt={img.alt} onClose={() => setImg(null)} />}
    </ImageViewerContext.Provider>
  )
}

function Lightbox({ src, alt, onClose }: { src: string; alt?: string; onClose: () => void }) {
  const [show, setShow] = useState(false)
  useEffect(() => {
    const raf = requestAnimationFrame(() => setShow(true))
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)
    return () => {
      cancelAnimationFrame(raf)
      window.removeEventListener('keydown', onKey)
    }
  }, [onClose])

  return createPortal(
    <div
      onClick={onClose}
      className={`fixed inset-0 z-[120] flex flex-col items-center justify-center gap-3 bg-black/85 p-4 backdrop-blur-sm transition-opacity duration-200 ${show ? 'opacity-100' : 'opacity-0'}`}
    >
      <button
        type="button"
        onClick={onClose}
        className="absolute right-4 top-4 grid h-10 w-10 place-items-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
      >
        <X className="h-5 w-5" />
      </button>
      <img
        src={src}
        alt={alt ?? ''}
        onClick={(e) => e.stopPropagation()}
        className={`max-h-[88vh] max-w-full rounded-xl object-contain shadow-2xl transition-transform duration-200 ${show ? 'scale-100' : 'scale-95'}`}
      />
      {alt && (
        <p onClick={(e) => e.stopPropagation()} className="max-w-2xl text-center text-sm leading-relaxed text-white/80">
          {alt}
        </p>
      )}
    </div>,
    document.body,
  )
}
