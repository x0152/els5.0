import { useMemo, useCallback } from "react";
import { DataEditor } from "@glideapps/glide-data-grid";
import type { GridColumn, Rectangle, Item } from "@glideapps/glide-data-grid";
import { DropdownCell } from "@glideapps/glide-data-grid-cells";
import { DatePickerCell } from "@glideapps/glide-data-grid-cells";
import { dateRenderer, badgeRenderer } from "./cells";
import { useDataTableContext } from "./context/DataTableContext";
import {
  useDataOperations,
  useCellOperations,
  useCopyPaste,
  useEditMode,
  useValidation,
} from "./hooks";
import { ColumnHeaders } from "./ColumnHeaders";
import { cellRenderer } from "./utils/cellRenderer";
import { ROW_HEIGHT, ROW_MARKER_WIDTH } from "./utils/constants";

export function DataTableCore<T extends Record<string, any>>() {
  const {
    state,
    columns,
    actions,
    getRowTheme,
    customCellRenderers = [],
    allowAddRows = true,
    showRowMarkers = true,
    brandColor,
    onCellActivated,
    keyField,
  } = useDataTableContext<T>();
  const { filteredItems, addItem } = useDataOperations<T>();
  const { getCellContent, onCellEdited } = useCellOperations<T>();
  const { onPaste } = useCopyPaste<T>();
  const { onGridSelectionChange } = useEditMode<T>();
  const { validateField } = useValidation();

  const drawCell = useCallback(
    (args: any, drawContent: () => void) => {
      cellRenderer(
        args,
        state.validationState,
        state.changedCells,
        state.editMode,
        drawContent,
        columns,
        filteredItems,
        validateField,
        state.newRowIds,
        keyField,
      );
    },
    [
      state.validationState,
      state.changedCells,
      state.editMode,
      columns,
      filteredItems,
      validateField,
      state.newRowIds,
      keyField,
    ],
  );

  const handleCellActivated = useCallback(
    (cell: Item) => {
      if (!onCellActivated) return;
      const [col, row] = cell;
      const item = filteredItems[row];
      const column = columns[col];
      if (item && column) onCellActivated(item, column.id);
    },
    [onCellActivated, filteredItems, columns],
  );

  const gridColumns: GridColumn[] = useMemo(
    () =>
      columns.map((col) => ({
        title: col.title,
        width: col.width,
        id: col.id,
      })),
    [columns],
  );

  const onVisibleRegionChanged = useCallback(
    (
      range: Rectangle,
      tx: number,
      _ty: number,
      _extras: { selected?: Item; freezeRegion?: Rectangle },
    ) => {
      // Compute the absolute offset:
      // 1. Sum widths of all hidden columns on the left (up to range.x)
      const hiddenColumnsWidth = columns
        .slice(0, range.x)
        .reduce((sum, col) => sum + col.width, 0);

      // 2. Keep the offset sign: for a negative value the header
      //    shifts in the opposite direction.
      const absoluteScrollX = hiddenColumnsWidth - tx;

      actions.setHeaderScrollX(absoluteScrollX);
    },
    [actions, columns],
  );

  // Create getRowThemeOverride for the Glide Data Grid API
  const getRowThemeOverride = useCallback(
    (row: number) => {
      if (!getRowTheme || row >= filteredItems.length) return undefined;

      const item = filteredItems[row];
      if (!item) return undefined;

      return getRowTheme(item, row);
    },
    [getRowTheme, filteredItems],
  );

  const gridTheme = useMemo(
    () => ({
      accentColor: brandColor,
      accentLight: `color-mix(in srgb, ${brandColor} 12%, transparent)`,
      textDark: "#111827",
      textMedium: "#6b7280",
      textLight: "#9ca3af",
      textBubble: "#111827",
      bgIconHeader: "#6b7280",
      fgIconHeader: "#ffffff",
      textHeader: "#374151",
      textHeaderSelected: "#ffffff",
      bgCell: "#ffffff",
      bgCellMedium: "#ffffff",
      bgHeader: "#f9fafb",
      bgHeaderHasFocus: "#f3f4f6",
      bgHeaderHovered: "#f3f4f6",
      bgBubble: "#f3f4f6",
      bgBubbleSelected: brandColor,
      bgSearchResult: "#fef9c3",
      borderColor: "#e5e7eb",
      horizontalBorderColor: "#f3f4f6",
      drilldownBorder: "#e5e7eb",
      linkColor: brandColor,
      cellHorizontalPadding: 12,
      cellVerticalPadding: 4,
      headerFontStyle: "600 13px",
      baseFontStyle: "13px",
      fontFamily:
        "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
    }),
    [brandColor],
  );

  const hasNoData = state.items.length === 0;
  const hasFilteredOut = !hasNoData && filteredItems.length === 0;

  return (
    <div className="flex-1 min-w-0 bg-white overflow-hidden relative flex flex-col">
      <ColumnHeaders />
      <div className="flex-1 overflow-hidden h-full min-w-0 relative">
        {hasNoData && (
          <div className="absolute inset-0 z-10 flex items-center justify-center pointer-events-none">
            <div className="text-center text-gray-500 text-sm pointer-events-auto">
              <svg
                className="w-10 h-10 mx-auto mb-2 text-gray-300"
                viewBox="0 0 20 20"
                fill="currentColor"
                aria-hidden
              >
                <path
                  fillRule="evenodd"
                  d="M4 5a2 2 0 012-2h8a2 2 0 012 2v10a2 2 0 01-2 2H6a2 2 0 01-2-2V5zm2 0v10h8V5H6z"
                  clipRule="evenodd"
                />
                <path d="M7 7h6v2H7V7zm0 4h6v2H7v-2z" opacity=".4" />
              </svg>
              <p className="font-medium text-gray-700">No data</p>
              {state.editMode === "edit" && allowAddRows && (
                <p className="mt-1 text-xs text-gray-400">
                  Click “+” in the last row to add a record
                </p>
              )}
            </div>
          </div>
        )}
        {hasFilteredOut && (
          <div className="absolute inset-0 z-10 flex items-center justify-center pointer-events-none">
            <div className="text-center text-sm pointer-events-auto">
              <p className="font-medium text-gray-700">
                Nothing found for the current filters
              </p>
              <button
                type="button"
                onClick={() => actions.setActiveFilters({})}
                className="mt-2 inline-flex items-center gap-1 px-3 py-1 rounded-md text-xs font-medium border border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
              >
                Reset filters
              </button>
            </div>
          </div>
        )}
        <div className={`h-full ${hasNoData || hasFilteredOut ? "blur-[2px]" : ""}`}>
        <DataEditor
          getCellContent={getCellContent}
          columns={gridColumns}
          rows={filteredItems.length}
          onCellEdited={onCellEdited}
          smoothScrollX={true}
          smoothScrollY={true}
          rowHeight={ROW_HEIGHT}
          headerHeight={0}
          height="100%"
          theme={gridTheme}
          keybindings={{
            downFill: true,
            rightFill: true,
            selectAll: true,
            selectRow: true,
            selectColumn: true,
            copy: true,
            paste: true,
            search: true,
          }}
          customRenderers={[
            DropdownCell,
            DatePickerCell,
            dateRenderer,
            badgeRenderer,
            ...customCellRenderers.map((ccr) => ccr.renderer),
          ]}
          drawCell={drawCell}
          trailingRowOptions={
            state.editMode === "edit" && allowAddRows
              ? {
                  hint: "",
                  sticky: false,
                  tint: true,
                }
              : undefined
          }
          onRowAppended={state.editMode === "edit" && allowAddRows ? addItem : undefined}
          rowMarkers={showRowMarkers ? "both" : "none"}
          gridSelection={state.gridSelection}
          onGridSelectionChange={onGridSelectionChange}
          rowMarkerWidth={showRowMarkers ? ROW_MARKER_WIDTH : 0}
          fillHandle={state.editMode === "edit"}
          onPaste={onPaste}
          getCellsForSelection={true}
          rightElementProps={{
            fill: state.editMode === "edit",
            sticky: false,
          }}
          onVisibleRegionChanged={onVisibleRegionChanged}
          getRowThemeOverride={getRowThemeOverride}
          onCellActivated={handleCellActivated}
        />
        </div>
      </div>
    </div>
  );
}
