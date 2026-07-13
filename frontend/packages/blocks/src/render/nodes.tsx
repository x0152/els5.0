import { useContext } from 'react'
import type { Node } from '../parse.ts'
import { Prose, Inline } from '../markdown.tsx'
import { ACCENT, BlockCtx, gridColsClass, scopeKey } from './context.ts'
import { GapLines, ItemText } from './gaps.tsx'
import { Illustration, ImagePlaceholder } from './illustrations.tsx'
import { WriteArea, WriteLines } from './write.tsx'
import { Timeline } from './timeline.tsx'
import { MatchConnect } from './match.tsx'
import { SortBuckets } from './sort.tsx'
import { HighlightText } from './highlight.tsx'
import { GlossText } from './gloss.tsx'
import { Fork } from './fork.tsx'

export type RenderVariant = 'theory' | 'practice'

export function RenderNodes({ nodes, variant = 'practice' }: { nodes: Node[]; variant?: RenderVariant }) {
  const parent = useContext(BlockCtx)
  const out: React.ReactNode[] = []
  for (let i = 0; i < nodes.length; ) {
    const n = nodes[i]!
    if (n.t === 'bubble') {
      let j = i
      while (j < nodes.length && nodes[j]!.t === 'bubble') j++
      const group = nodes.slice(i, j) as Extract<Node, { t: 'bubble' }>[]
      if (group.length > 1) {
        out.push(
          <BlockCtx.Provider key={i} value={{ ...parent, keyBase: scopeKey(parent.keyBase, i) }}>
            <Conversation bubbles={group} />
          </BlockCtx.Provider>,
        )
        i = j
        continue
      }
    }
    out.push(
      <BlockCtx.Provider key={i} value={{ ...parent, keyBase: scopeKey(parent.keyBase, i) }}>
        <RenderNode node={n} variant={variant} />
      </BlockCtx.Provider>,
    )
    i++
  }
  return <>{out}</>
}

function RenderNode({ node, variant = 'practice' }: { node: Node; variant?: RenderVariant }) {
  const parent = useContext(BlockCtx)
  switch (node.t) {
    case 'grid':
      return (
        <div className={`grid gap-3 ${gridColsClass(node.cols)}`}>
          {node.children.map((c, i) => (
            <BlockCtx.Provider key={i} value={{ ...parent, dense: node.cols > 1, keyBase: scopeKey(parent.keyBase, i) }}>
              <div className={`min-w-0 [display:flow-root] ${node.divider ? 'rounded-xl border border-neutral-200/90 bg-white p-3.5 shadow-sm' : ''}`}>
                <RenderNode node={c} variant={variant} />
              </div>
            </BlockCtx.Provider>
          ))}
        </div>
      )
    case 'stack':
      return (
        <div className={`min-w-0 space-y-2.5 ${node.center ? 'flex flex-col items-center text-center' : '[display:flow-root]'}`}>
          <RenderNodes nodes={node.children} variant={variant} />
        </div>
      )
    case 'text':
      return node.md.includes('{{') ? (
        <GapLines md={node.md} />
      ) : (
        <Prose md={node.md} variant={variant} />
      )
    case 'image':
      return <ImagePlaceholder prompt={node.prompt} size={node.size} align={node.align} />
    case 'images':
      return (
        <div className="flex flex-wrap gap-2.5">
          {node.prompts.map((prompt, i) => (
            <Illustration key={i} prompt={prompt} index={i + 1} aspect="square" className="aspect-square w-28 rounded-xl sm:w-32" />
          ))}
        </div>
      )
    case 'bubble':
      return <Bubble speaker={node.speaker} md={node.md} />
    case 'box':
      return <Box md={node.md} />
    case 'bank':
      return (
        <div className="flex flex-wrap gap-2 rounded-xl bg-brand-50/80 p-3.5 ring-1 ring-brand-200/80">
          {node.items.map((w, i) => (
            <span key={i} className="rounded-lg bg-white px-2.5 py-1 text-sm font-medium text-neutral-800 shadow-sm ring-1 ring-neutral-200">
              {w}
            </span>
          ))}
        </div>
      )
    case 'write':
      return node.mode === 'area' ? (
        <WriteArea prompt={node.prompt} sample={node.sample} rows={node.rows} />
      ) : (
        <WriteLines prompts={node.prompts} sample={node.sample} lines={node.lines} />
      )
    case 'timeline':
      return <Timeline raw={node.raw} />
    case 'match':
      return <MatchConnect block={node} />
    case 'sort':
      return <SortBuckets block={node} />
    case 'highlight':
      return <HighlightText block={node} />
    case 'gloss':
      return <GlossText block={node} />
    case 'fork':
      return <Fork block={node} />
  }
}

