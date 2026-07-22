import type { ReactNode } from 'react'
import ReactMarkdown, { type Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { iconify } from './emoji.tsx'

export const PROSE_CLS = 'min-w-0 break-words [overflow-wrap:anywhere] [&_*:first-child]:mt-0 [&_*:last-child]:mb-0'

function h(level: 1 | 2 | 3 | 4, className: string) {
  return ({ children }: { children?: ReactNode }) => {
    const props = { className, children: iconify(children) }
    if (level === 1) return <h1 {...props} />
    if (level === 2) return <h2 {...props} />
    if (level === 3) return <h3 {...props} />
    return <h4 {...props} />
  }
}

function makeMdComponents(variant: 'theory' | 'practice'): Components {
  const body = variant === 'theory' ? 'text-[15px] leading-[1.65] text-neutral-700' : 'text-sm leading-relaxed text-neutral-700'
  const gap = variant === 'theory' ? 'space-y-3' : 'space-y-2.5'

  return {
    h1: h(1, 'mb-3 mt-1 text-xl font-bold text-neutral-900'),
    h2: h(2, 'mb-2.5 mt-4 text-lg font-bold text-neutral-900 first:mt-0'),
    h3: h(3, 'mb-2 mt-3 text-base font-semibold text-neutral-800 first:mt-0'),
    h4: h(4, 'mb-1.5 mt-2.5 text-sm font-semibold text-neutral-800 first:mt-0'),
    p: ({ children }) => <p className={`${body} mb-2 last:mb-0`}>{iconify(children)}</p>,
    strong: ({ children }) => <strong className="font-semibold text-neutral-900">{iconify(children)}</strong>,
    em: ({ children }) => <em className="italic text-neutral-500">{iconify(children)}</em>,
    a: ({ href, children }) => (
      <a href={href} target="_blank" rel="noreferrer" className="font-medium text-brand-700 underline decoration-brand-300 underline-offset-2 hover:text-brand-800">
        {children}
      </a>
    ),
    ul: ({ children }) => (
      <ul
        className={`${gap} [&>li]:relative [&>li]:[display:flow-root] [&>li]:pl-4 [&>li]:before:absolute [&>li]:before:left-0 [&>li]:before:top-[0.55em] [&>li]:before:h-1.5 [&>li]:before:w-1.5 [&>li]:before:rounded-full [&>li]:before:bg-brand-400 [&>li]:before:content-['']`}
      >
        {children}
      </ul>
    ),
    ol: ({ children }) => (
      <ol className={`list-decimal ${gap} pl-6 marker:font-semibold marker:text-brand-500/70 [&>li]:pl-1`}>{children}</ol>
    ),
    li: ({ children }) => <li className={body}>{iconify(children)}</li>,
    hr: () => <hr className="my-4 border-0 border-t border-neutral-200" />,
    blockquote: ({ children }) => (
      <blockquote className="my-3 rounded-r-xl border-l-[3px] border-brand-400 bg-white px-4 py-2.5 text-sm leading-relaxed text-neutral-700 shadow-sm ring-1 ring-neutral-200/80 [display:flow-root]">
        {children}
      </blockquote>
    ),
    table: ({ children }) => (
      <div className="my-3 w-full max-w-full overflow-x-auto rounded-xl bg-white shadow-sm ring-1 ring-neutral-200">
        <table className="w-full min-w-[12rem] text-left text-sm text-neutral-700">{children}</table>
      </div>
    ),
    thead: ({ children }) => <thead className="bg-neutral-50">{children}</thead>,
    th: ({ children }) => <th className="whitespace-nowrap px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-neutral-500">{iconify(children)}</th>,
    td: ({ children }) => <td className="border-t border-neutral-100 px-3 py-2 align-top">{iconify(children)}</td>,
    tr: ({ children }) => <tr className="even:bg-neutral-50/60">{children}</tr>,
    code: ({ className, children }) => {
      const block = className?.includes('language-')
      if (block) {
        return <code className={`${className} font-mono text-[0.85em]`}>{children}</code>
      }
      return <code className="rounded-md bg-neutral-100 px-1.5 py-0.5 font-mono text-[0.85em] font-medium text-brand-800">{children}</code>
    },
    pre: ({ children }) => (
      <pre className="my-3 overflow-x-auto rounded-xl bg-neutral-900 p-3.5 text-[13px] leading-relaxed text-neutral-100">{children}</pre>
    ),
  }
}

const theoryComponents = makeMdComponents('theory')
const practiceComponents = makeMdComponents('practice')

export function Prose({ md, variant = 'practice' }: { md: string; variant?: 'theory' | 'practice' }) {
  return (
    <div className={PROSE_CLS}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={variant === 'theory' ? theoryComponents : practiceComponents}>
        {md}
      </ReactMarkdown>
    </div>
  )
}

export function Inline({ text }: { text: string }) {
  const parts = text.split(/(\*\*[^*]+\*\*|__[^_]+__|\*[^*\s][^*]*\*|_[^_]+_)/g)
  return (
    <>
      {parts.map((p, i) => {
        const bold = (p.startsWith('**') && p.endsWith('**')) || (p.startsWith('__') && p.endsWith('__'))
        if (bold && p.length > 4) {
          return (
            <strong key={i} className="font-semibold text-neutral-900">
              {iconify(p.slice(2, -2))}
            </strong>
          )
        }
        const italic = (p.startsWith('*') && p.endsWith('*')) || (p.startsWith('_') && p.endsWith('_'))
        if (italic && p.length > 2) {
          return (
            <em key={i} className="italic text-neutral-500">
              {iconify(p.slice(1, -1))}
            </em>
          )
        }
        return <span key={i}>{iconify(p)}</span>
      })}
    </>
  )
}
