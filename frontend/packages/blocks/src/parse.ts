export type Size = 'sm' | 'md' | 'lg' | 'full'

export type Node =
  | { t: 'grid'; cols: number; divider: boolean; children: Node[] }
  | { t: 'stack'; center: boolean; children: Node[] }
  | { t: 'text'; md: string }
  | { t: 'image'; prompt: string; size: Size; align?: 'left' | 'right' }
  | { t: 'images'; prompts: string[] }
  | { t: 'bubble'; speaker?: string; md: string }
  | { t: 'box'; md: string }
  | { t: 'bank'; items: string[] }
  | { t: 'write'; mode: 'area'; prompt: string; sample?: string; rows: number }
  | { t: 'write'; mode: 'lines'; prompts: string[]; sample?: string; lines: number }
  | { t: 'timeline'; raw: string }
  | { t: 'match'; left: { n: string; text: string; answer: string }[]; right: { l: string; text: string }[] }
  | { t: 'sort'; cats: { name: string; items: string[] }[] }
  | { t: 'highlight'; lines: string[] }
  | { t: 'gloss'; raw: string }
  | { t: 'fork'; stem: string; branches: string[] }

export type Gap = (
  | { type: 'text'; answers: string[] }
  | { type: 'choice'; choices: string[]; answers: string[] }
) & { ordinal?: number; fill?: string }

const GAP_RE = /\{\{([^}]*)\}\}/g

// Tags each gap with its source-order ordinal (invisible marker inside the braces),
// so a fill can later be written back into the exact gap of the original text.
export function indexGaps(text: string): string {
  let n = 0
  return text.replace(GAP_RE, (_, inside: string) => `{{${n++}\u2063${inside}}}`)
}

// Writes the user's fill into the ordinal-th gap of the original (unindexed) text: {{spec||fill}}.
export function fillGap(text: string, ordinal: number, answer: string): string {
  let n = 0
  return text.replace(GAP_RE, (m, inside: string) => {
    if (n++ !== ordinal) return m
    const cut = inside.indexOf('||')
    return `{{${cut === -1 ? inside : inside.slice(0, cut)}||${answer}}}`
  })
}

export function parseGap(inside: string): Gap {
  let ordinal: number | undefined
  const om = /^(\d+)\u2063([\s\S]*)$/.exec(inside)
  if (om) {
    ordinal = Number(om[1])
    inside = om[2] ?? ''
  }
  const cut = inside.indexOf('||')
  const spec = cut === -1 ? inside : inside.slice(0, cut)
  const fill = cut === -1 ? undefined : inside.slice(cut + 2).trim() || undefined
  const parts = spec.split('|').map((p) => p.trim())
  const starred = parts.filter((p) => p.startsWith('*'))
  if (starred.length > 0) {
    return { type: 'choice', choices: parts.map((p) => p.replace(/^\*/, '')), answers: starred.map((p) => p.slice(1)), ordinal, fill }
  }
  return { type: 'text', answers: parts, ordinal, fill }
}

export function gapPrompt(text: string): string {
  return text
    .replace(/\{\{[^}]*\}\}/g, '___')
    .replace(/\*\*|__|[*_]/g, '')
    .replace(/^\d+\.\s*/, '')
    .replace(/\s+/g, ' ')
    .trim()
}

export function parseBlocks(src: string): Node[] {
  return parseLines(src.replace(/\s+$/, '').split('\n'))
}

