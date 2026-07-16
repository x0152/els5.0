// Mid-sagittal articulation diagrams (CC0, Richard Wright & Dan McCloy,
// github.com/drammock/phonetics-teaching-assets).
const img = (name: string) => new URL(`./assets/phonemes/${name}.svg`, import.meta.url).href

const IMAGES: Record<string, string> = {
  i: img('i'),
  ɪ: img('ih'),
  ɛ: img('eh'),
  æ: img('ae'),
  ɑ: img('a'),
  a: img('a'),
  ʊ: img('uh'),
  u: img('u'),
  p: img('p'),
  b: img('b'),
  t: img('t'),
  d: img('d'),
  k: img('k'),
  ɡ: img('g'),
  f: img('f'),
  v: img('v'),
  θ: img('theta'),
  ð: img('eth'),
  s: img('s'),
  z: img('z'),
  ʃ: img('esh'),
  ʒ: img('ezh'),
  m: img('m'),
  n: img('n'),
  ŋ: img('ng'),
  ɹ: img('r'),
}

// Common variants -> canonical guide symbols (length marks are stripped first).
const CANONICAL: Record<string, string> = {
  'əʊ': 'oʊ',
  'ə˞': 'ɚ',
  ɜ: 'ɝ',
  ɒ: 'ɑ',
  e: 'ɛ',
  g: 'ɡ',
  ʧ: 'tʃ',
  ʤ: 'dʒ',
}

export function canonicalPhoneme(symbol: string): string {
  const s = symbol.replace(/ː/g, '')
  return CANONICAL[s] ?? s
}

export function phonemeImage(symbol: string): string | undefined {
  return IMAGES[canonicalPhoneme(symbol)]
}

const MULTI = ['tʃ', 'dʒ', 'eɪ', 'aɪ', 'ɔɪ', 'aʊ', 'oʊ', 'əʊ', 'iː', 'uː', 'ɑː', 'ɔː', 'ɜː', 'ə˞', 'ʈʂ', 'ʧ', 'ʤ']
const SINGLE = 'iɪɛæɑaɒɔʊuʌəɚɝɜeopbtdkɡgfvθðszʃʒhmnŋlɹrjwɾʔx'

export interface PhonemeGuideInfo {
  symbol: string
  kind?: string
  examples?: string
  description?: string
  pitfall?: string
}

export interface IpaToken {
  text: string
  symbol?: string
}

// Greedy longest-match split of an IPA string into clickable phonemes.
export function splitIpa(ipa: string): IpaToken[] {
  const tokens: IpaToken[] = []
  let i = 0
  while (i < ipa.length) {
    const two = ipa.slice(i, i + 2)
    if (MULTI.includes(two)) {
      tokens.push({ text: two, symbol: canonicalPhoneme(two) })
      i += 2
      continue
    }
    const one = ipa.charAt(i)
    if (SINGLE.includes(one)) {
      tokens.push({ text: one, symbol: canonicalPhoneme(one) })
    } else {
      tokens.push({ text: one })
    }
    i += 1
  }
  return tokens
}
