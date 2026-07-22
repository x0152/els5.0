import type { SVGProps } from 'react'

const FONT = "ui-rounded, 'SF Pro Rounded', 'Segoe UI', system-ui, sans-serif"
const COLORS = ['#065f46', '#047857', '#059669', '#10b981', '#34d399']
const CHARS = ['a', 'e', 't', 'h', 's', 'o', 'w', 'ə', 'ʃ', 'ŋ']
const N = CHARS.length

const CYCLE = '16s'
const KEY_TIMES = '0;0.26;0.34;0.58;0.66;0.92;1'
const SPLINES = Array(6).fill('0.4 0 0.2 1').join(';')

function scatter(i: number, rot: number): [number, number] {
  const a = i * 2.39996 + rot
  const r = 32 + 38 * Math.sqrt((i + 0.5) / N)
  return [r * Math.cos(a), r * Math.sin(a)]
}

function wave(i: number, phase: number): [number, number] {
  const x = -78 + (156 * i) / (N - 1)
  return [x, 16 * Math.sin(x / 26 + phase) + (i % 2 ? 16 : -16)]
}

function ring(i: number, rot: number): [number, number] {
  const a = (Math.PI * 2 * i) / N - Math.PI / 2 + rot
  return [62 * Math.cos(a), 62 * Math.sin(a)]
}

const fmt = ([x, y]: [number, number]) => `${x.toFixed(1)} ${y.toFixed(1)}`

const GLYPHS = CHARS.map((ch, i) => ({
  ch,
  travel: [
    scatter(i, 0),
    scatter(i, 0.7),
    wave(i, 0),
    wave(i, Math.PI),
    ring(i, 0),
    ring(i, 0.55),
    scatter(i, 0),
  ]
    .map(fmt)
    .join(';'),
  color: COLORS[(i * 2) % COLORS.length],
  size: [24, 30, 25, 32, 26, 28][i % 6],
  tilt: ((i * 53) % 28) - 14,
  floatAmp: 2 + (i % 3),
  floatDur: `${(2.2 + (i % 5) * 0.35).toFixed(2)}s`,
  floatBegin: `-${((i * 0.37) % 2).toFixed(2)}s`,
  twinkleDur: `${(2.6 + (i % 4) * 0.5).toFixed(2)}s`,
  twinkleBegin: `-${((i * 0.61) % 3).toFixed(2)}s`,
}))

export function Mascot(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 200 190" xmlns="http://www.w3.org/2000/svg" aria-hidden="true" {...props}>
      <g transform="translate(100 95)" fontFamily={FONT} fontWeight="800">
        {GLYPHS.map((g, i) => (
          <g key={i}>
            <animateTransform
              attributeName="transform"
              type="translate"
              values={g.travel}
              keyTimes={KEY_TIMES}
              keySplines={SPLINES}
              calcMode="spline"
              dur={CYCLE}
              repeatCount="indefinite"
            />
            <g>
              <animateTransform
                attributeName="transform"
                type="translate"
                values={`0 0;0 -${g.floatAmp};0 0`}
                dur={g.floatDur}
                begin={g.floatBegin}
                repeatCount="indefinite"
              />
              <text
                transform={`rotate(${g.tilt})`}
                fontSize={g.size}
                fill={g.color}
                textAnchor="middle"
                dominantBaseline="central"
              >
                <animate
                  attributeName="opacity"
                  values="0.75;1;0.75"
                  dur={g.twinkleDur}
                  begin={g.twinkleBegin}
                  repeatCount="indefinite"
                />
                {g.ch}
              </text>
            </g>
          </g>
        ))}
      </g>
    </svg>
  )
}
