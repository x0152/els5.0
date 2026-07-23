import { useCallback, useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useQueryClient } from '@tanstack/react-query'
import { BookOpen, BookPlus, Check, ChevronDown, CirclePlay, Film, Loader2, MessageCircleQuestion, Mic, Plus, Square, X } from 'lucide-react'
import { cn, CefrBadge, FrequencyBars, PhonemePopover, SpeakButton, anchorOf, canonicalPhoneme, useRecorder, type PhonemeAnchor, type PhonemeGuideInfo } from '@els/ui'
import { type Api, type SpeechComponents } from '@els/api-client'
import { SpotsDialog } from './SpotsDialog.tsx'
import { streamAnalyze, type AnalyzeStreamItem } from './analyzeStream.ts'

interface Picked {
  text: string
  context: string
  rect: { top: number; bottom: number; left: number; width: number }
}

type RowState = 'idle' | 'adding' | 'added' | 'dup' | 'error'

interface Spot {
  ref: number
  example: string
}

interface Occurrence {
  mediaId: string
  mediaType: string
  title: string
  kind: string
  seriesTitle: string
  season: number
  episode: number
  author: string
  count: number
  spots: Spot[]
}

function spotHref(m: Occurrence, ref: number): string {
  return m.mediaType === 'film' ? `/v1/films/${m.mediaId}?t=${ref}` : `/v1/reader/${m.mediaId}?pos=${ref}`
}

interface Row {
  text: string
  kind: string
  description: string
  translation: string
  frequency: number
  cefr: string
  checked: boolean
  state: RowState
  note?: string
  common: boolean
  total: number
  media: Occurrence[]
}

const KIND_LABEL: Record<string, string> = {
  word: 'word',
  phrase: 'phrase',
  phrasal_verb: 'phrasal verb',
  idiom: 'idiom',
}

function toRow(it: AnalyzeStreamItem): Row {
  return {
    text: it.text,
    kind: it.kind,
    description: it.description,
    translation: it.translation ?? '',
    frequency: it.frequency,
    cefr: it.cefr,
    checked: false,
    state: it.existing ? 'dup' : 'idle',
    common: it.common,
    total: it.total,
    media: (it.media ?? []).map((m) => ({
      mediaId: m.media_id,
      mediaType: m.media_type,
      title: m.title,
      kind: m.kind ?? '',
      seriesTitle: m.series_title ?? '',
      season: m.season ?? 0,
      episode: m.episode ?? 0,
      author: m.author ?? '',
      count: m.count,
      spots: (m.spots ?? []).map((s) => ({ ref: s.ref, example: s.example ?? '' })),
    })),
  }
}

function MediaChips({ media, onPick }: { media: Occurrence[]; onPick: (m: Occurrence) => void }) {
  return (
    <div className="flex flex-wrap gap-1">
      {media.map((m, mi) => {
        const chip = (
          <>
            {m.mediaType === 'film' ? <Film size={11} /> : <BookOpen size={11} />}
            <span className="max-w-[160px] truncate">{m.title || 'Untitled'}</span>
            {m.count > 1 && <span className="text-neutral-400">×{m.count}</span>}
          </>
        )
        const chipClass =
          'inline-flex items-center gap-1 rounded-md bg-white px-2 py-1 text-[11px] text-neutral-600 ring-1 ring-neutral-200 hover:bg-neutral-100'
        return m.count <= 1 ? (
          <a key={`${m.title}-${mi}`} href={spotHref(m, m.spots[0]?.ref ?? 0)} onClick={(e) => e.stopPropagation()} className={chipClass}>
            {chip}
          </a>
        ) : (
          <button
            type="button"
            key={`${m.title}-${mi}`}
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              onPick(m)
            }}
            className={chipClass}
          >
            {chip}
          </button>
        )
      })}
    </div>
  )
}

type AssessOutput = SpeechComponents['schemas']['AssessOutput']

