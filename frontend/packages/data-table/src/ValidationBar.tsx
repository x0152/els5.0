import { useState, useMemo, useEffect, useRef } from "react";
import { CompactSelection as CS } from "@glideapps/glide-data-grid";
import { useDataTableContext } from "./context/DataTableContext";
import { useDataOperations } from "./hooks/useDataOperations";
import { useUnifiedValidation, cellKeyOf } from "./utils/validationUtils";

export function ValidationBar<T extends Record<string, any>>() {
  const { state, columns, actions, keyField } = useDataTableContext<T>();
  const { filteredItems } = useDataOperations<T>();
  const { getAllErrors } = useUnifiedValidation();
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);

  const allErrors = useMemo(
    () =>
      getAllErrors(
        filteredItems,
        columns,
        state.validationState,
        state.deletedRowIds,
      ),
    [
      filteredItems,
      columns,
      state.validationState,
      state.deletedRowIds,
      getAllErrors,
    ],
  );

  const errorCount = allErrors.length;
  const affectedRows = useMemo(
    () => new Set(allErrors.map((e) => e.row)).size,
    [allErrors],
  );
  const changedCount = state.changedCells.size;
  const hasAnything = errorCount > 0 || changedCount > 0;

  useEffect(() => {
    if (errorCount === 0 && open) setOpen(false);
  }, [errorCount, open]);

  useEffect(() => {
    if (!open) return;
    const onClick = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onClick);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onClick);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  const selectCell = (row: number, col: number) => {
    actions.setGridSelection({
      columns: CS.empty(),
      rows: CS.empty(),
      current: {
        cell: [col, row] as [number, number],
        range: { x: col, y: row, width: 1, height: 1 },
        rangeStack: [],
      },
    });
  };

  const currentCellKey = (() => {
    const cur = state.gridSelection.current?.cell;
    if (!cur) return null;
    const item = filteredItems[cur[1]];
    const column = columns[cur[0]];
    if (!item || !column) return null;
    return cellKeyOf(item[keyField], column.id);
  })();
  const currentIdx = allErrors.findIndex((e) => e.cellKey === currentCellKey);
  const currentError = currentIdx !== -1 ? allErrors[currentIdx] : null;

  const gotoNext = () => {
    if (errorCount === 0) return;
    const next = currentIdx === -1 ? 0 : (currentIdx + 1) % errorCount;
    const e = allErrors[next];
    selectCell(e.row, e.col);
  };
  const gotoPrev = () => {
    if (errorCount === 0) return;
    const prev =
      currentIdx === -1
        ? errorCount - 1
        : (currentIdx - 1 + errorCount) % errorCount;
    const e = allErrors[prev];
    selectCell(e.row, e.col);
  };

  if (!hasAnything) return null;

  return (
    <div
      ref={rootRef}
      className="relative shrink-0 border-t border-gray-200 bg-white"
    >
      {open && (
        <div className="absolute left-0 right-0 bottom-full z-30 border-t border-gray-200 bg-white shadow-[0_-8px_24px_-12px_rgba(0,0,0,0.12)]">
          <div className="flex items-center justify-between px-4 py-2 border-b border-gray-100 bg-gray-50">
            <div className="text-sm font-medium text-gray-900">
              Validation errors
              <span className="ml-2 inline-flex items-center justify-center min-w-[1.5rem] h-5 px-1.5 text-xs font-semibold text-white bg-red-500 rounded-full">
                {errorCount}
              </span>
            </div>
            <button
              type="button"
              onClick={() => setOpen(false)}
              className="text-gray-400 hover:text-gray-600 p-1 rounded"
              aria-label="Close"
            >
              <svg
                className="w-4 h-4"
                viewBox="0 0 20 20"
                fill="currentColor"
                aria-hidden
              >
                <path
                  fillRule="evenodd"
                  d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          </div>
          <div className="max-h-[40vh] overflow-y-auto py-1">
            {allErrors.map((err, idx) => {
              const active = err.cellKey === currentCellKey;
              return (
                <button
                  type="button"
                  key={err.cellKey}
                  onClick={() => {
                    selectCell(err.row, err.col);
                    setOpen(false);
                  }}
                  className={`w-full text-left px-4 py-2 flex items-start gap-3 border-l-2 transition-colors ${
                    active
                      ? "bg-red-50 border-red-500"
                      : "border-transparent hover:bg-gray-50"
                  }`}
                >
                  <span className="mt-0.5 inline-flex items-center justify-center w-5 h-5 text-[10px] font-semibold text-red-700 bg-red-100 rounded">
                    {idx + 1}
                  </span>
                  <span className="flex-1 min-w-0">
                    <span className="flex items-center gap-2 text-xs text-gray-500">
                      <span className="font-medium text-gray-700">
                        {err.column?.title}
                      </span>
                      <span>•</span>
                      <span>row {err.row + 1}</span>
                      {err.value !== undefined && err.value !== null && err.value !== "" && (
                        <>
                          <span>•</span>
                          <span className="truncate max-w-[180px]">
                            "{String(err.value)}"
                          </span>
                        </>
                      )}
                    </span>
                    <span className="block text-sm text-gray-900 mt-0.5">
                      {err.message}
                    </span>
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      )}

      <div className="flex items-center justify-between px-4 py-2 text-sm">
        <div className="flex items-center gap-3 min-w-0">
          {errorCount > 0 ? (
            <span className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-red-50 text-red-700 border border-red-200">
              <svg
                className="w-4 h-4"
                viewBox="0 0 20 20"
                fill="currentColor"
                aria-hidden
              >
                <path
                  fillRule="evenodd"
                  d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l6.28 11.17c.75 1.335-.213 2.981-1.743 2.981H3.72c-1.53 0-2.493-1.646-1.743-2.98l6.28-11.171zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-7a1 1 0 00-1 1v3a1 1 0 102 0V7a1 1 0 00-1-1z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="font-medium">
                {errorCount}{" "}
                {pluralize(errorCount, "error", "errors")}
              </span>
              {affectedRows > 0 && (
                <span className="text-red-500/80">
                  · in {affectedRows}{" "}
                  {pluralize(affectedRows, "row", "rows")}
                </span>
              )}
            </span>
          ) : (
            <span className="inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-green-50 text-green-700 border border-green-200">
              <svg
                className="w-4 h-4"
                viewBox="0 0 20 20"
                fill="currentColor"
                aria-hidden
              >
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="font-medium">Data is valid</span>
            </span>
          )}

          {currentError ? (
            <span className="flex items-center gap-2 min-w-0 text-gray-700">
              <span className="text-gray-400">·</span>
              <span className="font-medium text-gray-900 shrink-0">
                {currentError.column?.title}
              </span>
              <span className="text-gray-400 shrink-0">
                (row {currentError.row + 1}):
              </span>
              <span className="truncate text-red-700">
                {currentError.message}
              </span>
              {currentError.value !== undefined &&
                currentError.value !== null &&
                currentError.value !== "" && (
                  <span className="hidden lg:inline text-xs text-gray-500 bg-gray-100 px-1.5 py-0.5 rounded truncate max-w-[180px]">
                    "{String(currentError.value)}"
                  </span>
                )}
            </span>
          ) : errorCount > 0 ? (
            <span className="text-gray-500 truncate hidden md:inline">
              · select a cell with an error or press
              <span className="mx-1 inline-flex items-center gap-0.5 align-[-1px]">
                <kbd className="px-1 py-0.5 text-[10px] font-semibold bg-gray-100 border border-gray-300 rounded">
                  →
                </kbd>
              </span>
            </span>
          ) : changedCount > 0 ? (
            <span className="text-gray-500 truncate">
              · unsaved changes: {changedCount}
            </span>
          ) : null}
        </div>

        <div className="flex items-center gap-1 shrink-0">
          {errorCount > 0 && (
            <>
              <button
                type="button"
                onClick={gotoPrev}
                title="Previous error"
                className="p-1.5 rounded hover:bg-gray-100 text-gray-500"
              >
                <svg
                  className="w-4 h-4"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                  aria-hidden
                >
                  <path
                    fillRule="evenodd"
                    d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
                    clipRule="evenodd"
                  />
                </svg>
              </button>
              <button
                type="button"
                onClick={gotoNext}
                title="Next error"
                className="p-1.5 rounded hover:bg-gray-100 text-gray-500"
              >
                <svg
                  className="w-4 h-4"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                  aria-hidden
                >
                  <path
                    fillRule="evenodd"
                    d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
                    clipRule="evenodd"
                  />
                </svg>
              </button>
              <button
                type="button"
                onClick={() => setOpen((v) => !v)}
                className={`ml-1 inline-flex items-center gap-1 px-2.5 py-1 rounded border text-xs font-medium transition-colors ${
                  open
                    ? "text-white hover:[background:var(--dt-brand-hover)]"
                    : "bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
                }`}
                style={
                  open
                    ? {
                        background: "var(--dt-brand)",
                        borderColor: "var(--dt-brand)",
                      }
                    : undefined
                }
              >
                {open ? "Hide list" : "Show list"}
                <svg
                  className={`w-3.5 h-3.5 transition-transform ${
                    open ? "rotate-180" : ""
                  }`}
                  viewBox="0 0 20 20"
                  fill="currentColor"
                  aria-hidden
                >
                  <path
                    fillRule="evenodd"
                    d="M14.707 12.707a1 1 0 01-1.414 0L10 9.414l-3.293 3.293a1 1 0 01-1.414-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 010 1.414z"
                    clipRule="evenodd"
                  />
                </svg>
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function pluralize(n: number, one: string, many: string) {
  return n === 1 ? one : many;
}
