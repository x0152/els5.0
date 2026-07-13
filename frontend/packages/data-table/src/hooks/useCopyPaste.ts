import { useCallback } from 'react'
import { GridCellKind } from '@glideapps/glide-data-grid'
import type { Item, EditableGridCell } from '@glideapps/glide-data-grid'
import { useDataTableContext } from '../context/DataTableContext'
import { useDataOperations } from './useDataOperations'
import { useCellOperations } from './useCellOperations'

export const useCopyPaste = <T extends Record<string, any>>() => {
  const { state, actions, columns, keyField, allowAddRows = true } = useDataTableContext<T>()
  const { filteredItems, generateNewRow } = useDataOperations<T>()
  const { updateCellValue } = useCellOperations<T>()

  const onPaste = useCallback((target: Item, values: readonly (readonly string[])[]) => {
    if (state.editMode !== 'edit') return false

    const [targetCol, targetRow] = target
    let pasteHeight = values.length
    const currentRowCount = filteredItems.length

    const newItemIds: string[] = []

    if (targetRow + pasteHeight > currentRowCount) {
      if (!allowAddRows) {
        pasteHeight = Math.max(0, currentRowCount - targetRow)
        if (pasteHeight === 0) return false
      } else {
        const newRowsNeeded = targetRow + pasteHeight - currentRowCount
        const updatedNewRowIds = new Set(state.newRowIds)
        const newItems: T[] = []

        for (let i = 0; i < newRowsNeeded; i++) {
          const newId = `${Date.now()}-${i}-${Math.random()}`
          const newItem = { ...generateNewRow(newId) } as unknown as T
          newItems.push(newItem)
          updatedNewRowIds.add(newId)
          newItemIds.push(newId)
        }

        actions.setNewRowIds(updatedNewRowIds)
        actions.setItems((prev: T[]) => [...prev, ...newItems])
      }
    }

    values.slice(0, pasteHeight).forEach((row, rowIndex) => {
      const actualRowIndex = targetRow + rowIndex
      let itemId: string

      if (actualRowIndex < currentRowCount) {
        itemId = filteredItems[actualRowIndex][keyField]
      } else {
        itemId = newItemIds[actualRowIndex - currentRowCount]
      }

      row.forEach((value, colIndex) => {
        if (targetCol + colIndex < columns.length) {
          const column = columns[targetCol + colIndex]
          if (column.readonly) return

          let newValue: EditableGridCell
          if (column.type === 'boolean') {
            const boolValue = value === 'true' || value === '1' || value === 'TRUE' || value === 'True'
            newValue = {
              kind: GridCellKind.Boolean as const,
              data: boolValue,
              allowOverlay: false,
            }
          } else if (column.type === 'number' || column.type === 'monetary') {
            const numValue = Number(value)
            newValue = {
              kind: GridCellKind.Number as const,
              data: isNaN(numValue) ? 0 : numValue,
              allowOverlay: true,
              displayData: value,
            }
          } else {
            newValue = {
              kind: GridCellKind.Text as const,
              data: value,
              allowOverlay: true,
              copyData: value,
              displayData: value,
            }
          }

          updateCellValue(column, newValue, itemId)
        }
      })
    })

    return true
  }, [state.editMode, state.newRowIds, filteredItems, columns, actions, updateCellValue, generateNewRow, keyField, allowAddRows])

  return {
    onPaste
  }
}
