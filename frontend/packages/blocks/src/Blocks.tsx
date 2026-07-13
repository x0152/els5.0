import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { Check, RotateCcw } from 'lucide-react'
import { parseBlocks } from './parse.ts'
import { mockCheckAnswer, type CheckFn } from './check.ts'
import { ImageApiCtx, type ImageApi } from './images.ts'
import {
  ProgressCtx,
  ProduceCtx,
  CheckFreeCtx,
  useProgress,
  type PracticeAnswer,
  type ProgressCtxValue,
  type ProduceEvent,
  type CheckFreeFn,
} from './state.ts'
import { PROSE_CLS } from './markdown.tsx'
import { BlockCtx } from './render/context.ts'
import { RenderNodes } from './render/nodes.tsx'

// Everything the DSL renderer needs from the host app. All optional:
// omitted check falls back to local comparison, omitted progress means ephemeral state.
export type BlocksAdapters = {
  check?: CheckFn
  checkFree?: CheckFreeFn
  images?: ImageApi
  produce?: ProduceEvent
  progress?: {
    answers?: Record<string, PracticeAnswer>
    onChange: (answers: Record<string, PracticeAnswer>, completed: boolean) => void
  }
}

export function BlocksProvider({ adapters = {}, children }: { adapters?: BlocksAdapters; children: ReactNode }) {
  const mapRef = useRef<Record<string, PracticeAnswer>>(adapters.progress?.answers ?? {})
  const registry = useRef<Set<string>>(new Set())
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [version, setVersion] = useState(0)
  const onChange = adapters.progress?.onChange

  useEffect(
    () => () => {
      if (saveTimer.current) clearTimeout(saveTimer.current)
    },
    [],
  )

  const progressValue = useMemo<ProgressCtxValue>(() => {
    const scheduleSave = () => {
      if (!onChange) return
      if (saveTimer.current) clearTimeout(saveTimer.current)
      saveTimer.current = setTimeout(() => {
        const keys = [...registry.current]
        const completed = keys.length > 0 && keys.every((key) => mapRef.current[key]?.correct)
        onChange(mapRef.current, completed)
      }, 600)
    }
    return {
      // Always track in-memory so per-exercise "completed" and reset work
      // even without a persistence adapter (e.g. chat); onChange only adds saving.
      enabled: true,
      version,
      get: (k) => mapRef.current[k],
      set: (k, v) => {
        mapRef.current = { ...mapRef.current, [k]: v }
        setVersion((n) => n + 1)
        scheduleSave()
      },
      register: (k) => {
        if (!registry.current.has(k)) {
          registry.current.add(k)
          setVersion((n) => n + 1)
        }
      },
      keys: () => [...registry.current],
      remove: (ks) => {
        if (!ks.length) return
        const next = { ...mapRef.current }
        for (const k of ks) delete next[k]
        mapRef.current = next
        setVersion((n) => n + 1)
        scheduleSave()
      },
    }
  }, [onChange, version])

  return (
    <ProduceCtx.Provider value={adapters.produce ?? null}>
      <ImageApiCtx.Provider value={adapters.images ?? null}>
        <CheckFreeCtx.Provider value={adapters.checkFree ?? null}>
          <ProgressCtx.Provider value={progressValue}>{children}</ProgressCtx.Provider>
        </CheckFreeCtx.Provider>
      </ImageApiCtx.Provider>
    </ProduceCtx.Provider>
  )
}

// Renders a bare DSL fragment (no `## N` exercise headings). The universal building block:
// chat inserts, previews, anything that just needs the DSL rendered.
export function Blocks({ md, adapters = {} }: { md: string; adapters?: BlocksAdapters }) {
  return (
    <BlocksProvider adapters={adapters}>
      <BlockCtx.Provider value={{ dense: false, check: adapters.check ?? mockCheckAnswer, onTheory: () => {}, keyBase: 'b' }}>
        <FragmentBody md={md} />
      </BlockCtx.Provider>
    </BlocksProvider>
  )
}

function FragmentBody({ md }: { md: string }) {
  const progress = useProgress()
  const [nonce, setNonce] = useState(0)
  const keys = progress.keys()
  const done = keys.length > 0 && keys.every((k) => progress.get(k)?.correct)
  const touched = keys.some((k) => progress.get(k))
  return (
    <div className="@container">
      {touched && (
        <div className="mb-1.5 flex items-center justify-between">
          <span className={`inline-flex items-center gap-1 text-xs font-medium ${done ? 'text-emerald-600' : 'text-neutral-400'}`}>
            {done && <Check className="h-3.5 w-3.5" />}
            {done ? 'Completed' : 'In progress'}
          </span>
          <button
            type="button"
            onClick={() => {
              progress.remove(keys)
              setNonce((n) => n + 1)
            }}
            title="Reset"
            className="inline-flex items-center gap-1 rounded-md px-1.5 py-1 text-xs font-medium text-neutral-400 transition-colors hover:bg-neutral-100 hover:text-neutral-700"
          >
            <RotateCcw className="h-3.5 w-3.5" />
          </button>
        </div>
      )}
      <div key={nonce} className={`space-y-3 [display:flow-root] ${PROSE_CLS}`}>
        <RenderNodes nodes={parseBlocks(md)} />
      </div>
    </div>
  )
}
