import type { Correction } from '../lib/types'

export function CorrectionDiff({ item, revealed = true }: { item: Correction; revealed?: boolean }) {
  const i = item.sentence.toLowerCase().indexOf(item.fragment.toLowerCase())
  if (!revealed || i < 0) {
    return <span>{item.sentence}</span>
  }
  return (
    <span>
      {item.sentence.slice(0, i)}
      <span className="rounded bg-red-100 px-1 text-red-700 line-through decoration-2">
        {item.sentence.slice(i, i + item.fragment.length)}
      </span>{' '}
      <span className="rounded bg-emerald-100 px-1 font-medium text-emerald-700">{item.correction}</span>
      {item.sentence.slice(i + item.fragment.length)}
    </span>
  )
}
