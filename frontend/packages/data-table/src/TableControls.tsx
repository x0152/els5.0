import { useDataTableContext } from "./context/DataTableContext";
import { useEditMode } from "./hooks";

export function TableControls<T extends Record<string, any>>() {
  const { state, showSaveControls = true } = useDataTableContext<T>();
  const {
    saveChanges,
    cancelChanges,
    hasChanges,
    hasValidationErrors,
    hasActiveFilters,
    clearAllFilters,
    saveError,
  } = useEditMode<T>();

  const activeFiltersCount = Object.keys(state.activeFilters).length;

  if (!showSaveControls || !hasChanges) {
    return null;
  }

  const saveDisabled = hasValidationErrors || hasActiveFilters;
  const saveTitle = hasValidationErrors
    ? "Fix the validation errors"
    : hasActiveFilters
    ? "Disable filters first — otherwise it's not clear what is being saved"
    : "Save changes";

  return (
    <div className="flex items-center gap-2">
      {saveError && (
        <span
          className="max-w-[280px] truncate text-xs font-medium text-red-600"
          title={saveError}
        >
          {saveError}
        </span>
      )}
      {hasActiveFilters && (
        <button
          type="button"
          onClick={clearAllFilters}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium bg-amber-50 border border-amber-200 text-amber-800 hover:bg-amber-100"
          title="Disable filters to see all rows before saving"
        >
          <svg
            className="w-3.5 h-3.5"
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
          Clear filters ({activeFiltersCount})
        </button>
      )}
      <button
        onClick={cancelChanges}
        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm bg-white border border-gray-300 text-gray-700 hover:bg-gray-50 transition-colors"
        title="Cancel all changes"
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
        Cancel
      </button>

      <button
        onClick={saveChanges}
        disabled={saveDisabled}
        className={`inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
          saveDisabled
            ? "bg-gray-100 text-gray-400 cursor-not-allowed border border-gray-200"
            : "text-white hover:[background:var(--dt-brand-hover)]"
        }`}
        style={
          saveDisabled
            ? undefined
            : {
                background: "var(--dt-brand)",
                borderColor: "var(--dt-brand)",
              }
        }
        title={saveTitle}
      >
        <svg
          className="w-3.5 h-3.5"
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden
        >
          <path
            fillRule="evenodd"
            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
            clipRule="evenodd"
          />
        </svg>
        Save
      </button>
    </div>
  );
}
