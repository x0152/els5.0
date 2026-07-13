import { useState } from "react";
import { ConfirmDialog } from "@els/ui";
import { useDataTableContext } from "./context/DataTableContext";
import { useDataOperations } from "./hooks";

export function SelectionInfo<T extends Record<string, any>>() {
  const { state, allowAddRows = true } = useDataTableContext<T>();
  const { filteredItems, clearSelection, deleteSelectedRows, addItem } =
    useDataOperations<T>();
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  const selectedCount = state.gridSelection.rows.length;
  const totalCount = state.items.length;
  const filteredCount = filteredItems.length;
  const activeFiltersCount = Object.keys(state.activeFilters).length;
  const hasSelection = selectedCount > 0;
  const isEdit = state.editMode === "edit";

  const showAdd = isEdit && allowAddRows && !hasSelection;
  const showSelectionActions = isEdit && hasSelection;

  return (
    <div className="flex items-center justify-between gap-3 px-4 py-2 bg-gray-50 border-t border-gray-200 text-sm">
      <div className="flex items-center gap-2 text-gray-600 min-w-0 tabular-nums">
        <span>
          Total:{" "}
          <span className="font-semibold text-gray-900">{totalCount}</span>
        </span>
        {activeFiltersCount > 0 && (
          <>
            <span className="text-gray-300">·</span>
            <span>
              Filtered:{" "}
              <span className="font-semibold text-gray-900">
                {filteredCount}
              </span>
            </span>
          </>
        )}
        {hasSelection && (
          <>
            <span className="text-gray-300">·</span>
            <span>
              Selected:{" "}
              <span className="font-semibold text-gray-900">
                {selectedCount}
              </span>
            </span>
          </>
        )}
      </div>

      <div className="flex items-center gap-1.5 shrink-0">
        {showAdd && (
          <button
            onClick={addItem}
            className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded-md border border-gray-300 bg-white text-gray-700 hover:bg-gray-100"
            title="Add row"
          >
            <svg
              className="w-3.5 h-3.5"
              viewBox="0 0 20 20"
              fill="currentColor"
              aria-hidden
            >
              <path
                fillRule="evenodd"
                d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z"
                clipRule="evenodd"
              />
            </svg>
            Add
          </button>
        )}

        {showSelectionActions && (
          <>
            <button
              onClick={clearSelection}
              className="inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-md border border-gray-300 bg-white text-gray-700 hover:bg-gray-100"
            >
              Reset
            </button>
            <button
              onClick={() => setConfirmingDelete(true)}
              className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded-md bg-red-600 text-white hover:bg-red-700"
            >
              <svg
                className="w-3.5 h-3.5"
                viewBox="0 0 20 20"
                fill="currentColor"
                aria-hidden
              >
                <path
                  fillRule="evenodd"
                  d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm4 0a1 1 0 112 0v6a1 1 0 11-2 0V8z"
                  clipRule="evenodd"
                />
              </svg>
              Delete ({selectedCount})
            </button>
          </>
        )}
      </div>

      {confirmingDelete && (
        <ConfirmDialog
          title="Delete records"
          description={`Delete ${selectedCount} selected record${selectedCount === 1 ? "" : "s"}? This cannot be undone.`}
          onConfirm={() => {
            deleteSelectedRows();
            setConfirmingDelete(false);
          }}
          onClose={() => setConfirmingDelete(false)}
        />
      )}
    </div>
  );
}
