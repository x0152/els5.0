import type { TableColumn, ValidationResult, ValidationState } from "../types";
import { cellKeyOf } from "./validationUtils";

/**
 * Runs the same validation logic used live in the grid to seed the initial
 * validation state. Keeps DataTable.tsx focused on orchestration.
 */
export function computeInitialValidationState<T extends Record<string, any>>(
  data: T[],
  columns: TableColumn[],
  keyField: string,
): ValidationState {
  if (data.length === 0 || columns.length === 0) return {};

  const state: ValidationState = {};

  data.forEach((item) => {
    columns.forEach((column) => {
      if (!column.validation && !column.required) return;

      const value = item[column.id];
      const result = validateOne(column, value);

      const cellKey = cellKeyOf(item[keyField], column.id);
      state[cellKey] = {
        isValid: result.isValid,
        isChanged: false,
        message: result.message,
      };
    });
  });

  return state;
}

function validateOne(
  column: TableColumn,
  value: unknown,
): ValidationResult {
  if (column.validation) return column.validation(value);

  if (!column.required) return { isValid: true };

  if (value === null || value === undefined) {
    return {
      isValid: false,
      message: `Field "${column.title}" is required`,
    };
  }
  if (typeof value === "string" && value.trim() === "") {
    return {
      isValid: false,
      message: `Field "${column.title}" cannot be empty`,
    };
  }
  if (column.type === "dropdown" && value === "<not set>") {
    return {
      isValid: false,
      message: `A value must be selected for field "${column.title}"`,
    };
  }
  return { isValid: true };
}
