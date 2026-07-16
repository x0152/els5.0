import type { SpeechComponents } from '@els/api-client'

type Schemas = SpeechComponents['schemas']

export type Assessment = Schemas['AssessOutput']
export type WordResult = Schemas['WordOutput']
export type PhonemeResult = Schemas['PhonemeOutput']
export type Feedback = Schemas['FeedbackOutput']
export type PhonemeInfo = Schemas['PhonemeInfoOutput']

export type Verdict = 'good' | 'close' | 'wrong' | 'missing'

export const VERDICT_LABELS: Record<Verdict, string> = {
  good: 'Correct',
  close: 'Close',
  wrong: 'Wrong sound',
  missing: 'Not pronounced',
}

export function buildIssues(assessment: Assessment): string[] {
  const issues: string[] = []
  for (const word of assessment.words ?? []) {
    for (const p of word.phonemes ?? []) {
      if (p.verdict === 'good') continue
      if (p.verdict === 'missing') {
        issues.push(`${word.word}: /${p.expected}/ was not pronounced`)
      } else {
        issues.push(`${word.word}: expected /${p.expected}/, heard /${p.heard ?? '?'}/`)
      }
    }
    if (word.extra?.length) {
      issues.push(`${word.word}: extra sounds [${word.extra.join(' ')}]`)
    }
  }
  return issues
}
