import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
  type ReactNode,
} from 'react'
import { createPortal } from 'react-dom'
import { GripVertical, Maximize2, Pause, Play, Volume2, VolumeX, X } from 'lucide-react'

export interface MiniPlayerPayload {
  id: string
  src: string
  title: string
  startMs: number
  playing: boolean
  returnTo: string
  onProgress?: (ms: number) => void
}

interface MiniPlayerCtxValue {
  open: (payload: MiniPlayerPayload) => void
  close: () => void
  activeId: string | null
}

const MiniPlayerCtx = createContext<MiniPlayerCtxValue | null>(null)

export function useMiniPlayer(): MiniPlayerCtxValue {
  const ctx = useContext(MiniPlayerCtx)
  if (!ctx) throw new Error('useMiniPlayer must be used within MiniPlayerProvider')
  return ctx
}

export function MiniPlayerProvider({ children, onNavigate }: { children: ReactNode; onNavigate: (to: string) => void }) {
  const [payload, setPayload] = useState<MiniPlayerPayload | null>(null)
  const open = useCallback((p: MiniPlayerPayload) => setPayload(p), [])
  const close = useCallback(() => setPayload(null), [])

  return (
    <MiniPlayerCtx.Provider value={{ open, close, activeId: payload?.id ?? null }}>
      {children}
      {payload && (
        <MiniPlayerWindow
          key={payload.id}
          payload={payload}
          onClose={close}
          onExpand={() => {
            onNavigate(payload.returnTo)
            close()
          }}
        />
      )}
    </MiniPlayerCtx.Provider>
  )
}

const W = 320
const H = 240
const clamp = (v: number, min: number, max: number) => Math.min(Math.max(v, min), max)

function MiniPlayerWindow({ payload, onClose, onExpand }: { payload: MiniPlayerPayload; onClose: () => void; onExpand: () => void }) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const [paused, setPaused] = useState(!payload.playing)
  const [muted, setMuted] = useState(false)
  const [pos, setPos] = useState(() => ({ x: window.innerWidth - W - 16, y: window.innerHeight - H - 16 }))
  const [dragging, setDragging] = useState(false)
  const offset = useRef({ dx: 0, dy: 0 })

  const payloadRef = useRef(payload)
  useEffect(() => {
    payloadRef.current = payload
  }, [payload])

  const saveProgress = useCallback(() => {
    const v = videoRef.current
    if (v) payloadRef.current.onProgress?.(Math.round(v.currentTime * 1000))
  }, [])

  useEffect(() => {
    const t = setInterval(() => {
      if (!videoRef.current?.paused) saveProgress()
    }, 5000)
    return () => {
      clearInterval(t)
      saveProgress()
    }
  }, [saveProgress])

  useEffect(() => {
    if (!dragging) return
    const move = (e: PointerEvent) => {
      setPos({
        x: clamp(e.clientX - offset.current.dx, 0, window.innerWidth - W),
        y: clamp(e.clientY - offset.current.dy, 0, window.innerHeight - H),
      })
    }
    const up = () => setDragging(false)
    window.addEventListener('pointermove', move)
    window.addEventListener('pointerup', up)
    return () => {
      window.removeEventListener('pointermove', move)
      window.removeEventListener('pointerup', up)
    }
  }, [dragging])

  const startDrag = (e: ReactPointerEvent) => {
    offset.current = { dx: e.clientX - pos.x, dy: e.clientY - pos.y }
    setDragging(true)
  }

  const onLoaded = () => {
    const v = videoRef.current
    if (!v) return
    if (payload.startMs > 0) v.currentTime = payload.startMs / 1000
    if (payload.playing) void v.play()
  }

  const togglePlay = () => {
    const v = videoRef.current
    if (!v) return
    if (v.paused) void v.play()
    else v.pause()
  }

  return createPortal(
    <div
      style={{ left: pos.x, top: pos.y, width: W }}
      className="fixed z-[2147483600] overflow-hidden rounded-xl bg-neutral-900 shadow-2xl ring-1 ring-black/20"
    >
      <div
        onPointerDown={startDrag}
        className="flex cursor-grab touch-none items-center gap-1.5 bg-neutral-800 px-2 py-1 text-white active:cursor-grabbing"
      >
        <GripVertical className="h-3.5 w-3.5 shrink-0 text-neutral-400" />
        <span className="min-w-0 flex-1 truncate text-[11px] font-medium">{payload.title}</span>
        <button type="button" onClick={onExpand} title="Open full player" className="rounded p-0.5 text-neutral-300 hover:bg-white/10 hover:text-white">
          <Maximize2 className="h-3.5 w-3.5" />
        </button>
        <button type="button" onClick={onClose} title="Close" className="rounded p-0.5 text-neutral-300 hover:bg-white/10 hover:text-white">
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
      <div className="group relative bg-black">
        <video
          ref={videoRef}
          src={payload.src}
          playsInline
          className="block max-h-[200px] w-full"
          autoPlay={payload.playing}
          onLoadedMetadata={onLoaded}
          onClick={togglePlay}
          onPlay={() => setPaused(false)}
          onPause={() => setPaused(true)}
          onVolumeChange={(e) => setMuted(e.currentTarget.muted)}
        />
        <div className="absolute inset-x-0 bottom-0 flex items-center gap-3 bg-gradient-to-t from-black/70 to-transparent px-2 py-1.5 text-white transition-opacity [@media(hover:hover)]:opacity-0 [@media(hover:hover)]:group-hover:opacity-100">
          <button type="button" onClick={togglePlay} className="transition-colors hover:text-brand-400">
            {paused ? <Play className="h-4 w-4" /> : <Pause className="h-4 w-4" />}
          </button>
          <button
            type="button"
            onClick={() => videoRef.current && (videoRef.current.muted = !videoRef.current.muted)}
            className="transition-colors hover:text-brand-400"
          >
            {muted ? <VolumeX className="h-4 w-4" /> : <Volume2 className="h-4 w-4" />}
          </button>
        </div>
      </div>
    </div>,
    document.body,
  )
}
