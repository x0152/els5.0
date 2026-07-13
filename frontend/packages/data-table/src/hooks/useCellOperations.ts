import { useCallback } from "react";
import { GridCellKind } from "@glideapps/glide-data-grid";
import type {
  GridCell,
  Item,
  EditableGridCell,
} from "@glideapps/glide-data-grid";
import { useDataTableContext } from "../context/DataTableContext";
import { createBadgeCell } from "../cells";
import { useDataOperations } from "./useDataOperations";
import { useValidation } from "./useValidation";
import { cellKeyOf } from "../utils/validationUtils";

export const useCellOperations = <T extends Record<string, any>>() => {
  const {
    state,
    actions,
    columns,
    onDataChange,
    customCellRenderers = [],
    keyField,
  } = useDataTableContext<T>();
  const { filteredItems } = useDataOperations<T>();
  const { validateField } = useValidation();

  // Function to normalize a Telegram username
  const normalizeTelegramValue = useCallback(
    (value: string | null): string | null => {
      if (!value || typeof value !== "string") return value;

      const trimmed = value.trim();
      if (!trimmed) return null;

      if (trimmed.startsWith("@")) {
        return trimmed;
      }

      if (trimmed.startsWith("https://t.me/")) {
        const username = trimmed.replace("https://t.me/", "");
        return `@${username}`;
      }

      if (trimmed.startsWith("t.me/")) {
        const username = trimmed.replace("t.me/", "");
        return `@${username}`;
      }

      if (trimmed.includes("@") && !trimmed.startsWith("@")) {
        // May be an email or another format — return as-is
        return trimmed;
      }

      if (!trimmed.startsWith("@")) {
        return `@${trimmed}`;
      }

      return trimmed;
    },
    [],
  );

  // Function to validate Telegram
  const validateTelegramField = useCallback(
    (value: string | null): { isValid: boolean; message?: string } => {
      if (value === null || value === "") {
        return { isValid: true };
      }

      const normalized = normalizeTelegramValue(value);

      if (!normalized) {
        return { isValid: true };
      }

      if (!normalized.startsWith("@")) {
        return {
          isValid: false,
          message:
            "Telegram must start with @ or be in the format https://t.me/username",
        };
      }

      const username = normalized.slice(1);

      // Check username length (Telegram limits: 5-32 characters)
      if (username.length < 5) {
        return {
          isValid: false,
          message: "Telegram username must contain at least 5 characters",
        };
      }

      if (username.length > 32) {
        return {
          isValid: false,
          message: "Telegram username cannot be longer than 32 characters",
        };
      }

      // Check allowed characters: letters, digits, underscore
      const telegramRegex = /^[a-zA-Z0-9_]+$/;
      if (!telegramRegex.test(username)) {
        return {
          isValid: false,
          message:
            "Telegram username can only contain Latin letters, digits, and underscores",
        };
      }

      return { isValid: true };
    },
    [normalizeTelegramValue],
  );

  // Function to normalize a phone number to +7 (XXX) XXX-XX-XX
  const normalizePhoneValue = useCallback(
    (value: string | null): string | null => {
      if (!value || typeof value !== "string") return value;

      const digits = value.replace(/\D/g, "");

      if (digits.length === 0) return null;

      if (digits.length >= 10) {
        let phoneDigits = digits;

        if (phoneDigits.length === 11) {
          if (phoneDigits[0] === "8" || phoneDigits[0] === "7") {
            phoneDigits = phoneDigits.substring(1);
          }
        }

        if (phoneDigits.length > 10) {
          phoneDigits = phoneDigits.slice(-10);
        }

        // Format: +7 (XXX) XXX-XX-XX
        const areaCode = phoneDigits.substring(0, 3);
        const firstPart = phoneDigits.substring(3, 6);
        const secondPart = phoneDigits.substring(6, 8);
        const thirdPart = phoneDigits.substring(8, 10);

        return `+7 (${areaCode}) ${firstPart}-${secondPart}-${thirdPart}`;
      }

      // If not enough digits, return as-is
      return value;
    },
    [],
  );

  // Function to validate a phone number
  const validatePhoneField = useCallback(
    (value: string | null): { isValid: boolean; message?: string } => {
      if (value === null || value === "") {
        return { isValid: true };
      }

      const normalized = normalizePhoneValue(value);

      if (!normalized) {
        return {
          isValid: false,
          message: "Invalid phone number format",
        };
      }

      const digits = normalized.replace(/\D/g, "");

      if (digits.length !== 11) {
        return {
          isValid: false,
          message: "Phone number must contain 10 digits after the country code",
        };
      }

      if (digits[0] !== "7") {
        return {
          isValid: false,
          message: "Phone number must start with +7",
        };
      }

      return { isValid: true };
    },
    [normalizePhoneValue],
  );

  // Function to format a money value
  const formatMonetaryValue = useCallback(
    (value: number | string | null): string => {
      if (value === null || value === undefined || value === "") {
        return "";
      }

      let numValue: number;
      if (typeof value === "string") {
        const cleanValue = value.replace(/[^\d.-]/g, "");
        numValue = parseFloat(cleanValue);
      } else {
        numValue = value;
      }

      if (isNaN(numValue)) {
        return "";
      }

      // Format with spaces between thousands
      const formatter = new Intl.NumberFormat("ru-RU", {
        maximumFractionDigits: 0,
        minimumFractionDigits: 0,
        useGrouping: true,
      });

      return formatter.format(numValue);
    },
    [],
  );

  // Function to parse a money value from a string
  const parseMonetaryValue = useCallback(
    (value: string | null): number | null => {
      if (!value || typeof value !== "string") return null;

      const cleanValue = value.replace(/[^\d.-]/g, "");
      const parsed = parseFloat(cleanValue);

      return isNaN(parsed) ? null : parsed;
    },
    [],
  );

  // Function to validate a money field
  const validateMonetaryField = useCallback(
    (value: number | string | null): { isValid: boolean; message?: string } => {
      if (value === null || value === undefined || value === "") {
        return { isValid: true };
      }

      let numValue: number;

      if (typeof value === "string") {
        const cleanValue = value.replace(/[^\d.-]/g, "");
        numValue = parseFloat(cleanValue);
      } else {
        numValue = value;
      }

      if (isNaN(numValue)) {
        return {
          isValid: false,
          message: "Enter a valid number",
        };
      }

      if (numValue < 0) {
        return {
          isValid: false,
          message: "Value cannot be negative",
        };
      }

      // Check for a value that is too large
      if (numValue > 1000000000) {
        return {
          isValid: false,
          message: "Value exceeds the maximum allowed (1,000,000,000)",
        };
      }

      return { isValid: true };
    },
    [],
  );

  const getCellContent = useCallback(
    (cell: Item): GridCell => {
      const [col, row] = cell;
      const item = filteredItems[row];
      if (!item)
        return {
          kind: GridCellKind.Text,
          data: "",
          allowOverlay: false,
          displayData: "",
        };

      const column = columns[col];
      const field = column.id;
      let value = item[field];

      // Apply normalization for special fields
      if (field === "telegram_login" || column.type === "telegram") {
        value = normalizeTelegramValue(value);
      }

      if (field === "phone_number" || column.type === "phone") {
        value = normalizePhoneValue(value);
      }

      const cellKey = cellKeyOf(item[keyField], field);
      const isChanged = state.changedCells.has(cellKey);

      let validationResult;
      if (field === "telegram_login" || column.type === "telegram") {
        validationResult =
          state.validationState[cellKey] ?? validateTelegramField(value);
      } else if (field === "phone_number" || column.type === "phone") {
        validationResult =
          state.validationState[cellKey] ?? validatePhoneField(value);
      } else if (column.type === "monetary") {
        // For money fields validation uses the original numeric value
        validationResult =
          state.validationState[cellKey] ?? validateMonetaryField(item[field]);
      } else {
        validationResult =
          state.validationState[cellKey] ?? validateField(column, value);
      }

      const isValid = validationResult.isValid;
      const isDeleted = state.deletedRowIds.has(item[keyField]);
      const isNewRow = state.newRowIds.has(item[keyField]);

      let textColor = undefined;
      if (isDeleted) {
        textColor = "#999999"; // Gray for deleted
      } else if (isNewRow) {
        // For new rows: blue if valid, red if invalid
        textColor = !isValid ? "#d32f2f" : "#2196f3";
      } else if (!isValid) {
        textColor = "#d32f2f"; // Red for errors in existing rows
      } else if (state.editMode === "edit" && isChanged && isValid) {
        textColor = "#4ea358"; // Green for valid changes in existing rows
      }

      const baseCell = {
        allowOverlay:
          state.editMode === "edit" && !isDeleted && !column.readonly,
        style: isDeleted ? ("faded" as const) : ("normal" as const),
        themeOverride: textColor
          ? {
              ...(isDeleted && { bgCell: "#f5f5f5" }),
              textDark: textColor,
            }
          : undefined,
      };

      if (column.type === "rowid") {
        return {
          kind: GridCellKind.RowID,
          data: value != null ? value.toString() : "",
          allowOverlay: false,
          readonly: true,
          style: "faded" as const,
          contentAlign: column.align || "center",
          themeOverride: {
            bgCell: "#f8f9fa",
            textDark: "#6c757d",
          },
        };
      }

      const customRenderer = customCellRenderers.find(
        (ccr) => ccr.kind === column.type,
      );
      if (customRenderer) {
        const customCell = customRenderer.cellCreator(value, column.readonly);

        let copyData = null;
        if (customRenderer.copyDataExtractor) {
          copyData = customRenderer.copyDataExtractor(customCell.data) || null;
        } else {
          const valueExtractor =
            customRenderer.valueExtractor ||
            ((cellData: any) => cellData.value);
          const extractedValue = valueExtractor(customCell.data);
          copyData = extractedValue != null ? extractedValue.toString() : null;
        }

        return {
          ...customCell,
          ...baseCell,
          copyData,
          contentAlign: column.align || "center",
          themeOverride: {
            ...customCell.themeOverride,
            ...baseCell.themeOverride,
          },
        };
      }

      if (column.type === "boolean") {
        return {
          kind: GridCellKind.Boolean,
          data: value != null ? Boolean(value) : false,
          allowOverlay: false,
          style: isDeleted ? ("faded" as const) : ("normal" as const),
          contentAlign: column.align || "center",
          themeOverride: textColor
            ? {
                ...(isDeleted && { bgCell: "#f5f5f5" }),
                textDark: textColor,
              }
            : undefined,
        };
      }

      if (column.type === "dropdown") {
        return {
          kind: GridCellKind.Custom,
          data: {
            kind: "dropdown-cell",
            allowedValues: ["<not set>", ...(column.options || [])],
            value: (value != null && value !== "") ? value : "<not set>",
            displayValue: (value != null && value !== "") ? value : "<not set>",
          },
          copyData: (value != null && value !== "") ? value : "<not set>",
          contentAlign: column.align || "left",
          ...baseCell,
        };
      }

      if (column.type === "badge") {
        return {
          ...createBadgeCell(value),
          contentAlign: column.align || "left",
          themeOverride: baseCell.themeOverride,
        };
      }

      if (column.type === "drilldown") {
        return {
          kind: GridCellKind.Drilldown,
          data: (value != null && value !== "") ? [{ text: value, img: undefined }] : [],
          contentAlign: column.align || "left",
          ...baseCell,
        };
      }

      if (column.type === "date") {
        const dateValue = (value != null && value !== "") ? value : null;
        return {
          kind: GridCellKind.Custom,
          data: {
            kind: "date-cell",
            value: dateValue,
            readonly: column.readonly,
            placeholder: "<empty>",
          },
          copyData: dateValue || "",
          contentAlign: column.align || "left",
          ...baseCell,
        };
      }

      if (column.type === "number") {
        const numEmpty = value === null || value === undefined;
        return {
          kind: GridCellKind.Number,
          data: numEmpty ? 0 : value,
          displayData: numEmpty ? "" : value.toString(),
          contentAlign: column.align || "right",
          ...baseCell,
        };
      }

      if (column.type === "uri") {
        return {
          kind: GridCellKind.Uri,
          data: value || "",
          hoverEffect: true,
          contentAlign: column.align || "center",
          ...baseCell,
        };
      }

      if (column.type === "monetary") {
        if (!state.areSecretsUnlocked) {
          const hasEncryptedValue =
            !isNewRow && value !== null && value !== undefined && typeof value === "number" && isNaN(value);
          return {
            kind: GridCellKind.Text,
            data: hasEncryptedValue ? "*****" : "",
            displayData: hasEncryptedValue ? "*****" : "",
            contentAlign: column.align || "right",
            allowOverlay: false,
            readonly: true,
            style: "normal" as const,
          };
        }
        const isKeyMismatch =
          !isNewRow && typeof value === "number" && isNaN(value);
        if (isKeyMismatch) {
          return {
            kind: GridCellKind.Text,
            data: "\u26a0\ufe0f \u041e\u0448\u0438\u0431\u043a\u0430 \u043a\u043b\u044e\u0447\u0430",
            displayData: "\u26a0\ufe0f \u041e\u0448\u0438\u0431\u043a\u0430 \u043a\u043b\u044e\u0447\u0430",
            contentAlign: column.align || "right",
            allowOverlay: false,
            readonly: true,
            style: "normal" as const,
            themeOverride: {
              textDark: "#ff4d4f",
            },
          };
        }
        const isEmpty = value === null || value === undefined;
        const formattedValue = formatMonetaryValue(value);
        return {
          kind: GridCellKind.Number,
          data: isEmpty ? 0 : value,
          displayData: isEmpty ? "" : `${formattedValue} ₽`,
          contentAlign: column.align || "right",
          allowOverlay:
            state.editMode === "edit" && !isDeleted && !column.readonly,
          readonly: false,
          style: isDeleted ? ("faded" as const) : ("normal" as const),
          themeOverride: {
            ...baseCell.themeOverride,
            fontFamily: "monospace",
          },
          copyData: isEmpty ? "" : formattedValue,
        };
      }

      if (field === "telegram_login" || column.type === "telegram") {
        return {
          kind: GridCellKind.Text,
          data: (value != null && value !== "") ? value : "",
          displayData: (value != null && value !== "") ? value : "<empty>",
          contentAlign: column.align || "left",
          ...baseCell,
        };
      }

      if (field === "phone_number" || column.type === "phone") {
        return {
          kind: GridCellKind.Text,
          data: (value != null && value !== "") ? value : "",
          displayData: (value != null && value !== "") ? value : "<empty>",
          contentAlign: column.align || "left",
          ...baseCell,
        };
      }

      return {
        kind: GridCellKind.Text,
        data: (value != null && value !== "") ? value : "",
        displayData: (value != null && value !== "") ? value : "<empty>",
        contentAlign: column.align || "left",
        ...baseCell,
      };
    },
    [
      filteredItems,
      columns,
      state,
      keyField,
      validateField,
      normalizeTelegramValue,
      validateTelegramField,
      normalizePhoneValue,
      validatePhoneField,
      formatMonetaryValue,
      validateMonetaryField,
    ],
  );

  const updateCellValue = useCallback(
    (
      column: any,
      newValue: EditableGridCell,
      itemId: string,
    ) => {
      let parsedValue: any;

      if (newValue.kind === GridCellKind.Boolean) {
        parsedValue = newValue.data;
      } else if (newValue.kind === GridCellKind.Number) {
        parsedValue = newValue.data != null ? newValue.data : null;
      } else if (newValue.kind === GridCellKind.Custom) {
        const customData = newValue.data as any;
        if (customData?.kind === "dropdown-cell") {
          parsedValue =
            customData.value != null && customData.value !== ""
              ? customData.value
              : null;
        } else if (
          customData?.kind === "date-picker-cell" ||
          customData?.kind === "date-cell"
        ) {
          if (customData.date) {
            parsedValue = customData.date.toISOString().split("T")[0];
          } else if (customData.value) {
            parsedValue = customData.value;
          } else {
            parsedValue = null;
          }
        } else {
          const customRenderer = customCellRenderers.find(
            (ccr) => ccr.kind === customData?.kind,
          );
          if (customRenderer) {
            const valueExtractor =
              customRenderer.valueExtractor ||
              ((cellData: any) => cellData.value);
            parsedValue = valueExtractor(customData);
          } else {
            parsedValue =
              customData?.value != null && customData?.value !== ""
                ? customData?.value
                : null;
          }
        }
      } else if ((newValue as any).kind === GridCellKind.Drilldown) {
        const drilldownData = (newValue as any).data as Array<{ text: string }>;
        parsedValue =
          drilldownData && drilldownData.length > 0
            ? drilldownData[0].text
            : null;
      } else {
        parsedValue =
          newValue.data != null && newValue.data !== "" ? newValue.data : null;
      }

      // Apply normalization for special fields
      if (column.id === "telegram_login" || column.type === "telegram") {
        parsedValue = normalizeTelegramValue(parsedValue);
      }

      if (column.id === "phone_number" || column.type === "phone") {
        parsedValue = normalizePhoneValue(parsedValue);
      }

      if (column.type === "monetary") {
        // For a money field parse the string into a number
        if (typeof parsedValue === "string") {
          parsedValue = parseMonetaryValue(parsedValue);
        }
        if (parsedValue === null || parsedValue === undefined) {
          parsedValue = null;
        }
      }

      let validationResult;
      if (column.id === "telegram_login" || column.type === "telegram") {
        validationResult = validateTelegramField(parsedValue);
      } else if (column.id === "phone_number" || column.type === "phone") {
        validationResult = validatePhoneField(parsedValue);
      } else if (column.type === "monetary") {
        validationResult = validateMonetaryField(parsedValue);
      } else {
        validationResult = validateField(column, parsedValue);
      }

      const cellKey = cellKeyOf(itemId, column.id);
      actions.setValidationState((prev) => ({
        ...prev,
        [cellKey]: {
          isValid: validationResult.isValid,
          isChanged: true,
          message: validationResult.message,
        },
      }));

      actions.setChangedCells((prev) => new Set(prev).add(cellKey));

      const field = column.id;
      actions.setItems((prev: T[]) => {
        const realIndex = prev.findIndex(
          (item: T) => item[keyField] === itemId,
        );
        if (realIndex === -1) return prev;

        const updated = prev.map((item: T, idx: number) =>
          idx === realIndex ? { ...item, [field]: parsedValue } : item,
        );
        onDataChange?.(updated);
        return updated;
      });
    },
    [
      actions,
      onDataChange,
      validateField,
      keyField,
      normalizeTelegramValue,
      validateTelegramField,
      normalizePhoneValue,
      validatePhoneField,
      parseMonetaryValue,
      validateMonetaryField,
    ],
  );

  const onCellEdited = useCallback(
    (cell: Item, newValue: EditableGridCell) => {
      if (state.editMode !== "edit") return;

      const [col, row] = cell;
      const column = columns[col];

      const filteredItem = filteredItems[row];
      if (!filteredItem) return;

      updateCellValue(column, newValue, filteredItem[keyField]);
    },
    [state.editMode, columns, filteredItems, updateCellValue, keyField],
  );

  return {
    getCellContent,
    onCellEdited,
    updateCellValue,
    normalizeTelegramValue,
    validateTelegramField,
    normalizePhoneValue,
    validatePhoneField,
    formatMonetaryValue,
    parseMonetaryValue,
    validateMonetaryField,
  };
};
