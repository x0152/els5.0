import { useEffect, useState } from 'react'
import { Flame, RotateCcw } from 'lucide-react'
import { Button } from '@els/ui'
import { DraftWorld, PlayerBubble, SceneTransition, SystemLine, TypingBubble, WorkingDots, useDelayedUnmount } from '../components/Chat.tsx'
import { SceneTransitionDialog } from '../components/SceneOverlay.tsx'
import { GrammarPanel } from './MissionPlay.tsx'
import { OverlayProvider } from '../components/Overlay.tsx'
import type { Mission, PartialWorld } from '../lib/types.ts'

const PREV_SCENE_IMAGE =
  'data:image/svg+xml,' +
  encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="360"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#065f46"/><stop offset="0.6" stop-color="#0f172a"/><stop offset="1" stop-color="#022c22"/></linearGradient></defs><rect width="640" height="360" fill="url(#g)"/><circle cx="140" cy="120" r="55" fill="#34d399" opacity="0.25"/><rect x="380" y="180" width="190" height="120" rx="8" fill="#10b981" opacity="0.18"/><rect x="60" y="230" width="240" height="14" rx="7" fill="#a7f3d0" opacity="0.3"/><rect x="60" y="256" width="170" height="14" rx="7" fill="#a7f3d0" opacity="0.2"/><text x="320" y="80" font-family="sans-serif" font-size="20" fill="#d1fae5" opacity="0.5" text-anchor="middle">office lobby, scene 1</text></svg>`,
  )

const mission = {
  id: 'demo',
  title: 'Animations Lab',
  characters: [{ name: 'Nina Brooks', role: 'HR coordinator' }],
  characterAvatars: {},
  sceneImages: { '0': PREV_SCENE_IMAGE },
  sceneImageStatus: { '1': 'generating' },
  scenes: [
    {
      stage: 0,
      summary: 'At the office lobby, Egor met Nina, got a coffee and learned Diana is running late.',
    },
  ],
} as unknown as Mission

const NARRATION = 'Nina gets up and moves to the coffee machine in the corner of the lobby.'
const REPLY =
  "It's Nina — I mentioned it when you came in, but no worries at all. How do you take your coffee? Milk, sugar, or just black like most developers I know?"

// Full turn scenario: grammar -> narration -> typing -> line -> evaluation.
type SimPhase =
  | { kind: 'checking' }
  | { kind: 'partial'; partial: PartialWorld; step: 'world' | 'evaluating' }
  | { kind: 'done' }

function buildTimeline(): Array<{ at: number; phase: SimPhase }> {
  const t: Array<{ at: number; phase: SimPhase }> = [{ at: 0, phase: { kind: 'checking' } }]
  let at = 2000

  const narrWords = NARRATION.split(' ')
  for (let i = 4; i <= narrWords.length; i += 4) {
    t.push({
      at,
      phase: {
        kind: 'partial',
        step: 'world',
        partial: { narration: narrWords.slice(0, i).join(' '), narrationDone: i >= narrWords.length },
      },
    })
    at += 260
  }

  t.push({
    at,
    phase: {
      kind: 'partial',
      step: 'world',
      partial: { narration: NARRATION, narrationDone: true, responses: [{ name: 'Nina Brooks' }] },
    },
  })
  at += 1400

  const replyWords = REPLY.split(' ')
  for (let i = 5; i <= replyWords.length; i += 5) {
    const doneReply = i >= replyWords.length
    t.push({
      at,
      phase: {
        kind: 'partial',
        step: 'world',
        partial: {
          narration: NARRATION,
          narrationDone: true,
          responses: [{ name: 'Nina Brooks', text: replyWords.slice(0, i).join(' '), done: doneReply }],
        },
      },
    })
    at += 300
  }

  t.push({
    at,
    phase: {
      kind: 'partial',
      step: 'evaluating',
      partial: {
        narration: NARRATION,
        narrationDone: true,
        responses: [{ name: 'Nina Brooks', text: REPLY, done: true }],
      },
    },
  })
  at += 2600
  t.push({ at, phase: { kind: 'done' } })
  return t
}

function useTurnSimulation() {
  const [phase, setPhase] = useState<SimPhase>({ kind: 'checking' })
  const [run, setRun] = useState(0)

  useEffect(() => {
    const timeline = buildTimeline()
    const timers = timeline.map(({ at, phase }) => setTimeout(() => setPhase(phase), at))
    const total = (timeline[timeline.length - 1]?.at ?? 0) + 1800
    const loop = setTimeout(() => setRun((n) => n + 1), total)
    return () => {
      timers.forEach(clearTimeout)
      clearTimeout(loop)
    }
  }, [run])

  return { phase, replay: () => setRun((n) => n + 1) }
}

