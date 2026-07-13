import type { ReactNode } from "react";

export interface ValidationResult {
  isValid: boolean;
  message?: string;
}

export interface ValidationState {
  [key: string]: {
    isValid: boolean;
    isChanged: boolean;
    message?: string;
  };
}

export type EditMode = "read" | "edit";

export type ColumnType =
  | "text"
  | "email"
  | "number"
  | "boolean"
  | "dropdown"
  | "drilldown"
  | "date"
  | "rowid"
  | "uri"
  | string;

export interface SortConfig {
  column: string;
  direction: "asc" | "desc" | null;
}

export interface TableColumn {
  id: string;
  title: string;
  width: number;
  type: ColumnType;
  required?: boolean;
  options?: string[];
  validation?: (value: any) => ValidationResult;
  default?: any;
  readonly?: boolean;
  align?: "left" | "center" | "right";
}

export type RowThemeCallback<T = any> = (
  item: T,
  rowIndex: number,
) =>
  | Partial<{
      bgCell: string;
      bgCellMedium: string;
      bgHeader: string;
      bgHeaderHasFocus: string;
      bgHeaderHovered: string;
      textDark: string;
      textMedium: string;
      textLight: string;
      textBubble: string;
      borderColor: string;
      drilldownBorder: string;
      linkColor: string;
      cellHorizontalPadding: number;
      cellVerticalPadding: number;
    }>
  | undefined;

export interface CustomCellRenderer {
  kind: string;
  renderer: any;
  cellCreator: (value: any, readonly?: boolean) => any;
  valueExtractor?: (cellData: any) => any;
  copyDataExtractor?: (cellData: any) => string;
}

export interface CellSelection {
  rowId: string;
  columnId: string;
}

export interface ServerValidationError {
  rowId: string;
  columnId: string;
  messages: string[];
}

export interface DataTableProps<T = any> {
  data: T[];
  columns: TableColumn[];
  keyField: string;
  areSecretsUnlocked?: boolean;
  onDataChange?: (data: T[]) => void;
  onSave?: (
    data: T[],
    deletedIds?: string[],
  ) => Promise<ServerValidationError[] | null> | ServerValidationError[] | null;
  className?: string;
  title?: ReactNode;
  toolbar?: ReactNode;
  showFilters?: boolean;
  showControls?: boolean;
  showRowMarkers?: boolean;
  showSaveControls?: boolean;
  showValidationPanel?: boolean;
  editMode?: "toggle" | "always" | "never";
  allowAddRows?: boolean;
  onSelectionChange?: (selectedItems: T[]) => void;
  onCellSelectionChange?: (selectedCells: CellSelection[]) => void;
  onCellActivated?: (item: T, columnId: string) => void;
  getRowTheme?: RowThemeCallback<T>;
  customCellRenderers?: CustomCellRenderer[];
  brandColor?: string;
}
