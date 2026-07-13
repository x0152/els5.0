import { useState, useMemo, useRef, useEffect, useLayoutEffect } from "react";
import { createPortal } from "react-dom";
import { useDataTableContext } from "./context/DataTableContext";
import { buildBrandStyle } from "./utils/brand";
import type { TableColumn } from "./types";

interface ColumnFilterProps {
  column: TableColumn;
  onClose: () => void;
  position: { x: number; y: number; anchorBottom: number };
}

const PANEL_WIDTH = 340;
const PANEL_MAX_HEIGHT = 460;
const GAP = 4;

export function ColumnFilter<T extends Record<string, any>>({
  column,
  onClose,
  position,
}: ColumnFilterProps) {
  const { state, actions, brandColor } = useDataTableContext<T>();
  const [searchTerm, setSearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");
  const [localSelectedValues, setLocalSelectedValues] = useState<Set<string>>(
    new Set(),
  );
  const dropdownRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  const counts = useMemo(() => {
    const map = new Map<string, number>();
    state.items.forEach((item) => {
      const raw = item[column.id];
      if (raw === undefined || raw === null || raw === "") return;
      const key = String(raw);
      map.set(key, (map.get(key) ?? 0) + 1);
    });
    return map;
  }, [state.items, column.id]);

  const allValues = useMemo(() => {
    const values = Array.from(counts.keys());
    values.sort((a, b) => {
      if (column.type === "salary" || column.type === "number") {
        const numA = parseFloat(a) || 0;
        const numB = parseFloat(b) || 0;
        return sortOrder === "asc" ? numA - numB : numB - numA;
      }
      return sortOrder === "asc" ? a.localeCompare(b) : b.localeCompare(a);
    });
    return values;
  }, [counts, column.type, sortOrder]);

  const columnValues = useMemo(() => {
    if (!searchTerm.trim()) return allValues;
    const q = searchTerm.toLowerCase();
    return allValues.filter((v) => v.toLowerCase().includes(q));
  }, [allValues, searchTerm]);

  useEffect(() => {
    const currentFilter = state.activeFilters[column.id];
    setLocalSelectedValues(
      currentFilter !== undefined
        ? new Set(currentFilter)
        : new Set(counts.keys()),
    );
  }, [column.id, counts, state.activeFilters]);

  useEffect(() => {
    const t = setTimeout(() => searchRef.current?.focus(), 0);
    return () => clearTimeout(t);
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleKey);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleKey);
    };
  }, [onClose]);

  const [coords, setCoords] = useState<{ left: number; top: number }>({
    left: position.x,
    top: position.y,
  });

  useLayoutEffect(() => {
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const left = Math.max(8, Math.min(position.x, vw - PANEL_WIDTH - 8));
    const estimatedHeight = Math.min(
      PANEL_MAX_HEIGHT,
      dropdownRef.current?.getBoundingClientRect().height ?? PANEL_MAX_HEIGHT,
    );
    const spaceBelow = vh - position.y - 8;
    const spaceAbove = position.anchorBottom - estimatedHeight - GAP;
    let top = position.y;
    if (spaceBelow < estimatedHeight && spaceAbove > 8) {
      top = position.anchorBottom - estimatedHeight - GAP;
    }
    top = Math.max(8, Math.min(top, vh - estimatedHeight - 8));
    setCoords({ left, top });
  }, [position.x, position.y, position.anchorBottom, columnValues.length]);

  const selectedCount = localSelectedValues.size;
  const totalCount = allValues.length;
  const visibleCount = columnValues.length;
  const allVisibleSelected =
    visibleCount > 0 &&
    columnValues.every((v) => localSelectedValues.has(v));
  const noneVisibleSelected =
    visibleCount > 0 &&
    columnValues.every((v) => !localSelectedValues.has(v));

  const toggleAllVisible = () => {
    setLocalSelectedValues((prev) => {
      const next = new Set(prev);
      if (allVisibleSelected) {
        columnValues.forEach((v) => next.delete(v));
      } else {
        columnValues.forEach((v) => next.add(v));
      }
      return next;
    });
  };

  const handleValueToggle = (value: string) => {
    setLocalSelectedValues((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(value)) newSet.delete(value);
      else newSet.add(value);
      return newSet;
    });
  };

  const handleApply = () => {
    actions.setActiveFilters((prev) => {
      const newFilters = { ...prev };
      if (localSelectedValues.size === totalCount) {
        delete newFilters[column.id];
      } else {
        newFilters[column.id] = Array.from(localSelectedValues);
      }
      return newFilters;
    });
    onClose();
  };

  const handleReset = () => {
    setLocalSelectedValues(new Set(allValues));
    setSearchTerm("");
    actions.setActiveFilters((prev) => {
      const newFilters = { ...prev };
      delete newFilters[column.id];
      return newFilters;
    });
    onClose();
  };

  const hasActiveFilter = state.activeFilters[column.id] !== undefined;

  const panel = (
    <div
      ref={dropdownRef}
      className="fixed z-50 flex flex-col rounded-xl border border-gray-200 bg-white shadow-xl ring-1 ring-black/5 outline-none overflow-hidden"
      style={{
        left: coords.left,
        top: coords.top,
        width: PANEL_WIDTH,
        maxHeight: PANEL_MAX_HEIGHT,
        ...buildBrandStyle(brandColor),
      }}
      onKeyDown={(e) => {
        if (e.key === "Enter") handleApply();
      }}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-gray-100">
        <svg
          className="w-4 h-4 text-gray-400"
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
        <span className="text-sm font-semibold text-gray-900 truncate flex-1">
          {column.title}
        </span>
        <button
          type="button"
          onClick={() =>
            setSortOrder((prev) => (prev === "asc" ? "desc" : "asc"))
          }
          className="inline-flex items-center gap-1 px-2 py-1 rounded text-xs text-gray-600 hover:bg-gray-100"
          title={
            sortOrder === "asc" ? "Ascending" : "Descending"
          }
        >
          <svg className="w-3.5 h-3.5" viewBox="0 0 20 20" fill="currentColor">
            {sortOrder === "asc" ? (
              <path
                fillRule="evenodd"
                d="M3 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h4a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h2a1 1 0 110 2H4a1 1 0 01-1-1zm12.707-1.707a1 1 0 00-1.414 0L13 11.586V4a1 1 0 10-2 0v7.586l-1.293-1.293a1 1 0 10-1.414 1.414l3 3a1 1 0 001.414 0l3-3a1 1 0 000-1.414z"
                clipRule="evenodd"
              />
            ) : (
              <path
                fillRule="evenodd"
                d="M3 4a1 1 0 011-1h6a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h4a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h2a1 1 0 110 2H4a1 1 0 01-1-1zm9-8a1 1 0 011 1v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 011.414-1.414L11 12.586V5a1 1 0 011-1z"
                clipRule="evenodd"
              />
            )}
          </svg>
          {sortOrder === "asc" ? "A–Z" : "Z–A"}
        </button>
        <button
          type="button"
          onClick={onClose}
          className="p-1 rounded text-gray-400 hover:text-gray-700 hover:bg-gray-100"
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

      <div className="px-4 py-2.5 border-b border-gray-100 space-y-2">
        <div className="relative">
          <svg
            className="w-4 h-4 text-gray-400 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none"
            viewBox="0 0 20 20"
            fill="currentColor"
            aria-hidden
          >
            <path
              fillRule="evenodd"
              d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.817-4.817A6 6 0 012 8z"
              clipRule="evenodd"
            />
          </svg>
          <input
            ref={searchRef}
            type="text"
            placeholder="Search values"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-8 pr-8 py-1.5 text-sm border border-gray-200 rounded-md bg-gray-50 focus:bg-white focus:outline-none focus:[border-color:var(--dt-brand-border)] focus:[box-shadow:0_0_0_3px_var(--dt-brand-soft)]"
          />
          {searchTerm && (
            <button
              type="button"
              onClick={() => setSearchTerm("")}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-700"
              aria-label="Clear search"
            >
              <svg
                className="w-3.5 h-3.5"
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
          )}
        </div>

        <div className="flex items-center justify-between text-xs">
          <label className="flex items-center gap-2 cursor-pointer select-none">
            <input
              type="checkbox"
              checked={allVisibleSelected}
              ref={(el) => {
                if (el)
                  el.indeterminate = !allVisibleSelected && !noneVisibleSelected;
              }}
              onChange={toggleAllVisible}
              disabled={visibleCount === 0}
              className="rounded border-gray-300 focus:ring-offset-0 focus:[box-shadow:0_0_0_3px_var(--dt-brand-soft)]"
              style={{ accentColor: brandColor }}
            />
            <span className="text-gray-700 font-medium">
              {searchTerm
                ? allVisibleSelected
                  ? "Deselect found"
                  : "Select found"
                : allVisibleSelected
                ? "Deselect all"
                : "Select all"}
            </span>
          </label>
          <span className="text-gray-500">
            {selectedCount} of {totalCount}
            {searchTerm && (
              <span className="ml-1 text-gray-400">
                · found {visibleCount}
              </span>
            )}
          </span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto max-h-[280px]">
        {columnValues.length === 0 ? (
          <div className="px-4 py-10 text-center text-sm text-gray-500">
            {searchTerm ? "Nothing found" : "No values"}
          </div>
        ) : (
          <ul className="py-1">
            {columnValues.map((value) => {
              const isSelected = localSelectedValues.has(value);
              const count = counts.get(value) ?? 0;
              return (
                <li key={value}>
                  <label
                    className={`flex items-center gap-2 px-4 py-1.5 cursor-pointer text-sm transition-colors ${
                      isSelected
                        ? "hover:[background:var(--dt-brand-softer)]"
                        : "hover:bg-gray-50 text-gray-500"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => handleValueToggle(value)}
                      className="rounded border-gray-300 focus:ring-offset-0 focus:[box-shadow:0_0_0_3px_var(--dt-brand-soft)]"
                      style={{ accentColor: brandColor }}
                    />
                    <span className="flex-1 truncate text-gray-900">
                      {value}
                    </span>
                    <span className="text-xs text-gray-400 tabular-nums">
                      {count}
                    </span>
                  </label>
                </li>
              );
            })}
          </ul>
        )}
      </div>

      <div className="flex items-center justify-between gap-2 px-3 py-2.5 border-t border-gray-100 bg-gray-50">
        <button
          type="button"
          onClick={handleReset}
          disabled={!hasActiveFilter && !searchTerm}
          className="text-xs font-medium text-gray-600 hover:text-gray-900 disabled:opacity-40 disabled:cursor-not-allowed"
        >
          Reset
        </button>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-100"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleApply}
            className="px-3 py-1.5 text-sm font-medium text-white rounded-md border hover:[background:var(--dt-brand-hover)]"
            style={{
              background: "var(--dt-brand)",
              borderColor: "var(--dt-brand)",
            }}
          >
            Apply
          </button>
        </div>
      </div>
    </div>
  );

  return createPortal(panel, document.body);
}
