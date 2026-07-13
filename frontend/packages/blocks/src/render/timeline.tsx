// Shrink a timeline label to fit its segment width; hard-clamp with textLength if still too long.
function fitLabel(label: string, maxW: number): { fontSize: number; textLength?: number; lengthAdjust?: 'spacingAndGlyphs' } {
  const natural = label.length * 0.56 * 21
  if (maxW <= 0) return { fontSize: 11, textLength: 8, lengthAdjust: 'spacingAndGlyphs' }
  if (natural <= maxW) return { fontSize: 21 }
  const scaled = Math.floor(21 * (maxW / natural))
  if (scaled >= 11) return { fontSize: scaled }
  return { fontSize: 11, textLength: maxW, lengthAdjust: 'spacingAndGlyphs' }
}

export function Timeline({ raw }: { raw: string }) {
  const ticks: { x: number; label: string }[] = []
  const notes: { x: number; lines: string[] }[] = []
  const segs: { shape: 'arrow' | 'arrow2' | 'box'; x1: number; x2: number; label: string }[] = []
  for (const line of raw.split('\n').map((l) => l.trim())) {
    const t = /^tick:\s*(\d+)\s+(.*)$/.exec(line)
    if (t) {
      ticks.push({ x: Number(t[1]), label: t[2] ?? '' })
      continue
    }
    const n = /^note:\s*(\d+)\s+(.*)$/.exec(line)
    if (n) {
      notes.push({ x: Number(n[1]), lines: (n[2] ?? '').split('|').map((s) => s.trim()) })
      continue
    }
    const m = /^(arrow2|arrow|box):\s*(\d+)\s+(\d+)\s*\|\s*(.*)$/.exec(line)
    if (m) segs.push({ shape: m[1] as 'arrow' | 'arrow2' | 'box', x1: Number(m[2]), x2: Number(m[3]), label: m[4] ?? '' })
  }

  const noteLines = notes.reduce((m, n) => Math.max(m, n.lines.length), 0)
  const noteH = noteLines ? noteLines * 24 + 12 : 0
  const rowH = 50
  const top = 12 + noteH
  const baselineY = top + segs.length * rowH + 8
  const H = baselineY + 44
  const X = (p: number) => 46 + (p / 100) * 908
  const head = 40

  return (
    <div className="my-2 w-full overflow-x-auto">
      <svg viewBox={`0 0 1000 ${H}`} className="h-auto w-full min-w-[26rem]">
        {notes.map((n, i) =>
          n.lines.map((ln, j) => (
            <text key={`${i}-${j}`} x={X(n.x)} y={16 + j * 24} textAnchor="middle" className="fill-neutral-500" fontSize={21} fontStyle="italic">
              {ln}
            </text>
          )),
        )}
        <line x1={X(0)} y1={baselineY} x2={X(100)} y2={baselineY} className="stroke-neutral-400" strokeWidth={2} />
        {ticks.map((t, i) => (
          <g key={i}>
            <line x1={X(t.x)} y1={baselineY - 10} x2={X(t.x)} y2={baselineY + 10} className="stroke-neutral-500" strokeWidth={2} />
            <text x={X(t.x)} y={baselineY + 32} textAnchor="middle" className="fill-neutral-500" fontSize={21} fontStyle="italic">
              {t.label}
            </text>
          </g>
        ))}
        {segs.map((s, i) => {
          const yMid = top + i * rowH + 16
          const x1 = X(s.x1)
          const x2 = X(s.x2)
          if (s.shape === 'box') {
            const fit = fitLabel(s.label, x2 - x1 - 12)
            return (
              <g key={i}>
                <rect x={x1} y={yMid - 15} width={x2 - x1} height={30} rx={4} className="fill-brand-500 stroke-brand-600" strokeWidth={1} />
                <text
                  x={(x1 + x2) / 2}
                  y={yMid + 7}
                  textAnchor="middle"
                  className="fill-white"
                  fontSize={fit.fontSize}
                  fontWeight={700}
                  textLength={fit.textLength}
                  lengthAdjust={fit.lengthAdjust}
                >
                  {s.label}
                </text>
              </g>
            )
          }
          const pts =
            s.shape === 'arrow'
              ? `${x1},${yMid - 13} ${x2 - head},${yMid - 13} ${x2 - head},${yMid - 22} ${x2},${yMid} ${x2 - head},${yMid + 22} ${x2 - head},${yMid + 13} ${x1},${yMid + 13}`
              : `${x1 + head},${yMid - 13} ${x2 - head},${yMid - 13} ${x2 - head},${yMid - 22} ${x2},${yMid} ${x2 - head},${yMid + 22} ${x2 - head},${yMid + 13} ${x1 + head},${yMid + 13} ${x1 + head},${yMid + 22} ${x1},${yMid} ${x1 + head},${yMid - 22}`
          const cx = s.shape === 'arrow' ? (x1 + (x2 - head)) / 2 : (x1 + x2) / 2
          const innerW = s.shape === 'arrow' ? x2 - head - x1 : x2 - head - (x1 + head)
          const fit = fitLabel(s.label, innerW - 12)
          return (
            <g key={i}>
              <polygon points={pts} className="fill-brand-500 stroke-brand-600" strokeWidth={1} />
              <text
                x={cx}
                y={yMid + 7}
                textAnchor="middle"
                className="fill-white"
                fontSize={fit.fontSize}
                fontWeight={700}
                textLength={fit.textLength}
                lengthAdjust={fit.lengthAdjust}
              >
                {s.label}
              </text>
            </g>
          )
        })}
      </svg>
    </div>
  )
}
