import { useMemo, useState } from 'react'
import {
  DataTable,
  type ColumnType,
  type ServerValidationError,
  type TableColumn,
} from '@els/data-table'
import { isApiError } from '@els/api-client'
import { Button, ErrorState, LoadingState } from '@els/ui'
import {
  useGridApply,
  useGridDescribe,
  useGridLookupSource,
  useInvalidateGridDescribe,
} from '../store/grid.ts'
import type {
  DescribeGridResponse,
  GridColumn,
  GridOp,
  GridRow,
} from '../lib/grid-client.ts'
import { useToast } from './Toasts.tsx'
import { SelectionPanel } from './SelectionPanel.tsx'

interface GridViewProps {
  basePath: string
  /** Title written into errors if a column could not be mapped. */
  fallbackColumnId?: string
}

const UNSET_LABEL = '<not set>'

type Row = {
  id: string
  base_version?: number
} & Record<string, unknown>

interface ColumnMaps {
  /** label → value (for enum/ref) */
  byLabel: Map<string, string>
  /** value(id) → label (for enum/ref) */
  byKey: Map<string, string>
}

export function GridView({ basePath, fallbackColumnId }: GridViewProps) {
  const toast = useToast()
  const describe = useGridDescribe(basePath)
  const apply = useGridApply(basePath)
  const invalidateDescribe = useInvalidateGridDescribe(basePath)

  const [selectedId, setSelectedId] = useState<string | null>(null)

  // Take the first ref source from the schema. Current grids have at most one (`client`).
  const refSource = useMemo<string | null>(() => {
    const cols = describe.data?.columns ?? []
    const first = cols.find((c) => c.type === 'ref' && c.ref)
    return first?.ref?.source ?? null
  }, [describe.data])
  const lookup = useGridLookupSource(basePath, refSource)

  const { columns: tableColumns, maps, rows, knownColumnIds } = useMemo(
    () => buildView(describe.data, lookup.data ?? []),
    [describe.data, lookup.data],
  )

  const selectedRow = useMemo(
    () => (selectedId ? rows.find((r) => r.id === selectedId) ?? null : null),
    [rows, selectedId],
  )

  if (describe.isLoading) {
    return <LoadingState className="flex-1 items-center py-0 text-neutral-400" />
  }

  if (describe.error) {
    return (
      <div className="flex-1 grid place-items-center p-6">
        <ErrorState
          className="w-full max-w-md"
          title="Failed to load"
          description={describe.error instanceof Error ? describe.error.message : 'error'}
          action={
            <Button variant="secondary" onClick={() => describe.refetch()}>
              Try again
            </Button>
          }
        />
      </div>
    )
  }

  if (!describe.data) return null

  const schemaVersion = describe.data.schema_version
  const schemaColumns = describe.data.columns
  const prevRowsById = new Map(rows.map((r) => [r.id, r]))

  return (
    <div className="flex-1 min-w-0 min-h-0 flex overflow-hidden">
      <div className="flex-1 min-w-0 min-h-0 p-4 overflow-hidden">
      <DataTable<Row>
        data={rows}
        columns={tableColumns}
        keyField="id"
        brandColor="#059669"
        showControls
        showSaveControls
        editMode="toggle"
        allowAddRows
        onSelectionChange={(items) => {
          const first = items[0]
          // For new rows (without base_version) the panel is not needed — save them first.
          if (!first || first.base_version === undefined) {
            setSelectedId(null)
            return
          }
          setSelectedId(first.id)
        }}
        onSave={async (next, deletedIds) => {
          const operations: GridOp[] = []
          const tempIdByRowId = new Map<string, string>()

          for (const current of next) {
            const prev = prevRowsById.get(current.id)
            if (!prev) {
              // New row — invent a temp_id.
              const tempId =
                typeof crypto !== 'undefined' && 'randomUUID' in crypto
                  ? crypto.randomUUID()
                  : `new-${current.id}`
              tempIdByRowId.set(current.id, tempId)
              const data = buildOpData(current, schemaColumns, maps, /* forCreate */ true)
              operations.push({ kind: 'create', temp_id: tempId, data })
              continue
            }
            const patch = diffPatch(prev, current, schemaColumns, maps)
            if (patch && Object.keys(patch).length > 0) {
              operations.push({
                kind: 'update',
                id: current.id,
                base_version: prev.base_version,
                data: patch,
              })
            }
          }

          for (const id of deletedIds ?? []) {
            const prev = prevRowsById.get(id)
            if (!prev) continue
            operations.push({
              kind: 'delete',
              id,
              base_version: prev.base_version,
            })
          }

          if (operations.length === 0) {
            toast.success('No changes')
            return null
          }

          try {
            const res = await apply.mutateAsync({
              schema_version: schemaVersion,
              operations,
            })

            if (res.failed.length > 0) {
              const serverErrors = mapFailedToValidationErrors(
                res.failed,
                tempIdByRowId,
                knownColumnIds,
                fallbackColumnId,
              )
              toast.error(summarizeFailed(res.failed))
              return serverErrors
            }

            const counts = countByKind(res.applied)
            toast.success(
              `Done${counts.create ? `, created: ${counts.create}` : ''}${
                counts.update ? `, updated: ${counts.update}` : ''
              }${counts.delete ? `, deleted: ${counts.delete}` : ''}`,
            )
            invalidateDescribe()
            return null
          } catch (e) {
            const msg = isApiError(e) ? e.message : e instanceof Error ? e.message : 'Error'
            toast.error(msg)
            return [
              {
                rowId: operations[0]?.id ?? operations[0]?.temp_id ?? '—',
                columnId: fallbackColumnId ?? firstEditableColumnId(schemaColumns) ?? 'id',
                messages: [msg],
              },
            ]
          }
        }}
      />
      </div>
      <SelectionPanel
        basePath={basePath}
        schemaVersion={schemaVersion}
        columns={schemaColumns}
        maps={maps}
        row={selectedRow}
        onClose={() => setSelectedId(null)}
        onApplied={() => {
          invalidateDescribe()
        }}
      />
    </div>
  )
}