function Bubble({ speaker, md }: { speaker?: string; md: string }) {
  return (
    <div className="[display:flow-root]">
      {speaker && <div className="mb-1 text-xs font-semibold uppercase tracking-wide text-neutral-500">{speaker}</div>}
      <div className={`relative w-fit max-w-full rounded-2xl rounded-bl-md px-4 py-2.5 text-sm leading-relaxed text-neutral-700 ${ACCENT}`}>
        {md.includes('{{') ? <ItemText text={md} lineIdx={0} /> : <Inline text={md} />}
        <span className="absolute -bottom-1.5 left-6 h-3 w-3 rotate-45 bg-brand-50" />
      </div>
    </div>
  )
}

function Avatar({ name }: { name: string }) {
  const hue = [...name].reduce((h, c) => (h * 31 + c.charCodeAt(0)) % 360, 7)
  const initials = name
    .trim()
    .split(/\s+/)
    .map((w) => w[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
  return (
    <div
      className="grid h-8 w-8 shrink-0 select-none place-items-center rounded-full text-xs font-bold text-white shadow-sm"
      style={{ backgroundColor: `hsl(${hue} 60% 52%)` }}
    >
      {initials}
    </div>
  )
}

function Conversation({ bubbles }: { bubbles: { speaker?: string; md: string }[] }) {
  const speakers = [...new Set(bubbles.map((b) => b.speaker).filter(Boolean) as string[])]
  return (
    <div className="[display:flow-root] space-y-2">
      {bubbles.map((b, i) => {
        const right = !!b.speaker && speakers.indexOf(b.speaker) % 2 === 1
        const firstOfRun = i === 0 || bubbles[i - 1]?.speaker !== b.speaker
        const hasGaps = b.md.includes('{{')
        return (
          <div key={i} className={`flex items-end gap-2 ${right ? 'flex-row-reverse' : ''}`}>
            <div className="w-8 shrink-0">{firstOfRun && b.speaker ? <Avatar name={b.speaker} /> : null}</div>
            <div className={`flex max-w-[85%] flex-col sm:max-w-[78%] ${right ? 'items-end' : 'items-start'}`}>
              {firstOfRun && b.speaker && <span className="mb-0.5 px-1 text-xs font-semibold text-neutral-500">{b.speaker}</span>}
              <div
                className={`w-fit rounded-2xl px-3.5 py-2 text-sm shadow-sm ${
                  right && !hasGaps ? 'rounded-br-sm bg-brand-600 text-white' : `${right ? 'rounded-br-sm' : 'rounded-bl-sm'} bg-white text-neutral-800 ring-1 ring-neutral-200`
                }`}
              >
                {hasGaps ? <ItemText text={b.md} lineIdx={i} /> : <Inline text={b.md} />}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}

function Box({ md }: { md: string }) {
  return (
    <div className="[display:flow-root] rounded-xl border border-brand-200/70 bg-white px-4 py-3 text-sm leading-relaxed text-neutral-700 shadow-sm">
      <Inline text={md} />
    </div>
  )
}
