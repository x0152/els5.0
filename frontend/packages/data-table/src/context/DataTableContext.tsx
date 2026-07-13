import { createContext, useContext } from "react";
import { CompactSelection } from "@glideapps/glide-data-grid";
import type {
  ValidationState,
  TableColumn,
  EditMode,
  RowThemeCallback,
  CustomCellRenderer,
  CellSelection,
} from "../types";

export interface SortConfig {
  column: string;
  direction: "asc" | "desc" | null;
}

export interface DataTableState<T = any> {
  items: T[];
  originalItems: T[];
  editMode: EditMode;
  validationState: ValidationState;
  changedCells: Set<string>;
  gridSelection: {
    columns: CompactSelection;
    rows: CompactSelection;
    current?: {
      cell: [number, number];
      range: {
        x: number;
        y: number;
        width: number;
        height: number;
      };
      rangeStack: any[];
    };
  };
  newRowIds: Set<string>;
  deletedRowIds: Set<string>;
  activeFilters: { [key: string]: string[] };
  sortConfig: SortConfig;
  headerScrollX: number;
  areSecretsUnlocked?: boolean;
}

export interface DataTableActions<T = any> {
  setItems: (items: T[] | ((prev: T[]) => T[])) => void;
  setOriginalItems: (items: T[]) => void;
  setEditMode: (mode: EditMode) => void;
  setValidationState: (
    state: ValidationState | ((prev: ValidationState) => ValidationState),
  ) => void;
  setChangedCells: (
    cells: Set<string> | ((prev: Set<string>) => Set<string>),
  ) => void;
  setGridSelection: (selection: any) => void;
  setNewRowIds: (
    ids: Set<string> | ((prev: Set<string>) => Set<string>),
  ) => void;
  setDeletedRowIds: (
    ids: Set<string> | ((prev: Set<string>) => Set<string>),
  ) => void;
  setActiveFilters: (
    filters:
      | { [key: string]: string[] }
      | ((prev: { [key: string]: string[] }) => { [key: string]: string[] }),
  ) => void;
  setSortConfig: (
    config: SortConfig | ((prev: SortConfig) => SortConfig),
  ) => void;
  updateColumnWidth: (columnId: string, newWidth: number) => void;
  setHeaderScrollX: (scrollX: number) => void;
}

export interface DataTableContextValue<T = any> {
  state: DataTableState<T>;
  actions: DataTableActions<T>;
  columns: TableColumn[];
  keyField: string;
  onDataChange?: (data: T[]) => void;
  onSelectionChange?: (selectedItems: T[]) => void;
  onCellSelectionChange?: (selectedCells: CellSelection[]) => void;
  onCellActivated?: (item: T, columnId: string) => void;
  onSave?: (data: T[], deletedIds?: string[]) => void;
  getRowTheme?: RowThemeCallback<T>;
  customCellRenderers?: CustomCellRenderer[];
  showSaveControls?: boolean;
  showRowMarkers?: boolean;
  allowAddRows?: boolean;
  brandColor: string;
}

const DataTableContext = createContext<DataTableContextValue | undefined>(
  undefined,
);

export function DataTableProvider<T = any>({
  children,
  value,
}: {
  children: React.ReactNode;
  value: DataTableContextValue<T>;
}) {
  return (
    <DataTableContext.Provider value={value}>
      {children}
    </DataTableContext.Provider>
  );
}

export function useDataTableContext<T = any>(): DataTableContextValue<T> {
  const context = useContext(DataTableContext);
  if (!context) {
    throw new Error(
      "useDataTableContext must be used within a DataTableProvider",
    );
  }
  return context as DataTableContextValue<T>;
}
