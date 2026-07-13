import { useEffect, useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Flame, Loader2, RotateCcw, Send, Sparkles, Target, TriangleAlert, X } from 'lucide-react'
import { Button, ErrorState, LoadingState, useAgentView } from '@els/ui'
import { Chat, Characters, DraftWorld, SceneTransition, TypingBubble, useDelayedUnmount } from '../components/Chat.tsx'
import { SceneTransitionDialog } from '../components/SceneOverlay.tsx'
import { RegenButton } from '../components/RegenButton.tsx'
import { OverlayProvider, useOverlay } from '../components/Overlay.tsx'
import { OUTCOME_LABEL, sceneGenerating } from '../lib/helpers.ts'
import { useMission, useResetMission, useRespond, useSuggestNativeReply } from '../store/missions.ts'
import { useMe } from '../store/me.ts'
import { emitWriting } from '../lib/events.ts'
import type { GrammarError, Mission } from '../lib/types.ts'

export function MissionPlay() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const missionQ = useMission(id)
  const respond = useRespond(id)
  const native = useSuggestNativeReply(id)
  const resetM = useResetMission(id)
  const meQ = useMe()

  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  const mission = missionQ.data?.mission
  const active = missionQ.data?.activeReply
  const running = active?.status === 'running'
  const pending = respond.isPending
  const sceneBusy = mission ? sceneGenerating(mission, active) : false
  const busy = running || pending || sceneBusy
  const complete = !!mission?.isComplete
  const showTransition = sceneBusy && !running
  const [transitionOpen, setTransitionOpen] = useState(true)
  useEffect(() => {
    if (showTransition) setTransitionOpen(true)
  }, [showTransition])
  const transitionDialogMounted = useDelayedUnmount(showTransition && transitionOpen, 700)

  useAgentView(
    mission
      ? {
          app: 'quest',
          screen: 'mission',
          title: mission.title,
          info: 'The user is playing a quest (dialogue mission). Details — read_quest with this missionId: part=info (plot/characters/goals) or part=dialogue (dialogue history).',
          ids: { missionId: id },
          state: { complete: complete ? 'yes' : 'no' },
        }
      : null,
  )

  const optimistic = pending ? respond.variables?.text ?? null : running ? active?.inputText ?? null : null
  const grammar: GrammarError[] | null =
    !busy && active?.status === 'done' && active.result && !active.result.grammarOk
      ? (active.result.errors ?? [])
      : null

  const [streak, setStreak] = useState(0)
  const streakJobRef = useRef<string | null>(null)
  useEffect(() => {
    if (active?.status !== 'done' || !active.result || active.jobId === streakJobRef.current) return
    streakJobRef.current = active.jobId
    setStreak((n) => (active.result?.grammarOk ? n + 1 : 0))
  }, [active?.status, active?.jobId, active?.result])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [mission?.history?.length, running, sceneBusy, grammar?.length, active?.partial?.narration, active?.partial?.responses?.length])

  async function submit() {
    const text = input.trim()
    if (!text || busy) return
    const firstAttempt = !grammar
    setInput('')
    try {
      await respond.mutateAsync({ text, strict: false })
      if (firstAttempt) emitWriting(text, { app: 'quest', mission_id: id })
    } catch {
      setInput(text)
    }
  }

  function suggestNative() {
    const text = input.trim()
    if (!text || busy || native.isPending) return
    native.mutate(text)
  }

  function reset() {
    setInput('')
    native.reset()
    resetM.mutate()
  }

  if (missionQ.isError) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <ErrorState
          title="Failed to load the mission"
          description={(missionQ.error as Error)?.message}
          action={
            <Button variant="secondary" onClick={() => navigate('..')}>
              <ArrowLeft className="h-4 w-4" /> Back to list
            </Button>
          }
          className="w-full max-w-md"
        />
      </div>
    )
  }

  if (missionQ.isLoading || !mission) {
    return <LoadingState className="h-full items-center py-0 text-neutral-400" />
  }

  if (mission.generationStatus === 'error') {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <ErrorState
          title="Failed to create the mission"
          description={mission.generationError}
          action={
            <Button variant="secondary" onClick={() => navigate('..')}>
              <ArrowLeft className="h-4 w-4" /> Back to list
            </Button>
          }
          className="w-full max-w-md"
        />
      </div>
    )
  }

  const outcome = mission.outcome ? OUTCOME_LABEL[mission.outcome] : undefined
  const playerAvatar = meQ.data?.picture_url || mission.playerAvatarImage || undefined

  return (
    <OverlayProvider>
    <div className="flex h-full min-h-0 flex-col bg-neutral-50">
      <header className="flex h-14 shrink-0 items-center gap-2 border-b border-neutral-200 bg-white/90 px-3 backdrop-blur">
        <Button variant="ghost" size="sm" onClick={() => navigate('..')}>
          <ArrowLeft className="h-4 w-4" /> Quests
        </Button>
        <span className="flex-1 truncate text-center text-sm font-semibold text-neutral-800">{mission.title}</span>
        <Button variant="ghost" size="sm" onClick={reset} disabled={resetM.isPending}>
          <RotateCcw className="h-4 w-4" /> Restart
        </Button>
      </header>

      {resetM.isError && (
        <div className="shrink-0 border-b border-rose-200 bg-rose-50 px-4 py-2 text-sm text-rose-600">
          {(resetM.error as Error)?.message || 'Failed to restart the mission'}
        </div>
      )}

      <div className="flex min-h-0 flex-1">
        <aside className="hidden w-80 shrink-0 flex-col gap-4 overflow-y-auto border-r border-neutral-200 bg-white p-4 lg:flex xl:w-96">
          <CoverCard mission={mission} />
          <GoalCard mission={mission} />
          <Characters mission={mission} />
        </aside>

        <div className="relative flex min-h-0 flex-1 flex-col">
          {transitionDialogMounted && (
            <SceneTransitionDialog
              sceneNumber={mission.currentStage + 2}
              summary={mission.scenes?.find((s) => s.stage === mission.currentStage)?.summary ?? undefined}
              backdropImage={mission.sceneImages?.[String(mission.currentStage)]}
              leaving={!(showTransition && transitionOpen)}
              onClose={() => setTransitionOpen(false)}
            />
          )}
          <div className="flex-1 min-h-0 overflow-y-auto">
            <div className="mx-auto max-w-3xl space-y-4 p-4 pb-8">
              <div className="space-y-4 lg:hidden">
                <CoverCard mission={mission} />
                <GoalCard mission={mission} />
                <Characters mission={mission} />
              </div>

              <Chat
                mission={mission}
                optimisticText={optimistic}
                playerAvatar={playerAvatar}
                optimisticState={running || pending ? (running && active?.step !== 'grammar' ? 'ok' : 'checking') : undefined}
              />

              {running && active?.partial && <DraftWorld mission={mission} partial={active.partial} step={active.step} />}

              {running && !active?.partial && active?.step !== 'grammar' && <TypingBubble />}

              {showTransition && <SceneTransition onClick={() => setTransitionOpen(true)} />}

              {grammar && grammar.length > 0 && (
                <GrammarPanel
                  text={active?.inputText ?? ''}
                  errors={grammar}
                  onEdit={() => setInput(active?.inputText ?? '')}
                />
              )}

              {complete && outcome && (
                <div className="rounded-2xl border border-amber-200 bg-amber-50 p-5 text-center">
                  <div className={`text-lg font-bold ${outcome.tone}`}>{outcome.label}</div>
                  <Button variant="secondary" className="mt-4" onClick={() => navigate('..')}>
                    <ArrowLeft className="h-4 w-4" /> Back to quests
                  </Button>
                </div>
              )}

              <div ref={bottomRef} />
            </div>
          </div>

          {!complete && (
            <div className="shrink-0 border-t border-neutral-200 bg-white p-3">
              <div className="mx-auto max-w-3xl space-y-2">
                {streak >= 2 && (
                  <div className="animate-pop-in flex justify-end">
                    <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-600">
                      <Flame className="h-3.5 w-3.5" />
                      Clean streak ×{streak}
                    </span>
                  </div>
                )}
                <NativeReplyPanel
                  pending={native.isPending}
                  error={native.isError ? (native.error as Error)?.message : null}
                  variants={native.data ?? null}
                  onClose={() => native.reset()}
                />
                <div className="flex items-end gap-2">
                  <textarea
                    rows={1}
                    value={input}
                    disabled={busy}
                    onChange={(e) => setInput(e.target.value)}
                    onPaste={(e) => e.preventDefault()}
                    onDrop={(e) => e.preventDefault()}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault()
                        submit()
                      }
                    }}
                    placeholder={grammar ? 'Fix the mistakes and try again…' : 'Your reply…'}
                    className="max-h-32 flex-1 resize-none rounded-xl border border-neutral-200 px-4 py-2.5 text-sm outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-100 disabled:bg-neutral-50"
                  />
                  <Button
                    variant="secondary"
                    size="icon"
                    title="How a native would say it"
                    onClick={suggestNative}
                    disabled={busy || native.isPending || !input.trim()}
                  >
                    {native.isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Sparkles className="h-4 w-4" />
                    )}
                  </Button>
                  <Button
                    variant="brand"
                    size="icon"
                    onClick={submit}
                    disabled={busy || !input.trim()}
                  >
                    {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
    </OverlayProvider>
  )
}

