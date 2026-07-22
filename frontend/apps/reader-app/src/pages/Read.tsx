import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { ArrowDownToLine, ArrowLeft, CheckCheck, RotateCcw } from 'lucide-react'
import { LoadingState, useAgentView } from '@els/ui'
import { useBook, useBookContent, useSaveProgress } from '../lib/books.ts'
import { emitReading } from '../lib/events.ts'

const CONTENT_STYLES = `
.book-content { max-width: 42rem; margin: 0 auto; color: #262626; line-height: 1.75; font-size: 1.05rem; }
.book-content img { max-width: 100%; height: auto; margin: 1rem auto; display: block; }
.book-content h1, .book-content h2, .book-content h3 { font-weight: 700; line-height: 1.3; margin: 1.5rem 0 0.75rem; color: #171717; }
.book-content h1 { font-size: 1.6rem; } .book-content h2 { font-size: 1.35rem; } .book-content h3 { font-size: 1.15rem; }
.book-content p { margin: 0 0 1rem; }
.book-content a { color: #059669; }
.book-content blockquote { border-left: 3px solid #a7f3d0; padding-left: 1rem; color: #525252; font-style: italic; }
`

type Para = { top: number; bottom: number; text: string }

const IDLE_MS = 3000

const BLOCK_TAG_RE = /<(p|h[1-6]|blockquote|li|td|th|figcaption|pre)\b[^>]*>/gi

function annotateOffsets(html: string): string {
  return html.replace(BLOCK_TAG_RE, (tag, _name, offset: number) => `${tag.slice(0, -1)} data-start="${offset + tag.length}">`)
}

export function Read() {
  const { id = '' } = useParams()
  return <ReadInner key={id} id={id} />
}

