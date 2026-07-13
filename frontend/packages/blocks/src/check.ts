export type CheckResult = {
  correct: boolean
  correction: string
  explanation: string
}

export type CheckFn = (input: { prompt: string; answers: string[]; answer: string }) => Promise<CheckResult>

function normalize(s: string): string {
  return s
    .toLowerCase()
    .replace(/[\u2018\u2019]/g, "'")
    .replace(/[.?!]+$/g, '')
    .replace(/\s+/g, ' ')
    .trim()
}

// Mock: local comparison with the reference answers (no backend / LLM yet).
export const mockCheckAnswer: CheckFn = async (input) => {
  await new Promise((r) => setTimeout(r, 350))
  const got = normalize(input.answer)
  const correct = input.answers.some((a) => normalize(a) === got)
  if (correct) return { correct: true, correction: '', explanation: 'Correct!' }
  return {
    correct: false,
    correction: input.answers[0] ?? '',
    explanation: 'Compare with the correct form (local check for now, no AI).',
  }
}
