import { useMemo, useCallback } from "react";
import { CompactSelection as CS } from "@glideapps/glide-data-grid";
import { useDataTableContext } from "../context/DataTableContext";
import dayjs from "dayjs";

export const useDataOperations = <T extends Record<string, any>>() => {
  const { state, actions, columns, keyField, onDataChange, onSelectionChange } =
    useDataTableContext<T>();

  const filteredItems = useMemo(() => {
    const activeFilterEntries = Object.entries(state.activeFilters);

    // Start with all rows (deleted rows will be shown with strikethrough)
    let filtered = state.items;

    // Filtering
    if (activeFilterEntries.length > 0) {
      filtered = filtered.filter((item) => {
        return activeFilterEntries.every(([columnId, filterValues]) => {
          // If a filter is set but nothing is selected — hide everything
          if (filterValues.length === 0) return false;

          const itemValue = item[columnId];

          // Check whether the filter is special
          if (filterValues.length === 1) {
            const filterValue = filterValues[0];

            // Handle "less than or equal" (≤) filter
            if (filterValue.startsWith("≤")) {
              const dateStr = filterValue.substring(1);
              const itemDate = dayjs(itemValue);
              const filterDate = dayjs(dateStr);

              return (
                itemDate.isValid() &&
                filterDate.isValid() &&
                (itemDate.isBefore(filterDate) ||
                  itemDate.isSame(filterDate, "day"))
              );
            }

            // Handle "greater than" (>) filter
            if (filterValue.startsWith(">")) {
              const dateStr = filterValue.substring(1);
              const itemDate = dayjs(itemValue);
              const filterDate = dayjs(dateStr);

              return (
                itemDate.isValid() &&
                filterDate.isValid() &&
                itemDate.isAfter(filterDate)
              );
            }

            // Handle "not equal null" filter (for termination date)
            if (filterValue === "not_null") {
              return (
                itemValue !== null &&
                itemValue !== undefined &&
                itemValue !== ""
              );
            }
          }

          let valueToCheck = "";

          if (itemValue !== undefined && itemValue !== null) {
            if (typeof itemValue === "number") {
              valueToCheck = itemValue.toString();
            } else {
              valueToCheck = String(itemValue);
            }
          }

          return filterValues.includes(valueToCheck);
        });
      });
    }

    // Sorting
    if (state.sortConfig.column && state.sortConfig.direction) {
      const column = columns.find((col) => col.id === state.sortConfig.column);
      if (column) {
        filtered = [...filtered].sort((a, b) => {
          const aValue = a[state.sortConfig.column];
          const bValue = b[state.sortConfig.column];

          // Handle different data types
          let comparison = 0;
          if (column.type === "number") {
            const numA =
              typeof aValue === "number" ? aValue : parseFloat(aValue);
            const numB =
              typeof bValue === "number" ? bValue : parseFloat(bValue);
            const aInvalid = Number.isNaN(numA);
            const bInvalid = Number.isNaN(numB);
            if (aInvalid || bInvalid) {
              return aInvalid && bInvalid ? 0 : aInvalid ? 1 : -1;
            }
            comparison = numA - numB;
          } else if (column.type === "boolean") {
            const boolA = Boolean(aValue);
            const boolB = Boolean(bValue);
            comparison = boolA === boolB ? 0 : boolA ? 1 : -1;
          } else if (
            column.type === "dropdown" ||
            column.type === "drilldown"
          ) {
            const strA = String(aValue != null ? aValue : "").toLowerCase();
            const strB = String(bValue != null ? bValue : "").toLowerCase();
            comparison = strA.localeCompare(strB);
          } else {
            const strA = String(aValue != null ? aValue : "").toLowerCase();
            const strB = String(bValue != null ? bValue : "").toLowerCase();
            comparison = strA.localeCompare(strB);
          }

          return state.sortConfig.direction === "asc"
            ? comparison
            : -comparison;
        });
      }
    }

    return filtered;
  }, [
    state.items,
    state.deletedRowIds,
    state.activeFilters,
    state.sortConfig,
    columns,
  ]);

  const generateNewRow = useCallback(
    (customId?: string) => {
      const newId = customId || Date.now().toString();

      const baseData = Object.fromEntries(
        columns.map((col) => {
          let defaultValue = col.default !== undefined ? col.default : null;

          // If default is a function, call it
          if (typeof defaultValue === "function") {
            defaultValue = defaultValue();
          }

          return [col.id, defaultValue];
        }),
      );

      baseData[keyField] = newId;

      return baseData;
    },
    [columns, keyField],
  );

  const addItem = useCallback(() => {
    const newItem = {
      ...generateNewRow(),
    } as unknown as T;

    const updatedItems = [...state.items, newItem];
    actions.setItems(updatedItems);
    actions.setNewRowIds((prev) => new Set(prev).add(newItem[keyField] as string));
    onDataChange?.(updatedItems);
  }, [state.items, generateNewRow, actions, onDataChange, keyField]);

  const deleteSelectedRows = useCallback(() => {
    const selectedRowsArray = state.gridSelection.rows.toArray();
    if (selectedRowsArray.length === 0) return;

    const selectedItems = selectedRowsArray
      .map((index) => filteredItems[index])
      .filter((item) => item != null);

    // Split into new and existing rows
    const newItemsToDelete = selectedItems.filter((item) =>
      state.newRowIds.has(item[keyField]),
    );
    const existingItemsToDelete = selectedItems.filter(
      (item) => !state.newRowIds.has(item[keyField]),
    );

    // Delete new rows directly
    let updatedItems = state.items;
    if (newItemsToDelete.length > 0) {
      updatedItems = state.items.filter(
        (item) =>
          !newItemsToDelete.find(
            (deleted) => deleted[keyField] === item[keyField],
          ),
      );
      actions.setItems(updatedItems);

      // Remove deleted new rows from newRowIds
      actions.setNewRowIds((prev) => {
        const newSet = new Set(prev);
        newItemsToDelete.forEach((item) => newSet.delete(item[keyField]));
        return newSet;
      });
    }

    // Mark existing rows as deleted
    if (existingItemsToDelete.length > 0) {
      actions.setDeletedRowIds((prev) => {
        const newSet = new Set(prev);
        existingItemsToDelete.forEach((item) => newSet.add(item[keyField]));
        return newSet;
      });
    }

    actions.setGridSelection({ columns: CS.empty(), rows: CS.empty() });
    onSelectionChange?.([]);

    // If new rows were deleted, notify about changes
    if (newItemsToDelete.length > 0) {
      onDataChange?.(updatedItems);
    }
  }, [
    state.gridSelection,
    state.items,
    state.newRowIds,
    filteredItems,
    keyField,
    actions,
    onDataChange,
    onSelectionChange,
  ]);

  const selectAllRows = useCallback(() => {
    const isAllSelected =
      state.gridSelection.rows.length === filteredItems.length;
    const newSelection = {
      columns: CS.empty(),
      rows: isAllSelected
        ? CS.empty()
        : CS.fromSingleSelection([0, filteredItems.length]),
    };
    actions.setGridSelection(newSelection);

    if (onSelectionChange) {
      const selectedItems = isAllSelected ? [] : filteredItems;
      onSelectionChange(selectedItems);
    }
  }, [state.gridSelection, filteredItems, actions, onSelectionChange]);

  const clearSelection = useCallback(() => {
    actions.setGridSelection({ columns: CS.empty(), rows: CS.empty() });
    onSelectionChange?.([]);
  }, [actions, onSelectionChange]);

  return {
    filteredItems,
    addItem,
    deleteSelectedRows,
    selectAllRows,
    clearSelection,
    generateNewRow,
  };
};