const VERDICT_STYLES: Record<string, string> = {
  good: 'bg-emerald-50 text-emerald-700 ring-emerald-200',
  close: 'bg-amber-50 text-amber-700 ring-amber-300',
  wrong: 'bg-red-50 text-red-700 ring-red-300',
  missing: 'bg-neutral-100 text-neutral-400 ring-neutral-200 line-through',
}

interface AnalyzeRowProps {
  row: Row
  showTranslations: boolean
  onCheck: (checked: boolean) => void
  onPickPlaces: (m: Occurrence) => void
  onSound: (symbol: string, anchor: PhonemeAnchor) => void
  assess: (text: string, blob: Blob) => Promise<AssessOutput | undefined>
}

function AnalyzeRow({ row, showTranslations, onCheck, onPickPlaces, onSound, assess }: AnalyzeRowProps) {
  const [open, setOpen] = useState(false)
  const [scoring, setScoring] = useState(false)
  const [result, setResult] = useState<AssessOutput | null>(null)
  const [failed, setFailed] = useState(false)
  const recorder = useRecorder((blob) => {
    setScoring(true)
    setFailed(false)
    assess(row.text, blob)
      .then((r) => setResult(r ?? null))
      .catch(() => setFailed(true))
      .finally(() => {
        setScoring(false)
        setOpen(true)
      })
  })

  const hasSources = !row.common && row.media.length > 0
  const disabled = row.state === 'added' || row.state === 'dup'
  const iconBtn = 'rounded-lg p-1.5 text-neutral-400 transition hover:bg-neutral-100 hover:text-neutral-600'

  return (
    <div className={cn('rounded-xl px-2 py-3 transition-colors', disabled ? 'opacity-60' : 'hover:bg-neutral-50')}>
      <div className="flex items-start gap-3">
        <input
          type="checkbox"
          checked={row.checked}
          disabled={disabled || row.state === 'adding'}
          onChange={(e) => onCheck(e.target.checked)}
          className="mt-1 h-4 w-4 shrink-0 accent-brand-600"
        />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate text-sm font-semibold text-neutral-900">{row.text}</span>
            {row.kind && KIND_LABEL[row.kind] && (
              <span className="shrink-0 rounded-full bg-neutral-100 px-1.5 py-0.5 text-[10px] font-medium text-neutral-500">
                {KIND_LABEL[row.kind]}
              </span>
            )}
            <CefrBadge level={row.cefr} className="shrink-0" />
            <FrequencyBars value={row.frequency} className="shrink-0" />
          </div>
          {row.description && <p className="text-sm text-neutral-700">{row.description}</p>}
          {showTranslations && row.translation && <p className="text-xs text-neutral-500">{row.translation}</p>}
          {row.note && <p className="text-xs text-amber-600">{row.note}</p>}
        </div>
        <div className="flex shrink-0 items-center gap-0.5">
          <span className="mr-1">
            {row.state === 'adding' && <Loader2 className="h-4 w-4 animate-spin text-neutral-400" />}
            {row.state === 'added' && <Check className="h-4 w-4 text-brand-600" />}
            {row.state === 'dup' && <span className="text-[10px] text-neutral-400">saved</span>}
          </span>
          <SpeakButton title="Listen" className={cn(iconBtn, 'rounded-lg p-1.5')} text={row.text} />
          <button
            type="button"
            onClick={recorder.state === 'recording' ? recorder.stop : recorder.start}
            disabled={scoring || recorder.state === 'unsupported'}
            title={recorder.state === 'recording' ? 'Stop recording' : 'Check my pronunciation'}
            className={cn(iconBtn, 'disabled:opacity-50', recorder.state === 'recording' && 'bg-red-50 text-red-600 hover:bg-red-100 hover:text-red-600')}
          >
            {recorder.state === 'recording' ? (
              <Square className="h-4 w-4" />
            ) : scoring ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Mic className="h-4 w-4" />
            )}
          </button>
          {recorder.blob && (
            <button type="button" onClick={recorder.play} disabled={recorder.state === 'recording'} title="Play my recording" className={cn(iconBtn, 'disabled:opacity-50')}>
              <CirclePlay className="h-4 w-4" />
            </button>
          )}
          {(hasSources || (row.common && row.total > 0) || result || failed) && (
            <button
              type="button"
              onClick={() => setOpen((v) => !v)}
              title={hasSources ? `Found in ${row.media.length} source${row.media.length === 1 ? '' : 's'}` : 'Details'}
              className={cn(iconBtn, 'flex items-center gap-0.5')}
            >
              {hasSources && <span className="text-[11px] tabular-nums">{row.media.length}</span>}
              <ChevronDown className={cn('h-4 w-4 transition-transform', open && 'rotate-180')} />
            </button>
          )}
        </div>
      </div>

      {recorder.state === 'recording' && (
        <p className="mt-2 pl-7 text-xs text-red-600">Recording… {recorder.elapsed}s — say “{row.text}” and press stop.</p>
      )}

      {open && (
        <div className="mt-2 space-y-2 pl-7">
          {failed && <p className="text-xs text-red-600">The pronunciation service did not respond. Try again.</p>}
          {result && (
            <div className="rounded-lg bg-neutral-50 p-2 ring-1 ring-neutral-200">
              <p className="text-xs font-medium text-neutral-900">
                Pronunciation:{' '}
                <span className={result.overall >= 85 ? 'text-emerald-600' : result.overall >= 60 ? 'text-amber-600' : 'text-red-600'}>
                  {result.overall}/100
                </span>
              </p>
              <div className="mt-1.5 flex flex-wrap items-center gap-1">
                {(result.words ?? []).flatMap((w, i) => [
                  ...(w.phonemes ?? []).map((p, j) => (
                    <button
                      key={`${i}-${j}`}
                      type="button"
                      onClick={(e) => onSound(p.expected, anchorOf(e.currentTarget))}
                      title={p.verdict === 'good' ? `/${p.expected}/` : `expected /${p.expected}/, heard /${p.heard ?? '—'}/`}
                      className={cn('rounded px-1.5 py-0.5 font-mono text-xs ring-1', VERDICT_STYLES[p.verdict] ?? VERDICT_STYLES.good)}
                    >
                      {p.expected}
                    </button>
                  )),
                  ...(w.extra ?? []).map((sym, j) => (
                    <span key={`${i}-x-${j}`} title="Extra sound" className="rounded bg-purple-50 px-1.5 py-0.5 font-mono text-xs text-purple-600 ring-1 ring-purple-200">
                      +{sym}
                    </span>
                  )),
                ])}
              </div>
            </div>
          )}
          {row.common && row.total > 0 && <p className="text-[11px] text-neutral-400">common word · seen {row.total}×</p>}
          {hasSources && <MediaChips media={row.media} onPick={onPickPlaces} />}
        </div>
      )}
    </div>
  )
}

