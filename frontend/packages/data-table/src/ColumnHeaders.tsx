import { useState, useRef, useCallback, useEffect } from "react";
import { useDataTableContext } from "./context/DataTableContext";
import { useDataOperations } from "./hooks";
import { ColumnFilter } from "./ColumnFilter";
import { MIN_COLUMN_WIDTH, ROW_MARKER_WIDTH } from "./utils/constants";
import type { SortConfig } from "./context/DataTableContext";

type SortDirection = SortConfig["direction"];

interface ResizingState {
  columnId: string;
  startX: number;
  startWidth: number;
}

interface FilterPosition {
  x: number;
  y: number;
  anchorBottom: number;
}

function nextSortDirection(
  current: SortDirection,
  isSameColumn: boolean,
): SortDirection {
  if (!isSameColumn) return "asc";
  if (current === "asc") return "desc";
  if (current === "desc") return null;
  return "asc";
}

function SortIcon({ direction }: { direction: SortDirection | "none" }) {
  if (direction === "asc") {
    return (
      <svg
        className="w-3.5 h-3.5 text-gray-900"
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
    );
  }
  if (direction === "desc") {
    return (
      <svg
        className="w-3.5 h-3.5 text-gray-900"
        viewBox="0 0 20 20"
        fill="currentColor"
        aria-hidden
      >
        <path
          fillRule="evenodd"
          d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z"
          clipRule="evenodd"
        />
      </svg>
    );
  }
  return (
    <svg
      className="w-3.5 h-3.5 text-gray-300 group-hover:text-gray-500"
      viewBox="0 0 20 20"
      fill="currentColor"
      aria-hidden
    >
      <path
        fillRule="evenodd"
        d="M10 3a1 1 0 01.707.293l3 3a1 1 0 01-1.414 1.414L10 5.414 7.707 7.707a1 1 0 01-1.414-1.414l3-3A1 1 0 0110 3zm-3.707 9.293a1 1 0 011.414 0L10 14.586l2.293-2.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z"
        clipRule="evenodd"
      />
    </svg>
  );
}

function FilterIcon({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 20 20"
      fill="currentColor"
      aria-hidden
    >
      <path
        fillRule="evenodd"
        d="M3 5a1 1 0 011-1h12a1 1 0 01.8 1.6l-4.8 6.4V16a1 1 0 01-1.45.894l-2-1A1 1 0 018 15v-3l-4.8-6.4A1 1 0 013 5z"
        clipRule="evenodd"
      />
    </svg>
  );
}

