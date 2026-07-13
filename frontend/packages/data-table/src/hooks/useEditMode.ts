import { useCallback, useState } from "react";
import { CompactSelection as CS } from "@glideapps/glide-data-grid";
import type { GridSelection } from "@glideapps/glide-data-grid";
import { useDataTableContext } from "../context/DataTableContext";
import { useDataOperations } from "./useDataOperations";
import { useUnifiedValidation, cellKeyOf } from "../utils/validationUtils";
import type { ServerValidationError } from "../types";

export const useEditMode = <T extends Record<string, any>>() => {
  const {
    state,
    actions,
    columns,
    keyField,
    onSelectionChange,
    onCellSelectionChange,
    onSave,
  } = useDataTableContext<T>();
  const { hasErrors, getAllErrors } = useUnifiedValidation();
  const { filteredItems } = useDataOperations<T>();
  const [saveError, setSaveError] = useState<string | null>(null);

  const toggleEditMode = useCallback(() => {
    if (state.editMode === "edit") {
      actions.setEditMode("read");
    } else {
      actions.setOriginalItems([...state.items]);
      actions.setEditMode("edit");
    }

    actions.setGridSelection({ columns: CS.empty(), rows: CS.empty() });
    onSelectionChange?.([]);
  }, [state.editMode, state.items, actions, onSelectionChange]);

  const saveChanges = useCallback(async () => {
    if (state.editMode === "edit") {
      if (Object.keys(state.activeFilters).length > 0) {
        return;
      }
      const hasInvalidCells = hasErrors(
        state.items,
        columns,
        state.validationState,
        state.deletedRowIds,
      );
      if (hasInvalidCells) {
        return;
      }

      const finalItems = state.items.filter(
        (item) => !state.deletedRowIds.has(item[keyField]),
      );

      setSaveError(null);
      if (onSave) {
        try {
          const deletedIds = Array.from(state.deletedRowIds);
          const result = await onSave(finalItems, deletedIds);

          if (Array.isArray(result) && result.length > 0) {
            const newValidationState = { ...state.validationState };

            result.forEach((error: ServerValidationError) => {
              newValidationState[cellKeyOf(error.rowId, error.columnId)] = {
                isValid: false,
                isChanged: false,
                message: error.messages.join("; "),
              };
            });

            actions.setValidationState(newValidationState);
            return;
          }
        } catch (err) {
          setSaveError(err instanceof Error ? err.message : "Save failed");
          return;
        }
      }

      actions.setOriginalItems([...finalItems]);
      actions.setItems([...finalItems]);

      actions.setChangedCells(new Set());
      actions.setValidationState({});
      actions.setNewRowIds(new Set());
      actions.setDeletedRowIds(new Set());
      actions.setGridSelection({ columns: CS.empty(), rows: CS.empty() });
      onSelectionChange?.([]);
    }
  }, [
    state.editMode,
    state.validationState,
    state.items,
    state.deletedRowIds,
    state.activeFilters,
    columns,
    keyField,
    actions,
    onSelectionChange,
    onSave,
    hasErrors,
  ]);

  const cancelChanges = useCallback(() => {
    if (state.editMode === "edit") {
      setSaveError(null);
      // Restore original data
      actions.setItems([...state.originalItems]);

      // Fully clear all edit states
      // actions.setEditMode("read");
      actions.setChangedCells(new Set());
      actions.setValidationState({});
      actions.setNewRowIds(new Set());
      actions.setDeletedRowIds(new Set());
      actions.setGridSelection({ columns: CS.empty(), rows: CS.empty() });

      onSelectionChange?.([]);
    }
  }, [state.editMode, state.originalItems, actions, onSelectionChange]);

  const onGridSelectionChange = useCallback(
    (selection: GridSelection) => {
      actions.setGridSelection(selection);

      if (onSelectionChange) {
        const selectedIndexes = selection.rows.toArray();
        const selectedItems = selectedIndexes
          .map((index) => filteredItems[index])
          .filter(Boolean);
        onSelectionChange(selectedItems);
      }

      if (onCellSelectionChange) {
        const selectedCells = [];
        const selectedRows = selection.rows.toArray();
        const selectedCols = selection.columns.toArray();

        for (const rowIndex of selectedRows) {
          const item = filteredItems[rowIndex];
          if (item) {
            for (const colIndex of selectedCols) {
              const column = columns[colIndex];
              if (column) {
                selectedCells.push({
                  rowId: item[keyField],
                  columnId: column.id,
                });
              }
            }
          }
        }

        if (selection.current?.range) {
          const { x, y, width, height } = selection.current.range;
          for (let row = y; row < y + height; row++) {
            const item = filteredItems[row];
            if (item) {
              for (let col = x; col < x + width; col++) {
                const column = columns[col];
                if (column) {
                  const exists = selectedCells.some(
                    (cell) =>
                      cell.rowId === item[keyField] &&
                      cell.columnId === column.id,
                  );
                  if (!exists) {
                    selectedCells.push({
                      rowId: item[keyField],
                      columnId: column.id,
                    });
                  }
                }
              }
            }
          }
        }

        onCellSelectionChange(selectedCells);
      }
    },
    [
      filteredItems,
      columns,
      keyField,
      actions,
      onSelectionChange,
      onCellSelectionChange,
    ],
  );

  // Use unified validation logic accounting for deleted rows
  const allValidationErrors = getAllErrors(
    filteredItems,
    columns,
    state.validationState,
    state.deletedRowIds,
  );
  const hasValidationErrors = allValidationErrors.length > 0;
  const errorCount = allValidationErrors.length;

  const hasActiveFilters = Object.keys(state.activeFilters).length > 0;
  const canSave = !hasValidationErrors && !hasActiveFilters;
  const hasChanges =
    state.changedCells.size > 0 ||
    state.newRowIds.size > 0 ||
    state.deletedRowIds.size > 0;

  const clearAllFilters = useCallback(() => {
    actions.setActiveFilters({});
  }, [actions]);

  return {
    toggleEditMode,
    saveChanges,
    cancelChanges,
    onGridSelectionChange,
    canSave,
    hasChanges,
    hasValidationErrors,
    errorCount,
    hasActiveFilters,
    clearAllFilters,
    saveError,
  };
};
