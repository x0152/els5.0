import React from 'react'
import { GridCellKind, type CustomCell, type CustomRenderer } from '@glideapps/glide-data-grid'
import { DatePicker } from 'antd'
import dayjs, { Dayjs } from 'dayjs'
import weekday from 'dayjs/plugin/weekday'
import localeData from 'dayjs/plugin/localeData'
import customParseFormat from 'dayjs/plugin/customParseFormat'
import weekOfYear from 'dayjs/plugin/weekOfYear'
import weekYear from 'dayjs/plugin/weekYear'
import advancedFormat from 'dayjs/plugin/advancedFormat'
import 'dayjs/locale/ru'

dayjs.extend(weekday)
dayjs.extend(localeData)
dayjs.extend(customParseFormat)
dayjs.extend(weekOfYear)
dayjs.extend(weekYear)
dayjs.extend(advancedFormat)
dayjs.locale('ru')

interface DateCellProps {
  readonly kind: "date-cell"
  readonly value: string | null
  readonly placeholder?: string
  readonly readonly?: boolean
}

export type DateCell = CustomCell<DateCellProps>

const formatDateToDDMMYYYY = (dateString: string | null): string => {
  if (!dateString) return ''
  try {
    return dayjs(dateString).format('DD.MM.YYYY')
  } catch {
    return ''
  }
}

export const dateRenderer: CustomRenderer<DateCell> = {
  kind: GridCellKind.Custom,
  isMatch: (c): c is DateCell => (c.data as any).kind === "date-cell",
  draw: (args, cell) => {
    const { ctx, theme, rect } = args
    const { value, placeholder } = cell.data

    const x = rect.x + theme.cellHorizontalPadding
    const y = rect.y + rect.height / 2

    ctx.save()
    ctx.font = `13px ${theme.fontFamily}`
    ctx.fillStyle = theme.textDark
    ctx.textAlign = "left"
    ctx.textBaseline = "middle"

    if (value) {
      ctx.fillText(formatDateToDDMMYYYY(value), x, y)
    } else {
      ctx.fillText(placeholder || "<empty>", x, y)
    }

    ctx.restore()
    return true
  },
  provideEditor: () => {
    return (p) => {
      const { data } = p.value
      const [selectedDate, setSelectedDate] = React.useState<Dayjs | null>(
        data.value ? dayjs(data.value) : null
      )
      const [isOpen, setIsOpen] = React.useState(true)

      const handleChange = (date: Dayjs | null) => {
        setSelectedDate(date)
        
        const newValue = {
          ...p.value,
          data: {
            ...data,
            value: date ? date.format('YYYY-MM-DD') : null,
          },
        }
        
        p.onChange(newValue)
        
        if (date) {
          setIsOpen(false)
          setTimeout(() => {
            p.onFinishedEditing?.(newValue)
          }, 100)
        } else if (date === null) {
          setTimeout(() => {
            p.onFinishedEditing?.(newValue)
          }, 100)
        }
      }

      const handleMouseDown = (e: React.MouseEvent) => {
        e.preventDefault()
        e.stopPropagation()
      }

      const handleClick = (e: React.MouseEvent) => {
        e.preventDefault()
        e.stopPropagation()
        if (!isOpen) {
          setIsOpen(true)
        }
      }

      return (
        <div 
          style={{ 
            width: '100%', 
            height: '100%', 
            display: 'flex', 
            alignItems: 'center',
            padding: '4px 8px'
          }}
          onMouseDown={handleMouseDown}
          onClick={handleClick}
          onPointerDown={handleMouseDown}
        >
          <DatePicker
            value={selectedDate}
            onChange={handleChange}
            placeholder="Select a date"
            format="DD.MM.YYYY"
            style={{ 
              width: '100%',
              border: 'none',
              boxShadow: 'none'
            }}
            popupStyle={{ zIndex: 10000 }}
            autoFocus
            open={isOpen}
            getPopupContainer={() => document.body}
            onMouseDown={handleMouseDown}
            onClick={handleClick}
          />
        </div>
      )
    }
  },
  onPaste: (v, d) => {
    const trimmedValue = v.trim()
    let parsedDate: Dayjs | null = null
    
    const formats = ['DD.MM.YYYY', 'DD-MM-YYYY', 'DD/MM/YYYY', 'YYYY-MM-DD']
    
    for (const format of formats) {
      parsedDate = dayjs(trimmedValue, format, true)
      if (parsedDate.isValid()) break
    }
    
    if (!parsedDate?.isValid()) {
      parsedDate = dayjs(trimmedValue)
    }
    
    return { 
      ...d, 
      value: parsedDate?.isValid() ? parsedDate.format('YYYY-MM-DD') : null 
    }
  },
}

export const createDateCell = (value: string | null, readonly?: boolean, placeholder?: string): DateCell => ({
  kind: GridCellKind.Custom,
  data: {
    kind: "date-cell",
    value,
    readonly,
    placeholder
  },
  allowOverlay: !readonly,
  copyData: value ? formatDateToDDMMYYYY(value) : '',
})
