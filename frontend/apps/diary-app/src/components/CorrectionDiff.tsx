import type { ReactNode } from 'react'
import type { Correction } from '../lib/types'

export function CorrectionDiff({ items, revealed = true }: { items: Correction[]; revealed?: boolean }) {
  const sentence = items[0]?.sentence ?? ''
  if (!revealed) return <span>{sentence}</span>

  const found = items
    .map((it) => ({ ...it, i: sentence.toLowerCase().indexOf(it.fragment.toLowerCase()) }))
    .filter((it) => it.i >= 0)
    .sort((a, b) => a.i - b.i)

  const parts: ReactNode[] = []
  let pos = 0
  for (const it of found) {
    if (it.i < pos) continue
    parts.push(sentence.slice(pos, it.i))
    parts.push(
      <span key={it.i} className="rounded bg-red-100 px-1 text-red-700 line-through decoration-2">
        {sentence.slice(it.i, it.i + it.fragment.length)}
      </span>,
      ' ',
      <span key={`${it.i}-fix`} className="rounded bg-emerald-100 px-1 font-medium text-emerald-700">
        {it.correction}
      </span>,
    )
    pos = it.i + it.fragment.length
  }
  parts.push(sentence.slice(pos))
  return <span>{parts}</span>
}
