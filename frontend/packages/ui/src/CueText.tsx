import { Fragment, type ReactNode } from 'react'

type Active = { i: boolean; b: boolean; u: boolean }

function parseLine(line: string): ReactNode[] {
  const active: Active = { i: false, b: false, u: false }
  const out: ReactNode[] = []
  line.split(/(<\/?[ibu]>)/i).forEach((part, idx) => {
    const tag = /^<(\/?)([ibu])>$/i.exec(part)
    if (tag) {
      active[tag[2]!.toLowerCase() as keyof Active] = !tag[1]
      return
    }
    if (!part) return
    let node: ReactNode = part
    if (active.i) node = <em>{node}</em>
    if (active.b) node = <strong>{node}</strong>
    if (active.u) node = <u>{node}</u>
    out.push(<Fragment key={idx}>{node}</Fragment>)
  })
  return out
}

export function CueText({ text }: { text: string }) {
  const cleaned = text.replace(/\{[^}]*\}/g, '').replace(/<\/?font[^>]*>/gi, '')
  const lines = cleaned.split(/\\N|\n/)
  return (
    <>
      {lines.map((line, i) => (
        <Fragment key={i}>
          {i > 0 && <br />}
          {parseLine(line)}
        </Fragment>
      ))}
    </>
  )
}