function CoverCard({ mission }: { mission: Mission }) {
  const { openImage } = useOverlay()
  return (
    <div className="overflow-hidden rounded-2xl bg-white ring-1 ring-neutral-200">
      <div className="relative aspect-[16/7] w-full bg-gradient-to-br from-brand-700 to-neutral-900">
        {mission.coverImage ? (
          <img
            src={mission.coverImage}
            alt={mission.title}
            onClick={() => openImage(mission.coverImage!, mission.title)}
            className="absolute inset-0 h-full w-full cursor-zoom-in object-cover"
          />
        ) : mission.coverImageStatus === 'generating' ? (
          <div className="flex h-full items-center justify-center gap-2 text-xs text-white/80">
            <Loader2 className="h-4 w-4 animate-spin" /> Drawing the cover…
          </div>
        ) : mission.coverImageStatus === 'error' ? (
          <div className="flex h-full flex-col items-center justify-center gap-2 text-xs text-white/80">
            <TriangleAlert className="h-5 w-5 text-rose-300" /> Cover failed
            <RegenButton missionId={mission.id} kind="cover" className="bg-white/10 text-white hover:bg-white/20" />
          </div>
        ) : (
          <div className="flex h-full items-center justify-center text-4xl">🎬</div>
        )}
      </div>
      <div className="p-4">
        <h1 className="text-base font-bold leading-snug text-neutral-900">{mission.title}</h1>
        {mission.description && <p className="mt-1 text-sm leading-relaxed text-neutral-500">{mission.description}</p>}
      </div>
    </div>
  )
}

