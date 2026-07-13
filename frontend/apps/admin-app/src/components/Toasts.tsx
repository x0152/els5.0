import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { CheckCircle2, AlertCircle, X } from 'lucide-react'
import { cn } from '@els/ui'

export type ToastKind = 'success' | 'error'

interface Toast {
  id: number
  kind: ToastKind
  message: string
}

interface ToastApi {
  success: (msg: string) => void
  error: (msg: string) => void
}

const Ctx = createContext<ToastApi | null>(null)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<Toast[]>([])
  const timers = useRef<ReturnType<typeof setTimeout>[]>([])

  useEffect(() => {
    const t = timers.current
    return () => t.forEach(clearTimeout)
  }, [])

  const push = useCallback((kind: ToastKind, message: string) => {
    const id = Date.now() + Math.random()
    setItems((prev) => [...prev, { id, kind, message }])
    timers.current.push(
      setTimeout(() => {
        setItems((prev) => prev.filter((t) => t.id !== id))
      }, 3500),
    )
  }, [])

  const api = useMemo<ToastApi>(
    () => ({
      success: (m) => push('success', m),
      error: (m) => push('error', m),
    }),
    [push],
  )

  return (
    <Ctx.Provider value={api}>
      {children}
      <div className="fixed bottom-5 right-5 z-50 flex flex-col gap-2 pointer-events-none">
        {items.map((t) => (
          <div
            key={t.id}
            className={cn(
              'pointer-events-auto flex items-center gap-3 min-w-[260px] max-w-sm px-4 py-3 rounded-lg shadow-lg ring-1 bg-white animate-[slideIn_.2s_ease-out]',
              t.kind === 'success' ? 'ring-emerald-200' : 'ring-red-200',
            )}
            role="status"
          >
            {t.kind === 'success' ? (
              <CheckCircle2 size={18} className="text-emerald-600 shrink-0" />
            ) : (
              <AlertCircle size={18} className="text-red-600 shrink-0" />
            )}
            <span className="text-sm text-neutral-800 flex-1">{t.message}</span>
            <button
              type="button"
              onClick={() => setItems((prev) => prev.filter((x) => x.id !== t.id))}
              className="text-neutral-400 hover:text-neutral-700"
              aria-label="Close"
            >
              <X size={14} />
            </button>
          </div>
        ))}
      </div>
      <style>{`
        @keyframes slideIn { from { opacity:0; transform: translateY(8px) } to { opacity:1; transform: translateY(0) } }
      `}</style>
    </Ctx.Provider>
  )
}

export function useToast(): ToastApi {
  const api = useContext(Ctx)
  if (!api) throw new Error('useToast must be used inside <ToastProvider>')
  return api
}
