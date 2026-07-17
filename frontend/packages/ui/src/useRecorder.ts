import { useCallback, useEffect, useRef, useState } from 'react'

export type RecorderState = 'idle' | 'recording' | 'unsupported'

export function useRecorder(onStop: (blob: Blob) => void) {
  const [state, setState] = useState<RecorderState>('idle')
  const [elapsed, setElapsed] = useState(0)
  const [blob, setBlob] = useState<Blob | null>(null)
  const recorderRef = useRef<MediaRecorder | null>(null)
  const timerRef = useRef<number | undefined>(undefined)
  const audioRef = useRef<HTMLAudioElement | null>(null)
  const urlRef = useRef<string | null>(null)
  const onStopRef = useRef(onStop)
  useEffect(() => {
    onStopRef.current = onStop
  }, [onStop])

  const releasePlayback = useCallback(() => {
    audioRef.current?.pause()
    audioRef.current = null
    if (urlRef.current) {
      URL.revokeObjectURL(urlRef.current)
      urlRef.current = null
    }
  }, [])

  const stop = useCallback(() => {
    recorderRef.current?.stop()
  }, [])

  const play = useCallback(() => {
    if (!blob) return
    releasePlayback()
    const url = URL.createObjectURL(blob)
    urlRef.current = url
    const audio = new Audio(url)
    audioRef.current = audio
    audio.onended = releasePlayback
    void audio.play().catch(releasePlayback)
  }, [blob, releasePlayback])

  const start = useCallback(async () => {
    if (!navigator.mediaDevices?.getUserMedia) {
      setState('unsupported')
      return
    }
    try {
      releasePlayback()
      setBlob(null)
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mimeType = MediaRecorder.isTypeSupported('audio/webm') ? 'audio/webm' : ''
      const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined)
      const chunks: Blob[] = []
      recorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunks.push(e.data)
      }
      recorder.onstop = () => {
        stream.getTracks().forEach((t) => t.stop())
        window.clearInterval(timerRef.current)
        setState('idle')
        const next = new Blob(chunks, { type: recorder.mimeType || 'audio/webm' })
        setBlob(next)
        onStopRef.current(next)
      }
      recorderRef.current = recorder
      recorder.start()
      setElapsed(0)
      timerRef.current = window.setInterval(() => setElapsed((s) => s + 1), 1000)
      setState('recording')
    } catch {
      setState('unsupported')
    }
  }, [releasePlayback])

  const clear = useCallback(() => {
    releasePlayback()
    setBlob(null)
  }, [releasePlayback])

  useEffect(
    () => () => {
      window.clearInterval(timerRef.current)
      releasePlayback()
      if (recorderRef.current?.state === 'recording') recorderRef.current.stop()
    },
    [releasePlayback],
  )

  return { state, elapsed, blob, start, stop, play, clear }
}
