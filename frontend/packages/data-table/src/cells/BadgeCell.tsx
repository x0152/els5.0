import { GridCellKind, type CustomCell, type CustomRenderer } from '@glideapps/glide-data-grid'

interface BadgeCellProps {
  readonly kind: 'badge-cell'
  readonly value: string
}

export type BadgeCell = CustomCell<BadgeCellProps>

function hue(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) % 360
  return h
}

export const badgeRenderer: CustomRenderer<BadgeCell> = {
  kind: GridCellKind.Custom,
  isMatch: (c): c is BadgeCell => (c.data as { kind?: string }).kind === 'badge-cell',
  draw: (args, cell) => {
    const { ctx, theme, rect } = args
    const text = cell.data.value
    if (!text) return true

    const h = hue(text)
    ctx.save()
    ctx.font = `12px ${theme.fontFamily}`
    const height = 20
    const x = rect.x + theme.cellHorizontalPadding
    const y = rect.y + (rect.height - height) / 2
    const width = ctx.measureText(text).width + 16

    ctx.beginPath()
    ctx.roundRect(x, y, width, height, 6)
    ctx.fillStyle = `hsl(${h} 70% 92%)`
    ctx.fill()

    ctx.fillStyle = `hsl(${h} 50% 30%)`
    ctx.textAlign = 'left'
    ctx.textBaseline = 'middle'
    ctx.fillText(text, x + 8, y + height / 2 + 0.5)
    ctx.restore()
    return true
  },
}

export const createBadgeCell = (value: unknown): BadgeCell => ({
  kind: GridCellKind.Custom,
  data: { kind: 'badge-cell', value: value != null ? String(value) : '' },
  allowOverlay: false,
  copyData: value != null ? String(value) : '',
})