function GoalCard({ mission }: { mission: Mission }) {
  const goal = mission.resolution?.goal
  const points = mission.plotPoints ?? []
  if (!goal && points.length === 0) return null

  const found = points.filter((p) => p.delivered).length
  const required = points.filter((p) => p.required)
  const ready = required.length > 0 && required.every((p) => p.delivered)

  return (
    <div className="rounded-2xl bg-white p-4 ring-1 ring-neutral-200">
      {goal && (
        <div className="flex items-start gap-2">
          <Target className="mt-0.5 h-4 w-4 shrink-0 text-brand-600" />
          <div>
            <div className="text-xs font-semibold uppercase tracking-wide text-neutral-400">Your goal</div>
            <p className="mt-0.5 text-sm leading-relaxed text-neutral-800">{goal}</p>
          </div>
        </div>
      )}
      {points.length > 0 && (
        <div className={goal ? 'mt-3' : ''}>
          <div className="flex items-center justify-between text-xs text-neutral-500">
            <span>Discoveries</span>
            <span className="font-medium text-neutral-700">{found} / {points.length}</span>
          </div>
          <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-neutral-100">
            <div
              className="h-full rounded-full bg-brand-500 transition-all"
              style={{ width: `${points.length ? (found / points.length) * 100 : 0}%` }}
            />
          </div>
          {ready && !mission.isComplete && (
            <p className="mt-2 text-xs font-medium text-emerald-600">You know enough to act — make your move.</p>
          )}
        </div>
      )}
    </div>
  )
}