function isEditable(node: Node | null): boolean {
  let el: Element | null = node instanceof Element ? node : (node?.parentElement ?? null)
  while (el) {
    const tag = el.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA') return true
    if (el instanceof HTMLElement && el.isContentEditable) return true
    el = el.parentElement
  }
  return false
}

function readSelection(): Picked | null {
  const sel = window.getSelection()
  if (!sel || sel.isCollapsed || sel.rangeCount === 0) return null
  if (isEditable(sel.anchorNode)) return null
  const text = sel.toString().replace(/\s+/g, ' ').trim()
  if (text.length < 2 || text.length > 1000) return null
  if (!/[A-Za-z]/.test(text)) return null
  let rect: DOMRect = sel.getRangeAt(0).getBoundingClientRect()
  if (sel.focusNode) {
    const caret = document.createRange()
    try {
      caret.setStart(sel.focusNode, sel.focusOffset)
      caret.collapse(true)
      const list = caret.getClientRects()
      if (list.length) rect = list[list.length - 1]!
    } catch {
      // keep bounding rect
    }
  }
  if (!rect || (rect.width === 0 && rect.height === 0)) return null
  const host = sel.anchorNode instanceof Element ? sel.anchorNode : sel.anchorNode?.parentElement
  const context = host?.textContent?.replace(/\s+/g, ' ').trim().slice(0, 500) ?? ''
  return { text, context, rect: { top: rect.top, bottom: rect.bottom, left: rect.left, width: rect.width } }
}