export function ColumnHeaders<T extends Record<string, any>>() {
  const { columns, state, actions, brandColor, showRowMarkers = true } = useDataTableContext<T>();
  const { filteredItems, selectAllRows, clearSelection } =
    useDataOperations<T>();
  const [activeFilter, setActiveFilter] = useState<string | null>(null);
  const [filterPosition, setFilterPosition] = useState<FilterPosition>({
    x: 0,
    y: 0,
    anchorBottom: 0,
  });
  const [resizing, setResizing] = useState<ResizingState | null>(null);
  const headerRefs = useRef<{ [key: string]: HTMLButtonElement | null }>({});

  const handleResizeStart = useCallback(
    (columnId: string, event: React.MouseEvent) => {
      event.preventDefault();
      const column = columns.find((col) => col.id === columnId);
      if (!column) return;
      setResizing({
        columnId,
        startX: event.clientX,
        startWidth: column.width,
      });
    },
    [columns],
  );

  const handleResizeMove = useCallback(
    (event: MouseEvent) => {
      if (!resizing) return;
      const deltaX = event.clientX - resizing.startX;
      const newWidth = Math.max(
        MIN_COLUMN_WIDTH,
        resizing.startWidth + deltaX,
      );
      actions.updateColumnWidth(resizing.columnId, newWidth);
    },
    [resizing, actions],
  );

  const handleResizeEnd = useCallback(() => setResizing(null), []);

  useEffect(() => {
    if (!resizing) return;
    document.addEventListener("mousemove", handleResizeMove);
    document.addEventListener("mouseup", handleResizeEnd);
    document.body.style.userSelect = "none";
    document.body.style.cursor = "col-resize";
    return () => {
      document.removeEventListener("mousemove", handleResizeMove);
      document.removeEventListener("mouseup", handleResizeEnd);
      document.body.style.userSelect = "";
      document.body.style.cursor = "";
    };
  }, [resizing, handleResizeMove, handleResizeEnd]);

  const handleFilterClick = (
    columnId: string,
    event: React.MouseEvent<HTMLButtonElement>,
  ) => {
    const rect = event.currentTarget.getBoundingClientRect();
    setFilterPosition({
      x: rect.left,
      y: rect.bottom + 6,
      anchorBottom: rect.top,
    });
    setActiveFilter((current) => (current === columnId ? null : columnId));
  };

  const hasActiveFilter = (columnId: string) =>
    (state.activeFilters[columnId]?.length ?? 0) > 0;

  const handleSort = (columnId: string) => {
    const isSame = state.sortConfig.column === columnId;
    const direction = nextSortDirection(state.sortConfig.direction, isSame);
    actions.setSortConfig({
      column: direction === null ? "" : columnId,
      direction,
    });
  };

  const getSortDirection = (columnId: string): SortDirection | "none" => {
    if (state.sortConfig.column !== columnId) return "none";
    return state.sortConfig.direction;
  };

  const totalRows = filteredItems.length;
  const selectedRows = state.gridSelection.rows.length;
  const allSelected = totalRows > 0 && selectedRows === totalRows;
  const someSelected = selectedRows > 0 && !allSelected;
  const toggleAllRows = () => {
    if (allSelected) {
      clearSelection();
    } else {
      selectAllRows();
    }
  };

  return (
    <>
      <div className="flex bg-gray-50 border-b border-gray-200 overflow-hidden min-w-0">
        {showRowMarkers && (
          <div
            className="z-10 sticky left-0 bg-gray-50 border-r border-gray-200 flex-shrink-0 flex items-center justify-center"
            style={{ width: ROW_MARKER_WIDTH }}
          >
            <input
              type="checkbox"
              checked={allSelected}
              ref={(el) => {
                if (el) el.indeterminate = someSelected;
              }}
              onChange={toggleAllRows}
              disabled={totalRows === 0}
              className="rounded border-gray-300 cursor-pointer focus:ring-offset-0 focus:[box-shadow:0_0_0_3px_var(--dt-brand-soft)] disabled:cursor-not-allowed disabled:opacity-40"
              style={{ accentColor: brandColor }}
              title={allSelected ? "Deselect all" : "Select all rows"}
              aria-label={allSelected ? "Deselect all" : "Select all"}
            />
          </div>
        )}

        <div
          className="flex flex-1 min-w-0"
          style={{
            transform: `translateX(${-state.headerScrollX}px)`,
            transition: "none",
          }}
        >
          {columns.map((column) => {
            const isFiltered = hasActiveFilter(column.id);
            return (
              <div key={column.id} className="relative">
                <div
                  className="flex items-center justify-between gap-1 px-3 py-2 border-r border-gray-200 bg-gray-50"
                  style={{
                    width: column.width,
                    minWidth: column.width,
                    maxWidth: column.width,
                  }}
                >
                  <button
                    type="button"
                    onClick={() => handleSort(column.id)}
                    className="group flex items-center gap-1.5 flex-1 min-w-0 rounded px-1 -mx-1 py-0.5 hover:bg-gray-200/60 transition-colors"
                    title={`Sort by “${column.title}”`}
                  >
                    <span className="text-sm font-medium text-gray-700 truncate">
                      {column.title}
                      {column.required && (
                        <span className="text-red-500 ml-0.5">*</span>
                      )}
                    </span>
                    <SortIcon direction={getSortDirection(column.id)} />
                  </button>

                  <button
                    ref={(el) => {
                      headerRefs.current[column.id] = el;
                    }}
                    onClick={(e) => handleFilterClick(column.id, e)}
                    className={`shrink-0 p-1 rounded transition-colors ${
                      isFiltered
                        ? "text-white hover:[background:var(--dt-brand-hover)]"
                        : "text-gray-400 hover:text-gray-700 hover:bg-gray-200"
                    }`}
                    style={
                      isFiltered ? { background: "var(--dt-brand)" } : undefined
                    }
                    title={`Filter “${column.title}”`}
                    aria-label={`Filter ${column.title}`}
                  >
                    <FilterIcon className="w-3.5 h-3.5" />
                  </button>
                </div>

                <div
                  className="absolute top-0 right-0 w-1.5 h-full cursor-col-resize z-10 group flex justify-end"
                  title="Drag to resize"
                  onMouseDown={(e) => handleResizeStart(column.id, e)}
                >
                  <div
                    className={`w-0.5 h-full transition-colors ${
                      resizing?.columnId === column.id
                        ? "bg-gray-900"
                        : "bg-transparent group-hover:bg-gray-400"
                    }`}
                  />
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {activeFilter && (
        <ColumnFilter
          column={columns.find((col) => col.id === activeFilter)!}
          onClose={() => setActiveFilter(null)}
          position={filterPosition}
        />
      )}
    </>
  );
}