function NativeReplyPanel({
  pending,
  error,
  variants,
  onClose,
}: {
  pending: boolean
  error: string | null
  variants: string[] | null
  onClose: () => void
}) {
  if (!pending && !error && (!variants || variants.length === 0)) return null

  return (
    <div className="rounded-2xl border border-brand-200 bg-brand-50/60 p-3">
      <div className="mb-2 flex items-center justify-between">
        <div className="flex items-center gap-2 text-sm font-semibold text-brand-700">
          <Sparkles className="h-4 w-4" />
          How a native would say it
        </div>
        <button
          onClick={onClose}
          className="rounded-md p-1 text-neutral-400 hover:bg-brand-100 hover:text-neutral-600"
          aria-label="Close"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {pending && (
        <div className="flex items-center gap-2 text-sm text-neutral-500">
          <Loader2 className="h-4 w-4 animate-spin" /> Thinking…
        </div>
      )}

      {!pending && error && <p className="text-sm text-rose-600">{error}</p>}

      {!pending && !error && variants && (
        <ul
          className="select-none space-y-1.5"
          onCopy={(e) => e.preventDefault()}
          onCut={(e) => e.preventDefault()}
          onContextMenu={(e) => e.preventDefault()}
        >
          {variants.map((variant, i) => (
            <li
              key={i}
              className="rounded-xl bg-white px-3 py-2 text-sm text-neutral-800 ring-1 ring-brand-100"
            >
              {variant}
            </li>
          ))}
        </ul>
      )}

      {!pending && !error && (
        <p className="mt-2 text-xs text-neutral-400">Read it, then type your own reply by hand.</p>
      )}
    </div>
  )
}

function highlightErrors(text: string, errors: GrammarError[]) {
  const ranges: Array<[number, number]> = []
  const lower = text.toLowerCase()
  for (const e of errors) {
    const frag = e.original?.trim()
    if (!frag) continue
    const idx = lower.indexOf(frag.toLowerCase())
    if (idx < 0) continue
    if (ranges.some(([s, en]) => idx < en && idx + frag.length > s)) continue
    ranges.push([idx, idx + frag.length])
  }
  if (ranges.length === 0) return [{ text, marked: false }]
  ranges.sort((a, b) => a[0] - b[0])

  const segments: Array<{ text: string; marked: boolean }> = []
  let pos = 0
  for (const [start, end] of ranges) {
    if (start > pos) segments.push({ text: text.slice(pos, start), marked: false })
    segments.push({ text: text.slice(start, end), marked: true })
    pos = end
  }
  if (pos < text.length) segments.push({ text: text.slice(pos), marked: false })
  return segments
}

export function GrammarPanel({ text, errors, onEdit }: { text: string; errors: GrammarError[]; onEdit: () => void }) {
  return (
    <div className="rounded-2xl border border-rose-200 bg-rose-50 p-4">
      <div className="mb-2 flex items-center gap-2 text-sm font-semibold text-rose-700">
        <TriangleAlert className="h-4 w-4" />
        Mistakes found
      </div>
      {text && (
        <p className="mb-3 rounded-xl bg-white/70 px-3 py-2 text-sm leading-relaxed text-neutral-700 ring-1 ring-rose-100">
          {highlightErrors(text, errors).map((seg, i) =>
            seg.marked ? (
              <mark key={i} className="rounded bg-rose-100 px-0.5 font-medium text-rose-700 underline decoration-rose-400 decoration-wavy underline-offset-2">
                {seg.text}
              </mark>
            ) : (
              <span key={i}>{seg.text}</span>
            ),
          )}
        </p>
      )}
      <ul className="space-y-2">
        {errors.map((e, i) => (
          <li key={i} className="text-sm">
            <span className="text-rose-500 line-through">{e.original}</span>
            <span className="mx-1.5 text-neutral-400">→</span>
            <span className="font-medium text-brand-700">{e.correction}</span>
            {e.explanation && <p className="mt-0.5 text-xs text-neutral-500">{e.explanation}</p>}
          </li>
        ))}
      </ul>
      <Button variant="ghost" size="sm" className="mt-3 text-rose-700 hover:bg-rose-100" onClick={onEdit}>
        Restore my text
      </Button>
    </div>
  )
}
