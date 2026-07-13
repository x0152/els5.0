import { memo } from 'react'
import ReactMarkdown, { type Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { Maximize2 } from 'lucide-react'
import { Blocks, PracticeSheet, type BlocksAdapters, type ImageApi, type IllustrationStatus } from '@els/blocks'
import { api } from '../lib/api'
import { useImageViewer } from './imageViewerContext'

const imageApi: ImageApi = async (prompt, trigger, aspect) =>
  (await api.learn.ensureIllustration({ body: { prompt, trigger, aspect } })) as IllustrationStatus

const adapters: BlocksAdapters = { images: imageApi }

function ChatImage({ src, alt }: { src?: string; alt?: string }) {
  const open = useImageViewer()
  if (!src) return null
  return (
    <button
      type="button"
      onClick={() => open(src, alt)}
      className="group relative my-2 block cursor-zoom-in overflow-hidden rounded-xl ring-1 ring-neutral-200"
    >
      <img src={src} alt={alt} loading="lazy" className="max-h-80 w-full object-cover transition-transform duration-200 group-hover:scale-[1.03]" />
      <span className="absolute right-2 top-2 grid h-7 w-7 place-items-center rounded-full bg-black/45 text-white opacity-0 backdrop-blur-sm transition-opacity group-hover:opacity-100">
        <Maximize2 className="h-3.5 w-3.5" />
      </span>
    </button>
  )
}

const COMPONENTS: Components = {
  p: ({ children }) => <p className="my-1.5 first:mt-0 last:mb-0 leading-relaxed">{children}</p>,
  a: ({ href, children }) => (
    <a
      href={href}
      target="_blank"
      rel="noreferrer noopener"
      className="text-brand-700 underline underline-offset-2 hover:text-brand-800"
    >
      {children}
    </a>
  ),
  ul: ({ children }) => <ul className="my-2 ml-5 list-disc space-y-1">{children}</ul>,
  ol: ({ children }) => <ol className="my-2 ml-5 list-decimal space-y-1">{children}</ol>,
  li: ({ children }) => <li className="leading-relaxed">{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className="my-2 border-l-2 border-neutral-300 pl-3 text-neutral-600 italic">{children}</blockquote>
  ),
  h1: ({ children }) => <h1 className="my-2 text-base font-semibold">{children}</h1>,
  h2: ({ children }) => <h2 className="my-2 text-[15px] font-semibold">{children}</h2>,
  h3: ({ children }) => <h3 className="my-2 text-sm font-semibold">{children}</h3>,
  table: ({ children }) => (
    <div className="my-2 overflow-x-auto">
      <table className="border-collapse text-sm">{children}</table>
    </div>
  ),
  th: ({ children }) => (
    <th className="border border-neutral-200 bg-neutral-50 px-2 py-1 text-left font-semibold">{children}</th>
  ),
  td: ({ children }) => <td className="border border-neutral-200 px-2 py-1 align-top">{children}</td>,
  hr: () => <hr className="my-3 border-neutral-200" />,
  img: ({ src, alt }) => <ChatImage src={typeof src === 'string' ? src : undefined} alt={alt} />,
  code: ({ className, children }) => {
    if (!className) {
      return (
        <code className="rounded bg-neutral-100 px-1 py-0.5 text-[12.5px] font-mono text-neutral-800">{children}</code>
      )
    }
    return <code className={`${className} font-mono text-[12.5px]`}>{children}</code>
  },
  pre: ({ children }) => (
    <pre className="my-2 overflow-x-auto rounded-xl bg-neutral-50 p-3 text-[12.5px] leading-relaxed text-neutral-800 ring-1 ring-neutral-200 [&_code]:rounded-none [&_code]:bg-transparent [&_code]:p-0 [&_code]:text-inherit">
      {children}
    </pre>
  ),
}

type Part = { type: 'md' | 'gaps' | 'pending'; content: string }

// DSL leaf types that render interactively even when the model forgets the ```blocks wrapper.
const DSL_TYPES = new Set(['bank', 'box', 'bubble', 'write', 'image', 'images', 'timeline', 'match', 'sort', 'highlight', 'gloss', 'fork'])

function splitGapBlocks(text: string): Part[] {
  const lines = text.split('\n')
  const parts: Part[] = []
  let md: string[] = []
  const flushMd = () => {
    if (md.join('\n').trim()) parts.push({ type: 'md', content: md.join('\n') })
    md = []
  }
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i] ?? ''
    const t = line.trim()
    if (/^```(?:gaps|blocks)[ \t]*$/.test(t)) {
      flushMd()
      const body: string[] = []
      let closed = false
      for (i++; i < lines.length; i++) {
        if (/^```[ \t]*$/.test((lines[i] ?? '').trim())) {
          closed = true
          break
        }
        body.push(lines[i] ?? '')
      }
      parts.push(closed ? { type: 'gaps', content: body.join('\n') } : { type: 'pending', content: '' })
      continue
    }
    if (/^```/.test(t)) {
      md.push(line)
      for (i++; i < lines.length; i++) {
        md.push(lines[i] ?? '')
        if (/^```[ \t]*$/.test((lines[i] ?? '').trim())) break
      }
      continue
    }
    const dsl = /^~~~(\w[\w-]*)/.exec(t)
    if (dsl && DSL_TYPES.has(dsl[1] ?? '')) {
      flushMd()
      const body: string[] = [line]
      let closed = false
      for (i++; i < lines.length; i++) {
        body.push(lines[i] ?? '')
        if ((lines[i] ?? '').trim() === '~~~') {
          closed = true
          break
        }
      }
      parts.push(closed ? { type: 'gaps', content: body.join('\n') } : { type: 'pending', content: '' })
      continue
    }
    md.push(line)
  }
  flushMd()
  return parts
}

function GapsPending() {
  return (
    <div className="my-1.5 animate-pulse space-y-2 rounded-xl border border-neutral-200 bg-neutral-50 p-3">
      <div className="h-3 w-3/4 rounded bg-neutral-200" />
      <div className="h-3 w-1/2 rounded bg-neutral-200" />
    </div>
  )
}

export const Markdown = memo(function Markdown({ text }: { text: string }) {
  const parts =
    text.includes('```gaps') || text.includes('```blocks') || /^~~~\w/m.test(text)
      ? splitGapBlocks(text)
      : [{ type: 'md' as const, content: text }]
  return (
    <div className="text-sm text-neutral-800 min-w-0 break-words [overflow-wrap:anywhere]">
      {parts.map((p, i) => {
        if (p.type === 'gaps') {
          const full = /^##\s/m.test(p.content)
          return (
            <div key={i} className="my-2 min-w-0 rounded-xl border border-brand-100 bg-brand-50/40 p-3">
              {full ? <PracticeSheet exercises={p.content} adapters={adapters} /> : <Blocks md={p.content} adapters={adapters} />}
            </div>
          )
        }
        if (p.type === 'pending') return <GapsPending key={i} />
        return (
          <ReactMarkdown key={i} remarkPlugins={[remarkGfm]} components={COMPONENTS}>
            {p.content}
          </ReactMarkdown>
        )
      })}
    </div>
  )
})