// Cinematic transition dialog: background blurs, centered card with scene number and
// previous outcome; closes on click — an inline loading banner remains,
// clicking it opens the dialog again.
function SceneOverlayCard() {
  const [run, setRun] = useState(0)
  const [open, setOpen] = useState(true)
  const [ready, setReady] = useState(false)
  const dialogMounted = useDelayedUnmount(open && !ready, 700)

  useEffect(() => {
    setOpen(true)
    setReady(false)
    const timers = [
      setTimeout(() => setReady(true), 9000),
      setTimeout(() => setRun((n) => n + 1), 12500),
    ]
    return () => timers.forEach(clearTimeout)
  }, [run])

  return (
    <Card title="Cinematic transition dialog (click/X — hide, banner click — open)" onReplay={() => setRun((n) => n + 1)}>
      <div className="relative h-[460px] overflow-hidden rounded-2xl bg-neutral-50 p-4 ring-1 ring-neutral-200">
        <div className="space-y-4">
          <p className="border-l-2 border-brand-200 pl-4 text-sm italic leading-relaxed text-neutral-600">
            Nina sets the black coffee on the edge of the desk in front of Egor. The lobby is quiet except for the hum
            of the coffee machine.
          </p>
          <PlayerBubble text="Thank you, Nina. I think I have everything I need for now." state="ok" />
          <p className="border-l-2 border-brand-200 pl-4 text-sm italic leading-relaxed text-neutral-600">
            Nina nods and returns to her paperwork, glancing at the clock above the door.
          </p>
          {!ready && <SceneTransition onClick={() => setOpen(true)} />}
          {ready && <SystemLine mission={mission} text="- Scene 2 -" scene={1} />}
        </div>
        {dialogMounted && (
          <SceneTransitionDialog
            sceneNumber={2}
            summary="At the office lobby, Egor met Nina, got a coffee and learned Diana is running late."
            backdropImage={PREV_SCENE_IMAGE}
            leaving={!open || ready}
            onClose={() => setOpen(false)}
          />
        )}
      </div>
    </Card>
  )
}

function Card({ title, children, onReplay }: { title: string; children: React.ReactNode; onReplay?: () => void }) {
  return (
    <section className="rounded-2xl bg-white p-5 ring-1 ring-neutral-200">
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-sm font-semibold text-neutral-700">{title}</h2>
        {onReplay && (
          <Button variant="ghost" size="sm" onClick={onReplay}>
            <RotateCcw className="h-3.5 w-3.5" /> Replay
          </Button>
        )}
      </div>
      <div className="space-y-4">{children}</div>
    </section>
  )
}

function Remount({ id, children }: { id: number; children: React.ReactNode }) {
  return <div key={id}>{children}</div>
}

export function AnimationsLab() {
  const sim = useTurnSimulation()
  const [replayAll, setReplayAll] = useState(0)

  return (
    <OverlayProvider>
      <div className="min-h-full bg-neutral-50 p-6">
        <div className="mx-auto max-w-3xl space-y-4">
          <header className="flex items-center justify-between">
            <div>
              <h1 className="text-lg font-bold text-neutral-900">Quest Animations Lab</h1>
              <p className="text-sm text-neutral-500">All turn animation states. Production components, mock data.</p>
            </div>
            <Button variant="secondary" size="sm" onClick={() => setReplayAll((n) => n + 1)}>
              <RotateCcw className="h-4 w-4" /> Replay all
            </Button>
          </header>

          <Card title="Full turn (auto-loop)" onReplay={sim.replay}>
            <PlayerBubble
              text="Coffee please! What is your name?"
              state={sim.phase.kind === 'checking' ? 'checking' : 'ok'}
            />
            {sim.phase.kind === 'partial' && (
              <DraftWorld mission={mission} partial={sim.phase.partial} step={sim.phase.step} />
            )}
            {sim.phase.kind === 'done' && (
              <p className="border-l-2 border-brand-200 pl-4 text-xs italic text-neutral-400">
                — turn finished, draft replaced by final history —
              </p>
            )}
          </Card>

          <Remount id={replayAll}>
            <Card title="Player bubble: checking (scan + caption)">
              <PlayerBubble text="Yesterday I went to the yard and saw nothing there." state="checking" />
            </Card>

            <Card title="Player bubble: checked (checkmark)">
              <PlayerBubble text="Yesterday I went to the yard and saw nothing there." state="ok" />
            </Card>

            <Card title="Grammar: errors with highlight">
              <GrammarPanel
                text="Yesterday I goed to the yard and seen nothing there"
                errors={[
                  { original: 'goed', correction: 'went', explanation: "Past simple of 'go' is 'went'.", type: 'grammar' },
                  { original: 'seen', correction: 'saw', explanation: "Past simple of 'see' is 'saw'.", type: 'grammar' },
                ]}
                onEdit={() => {}}
              />
            </Card>

            <Card title="Typing: name unknown / known">
              <TypingBubble />
              <TypingBubble mission={mission} name="Nina Brooks" />
            </Card>

            <Card title="Quiet processing (stream pause, turn evaluation)">
              <WorkingDots />
            </Card>

            <SceneOverlayCard />

            <Card title="Scene generation banner (inline)">
              <SceneTransition />
            </Card>

            <Card title="Scene transition: outcome + title with image">
              <SystemLine mission={mission} text="- Scene completed -" scene={0} />
              <SystemLine mission={mission} text="- Scene 2 -" scene={1} />
            </Card>

            <Card title="Clean streak">
              <div className="flex justify-end">
                <span className="animate-pop-in inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-xs font-medium text-amber-600">
                  <Flame className="h-3.5 w-3.5" />
                  Clean streak ×4
                </span>
              </div>
            </Card>
          </Remount>
        </div>
      </div>
    </OverlayProvider>
  )
}
