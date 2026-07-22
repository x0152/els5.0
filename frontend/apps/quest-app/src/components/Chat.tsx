import { useEffect, useMemo, useState } from 'react'
import { Check, ImageIcon, Loader2, Maximize2, RotateCcw, Trophy, TriangleAlert } from 'lucide-react'
import { SpeakButton as UiSpeakButton, cn } from '@els/ui'
import { avatarKey } from '../lib/helpers.ts'
import type { Character, Mission, PartialWorld } from '../lib/types.ts'
import { useRegenerateImage } from '../store/missions.ts'
import { useOverlay } from './Overlay.tsx'
import { RegenButton } from './RegenButton.tsx'

export type PlayerTurnState = 'checking' | 'ok'

interface Props {
  mission: Mission
  optimisticText: string | null
  playerAvatar?: string
  optimisticState?: PlayerTurnState
}

function findCharacter(mission: Mission, name: string): Character | undefined {
  return mission.characters?.find((c) => avatarKey(c.name) === avatarKey(name))
}

export function Chat({ mission, optimisticText, playerAvatar, optimisticState }: Props) {
  const history = mission.history ?? []

  return (
    <div className="space-y-4">
      {history.map((turn, i) => {
        const key = `${i}-${turn.speaker}`
        if (turn.speaker === 'system') return <SystemLine key={key} mission={mission} text={turn.text} scene={turn.scene} />
        if (turn.speaker === 'narrator') return <Narration key={key} text={turn.text} voice={turn.voice || mission.narratorVoice} />
        if (turn.speaker === 'player') return <PlayerBubble key={key} text={turn.text} avatar={playerAvatar} />
        return <NpcBubble key={key} mission={mission} name={turn.speaker} text={turn.text} voice={turn.voice} />
      })}

      {optimisticText && <PlayerBubble text={optimisticText} avatar={playerAvatar} state={optimisticState} />}
    </div>
  )
}

// StreamingText — “typewriter”: the LLM delivers text in bursts between
// polls, so the arrived buffer appears word-by-word at a steady
// pace, speeding up when far behind the stream.
function StreamingText({ text }: { text: string }) {
  const words = useMemo(() => text.split(/(\s+)/), [text])
  const [visible, setVisible] = useState(0)
  const done = visible >= words.length

  useEffect(() => {
    if (done) return
    const timer = setInterval(() => {
      setVisible((v) => {
        const backlog = words.length - v
        return Math.min(v + 1 + Math.floor(backlog / 25), words.length)
      })
    }, 30)
    return () => clearInterval(timer)
  }, [words.length, done])

  return (
    <>
      {words.slice(0, visible).map((w, i) => (
        <span key={i} className="animate-word-in inline-block whitespace-pre-wrap">
          {w}
        </span>
      ))}
    </>
  )
}

