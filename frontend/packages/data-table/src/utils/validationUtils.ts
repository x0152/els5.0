import type { ValidationResult, ValidationState, TableColumn } from '../types'
import { useValidation } from '../hooks/useValidation'
import { useDataTableContext } from '../context/DataTableContext'

export const cellKeyOf = (rowId: unknown, columnId: string) => `${rowId}:${columnId}`

export interface ValidationError {
  cellKey: string
  row: number
  col: number
  column: TableColumn
  item: any
  message: string
  value: any
}

/**
 * Gets all validation errors for the data
 */
export const getAllValidationErrors = (
  items: any[],
  columns: TableColumn[],
  validationState: ValidationState,
  deletedRowIds: Set<string>,
  validateField: (column: TableColumn, value: any) => ValidationResult,
  keyField: string,
): ValidationError[] => {
  const allErrors: ValidationError[] = []

  // Add errors from state.validationState
  Object.entries(validationState)
    .filter(([, validation]) => !validation.isValid && validation.message)
    .forEach(([cellKey, validation]) => {
      const sep = cellKey.indexOf(':')
      const rowId = cellKey.slice(0, sep)
      const columnId = cellKey.slice(sep + 1)
      const col = columns.findIndex((c) => c.id === columnId)
      const row = items.findIndex((item) => String(item[keyField]) === rowId)
      if (col === -1 || row === -1) return
      const column = columns[col]
      const item = items[row]

      // Skip rows marked for deletion
      if (deletedRowIds.has(item[keyField])) return

      allErrors.push({
        cellKey,
        row,
        col,
        column,
        item,
        message: validation.message!,
        value: item[column.id]
      })
    })

  // Check all cells for validation errors (including required fields)
  items.forEach((item, rowIndex) => {
    // Skip rows marked for deletion
    if (deletedRowIds.has(item[keyField])) return

    columns.forEach((column, colIndex) => {
      const cellKey = cellKeyOf(item[keyField], column.id)
      const value = item[column.id]

      // Skip if already present in state.validationState
      if (validationState[cellKey]) return

      // Check validation via validateField
      const validationResult = validateField(column, value)
      if (!validationResult.isValid && validationResult.message) {
        allErrors.push({
          cellKey,
          row: rowIndex,
          col: colIndex,
          column,
          item,
          message: validationResult.message,
          value
        })
      }
    })
  })

  return allErrors
}

/**
 * Checks whether there are validation errors in the data
 */
export const hasValidationErrors = (
  items: any[],
  columns: TableColumn[],
  validationState: ValidationState,
  deletedRowIds: Set<string>,
  validateField: (column: TableColumn, value: any) => ValidationResult,
  keyField: string,
): boolean => {
  const errors = getAllValidationErrors(items, columns, validationState, deletedRowIds, validateField, keyField)
  return errors.length > 0
}

/**
 * Hook for obtaining unified validation logic
 */
export const useUnifiedValidation = () => {
  const { validateField } = useValidation()
  const { keyField } = useDataTableContext()

  const getAllErrors = (
    items: any[], 
    columns: TableColumn[], 
    validationState: ValidationState,
    deletedRowIds: Set<string>
  ): ValidationError[] => {
    return getAllValidationErrors(items, columns, validationState, deletedRowIds, validateField, keyField)
  }

  const hasErrors = (
    items: any[], 
    columns: TableColumn[], 
    validationState: ValidationState,
    deletedRowIds: Set<string>
  ): boolean => {
    return hasValidationErrors(items, columns, validationState, deletedRowIds, validateField, keyField)
  }

  return { getAllErrors, hasErrors, validateField }
}
