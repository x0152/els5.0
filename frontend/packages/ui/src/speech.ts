export const VOICES = ['Bella', 'Jasper', 'Luna', 'Bruno', 'Rosie', 'Hugo', 'Kiki', 'Leo']

const cache = new Map<string, string>()
let current: HTMLAudioElement | null = null
let seq = 0

// Deterministic per text, so replaying the same sentence keeps the same voice.
function pickVoice(text: string) {
  let h = 0
  for (let i = 0; i < text.length; i++) h = (h * 31 + text.charCodeAt(i)) | 0
  return VOICES[Math.abs(h) % VOICES.length]!
}

function stop() {
  current?.pause()
  current = null
  if ('speechSynthesis' in window) speechSynthesis.cancel()
}

async function fetchAudio(text: string, voice: string, speed: number): Promise<string> {
  const key = `${voice}|${speed}|${text}`
  const cached = cache.get(key)
  if (cached) return cached
  const token = localStorage.getItem('els.auth.token')
  const res = await fetch('/api/v1/speech/tts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: JSON.stringify({ text, voice, speed }),
  })
  if (!res.ok) throw new Error(`tts ${res.status}`)
  const { data } = await res.json()
  const bytes = Uint8Array.from(atob(data.audio_base64), (c) => c.charCodeAt(0))
  const url = URL.createObjectURL(new Blob([bytes], { type: 'audio/wav' }))
  cache.set(key, url)
  return url
}

function speakNative(text: string, opts?: { rate?: number; onEnd?: () => void }) {
  if (!('speechSynthesis' in window)) return
  const utterance = new SpeechSynthesisUtterance(text)
  utterance.lang = 'en-US'
  if (opts?.rate) utterance.rate = opts.rate
  if (opts?.onEnd) utterance.onend = opts.onEnd
  speechSynthesis.cancel()
  speechSynthesis.speak(utterance)
}

export function speak(text: string, opts?: { rate?: number; voice?: string; onEnd?: () => void }) {
  const id = ++seq
  stop()
  fetchAudio(text, opts?.voice ?? pickVoice(text), opts?.rate ?? 1)
    .then((url) => {
      if (id !== seq) return
      const audio = new Audio(url)
      current = audio
      if (opts?.onEnd) audio.onended = opts.onEnd
      return audio.play()
    })
    .catch(() => {
      if (id === seq) speakNative(text, opts)
    })
}