function ReadInner({ id }: { id: string }) {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { data: book, isLoading } = useBook(id)
  const { data: html, isLoading: contentLoading } = useBookContent(book?.content_url)
  const annotatedHtml = useMemo(() => (html ? annotateOffsets(html) : html), [html])
  const saveProgress = useSaveProgress()
  const jumped = useRef(false)
  const jumpPos = searchParams.has('pos') ? Number(searchParams.get('pos')) : null

  const scrollRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  const saveTimer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const idleTimer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const userScrolled = useRef(false)
  const programmatic = useRef(false)
  const inited = useRef(false)

  const parasRef = useRef<Para[]>([])
  const lineIdx = useRef(-1) // last paragraph the reading line has passed (read up to here)
  const emittedIdx = useRef(-1) // last paragraph already sent as a reading event
  const dragging = useRef(false)
  const [lineTop, setLineTop] = useState<number | null>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [lineAbove, setLineAbove] = useState(false) // line scrolled out of view above
  const [atEnd, setAtEnd] = useState(false) // scrolled to the bottom (near "Mark as read")
  const [summary, setSummary] = useState<{ words: number; paragraphs: number } | null>(null)

  useAgentView(
    book
      ? {
          app: 'reader',
          screen: 'read',
          title: book.title,
          info: 'The user is reading a book. To read text around a position — read_book_text with bookId and position.',
          ids: { bookId: id },
          state: { position: book.position },
        }
      : null,
  )

  const contentTop = () => scrollRef.current!.getBoundingClientRect().top - scrollRef.current!.scrollTop

  const measure = () => {
    const content = contentRef.current
    if (!scrollRef.current || !content) return
    const base = contentTop()
    parasRef.current = Array.from(content.querySelectorAll('p'))
      .map((p) => {
        const r = p.getBoundingClientRect()
        return { top: r.top - base, bottom: r.bottom - base, text: (p.textContent ?? '').trim() }
      })
      .filter((p) => p.text.length > 0)
  }

  const lineY = () => {
    const ps = parasRef.current
    const i = lineIdx.current
    return i >= 0 ? ps[i]!.bottom : ps[0] ? ps[0]!.top : 0
  }

  const renderLine = () => {
    if (!parasRef.current.length) return
    setLineTop(lineIdx.current < 0 ? null : lineY())
    const el = scrollRef.current
    if (el) setLineAbove(lineY() < el.scrollTop + 8)
  }

  const chars = () => parasRef.current.reduce((s, p) => s + p.text.length, 0) || 1

  // Restore the line from the saved character offset, independent of layout.
  const initLine = () => {
    const ps = parasRef.current
    if (!ps.length || !book) return
    const target = (book.text_length > 0 ? Math.min(1, book.position / book.text_length) : 0) * chars()
    let acc = 0
    let idx = -1
    for (let i = 0; i < ps.length; i++) {
      acc += ps[i]!.text.length
      if (acc <= target) idx = i
      else break
    }
    lineIdx.current = idx
    emittedIdx.current = idx
    inited.current = true
  }

  const jumpToOffset = (pos: number): boolean => {
    const content = contentRef.current
    const el = scrollRef.current
    if (!content || !el) return false
    let target: HTMLElement | null = null
    for (const node of Array.from(content.querySelectorAll<HTMLElement>('[data-start]'))) {
      if (Number(node.dataset.start) <= pos) target = node
      else break
    }
    if (!target) return false
    const node = target
    programmatic.current = true
    el.scrollTop = Math.max(0, node.getBoundingClientRect().top - contentTop() - el.clientHeight / 3)
    node.style.transition = 'background-color 0.5s'
    node.style.backgroundColor = 'rgba(250, 204, 21, 0.4)'
    setTimeout(() => {
      node.style.backgroundColor = ''
    }, 2000)
    return true
  }

  const flushReading = () => {
    const ps = parasRef.current
    const to = lineIdx.current
    if (to <= emittedIdx.current) return
    const texts: string[] = []
    for (let i = emittedIdx.current + 1; i <= to; i++) if (ps[i]) texts.push(ps[i]!.text)
    if (texts.length) emitReading(texts, { app: 'reader', book_id: id })
    emittedIdx.current = to
  }
  const restartIdle = () => {
    clearTimeout(idleTimer.current)
    idleTimer.current = setTimeout(flushReading, IDLE_MS)
  }

  const savePosition = () => {
    if (!book) return
    const ps = parasRef.current
    let acc = 0
    for (let i = 0; i <= lineIdx.current && i < ps.length; i++) acc += ps[i]!.text.length
    const pos = Math.round((acc / chars()) * book.text_length)
    clearTimeout(saveTimer.current)
    saveTimer.current = setTimeout(() => {
      if (id) saveProgress.mutate({ id, position: pos })
    }, 600)
  }

  useEffect(() => {
    if (!html || !book || !scrollRef.current || !contentRef.current) return
    const el = scrollRef.current
    // Re-run on each content resize (late-loading images) until the user
    // scrolls: keeps the line and the scroll position aligned.
    const align = () => {
      measure()
      if (!parasRef.current.length) return
      if (!inited.current) initLine()
      if (jumpPos != null && Number.isFinite(jumpPos) && !jumped.current && jumpToOffset(jumpPos)) {
        jumped.current = true
        userScrolled.current = true
        renderLine()
        return
      }
      if (!userScrolled.current) {
        const i = lineIdx.current
        const y = i >= 0 ? parasRef.current[i]!.bottom : 0
        programmatic.current = true
        el.scrollTop = Math.max(0, y - 24)
      }
      renderLine()
    }
    const raf = requestAnimationFrame(align)
    const ro = new ResizeObserver(align)
    ro.observe(contentRef.current)
    return () => {
      cancelAnimationFrame(raf)
      ro.disconnect()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [html, book])

  useEffect(() => {
    const onResize = () => {
      measure()
      renderLine()
    }
    window.addEventListener('resize', onResize)
    const move = (e: PointerEvent) => {
      if (!dragging.current) return
      const y = Math.max(0, e.clientY - contentTop())
      setLineTop(y)
    }
    const up = (e: PointerEvent) => {
      if (!dragging.current) return
      dragging.current = false
      setIsDragging(false)
      const ps = parasRef.current
      const y = Math.max(0, e.clientY - contentTop())
      let idx = -1
      for (let i = 0; i < ps.length && ps[i]!.bottom <= y; i++) idx = i
      lineIdx.current = idx
      renderLine()
      restartIdle()
      savePosition()
    }
    window.addEventListener('pointermove', move)
    window.addEventListener('pointerup', up)
    return () => {
      window.removeEventListener('resize', onResize)
      window.removeEventListener('pointermove', move)
      window.removeEventListener('pointerup', up)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(
    () => () => {
      clearTimeout(idleTimer.current)
      flushReading()
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  )

  const onScroll = () => {
    if (programmatic.current) programmatic.current = false
    else userScrolled.current = true
    const el = scrollRef.current
    if (el && parasRef.current.length) setLineAbove(lineY() < el.scrollTop + 8)
    if (el) setAtEnd(el.scrollHeight - el.scrollTop - el.clientHeight < 96)
  }

  const moveLineHere = () => {
    const el = scrollRef.current
    const ps = parasRef.current
    if (!el || !ps.length) return
    const top = el.scrollTop
    let idx = ps.length - 1
    for (let i = 0; i < ps.length; i++) {
      if (ps[i]!.bottom > top + 8) {
        idx = i
        break
      }
    }
    if (idx > lineIdx.current) {
      lineIdx.current = idx
      restartIdle()
      savePosition()
    }
    renderLine()
  }

  const startDrag = (e: React.PointerEvent) => {
    e.preventDefault()
    dragging.current = true
    setIsDragging(true)
  }

  const finish = () => {
    const ps = parasRef.current
    if (!ps.length || !book) return
    lineIdx.current = ps.length - 1
    clearTimeout(idleTimer.current)
    flushReading()
    renderLine()
    if (id) saveProgress.mutate({ id, position: book.text_length })
    const words = ps.reduce((s, p) => s + (p.text.trim() ? p.text.trim().split(/\s+/).length : 0), 0)
    setSummary({ words, paragraphs: ps.length })
  }

  const resetProgress = () => {
    lineIdx.current = -1
    emittedIdx.current = -1
    userScrolled.current = true
    programmatic.current = true
    clearTimeout(idleTimer.current)
    scrollRef.current?.scrollTo({ top: 0 })
    renderLine()
    if (id) saveProgress.mutate({ id, position: 0 })
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col bg-neutral-50">
      <style>{CONTENT_STYLES}</style>
      <header className="flex h-14 shrink-0 items-center gap-3 border-b border-neutral-200 bg-white px-5">
        <Link to=".." className="rounded-lg p-1.5 text-neutral-500 hover:bg-neutral-100 hover:text-neutral-800">
          <ArrowLeft size={18} />
        </Link>
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-semibold text-neutral-900">{book?.title ?? 'Reading'}</p>
          {book?.author && <p className="truncate text-xs text-neutral-500">{book.author}</p>}
        </div>
        <button
          onClick={resetProgress}
          title="Reset reading progress"
          className="flex shrink-0 items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs font-medium text-neutral-500 hover:bg-neutral-100 hover:text-neutral-800"
        >
          <RotateCcw size={15} />
          Reset
        </button>
      </header>

      {isLoading || contentLoading ? (
        <LoadingState className="flex-1 items-center py-0" />
      ) : !html ? (
        <p className="flex flex-1 items-center justify-center text-sm text-neutral-500">Book is not available.</p>
      ) : (
        <div ref={scrollRef} onScroll={onScroll} className="relative min-h-0 flex-1 overflow-y-auto px-6 py-8">
          <div className="mx-auto mb-8 max-w-[42rem] border-b border-neutral-200 pb-6">
            <h1 className="text-3xl font-extrabold leading-tight text-neutral-900">{book?.title}</h1>
            {book?.author && <p className="mt-2 text-sm text-neutral-500">{book.author}</p>}
          </div>
          <div ref={contentRef} className="book-content" dangerouslySetInnerHTML={{ __html: annotatedHtml ?? '' }} />
          <div className="mx-auto mt-10 flex max-w-[42rem] justify-center border-t border-neutral-200 pt-8">
            <button
              onClick={finish}
              className="inline-flex items-center gap-2 rounded-full bg-emerald-600 px-5 py-2.5 text-sm font-semibold text-white shadow-sm hover:bg-emerald-700"
            >
              <CheckCheck size={16} /> Mark as read
            </button>
          </div>
          {lineTop != null && (
            <div
              className="pointer-events-none absolute inset-x-0 z-10"
              style={{ top: lineTop, transition: isDragging ? 'none' : 'top 150ms ease' }}
            >
              <div className="mx-auto flex max-w-[42rem] items-center px-6">
                <span
                  onPointerDown={startDrag}
                  title="Drag to where you've read"
                  className="pointer-events-auto -ml-3 grid h-6 w-6 cursor-grab touch-none place-items-center rounded-full bg-brand-500 shadow ring-2 ring-white active:cursor-grabbing"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-white" />
                </span>
                <span className="h-0.5 flex-1 rounded bg-brand-400/70" />
              </div>
            </div>
          )}
          {lineAbove && !atEnd && (
            <div className="pointer-events-none sticky bottom-5 z-20 flex justify-center">
              <button
                onClick={moveLineHere}
                className="pointer-events-auto flex items-center gap-1.5 rounded-full bg-brand-500 px-3.5 py-2 text-xs font-medium text-white shadow-lg ring-1 ring-black/5 hover:bg-brand-600"
              >
                <ArrowDownToLine size={14} />
                Move line here
              </button>
            </div>
          )}
        </div>
      )}

      {summary && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4"
          onClick={() => navigate('..')}
        >
          <div className="w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="mb-3 flex items-center gap-2 text-emerald-600">
              <CheckCheck size={22} />
              <h2 className="text-lg font-bold">Finished</h2>
            </div>
            <p className="text-sm font-semibold text-neutral-900">{book?.title}</p>
            {book?.author && <p className="text-xs text-neutral-500">{book.author}</p>}
            <div className="mt-4 grid grid-cols-2 gap-3 text-center">
              <div className="rounded-xl bg-neutral-50 p-3">
                <p className="text-xl font-bold text-neutral-900">{summary.paragraphs}</p>
                <p className="text-xs text-neutral-500">paragraphs</p>
              </div>
              <div className="rounded-xl bg-neutral-50 p-3">
                <p className="text-xl font-bold text-neutral-900">{summary.words}</p>
                <p className="text-xs text-neutral-500">words</p>
              </div>
            </div>
            <button
              onClick={() => navigate('..')}
              className="mt-5 w-full rounded-lg bg-brand-600 py-2 text-sm font-semibold text-white hover:bg-brand-700"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