/* --------------------------------- builders ------------------------------- */

interface BuiltView {
  columns: TableColumn[]
  maps: Record<string, ColumnMaps>
  rows: Row[]
  knownColumnIds: string[]
}

const HIDDEN_COLUMN_IDS = new Set<string>(['picture_url', 'account_id'])

function buildView(
  data: DescribeGridResponse | undefined,
  lookupItems: { key: string; label: string }[],
): BuiltView {
  if (!data) return { columns: [], maps: {}, rows: [], knownColumnIds: [] }

  const maps: Record<string, ColumnMaps> = {}
  const columns: TableColumn[] = []
  const knownColumnIds: string[] = []

  for (const col of data.columns) {
    knownColumnIds.push(col.id)
    const { tableCol, map } = toTableColumn(col, data.refs_hydrated, lookupItems)
    if (map) maps[col.id] = map
    if (!HIDDEN_COLUMN_IDS.has(col.id)) columns.push(tableCol)
  }

  const rows: Row[] = data.rows.map((serverRow) =>
    toRow(serverRow, data.columns, maps, data.refs_hydrated),
  )

  return { columns, maps, rows, knownColumnIds }
}

function toTableColumn(
  col: GridColumn,
  refsHydrated: DescribeGridResponse['refs_hydrated'],
  lookupItems: { key: string; label: string }[],
): { tableCol: TableColumn; map?: ColumnMaps } {
  const base: TableColumn = {
    id: col.id,
    title: col.title || col.id,
    type: mapType(col.type),
    width: defaultWidthFor(col),
    required: col.required,
    readonly: col.readonly,
  }

  if (col.type === 'enum' && col.enum?.length) {
    const byLabel = new Map<string, string>()
    const byKey = new Map<string, string>()
    const labels: string[] = []
    for (const opt of col.enum) {
      labels.push(opt.label)
      byLabel.set(opt.label, opt.value)
      byKey.set(opt.value, opt.label)
    }
    return { tableCol: { ...base, type: 'dropdown', options: labels }, map: { byLabel, byKey } }
  }

  if (col.type === 'ref' && col.ref) {
    const byLabel = new Map<string, string>()
    const byKey = new Map<string, string>()
    const labels: string[] = []
    const hydrated = refsHydrated?.[col.ref.source] ?? {}
    for (const [key, label] of Object.entries(hydrated)) {
      if (!byKey.has(key)) byKey.set(key, label)
      if (!byLabel.has(label)) {
        byLabel.set(label, key)
        labels.push(label)
      }
    }
    for (const item of lookupItems) {
      if (!byKey.has(item.key)) byKey.set(item.key, item.label)
      if (!byLabel.has(item.label)) {
        byLabel.set(item.label, item.key)
        labels.push(item.label)
      }
    }
    return { tableCol: { ...base, type: 'dropdown', options: labels }, map: { byLabel, byKey } }
  }

  return { tableCol: base }
}

function toRow(
  serverRow: GridRow,
  columns: GridColumn[],
  maps: Record<string, ColumnMaps>,
  refsHydrated: DescribeGridResponse['refs_hydrated'],
): Row {
  const row: Row = { id: serverRow.id, base_version: serverRow.base_version }
  for (const col of columns) {
    const raw = serverRow.cells[col.id]
    row[col.id] = toCellDisplay(col, raw, maps[col.id], refsHydrated)
  }
  return row
}

function toCellDisplay(
  col: GridColumn,
  raw: unknown,
  map: ColumnMaps | undefined,
  refsHydrated: DescribeGridResponse['refs_hydrated'],
): unknown {
  if (raw === null || raw === undefined) {
    if (col.type === 'bool') return false
    return ''
  }
  if (col.type === 'enum') {
    const value = String(raw)
    return map?.byKey.get(value) ?? value
  }
  if (col.type === 'ref' && col.ref) {
    const key = String(raw)
    const fromMap = map?.byKey.get(key)
    if (fromMap) return fromMap
    const hydrated = refsHydrated?.[col.ref.source]?.[key]
    return hydrated ?? key
  }
  if (col.type === 'bool') return Boolean(raw)
  if (col.type === 'int' || col.type === 'float') return Number(raw)
  return String(raw)
}