export function VocabLookupProvider({ api }: { api: Pick<Api, 'vocab' | 'account' | 'speech' | 'core'> }) {
  const qc = useQueryClient()
  const rootRef = useRef<HTMLDivElement>(null)
  const [showTranslations, setShowTranslations] = useState(true)
  const [pill, setPill] = useState<Picked | null>(null)
  const [picked, setPicked] = useState<Picked | null>(null)
  const [loading, setLoading] = useState(false)
  const [streaming, setStreaming] = useState(false)
  const [error, setError] = useState('')
  const [rows, setRows] = useState<Row[]>([])
  const [places, setPlaces] = useState<Occurrence | null>(null)
  const [sound, setSound] = useState<{ symbol: string; anchor: PhonemeAnchor } | null>(null)
  const [addedToast, setAddedToast] = useState(false)
  const [guide, setGuide] = useState<Map<string, PhonemeGuideInfo> | null>(null)
  const [fsEl, setFsEl] = useState<Element | null>(null)
  const abortRef = useRef<AbortController | null>(null)
  const pickedOpenRef = useRef(false)
  const touchUi = typeof window !== 'undefined' && window.matchMedia('(pointer: coarse)').matches

  useEffect(() => {
    pickedOpenRef.current = !!picked
  }, [picked])

  useEffect(() => {
    if (!picked) return
    api.account
      .accountMe()
      .then((me) => setShowTranslations(me?.show_translations ?? true))
      .catch(() => {})
  }, [api, picked])

  useEffect(() => {
    const onFs = () => setFsEl(document.fullscreenElement)
    document.addEventListener('fullscreenchange', onFs)
    return () => document.removeEventListener('fullscreenchange', onFs)
  }, [])

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | undefined
    const refresh = (immediate = false) => {
      clearTimeout(timer)
      const run = () => {
        if (pickedOpenRef.current) return
        const sel = window.getSelection()
        if (sel?.anchorNode && rootRef.current?.contains(sel.anchorNode)) return
        setPill(readSelection())
      }
      if (immediate) run()
      else timer = setTimeout(run, 150)
    }
    const onMouseUp = (e: Event) => {
      const target = e.target as Node | null
      if (rootRef.current && target && rootRef.current.contains(target)) return
      refresh(true)
    }
    const onKeyUp = (e: Event) => {
      const target = e.target as Node | null
      if (rootRef.current && target && rootRef.current.contains(target)) return
      refresh(true)
    }
    const onSelectionChange = () => refresh(false)
    const onTouchEnd = () => refresh(false)
    const onScroll = () => {
      setPill(null)
      refresh(false)
    }
    document.addEventListener('mouseup', onMouseUp)
    document.addEventListener('keyup', onKeyUp)
    document.addEventListener('selectionchange', onSelectionChange)
    document.addEventListener('touchend', onTouchEnd)
    window.addEventListener('scroll', onScroll, true)
    return () => {
      clearTimeout(timer)
      document.removeEventListener('mouseup', onMouseUp)
      document.removeEventListener('keyup', onKeyUp)
      document.removeEventListener('selectionchange', onSelectionChange)
      document.removeEventListener('touchend', onTouchEnd)
      window.removeEventListener('scroll', onScroll, true)
    }
  }, [])

  const open = useCallback((p: Picked) => {
    abortRef.current?.abort()
    const ac = new AbortController()
    abortRef.current = ac
    setPicked(p)
    setPill(null)
    setRows([])
    setError('')
    setLoading(true)
    setStreaming(true)
    let count = 0
    void streamAnalyze(
      p.text,
      p.context,
      {
        onItem: (it) => {
          count++
          setLoading(false)
          setRows((prev) => [...prev, toRow(it)])
        },
        onError: () => {
          if (count === 0) setError('Could not analyze the selection')
          setLoading(false)
          setStreaming(false)
        },
        onDone: () => {
          setLoading(false)
          setStreaming(false)
          if (count === 0) {
            setRows([{ text: p.text, kind: '', description: '', translation: '', frequency: 0, cefr: '', checked: false, state: 'idle', common: false, total: 0, media: [] }])
          }
        },
      },
      ac.signal,
    )
  }, [])

  useEffect(() => {
    const onAnalyze = (e: Event) => {
      const text = ((e as CustomEvent<string>).detail ?? '').replace(/\s+/g, ' ').trim()
      if (text.length >= 2) open({ text, context: text, rect: { top: 0, bottom: 0, left: 0, width: 0 } })
    }
    document.addEventListener('els:analyze', onAnalyze)
    return () => document.removeEventListener('els:analyze', onAnalyze)
  }, [open])

  const close = useCallback(() => {
    abortRef.current?.abort()
    setPicked(null)
    setRows([])
    setError('')
    setStreaming(false)
  }, [])

  useEffect(() => {
    if (!picked || sound) return
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && close()
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [picked, sound, close])

  const addSelected = useCallback(() => {
    const targets = rows.filter((r) => r.checked && (r.state === 'idle' || r.state === 'error'))
    if (targets.length === 0) return
    const texts = targets.map((t) => t.text)
    qc.setQueryData<string[]>(['vocab', 'pending-adds'], (old = []) => [...new Set([...old, ...texts])])
    for (const text of texts) {
      void api.vocab
        .addVocabUnit({ body: { text } })
        .catch(() => {})
        .finally(() => {
          void qc.invalidateQueries({ queryKey: ['vocab'] }).then(() => {
            qc.setQueryData<string[]>(['vocab', 'pending-adds'], (old = []) => old.filter((t) => t !== text))
          })
        })
    }
    setAddedToast(true)
    window.setTimeout(() => setAddedToast(false), 3000)
    close()
  }, [api, close, qc, rows])

  const pendingCount = rows.filter((r) => r.checked && (r.state === 'idle' || r.state === 'error')).length

  const assess = useCallback(
    async (text: string, blob: Blob) => {
      const form = new FormData()
      form.append('audio', blob, 'recording.webm')
      form.append('text', text)
      const result = await api.speech.assessSpeech({ body: form as unknown as never })
      if (result) {
        void api.core
          .ingestCoreEvents({
            body: {
              events: [
                {
                  target: text,
                  outcome: result.overall >= 60 ? 'ok' : 'fail',
                  skill: 'speaking',
                  source: { app: 'lookup', feature: 'analyze' },
                },
              ],
            },
          })
          .catch(() => {})
      }
      return result
    },
    [api],
  )

  useEffect(() => {
    if (!sound || guide) return
    api.speech
      .listSpeechPhonemes()
      .then((d) => setGuide(new Map((d?.items ?? []).map((p) => [p.symbol, p]))))
      .catch(() => {})
  }, [sound, guide, api])

  const pillBelow = pill ? touchUi || pill.rect.top < 56 : false
  const pillGap = touchUi ? 36 : 8
  const pillLeft = pill ? Math.min(Math.max(pill.rect.left + pill.rect.width / 2, 64), window.innerWidth - 64) : 0

  return createPortal(
    <div ref={rootRef}>
      {addedToast && (
        <div
          style={{ zIndex: 2147483646 }}
          className="fixed bottom-6 left-1/2 flex -translate-x-1/2 items-center gap-2 rounded-full bg-neutral-900 px-4 py-2 text-sm font-medium text-white shadow-lg ring-1 ring-black/10"
        >
          <Check size={15} className="text-brand-400" />
          Added to My Vocabulary
        </div>
      )}
      {pill && (
        <div
          style={{
            position: 'fixed',
            top: pillBelow ? pill.rect.bottom + pillGap : pill.rect.top - pillGap,
            left: pillLeft,
            transform: `translate(-50%, ${pillBelow ? '0' : '-100%'})`,
            zIndex: 2147483646,
          }}
          className="flex items-center overflow-hidden rounded-full bg-neutral-900 text-xs font-medium text-white shadow-lg ring-1 ring-black/10"
        >
          <button
            type="button"
            onPointerDown={(e) => e.preventDefault()}
            onPointerUp={(e) => {
              e.preventDefault()
              void open(pill)
            }}
            className="flex items-center gap-1.5 px-3 py-1.5 transition-colors hover:bg-white/10"
          >
            <BookPlus size={14} />
            Analyze
          </button>
          <button
            type="button"
            onPointerDown={(e) => e.preventDefault()}
            onPointerUp={(e) => {
              e.preventDefault()
              const text = pill.text
              setPill(null)
              document.dispatchEvent(new CustomEvent('els:ask', { detail: text }))
            }}
            className="flex items-center gap-1.5 border-l border-white/15 px-3 py-1.5 transition-colors hover:bg-white/10"
          >
            <MessageCircleQuestion size={14} />
            Ask
          </button>
        </div>
      )}

      {picked && (
        <div
          className="fixed inset-0 z-[2147483647] flex items-center justify-center bg-black/40 p-4"
          onClick={close}
        >
          <div
            className="flex max-h-[80vh] w-full max-w-md flex-col overflow-hidden rounded-2xl bg-white shadow-2xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-start justify-between gap-3 border-b border-neutral-100 px-5 py-4">
              <div className="min-w-0">
                <p className="text-xs font-bold uppercase tracking-wider text-neutral-400">Analyze</p>
                <p className="mt-1 truncate text-sm font-medium text-neutral-800">«{picked.text}»</p>
              </div>
              <button type="button" onClick={close} className="rounded-lg p-1 text-neutral-400 hover:bg-neutral-100 hover:text-neutral-700">
                <X size={18} />
              </button>
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto px-3 py-3">
              {loading && (
                <div className="flex items-center justify-center gap-2 py-10 text-sm text-neutral-500">
                  <Loader2 className="h-5 w-5 animate-spin" /> Analyzing…
                </div>
              )}
              {error && <p className="px-2 py-6 text-center text-sm text-red-600">{error}</p>}
              {!loading && !error &&
                rows.map((row, i) => (
                  <div key={`${row.text}-${i}`} className={cn(i > 0 && 'border-t border-neutral-100')}>
                    <AnalyzeRow
                      row={row}
                      showTranslations={showTranslations}
                      onCheck={(checked) => setRows((prev) => prev.map((r, idx) => (idx === i ? { ...r, checked } : r)))}
                      onPickPlaces={setPlaces}
                      onSound={(symbol, anchor) => setSound({ symbol, anchor })}
                      assess={assess}
                    />
                  </div>
                ))}
              {streaming && !loading && !error && (
                <div className="flex items-center gap-2 px-3 py-2 text-xs text-neutral-400">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" /> Analyzing…
                </div>
              )}
            </div>

            {!loading && !error && (
              <div className="border-t border-neutral-100 px-5 py-3">
                <button
                  type="button"
                  onClick={() => void addSelected()}
                  disabled={pendingCount === 0}
                  className="flex w-full items-center justify-center gap-2 rounded-xl bg-brand-600 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-brand-700 disabled:bg-neutral-200 disabled:text-neutral-400"
                >
                  <Plus size={16} />
                  Add{pendingCount > 0 ? ` (${pendingCount})` : ''}
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {sound && (
        <PhonemePopover
          symbol={canonicalPhoneme(sound.symbol)}
          info={guide?.get(canonicalPhoneme(sound.symbol))}
          anchor={sound.anchor}
          onClose={() => setSound(null)}
        />
      )}

      {places && (
        <SpotsDialog
          title={places.title}
          mediaType={places.mediaType}
          kind={places.kind}
          seriesTitle={places.seriesTitle}
          season={places.season}
          episode={places.episode}
          author={places.author}
          spots={places.spots}
          hrefFor={(ref) => spotHref(places, ref)}
          onClose={() => setPlaces(null)}
        />
      )}
    </div>,
    fsEl ?? document.body,
  )
}
