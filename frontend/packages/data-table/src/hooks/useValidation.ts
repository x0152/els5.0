import { useCallback } from 'react'
import type { TableColumn, ValidationResult } from '../types'

export const useValidation = () => {
  const validateField = useCallback((column: TableColumn, value: any): ValidationResult => {
    // First check custom validation if present
    if (column.validation) {
      return column.validation(value)
    }

    // Built-in validation only for required fields
    if (column.required) {
      if (value === null || value === undefined) {
        return { isValid: false, message: `Field "${column.title}" is required` }
      }
      if (typeof value === 'string' && value.trim() === '') {
        return { isValid: false, message: `Field "${column.title}" cannot be empty` }
      }
      // For dropdown check that the value is not empty
      if (column.type === 'dropdown' && value === '<not set>') {
        return { isValid: false, message: `A value must be selected for field "${column.title}"` }
      }
    }

    // Treat all other cases as valid
    return { isValid: true }
  }, [])

  return { validateField }
}
