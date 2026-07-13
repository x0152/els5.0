import { GridCellKind } from "@glideapps/glide-data-grid";
import type { GridCell, Rectangle, Theme } from "@glideapps/glide-data-grid";
import type { ValidationState, TableColumn } from '../types';
import { cellKeyOf } from './validationUtils';

export const cellRenderer = (
  args: {
    ctx: CanvasRenderingContext2D;
    cell: GridCell;
    theme: Theme;
    rect: Rectangle;
    col: number;
    row: number;
    hoverAmount: number;
    hoverX: number | undefined;
    hoverY: number | undefined;
    highlighted: boolean;
  },
  validationState: ValidationState,
  changedCells: Set<string>,
  editMode: string,
  drawContent: () => void,
  columns: TableColumn[],
  items: any[],
  validateField: (column: TableColumn, value: any) => { isValid: boolean; message?: string },
  newRowIds: Set<string>,
  keyField: string
) => {
  const { ctx, rect, col, row, theme } = args;

  drawContent();

  ctx.save();

  if (args.cell.kind === GridCellKind.Custom) {
    const cellData = (args.cell as any).data;
    if (cellData?.kind === "date-cell" && !cellData.value) {
      const placeholderText = cellData.placeholder || "<empty>";
      ctx.font = `13px ${theme.fontFamily}`;
      ctx.fillStyle = theme.textDark;
      ctx.textAlign = "left";
      ctx.textBaseline = "middle";
      ctx.fillText(
        placeholderText,
        rect.x + theme.cellHorizontalPadding,
        rect.y + rect.height / 2
      );
    }
  }

  const column = columns[col];
  const isEncryptedMonetary = column?.type === "monetary"
    && args.cell.kind === GridCellKind.Text
    && (args.cell as any).displayData === "";

  if (isEncryptedMonetary) {
    ctx.fillStyle = "#f3f4f6";
    ctx.fillRect(rect.x + 1, rect.y + 1, rect.width - 2, rect.height - 2);

    const centerY = rect.y + rect.height / 2;
    const centerX = rect.x + rect.width / 2;
    const dotRadius = 2;
    const dotSpacing = 7;
    const dotCount = 5;
    const totalWidth = (dotCount - 1) * dotSpacing;
    const startX = centerX - totalWidth / 2;

    ctx.fillStyle = "#c0c0c0";
    for (let i = 0; i < dotCount; i++) {
      ctx.beginPath();
      ctx.arc(startX + i * dotSpacing, centerY, dotRadius, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  const { x, y, height } = rect;
  const item = items[row];
  const cellKey = item && columns[col] ? cellKeyOf(item[keyField], columns[col].id) : "";
  const isChanged = changedCells.has(cellKey);

  // Check whether the row is new
  const isNewRow = item && newRowIds.has(item[keyField]);

  // Get validation with fallback logic (as in getCellContent)
  const validation = validationState[cellKey];
  let isValid = true;
  
  if (validation) {
    isValid = validation.isValid;
  } else {
    // Fallback for new rows or cells without validation in state
    if (item && columns[col]) {
      const column = columns[col];
      const value = item[column.id];
      const validationResult = validateField(column, value);
      isValid = validationResult.isValid;
    }
  }

  // Show borders
  if (editMode === 'edit') {
    if (isNewRow) {
      // For new rows: blue if valid, red if invalid
      if (!isValid) {
        ctx.fillStyle = "#d32f2f"; // Red for invalid cells in new rows
      } else {
        ctx.fillStyle = "#2196f3"; // Blue for valid cells in new rows
      }
      ctx.beginPath();
      if (ctx.roundRect) {
        ctx.roundRect(x, y + 2, 3, height - 4, [0, 2, 2, 0]);
      } else {
        ctx.rect(x, y + 2, 3, height - 4);
      }
      ctx.fill();
    } else {
      // For existing rows: green if changed and valid, red if invalid
      if (isChanged && isValid) {
        // Green border for valid changes in existing rows
        ctx.fillStyle = "#4ea358";
        ctx.beginPath();
        if (ctx.roundRect) {
          ctx.roundRect(x, y + 2, 3, height - 4, [0, 2, 2, 0]);
        } else {
          ctx.rect(x, y + 2, 3, height - 4);
        }
        ctx.fill();
      } else if (!isValid) {
        // Red border for invalid cells in existing rows
        ctx.fillStyle = "#d32f2f";
        ctx.beginPath();
        if (ctx.roundRect) {
          ctx.roundRect(x, y + 2, 3, height - 4, [0, 2, 2, 0]);
        } else {
          ctx.rect(x, y + 2, 3, height - 4);
        }
        ctx.fill();
      }
    }
  }

  ctx.restore();
};