export function Characters({ mission }: { mission: Mission }) {
  const { openImage, openCharacter } = useOverlay()
  const regen = useRegenerateImage(mission.id)
  const chars = mission.characters ?? []
  if (chars.length === 0) return null
  return (
    <div className="rounded-2xl bg-white p-4 ring-1 ring-neutral-200">
      <div className="mb-3 text-xs font-semibold uppercase tracking-wide text-neutral-400">Characters</div>
      <div className="flex flex-wrap gap-x-4 gap-y-3 lg:flex-col">
        {chars.map((c) => {
          const key = avatarKey(c.name)
          const url = mission.characterAvatars?.[key]
          const generating = mission.characterAvatarStatus?.[key] === 'generating'
          return (
            <div
              key={c.name}
              onClick={() => openCharacter(c, mission.id)}
              className="-m-1 flex max-w-full cursor-pointer items-center gap-2.5 rounded-lg p-1 text-left transition-colors hover:bg-neutral-50"
            >
              <Avatar
                url={url}
                name={c.name}
                size={40}
                generating={generating}
                onClick={url ? (e) => { e.stopPropagation(); openImage(url, c.name) } : undefined}
                onRegenerate={() => regen.mutate({ kind: 'avatar', key: c.name })}
              />
              <div className="min-w-0 flex-1">
                <div className="truncate text-sm font-medium text-neutral-900">{c.name}</div>
                <div className="line-clamp-2 text-xs leading-snug text-neutral-500">{c.role}</div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export function DraftWorld({ mission, partial, step }: { mission: Mission; partial: PartialWorld; step?: string }) {
  const narration = partial.narration?.trim()
  const allResponses = (partial.responses ?? []).filter((r) => r.name?.trim())
  if (!narration && allResponses.length === 0) return null

  const streaming = step === 'world'
  const last = allResponses[allResponses.length - 1]
  // The next speaker name is already known, but text has not started —
  // show a named typing bubble instead of an empty line.
  const lastIsPending = streaming && !!last && !last.done && !(last.text ?? '').trim()
  const responses = lastIsPending ? allResponses.slice(0, -1) : allResponses
  const narrationTyping = streaming && !partial.narrationDone && allResponses.length === 0
  const lastTyping = streaming && !lastIsPending && !!last && !last.done
  // Nothing is typing right now (stream pause or evaluation phase) —
  // quiet dots show that the turn is still being processed.
  const idle = !narrationTyping && !lastTyping && !lastIsPending

  return (
    <div className="space-y-4">
      {narration && <Narration text={narration} streaming={narrationTyping} animate />}
      {responses.map((r, i) => (
        <NpcBubble
          key={`${i}-${r.name}`}
          mission={mission}
          name={r.name}
          text={r.text ?? ''}
          streaming={streaming && i === responses.length - 1 && !r.done}
          animate
        />
      ))}
      {lastIsPending && <TypingBubble mission={mission} name={last.name} />}
      {idle && <WorkingDots />}
    </div>
  )
}

function StreamingCursor() {
  return <span className="ml-0.5 inline-block h-3.5 w-0.5 animate-pulse rounded bg-neutral-400 align-middle" />
}

export function TypingBubble({ mission, name }: { mission?: Mission; name?: string }) {
  const url = mission && name ? mission.characterAvatars?.[avatarKey(name)] : undefined
  return (
    <div className="animate-pop-in flex items-center gap-2">
      {name ? (
        <Avatar url={url} name={name} size={36} />
      ) : (
        <div className="h-9 w-9 shrink-0 animate-pulse rounded-full bg-neutral-200" />
      )}
      <div>
        {name && <div className="mb-0.5 text-xs font-medium text-neutral-500">{name} is typing…</div>}
        <div className="w-fit rounded-2xl rounded-tl-sm bg-white px-4 py-3.5 ring-1 ring-neutral-200">
          <span className="flex items-center gap-1">
            {[0, 160, 320].map((delay) => (
              <span
                key={delay}
                className="h-1.5 w-1.5 animate-bounce rounded-full bg-neutral-400"
                style={{ animationDelay: `${delay}ms` }}
              />
            ))}
          </span>
        </div>
      </div>
    </div>
  )
}

export function SystemLine({ mission, text, scene }: { mission: Mission; text: string; scene?: number }) {
  const sceneMatch = text.match(/Scene\s+(\d+)/i)
  if (sceneMatch) {
    const stage = Number(sceneMatch[1]) - 1
    return (
      <div className="animate-rise-in space-y-3 pt-2">
        <div className="flex items-center gap-3">
          <div className="h-px flex-1 bg-neutral-200" />
          <span className="text-xs font-semibold uppercase tracking-widest text-neutral-400">Scene {sceneMatch[1]}</span>
          <div className="h-px flex-1 bg-neutral-200" />
        </div>
        <SceneImage mission={mission} stage={stage} />
      </div>
    )
  }

  const complete = /mission complete/i.test(text)
  const summary = mission.scenes?.find((s) => s.stage === scene)?.summary?.trim()
  return (
    <div className="animate-rise-in space-y-1.5 py-1 text-center">
      <span
        className={cn(
          'inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-xs font-medium',
          complete ? 'bg-amber-50 text-amber-700' : 'bg-brand-50 text-brand-700',
        )}
      >
        {complete && <Trophy className="h-3.5 w-3.5" />}
        {complete ? 'Mission complete' : 'Scene cleared'}
      </span>
      {!complete && summary && (
        <p className="mx-auto max-w-md text-xs italic leading-relaxed text-neutral-400">{summary}</p>
      )}
    </div>
  )
}

function SceneImage({ mission, stage }: { mission: Mission; stage: number }) {
  const { openImage } = useOverlay()
  const [loaded, setLoaded] = useState(false)
  const url = mission.sceneImages?.[String(stage)]
  const status = mission.sceneImageStatus?.[String(stage)]

  if (url) {
    return (
      <button
        type="button"
        onClick={() => openImage(url, `Scene ${stage + 1}`)}
        className="group relative block w-full cursor-zoom-in overflow-hidden rounded-xl ring-1 ring-neutral-200"
      >
        <img
          src={url}
          alt={`Scene ${stage + 1}`}
          onLoad={() => setLoaded(true)}
          className={cn(
            'aspect-video w-full object-cover transition-all duration-700 group-hover:scale-105',
            loaded ? 'scale-100 blur-0' : 'scale-105 blur-md',
          )}
        />
        <span className="absolute right-2 top-2 grid h-7 w-7 place-items-center rounded-full bg-black/45 text-white opacity-0 backdrop-blur-sm transition-opacity group-hover:opacity-100">
          <Maximize2 className="h-3.5 w-3.5" />
        </span>
      </button>
    )
  }
  if (status === 'generating') {
    return (
      <div className="aspect-video w-full overflow-hidden rounded-xl ring-1 ring-neutral-200">
        <Placeholder text="Drawing the scene…" busy />
      </div>
    )
  }
  if (status === 'error') {
    return (
      <div className="aspect-video w-full overflow-hidden rounded-xl ring-1 ring-rose-200">
        <div className="flex h-full w-full flex-col items-center justify-center gap-2 bg-rose-50 text-rose-600">
          <TriangleAlert className="h-5 w-5" />
          <span className="text-xs">Image failed</span>
          <RegenButton missionId={mission.id} kind="scene" itemKey={String(stage)} />
        </div>
      </div>
    )
  }
  return null
}

function SpeakButton({ text, voice, className }: { text: string; voice: string; className?: string }) {
  return (
    <UiSpeakButton
      title={`Listen (${voice})`}
      text={text}
      voice={voice}
      className={cn('p-0 text-neutral-300 hover:bg-transparent hover:text-brand-600', className)}
      iconClassName="h-3.5 w-3.5"
    />
  )
}

function Narration({ text, voice, streaming, animate }: { text: string; voice?: string; streaming?: boolean; animate?: boolean }) {
  return (
    <p className="group border-l-2 border-brand-200 pl-4 text-sm italic leading-relaxed text-neutral-600">
      {animate ? <StreamingText text={text} /> : text}
      {streaming && <StreamingCursor />}
      {voice && !streaming && <SpeakButton text={text} voice={voice} className="ml-2 inline-flex align-middle opacity-0 group-hover:opacity-100" />}
    </p>
  )
}

// Player turn state is visible on the bubble itself: while text is being checked —
// it is dimmed, a light wave runs across it, and a small caption sits below;
// checked — full color and a green checkmark.
export function PlayerBubble({ text, avatar, state }: { text: string; avatar?: string; state?: PlayerTurnState }) {
  return (
    <div className="flex justify-end gap-2">
      <div className="relative max-w-[80%]">
        <div
          className={cn(
            'relative overflow-hidden rounded-2xl rounded-br-sm px-4 py-2.5 text-sm text-white transition-colors duration-500',
            state === 'checking' ? 'bg-brand-600/60' : 'bg-brand-600',
          )}
        >
          {text}
          {state === 'checking' && (
            <span className="pointer-events-none absolute inset-y-0 left-0 w-1/3 animate-scan bg-gradient-to-r from-transparent via-white/30 to-transparent" />
          )}
        </div>
        {state === 'checking' && (
          <div className="mt-0.5 animate-pulse text-right text-[11px] text-neutral-400">checking…</div>
        )}
        {state === 'ok' && (
          <span className="animate-pop-in absolute -bottom-1 -left-1 grid h-4 w-4 place-items-center rounded-full bg-white text-brand-600 shadow ring-1 ring-brand-200">
            <Check className="h-2.5 w-2.5" strokeWidth={3} />
          </span>
        )}
      </div>
      {avatar ? (
        <img src={avatar} alt="You" className="h-9 w-9 shrink-0 rounded-full object-cover ring-1 ring-brand-200" />
      ) : (
        <div className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-brand-100 text-xs font-semibold text-brand-700">
          You
        </div>
      )}
    </div>
  )
}

const TRANSITION_PHRASES = ['The story moves on…', 'A new scene is taking shape…', 'Meanwhile…']

// useDelayedUnmount keeps the element in the DOM while the exit animation plays.
export function useDelayedUnmount(show: boolean, ms = 600) {
  const [mounted, setMounted] = useState(show)
  useEffect(() => {
    if (show) {
      setMounted(true)
      return
    }
    const timer = setTimeout(() => setMounted(false), ms)
    return () => clearTimeout(timer)
  }, [show, ms])
  return mounted
}

// SceneTransition — “next scene is generating” banner in the chat feed.
// Click (if onClick is passed) opens the cinematic transition dialog.
export function SceneTransition({ onClick }: { onClick?: () => void }) {
  const [phrase, setPhrase] = useState(0)
  useEffect(() => {
    const timer = setInterval(() => setPhrase((n) => (n + 1) % TRANSITION_PHRASES.length), 2600)
    return () => clearInterval(timer)
  }, [])

  return (
    <div
      onClick={onClick}
      className={cn(
        'animate-rise-in relative overflow-hidden rounded-2xl bg-gradient-to-br from-brand-900 via-brand-800 to-brand-900 px-6 py-8 text-center',
        onClick && 'cursor-pointer',
      )}
    >
      <span
        className="pointer-events-none absolute inset-y-0 left-0 w-1/3 animate-scan bg-gradient-to-r from-transparent via-brand-200/15 to-transparent"
        style={{ animationDuration: '3s' }}
      />
      <div key={phrase} className="animate-word-in text-sm italic tracking-wide text-brand-100/80">
        {TRANSITION_PHRASES[phrase]}
      </div>
      <div className="mt-3 flex items-center justify-center gap-1.5">
        {[0, 160, 320].map((delay) => (
          <span
            key={delay}
            className="h-1 w-1 animate-bounce rounded-full bg-brand-300/50"
            style={{ animationDelay: `${delay}ms` }}
          />
        ))}
      </div>
    </div>
  )
}

// WorkingDots — “system still working”: quiet dots without a bubble when no
// line is currently being typed.
export function WorkingDots() {
  return (
    <div className="flex items-center gap-1 pl-11">
      {[0, 160, 320].map((delay) => (
        <span
          key={delay}
          className="h-1 w-1 animate-bounce rounded-full bg-neutral-300"
          style={{ animationDelay: `${delay}ms` }}
        />
      ))}
    </div>
  )
}

function NpcBubble({ mission, name, text, voice, streaming, animate }: { mission: Mission; name: string; text: string; voice?: string; streaming?: boolean; animate?: boolean }) {
  const { openCharacter } = useOverlay()
  const url = mission.characterAvatars?.[avatarKey(name)]
  const character = findCharacter(mission, name)
  const open = character ? () => openCharacter(character, mission.id) : undefined
  const lineVoice = voice || character?.voice
  return (
    <div className="group flex gap-2">
      <Avatar url={url} name={name} size={36} onClick={open} />
      <div className="max-w-[80%]">
        <span className="mb-0.5 flex items-center gap-1.5">
          <button
            type="button"
            onClick={open}
            disabled={!open}
            className="text-xs font-medium text-neutral-500 enabled:hover:text-brand-700"
          >
            {streaming ? `${name} is typing…` : name}
          </button>
          {lineVoice && !streaming && <SpeakButton text={text} voice={lineVoice} className="opacity-0 group-hover:opacity-100" />}
        </span>
        <div className="rounded-2xl rounded-tl-sm bg-white px-4 py-2.5 text-sm text-neutral-800 ring-1 ring-neutral-200">
          {animate ? <StreamingText text={text} /> : text}
          {streaming && <StreamingCursor />}
        </div>
      </div>
    </div>
  )
}

function Avatar({
  url,
  name,
  size,
  onClick,
  generating,
  onRegenerate,
}: {
  url?: string
  name: string
  size: number
  onClick?: (e: React.MouseEvent) => void
  generating?: boolean
  onRegenerate?: (e: React.MouseEvent) => void
}) {
  const style = { width: size, height: size }
  const clickable = onClick ? 'cursor-pointer' : ''
  return (
    <div className="group/avatar relative shrink-0" style={style}>
      {url ? (
        <img
          src={url}
          alt={name}
          style={style}
          onClick={onClick}
          className={cn('rounded-full object-cover ring-1 ring-neutral-200', clickable)}
        />
      ) : (
        <div
          style={style}
          onClick={onClick}
          className={cn('grid place-items-center rounded-full bg-neutral-200 text-xs font-semibold text-neutral-600', clickable)}
        >
          {name.slice(0, 1).toUpperCase()}
        </div>
      )}
      {generating ? (
        <div className="absolute inset-0 grid place-items-center rounded-full bg-black/40">
          <Loader2 className="h-4 w-4 animate-spin text-white" />
        </div>
      ) : (
        onRegenerate && (
          <button
            type="button"
            title={url ? 'Regenerate avatar' : 'Generate avatar'}
            onClick={(e) => {
              e.stopPropagation()
              onRegenerate(e)
            }}
            className={cn(
              'absolute -bottom-1 -right-1 grid h-5 w-5 place-items-center rounded-full bg-brand-600 text-white shadow ring-2 ring-white transition hover:bg-brand-700',
              url ? 'opacity-0 group-hover/avatar:opacity-100' : 'opacity-100',
            )}
          >
            <RotateCcw className="h-2.5 w-2.5" />
          </button>
        )
      )}
    </div>
  )
}

function Placeholder({ text, busy }: { text: string; busy?: boolean }) {
  return (
    <div className="flex h-full w-full flex-col items-center justify-center gap-2 bg-neutral-100 text-neutral-400">
      {busy ? <Loader2 className="h-5 w-5 animate-spin" /> : <ImageIcon className="h-5 w-5" />}
      <span className="text-xs">{text}</span>
    </div>
  )
}
