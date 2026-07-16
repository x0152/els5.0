import { useCallback, useEffect, useRef, useState } from 'react'

export type RecorderState = 'idle' | 'recording' | 'unsupported'

export function useRecorder(onStop: (blob: Blob) => void) {
  const [state, setState] = useState<RecorderState>('idle')
  const [elapsed, setElapsed] = useState(0)
  const recorderRef = useRef<MediaRecorder | null>(null)
  const timerRef = useRef<number | undefined>(undefined)
  const onStopRef = useRef(onStop)
  useEffect(() => {
    onStopRef.current = onStop
  }, [onStop])

  const stop = useCallback(() => {
    recorderRef.current?.stop()
  }, [])

  const start = useCallback(async () => {
    if (!navigator.mediaDevices?.getUserMedia) {
      setState('unsupported')
      return
    }
    try {
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
        onStopRef.current(new Blob(chunks, { type: recorder.mimeType || 'audio/webm' }))
      }
      recorderRef.current = recorder
      recorder.start()
      setElapsed(0)
      timerRef.current = window.setInterval(() => setElapsed((s) => s + 1), 1000)
      setState('recording')
    } catch {
      setState('unsupported')
    }
  }, [])

  useEffect(
    () => () => {
      window.clearInterval(timerRef.current)
      if (recorderRef.current?.state === 'recording') recorderRef.current.stop()
    },
    [],
  )

  return { state, elapsed, start, stop }
}
