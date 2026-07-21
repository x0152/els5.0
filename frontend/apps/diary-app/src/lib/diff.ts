export type DiffToken = { text: string; kind: 'same' | 'removed' | 'added' }

export function diffWords(a: string, b: string): DiffToken[] {
  const aw = a.split(/\s+/).filter(Boolean)
  const bw = b.split(/\s+/).filter(Boolean)
  const n = aw.length
  const m = bw.length
  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array<number>(m + 1).fill(0))
  for (let i = n - 1; i >= 0; i--)
    for (let j = m - 1; j >= 0; j--)
      dp[i]![j] = aw[i] === bw[j] ? dp[i + 1]![j + 1]! + 1 : Math.max(dp[i + 1]![j]!, dp[i]![j + 1]!)

  const out: DiffToken[] = []
  let i = 0
  let j = 0
  while (i < n && j < m) {
    if (aw[i] === bw[j]) {
      out.push({ text: aw[i]!, kind: 'same' })
      i++
      j++
    } else if (dp[i + 1]![j]! >= dp[i]![j + 1]!) {
      out.push({ text: aw[i]!, kind: 'removed' })
      i++
    } else {
      out.push({ text: bw[j]!, kind: 'added' })
      j++
    }
  }
  while (i < n) out.push({ text: aw[i++]!, kind: 'removed' })
  while (j < m) out.push({ text: bw[j++]!, kind: 'added' })
  return out
}
