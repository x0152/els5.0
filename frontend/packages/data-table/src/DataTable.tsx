import "@glideapps/glide-data-grid/dist/index.css";
import "@glideapps/glide-data-grid-cells/dist/index.css";
import { useState, useMemo, useEffect, useCallback } from "react";
import { CompactSelection as CS } from "@glideapps/glide-data-grid";
import { DataTableProvider } from "./context/DataTableContext";
import type { SortConfig } from "./context/DataTableContext";
import { DataTableCore } from "./DataTableCore";
import { TableControls } from "./TableControls";
import { SelectionInfo } from "./SelectionInfo";
import { ValidationBar } from "./ValidationBar";
import { FilterBadges } from "./FilterBadges";
import { buildBrandStyle } from "./utils/brand";
import { computeInitialValidationState } from "./utils/initialValidation";
import type { DataTableProps, EditMode, ValidationState } from "./types";

export function DataTable<T extends Record<string, any>>({
  data,
  columns,
  keyField,
  onDataChange,
  onSave,
  className = "",
  title,
  toolbar,
  showControls = true,
  showRowMarkers = true,
  showSaveControls = true,
  showValidationPanel = true,
  editMode: controlledEditMode = "toggle",
  allowAddRows = true,
  onSelectionChange,
  onCellSelectionChange,
  onCellActivated,
  getRowTheme,
  customCellRenderers = [],
  areSecretsUnlocked,
  brandColor = "#111827",
}: DataTableProps<T>) {
  const [items, setItems] = useState<T[]>(data);
  const [originalItems, setOriginalItems] = useState<T[]>(data);
  const [editMode, setEditMode] = useState<EditMode>(
    controlledEditMode === "never" ? "read" : "edit",
  );
  const [validationState, setValidationState] = useState<ValidationState>({});
  const [changedCells, setChangedCells] = useState<Set<string>>(new Set());
  const [gridSelection, setGridSelection] = useState({
    columns: CS.empty(),
    rows: CS.empty(),
  });
  const [newRowIds, setNewRowIds] = useState<Set<string>>(new Set());
  const [deletedRowIds, setDeletedRowIds] = useState<Set<string>>(new Set());
  const [activeFilters, setActiveFilters] = useState<{
    [key: string]: string[];
  }>({});
  const [sortConfig, setSortConfig] = useState<SortConfig>({
    column: "",
    direction: null,
  });
  const [currentColumns, setCurrentColumns] = useState(columns);
  const [headerScrollX, setHeaderScrollX] = useState<number>(0);

  const handleSetItems = useCallback(
    (newItems: T[] | ((prev: T[]) => T[])) => {
      setItems(newItems);
    },
    [],
  );

  useEffect(() => {
    setItems(data);
    setOriginalItems(data);
    setChangedCells(new Set());
    setNewRowIds(new Set());
    setDeletedRowIds(new Set());
    setValidationState(computeInitialValidationState(data, columns, keyField));
  }, [data, columns, keyField]);

  useEffect(() => {
    setCurrentColumns(columns);
  }, [columns]);

  const updateColumnWidth = useCallback(
    (columnId: string, newWidth: number) => {
      setCurrentColumns((prev) =>
        prev.map((col) =>
          col.id === columnId ? { ...col, width: newWidth } : col,
        ),
      );
    },
    [],
  );

  const contextValue = useMemo(
    () => ({
      state: {
        items,
        originalItems,
        editMode,
        validationState,
        changedCells,
        gridSelection,
        newRowIds,
        deletedRowIds,
        activeFilters,
        sortConfig,
        headerScrollX,
        areSecretsUnlocked,
      },
      actions: {
        setItems: handleSetItems,
        setOriginalItems,
        setEditMode,
        setValidationState,
        setChangedCells,
        setGridSelection,
        setNewRowIds,
        setDeletedRowIds,
        setActiveFilters,
        setSortConfig,
        updateColumnWidth,
        setHeaderScrollX,
      },
      columns: currentColumns,
      keyField,
      onDataChange,
      onSelectionChange,
      onCellSelectionChange,
      onCellActivated,
      onSave,
      getRowTheme,
      customCellRenderers,
      showSaveControls,
      showRowMarkers,
      allowAddRows,
      brandColor,
    }),
    [
      items,
      originalItems,
      editMode,
      validationState,
      changedCells,
      gridSelection,
      newRowIds,
      deletedRowIds,
      activeFilters,
      sortConfig,
      headerScrollX,
      areSecretsUnlocked,
      currentColumns,
      keyField,
      onDataChange,
      onSelectionChange,
      onCellSelectionChange,
      onCellActivated,
      onSave,
      handleSetItems,
      updateColumnWidth,
      getRowTheme,
      customCellRenderers,
      showSaveControls,
      showRowMarkers,
      allowAddRows,
      brandColor,
    ],
  );

  const showControlsSlot = showControls && controlledEditMode !== "never";
  const brandStyle = useMemo(() => buildBrandStyle(brandColor), [brandColor]);

  return (
    <DataTableProvider value={contextValue}>
      <div
        className={`h-full flex flex-col gap-2 ${className}`}
        style={brandStyle}
      >
        <div className="flex-1 bg-white rounded-lg border border-gray-200 overflow-hidden flex flex-col min-w-0 min-h-0 shadow-sm">
          <div className="flex items-center justify-between gap-3 px-3 border-b border-gray-200 bg-gray-50/60 shrink-0 h-12">
            <div className="flex flex-1 min-w-0 items-center gap-3">
              {title && (
                <h2 className="shrink-0 text-base font-semibold text-gray-900 truncate">
                  {title}
                </h2>
              )}
              <FilterBadges />
            </div>
            <div className="flex shrink-0 items-center gap-2">
              {toolbar}
              {showControlsSlot && <TableControls />}
            </div>
          </div>
          <DataTableCore />
          <SelectionInfo />
          {showValidationPanel && <ValidationBar />}
        </div>
      </div>
    </DataTableProvider>
  );
}