function parseLines(lines: string[]): Node[] {
  const nodes: Node[] = []
  let text: string[] = []
  const flush = () => {
    const md = text.join('\n').trim()
    if (md) nodes.push({ t: 'text', md })
    text = []
  }

  for (let i = 0; i < lines.length; i++) {
    const trimmed = (lines[i] ?? '').trim()

    const open = /^:::(\w+)(.*)$/.exec(trimmed)
    if (open) {
      flush()
      const inner: string[] = []
      let depth = 1
      i++
      for (; i < lines.length; i++) {
        const t = (lines[i] ?? '').trim()
        if (/^:::\w/.test(t)) depth++
        else if (t === ':::') {
          depth--
          if (depth === 0) break
        }
        inner.push(lines[i] ?? '')
      }
      nodes.push(container(open[1] ?? '', (open[2] ?? '').trim(), parseLines(inner)))
      continue
    }

    const fence = /^~~~(\w[\w-]*)(.*)$/.exec(trimmed)
    if (fence) {
      flush()
      const content: string[] = []
      i++
      for (; i < lines.length && (lines[i] ?? '').trim() !== '~~~'; i++) content.push(lines[i] ?? '')
      nodes.push(leaf(fence[1] ?? '', (fence[2] ?? '').trim(), content))
      continue
    }

    text.push(lines[i] ?? '')
  }
  flush()
  return nodes
}

function container(type: string, args: string, children: Node[]): Node {
  if (type === 'grid') {
    return { t: 'grid', cols: Number(/cols=(\d+)/.exec(args)?.[1]) || 2, divider: /\bdivider\b/.test(args), children }
  }
  return { t: 'stack', center: /\bcenter\b/.test(args), children }
}

function leaf(type: string, args: string, content: string[]): Node {
  const tokens = args.split(/\s+/).filter(Boolean)
  if (type === 'image') {
    const align = tokens.includes('right') ? 'right' : tokens.includes('left') ? 'left' : undefined
    const size = (['sm', 'md', 'lg', 'full'] as const).find((s) => tokens.includes(s)) ?? 'md'
    return { t: 'image', prompt: content.join(' ').trim(), size, align }
  }
  if (type === 'images') return { t: 'images', prompts: content.map((l) => l.trim()).filter(Boolean) }
  if (type === 'bubble') {
    const speaker = content[0]?.trim().startsWith('@') ? content[0].trim().slice(1).trim() : undefined
    const md = (speaker ? content.slice(1) : content).join(' ').trim()
    return { t: 'bubble', speaker, md }
  }
  if (type === 'box') return { t: 'box', md: content.join('\n').trim() }
  if (type === 'bank') return { t: 'bank', items: content.join(',').split(/[,\n]/).map((s) => s.trim()).filter(Boolean) }
  if (type === 'write') {
    const sample = content.filter((l) => l.trim().startsWith('>')).map((l) => l.trim().replace(/^>\s?/, '')).join(' ').trim() || undefined
    const rest = content.filter((l) => l.trim() && !l.trim().startsWith('>')).map((l) => l.trim())
    const lineMatch = /lines=(\d+)/.exec(args)
    if (lineMatch) {
      const prompts = rest.map((l) => l.replace(/^\d+\.\s*/, ''))
      return { t: 'write', mode: 'lines', prompts, sample, lines: Number(lineMatch[1]) || prompts.length || 4 }
    }
    return { t: 'write', mode: 'area', prompt: rest.join(' '), sample, rows: Number(/rows=(\d+)/.exec(args)?.[1]) || 4 }
  }
  if (type === 'timeline') return { t: 'timeline', raw: content.join('\n') }
  if (type === 'sort') {
    const cats = content
      .map((l) => /^([^:]+):\s*(.+)$/.exec(l.trim()))
      .filter((m): m is RegExpExecArray => !!m)
      .map((m) => ({ name: (m[1] ?? '').trim(), items: (m[2] ?? '').split(',').map((s) => s.trim()).filter(Boolean) }))
    return { t: 'sort', cats }
  }
  if (type === 'highlight') return { t: 'highlight', lines: content.map((l) => l.trim()).filter(Boolean) }
  if (type === 'gloss') return { t: 'gloss', raw: content.join('\n').trim() }
  if (type === 'fork') {
    // Several forks may share one block: groups split on blank lines, or on a
    // stem-looking line (no commas, no gaps) after a previous stem + branches.
    const groups: string[][] = []
    let cur: string[] = []
    for (const raw of content) {
      const l = raw.trim()
      if (!l) {
        if (cur.length) groups.push(cur)
        cur = []
        continue
      }
      const prev = cur[cur.length - 1] ?? ''
      if (cur.length >= 2 && !l.includes(',') && !l.includes('{{') && prev.includes(',')) {
        groups.push(cur)
        cur = []
      }
      cur.push(l)
    }
    if (cur.length) groups.push(cur)
    const forks = groups.map((g) => ({
      t: 'fork' as const,
      stem: g[0] ?? '',
      branches: g.slice(1).join('\n').split(/[,\n]/).map((s) => s.trim()).filter(Boolean),
    }))
    if (forks.length > 1) return { t: 'stack', center: false, children: forks }
    return forks[0] ?? { t: 'fork', stem: '', branches: [] }
  }
  if (type === 'match') {
    const sep = content.findIndex((l) => l.trim() === '---')
    const leftLines = sep === -1 ? content : content.slice(0, sep)
    const rightLines = sep === -1 ? [] : content.slice(sep + 1)
    const left = leftLines
      .map((l) => /^(\d+)\.\s+(.*?)\s*::\s*(\S+)\s*$/.exec(l.trim()))
      .filter((m): m is RegExpExecArray => !!m)
      .map((m) => ({ n: m[1] ?? '', text: m[2] ?? '', answer: m[3] ?? '' }))
    const right = rightLines
      .map((l) => /^([a-z])\.\s+(.*)$/.exec(l.trim()))
      .filter((m): m is RegExpExecArray => !!m)
      .map((m) => ({ l: m[1] ?? '', text: m[2] ?? '' }))
    return { t: 'match', left, right }
  }
  return { t: 'text', md: content.join('\n') }
}