function mapType(type: GridColumn['type']): ColumnType {
  switch (type) {
    case 'text':
      return 'text'
    case 'email':
      return 'email'
    case 'int':
    case 'float':
      return 'number'
    case 'bool':
      return 'boolean'
    case 'date':
    case 'datetime':
      return 'date'
    case 'enum':
    case 'ref':
      return 'dropdown'
    default:
      return 'text'
  }
}

function defaultWidthFor(col: GridColumn): number {
  switch (col.type) {
    case 'email':
      return 260
    case 'bool':
      return 110
    case 'enum':
      return 180
    case 'ref':
      return 220
    case 'date':
    case 'datetime':
      return 160
    case 'int':
    case 'float':
      return 120
    default:
      return col.id.length > 18 ? 240 : 180
  }
}

/* --------------------------------- diffing -------------------------------- */

function buildOpData(
  row: Row,
  columns: GridColumn[],
  maps: Record<string, ColumnMaps>,
  forCreate: boolean,
): Record<string, unknown> {
  const data: Record<string, unknown> = {}
  for (const col of columns) {
    if (col.readonly) continue
    const display = row[col.id]
    const backendValue = toBackendValue(col, display, maps[col.id])
    if (forCreate) {
      // For create, send all editable fields (including null for optional refs).
      data[col.id] = backendValue
    } else {
      data[col.id] = backendValue
    }
  }
  return data
}

function diffPatch(
  prev: Row,
  current: Row,
  columns: GridColumn[],
  maps: Record<string, ColumnMaps>,
): Record<string, unknown> | null {
  const patch: Record<string, unknown> = {}
  for (const col of columns) {
    if (col.readonly) continue
    const prevValue = toBackendValue(col, prev[col.id], maps[col.id])
    const currValue = toBackendValue(col, current[col.id], maps[col.id])
    if (!eqBackendValue(prevValue, currValue)) {
      patch[col.id] = currValue
    }
  }
  return Object.keys(patch).length > 0 ? patch : null
}

function toBackendValue(
  col: GridColumn,
  display: unknown,
  map: ColumnMaps | undefined,
): unknown {
  if (display === null || display === undefined || display === '' || display === UNSET_LABEL) {
    if (col.type === 'bool') return Boolean(display)
    return null
  }
  if (col.type === 'enum') {
    const label = String(display)
    return map?.byLabel.get(label) ?? label
  }
  if (col.type === 'ref') {
    const label = String(display)
    return map?.byLabel.get(label) ?? null
  }
  if (col.type === 'bool') return Boolean(display)
  if (col.type === 'int') return Math.trunc(Number(display))
  if (col.type === 'float') return Number(display)
  return String(display)
}

function eqBackendValue(a: unknown, b: unknown): boolean {
  if (a === b) return true
  if (a == null && b == null) return true
  return false
}

/* --------------------------------- errors --------------------------------- */

function mapFailedToValidationErrors(
  failed: { temp_id?: string; id?: string; field?: string; message: string }[],
  tempIdByRowId: Map<string, string>,
  knownColumnIds: string[],
  fallbackColumnId: string | undefined,
): ServerValidationError[] {
  // Reverse map temp_id → table rowId to highlight the right row.
  const rowIdByTempId = new Map<string, string>()
  for (const [rowId, tempId] of tempIdByRowId) rowIdByTempId.set(tempId, rowId)

  const fallback = fallbackColumnId ?? knownColumnIds[0] ?? 'id'
  return failed.map((f) => {
    const rowId = f.id ?? (f.temp_id ? rowIdByTempId.get(f.temp_id) : undefined) ?? '—'
    const rawField = f.field ?? ''
    const columnId = rawField.replace(/^data\./, '').replace(/^body\./, '')
    const target = columnId && knownColumnIds.includes(columnId) ? columnId : fallback
    return {
      rowId,
      columnId: target,
      messages: [f.message],
    }
  })
}

function summarizeFailed(
  failed: { code: string; field?: string; message: string }[],
): string {
  const first = failed[0]
  if (!first) return 'Save error'
  if (failed.length === 1) return first.message
  return `${first.message} (${failed.length - 1} more error(s))`
}

function countByKind(applied: { kind: 'create' | 'update' | 'delete' }[]): {
  create: number
  update: number
  delete: number
} {
  const out = { create: 0, update: 0, delete: 0 }
  for (const a of applied) out[a.kind] += 1
  return out
}

function firstEditableColumnId(columns: GridColumn[]): string | undefined {
  return columns.find((c) => !c.readonly)?.id
}
