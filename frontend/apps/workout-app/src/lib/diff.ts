export function tokenize(text: string): string[] {
  return (text.toLowerCase().match(/[a-z][a-z''-]*/g) ?? []).map((w) => w.replace(/^'+|'+$/g, ''))
}

export function alignWords(reference: string[], attempt: string[]): boolean[] {
  const n = reference.length
  const m = attempt.length
  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array<number>(m + 1).fill(0))
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      dp[i]![j] = reference[i] === attempt[j] ? dp[i + 1]![j + 1]! + 1 : Math.max(dp[i + 1]![j]!, dp[i]![j + 1]!)
    }
  }
  const heard = new Array<boolean>(n).fill(false)
  let i = 0
  let j = 0
  while (i < n && j < m) {
    if (reference[i] === attempt[j]) {
      heard[i] = true
      i++
      j++
    } else if (dp[i + 1]![j]! >= dp[i]![j + 1]!) i++
    else j++
  }
  return heard
}

export function accuracy(reference: string, attempt: string): { heard: boolean[]; words: string[]; percent: number } {
  const words = tokenize(reference)
  const heard = alignWords(words, tokenize(attempt))
  const hit = heard.filter(Boolean).length
  return { heard, words, percent: words.length ? Math.round((hit / words.length) * 100) : 0 }
}