export type TheorySection = { letter: string; title?: string; nodes: Node[] }

export function parseTheory(md: string): TheorySection[] {
  const trimmed = md.replace(/^\s+/, '')
  const firstHeading = trimmed.search(/^##\s+/m)
  const preamble = (firstHeading === -1 ? trimmed : trimmed.slice(0, firstHeading)).trim()
  const sections: TheorySection[] = []
  if (preamble) sections.push({ letter: '', nodes: parseBlocks(preamble) })
  ;(firstHeading === -1 ? '' : trimmed.slice(firstHeading))
    .split(/^##\s+/m)
    .map((s) => s.replace(/\s+$/, ''))
    .filter((s) => s.trim())
    .forEach((chunk) => {
      const nl = chunk.indexOf('\n')
      const head = (nl === -1 ? chunk : chunk.slice(0, nl)).trim()
      const body = nl === -1 ? '' : chunk.slice(nl + 1)
      const letter = head.split(/\s*—\s*/)[0]?.trim() ?? head
      const title = head.includes('—') ? head.split(/\s*—\s*/).slice(1).join(' — ') : undefined
      sections.push({ letter, title, nodes: parseBlocks(body) })
    })
  return sections
}

export type Exercise = { id: string; section?: string; instruction: string; lead?: string; nodes: Node[] }

export function parseExercises(md: string): Exercise[] {
  return md
    .split(/^##\s+/m)
    .map((s) => s.replace(/\s+$/, ''))
    .filter((s) => s.trim())
    .map((raw) => {
      let lead: string | undefined
      const chunk = raw.replace(/^~~~lead\n([\s\S]*?)\n~~~$/m, (_, text: string) => {
        lead = text.trim()
        return ''
      })
      const nl = chunk.indexOf('\n')
      const head = (nl === -1 ? chunk : chunk.slice(0, nl)).trim()
      const bodyLines = nl === -1 ? [] : chunk.slice(nl + 1).split('\n')
      const [id = '', section] = head.split(/\s*(?:→|->)\s*/)

      const instr: string[] = []
      let k = 0
      for (; k < bodyLines.length; k++) {
        const t = (bodyLines[k] ?? '').trim()
        if (!t) {
          if (instr.length) {
            k++
            break
          }
          continue
        }
        if (/^:::/.test(t) || /^~~~/.test(t) || t.includes('{{') || /^\d+\.\s/.test(t) || /^-\s/.test(t)) break
        instr.push(t)
      }
      return { id: id.trim(), section: section?.trim(), instruction: instr.join(' '), lead, nodes: parseBlocks(bodyLines.slice(k).join('\n')) }
    })
}
