import { useDataTableContext } from "./context/DataTableContext";

const FILTER_ICON = (
  <svg
    className="w-3.5 h-3.5 text-gray-400 shrink-0"
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

export function FilterBadges<T extends Record<string, any>>() {
  const { state, actions, columns } = useDataTableContext<T>();

  const activeFilters = Object.entries(state.activeFilters);

  const removeFilter = (columnId: string) => {
    actions.setActiveFilters((prev) => {
      const newFilters = { ...prev };
      delete newFilters[columnId];
      return newFilters;
    });
  };

  const clearAll = () => actions.setActiveFilters({});

  const getColumnTitle = (columnId: string) => {
    const column = columns.find((col) => col.id === columnId);
    return column?.title || columnId;
  };

  const getTotalValuesCount = (columnId: string) => {
    const uniqueValues = new Set<string>();
    state.items.forEach((item) => {
      const value = item[columnId];
      if (value !== undefined && value !== null && value !== "") {
        uniqueValues.add(String(value));
      }
    });
    return uniqueValues.size;
  };

  return (
    <div className="flex items-center gap-2 min-w-0">
      {FILTER_ICON}
      {activeFilters.length === 0 ? (
        <span className="text-xs text-gray-400 italic">
          No filters set — click the icon in a column header to filter
        </span>
      ) : (
        <>
          <div className="flex gap-1.5 flex-wrap flex-1 min-w-0">
            {activeFilters.map(([columnId, values]) => {
              const totalCount = getTotalValuesCount(columnId);
              const selectedCount = values.length;
              return (
                <span
                  key={columnId}
                  className="inline-flex items-center gap-1.5 pl-2 pr-1 py-0.5 bg-gray-100 border border-gray-200 rounded-full text-xs"
                >
                  <span className="font-medium text-gray-900">
                    {getColumnTitle(columnId)}
                  </span>
                  <span className="text-gray-500">
                    {selectedCount === 0
                      ? "all hidden"
                      : `${selectedCount}/${totalCount}`}
                  </span>
                  <button
                    type="button"
                    onClick={() => removeFilter(columnId)}
                    className="p-0.5 rounded-full text-gray-400 hover:text-gray-900 hover:bg-gray-200"
                    title={`Remove filter “${getColumnTitle(columnId)}”`}
                  >
                    <svg
                      className="w-3 h-3"
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
                </span>
              );
            })}
          </div>
          {activeFilters.length > 1 && (
            <button
              type="button"
              onClick={clearAll}
              className="text-xs font-medium text-gray-500 hover:text-gray-900 shrink-0"
            >
              Clear all
            </button>
          )}
        </>
      )}
    </div>
  );
}
