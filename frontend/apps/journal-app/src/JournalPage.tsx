import { useState } from 'react'
import { QueryClient, QueryClientProvider, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { RefreshCw, Trash2 } from 'lucide-react'
import ReactMarkdown, { type Components } from 'react-markdown'
import { DataTable, type TableColumn } from '@els/data-table'
import { Button, ConfirmDialog } from '@els/ui'
import { api } from './lib/api'

const queryClient = new QueryClient()

type Row = Record<string, unknown> & { id: string }

const PENDING_COLUMNS: TableColumn[] = [
  { id: 'id', title: 'ID', type: 'text', width: 80, readonly: true },
  { id: 'client_id', title: 'Client ID', type: 'text', width: 160, readonly: true },
  { id: 'skill', title: 'Skill', type: 'badge', width: 130, readonly: true },
  { id: 'text', title: 'Text', type: 'text', width: 360, readonly: true },
  { id: 'target', title: 'Target', type: 'text', width: 200, readonly: true },
  { id: 'outcome', title: 'Outcome', type: 'badge', width: 120, readonly: true },
  { id: 'context', title: 'Context', type: 'text', width: 220, readonly: true },
  { id: 'source', title: 'Source', type: 'text', width: 240, readonly: true },
  { id: 'meta', title: 'Meta', type: 'text', width: 240, readonly: true },
  { id: 'occurred_at', title: 'Occurred', type: 'text', width: 200, readonly: true },
  { id: 'created_at', title: 'Created', type: 'text', width: 200, readonly: true },
]

const PROCESSED_COLUMNS: TableColumn[] = [
  { id: 'id', title: 'ID', type: 'text', width: 80, readonly: true },
  { id: 'raw_event_id', title: 'Raw ID', type: 'text', width: 80, readonly: true },
  { id: 'client_id', title: 'Client ID', type: 'text', width: 160, readonly: true },
  { id: 'skill', title: 'Skill', type: 'badge', width: 120, readonly: true },
  { id: 'context', title: 'Context', type: 'text', width: 220, readonly: true },
  { id: 'action', title: 'Action', type: 'badge', width: 170, readonly: true },
  { id: 'lemma', title: 'Lemma', type: 'text', width: 130, readonly: true },
  { id: 'pos', title: 'POS', type: 'badge', width: 140, readonly: true },
  { id: 'grammar_key', title: 'Grammar', type: 'text', width: 240, readonly: true },
  { id: 'outcome', title: 'Outcome', type: 'badge', width: 120, readonly: true },
  { id: 'error', title: 'Error', type: 'text', width: 320, readonly: true },
  { id: 'source', title: 'Source', type: 'text', width: 240, readonly: true },
  { id: 'meta', title: 'Meta', type: 'text', width: 240, readonly: true },
  { id: 'occurred_at', title: 'Occurred', type: 'text', width: 200, readonly: true },
  { id: 'created_at', title: 'Created', type: 'text', width: 200, readonly: true },
]

const FAILED_COLUMNS: TableColumn[] = [
  { id: 'id', title: 'ID', type: 'text', width: 80, readonly: true },
  { id: 'client_id', title: 'Client ID', type: 'text', width: 160, readonly: true },
  { id: 'skill', title: 'Skill', type: 'badge', width: 130, readonly: true },
  { id: 'text', title: 'Text', type: 'text', width: 360, readonly: true },
  { id: 'error', title: 'Error', type: 'text', width: 420, readonly: true },
  { id: 'context', title: 'Context', type: 'text', width: 220, readonly: true },
  { id: 'source', title: 'Source', type: 'text', width: 240, readonly: true },
  { id: 'meta', title: 'Meta', type: 'text', width: 240, readonly: true },
  { id: 'occurred_at', title: 'Occurred', type: 'text', width: 200, readonly: true },
  { id: 'created_at', title: 'Created', type: 'text', width: 200, readonly: true },
]

const RAW_COLUMNS: TableColumn[] = [
  { id: 'id', title: 'ID', type: 'text', width: 80, readonly: true },
  { id: 'status', title: 'Status', type: 'badge', width: 130, readonly: true },
  { id: 'skill', title: 'Skill', type: 'badge', width: 130, readonly: true },
  { id: 'text', title: 'Text', type: 'text', width: 360, readonly: true },
  { id: 'context', title: 'Context', type: 'text', width: 220, readonly: true },
  { id: 'source', title: 'Source', type: 'text', width: 240, readonly: true },
  { id: 'meta', title: 'Meta', type: 'text', width: 240, readonly: true },
  { id: 'error', title: 'Error', type: 'text', width: 320, readonly: true },
  { id: 'occurred_at', title: 'Occurred', type: 'text', width: 200, readonly: true },
  { id: 'created_at', title: 'Created', type: 'text', width: 200, readonly: true },
]

const WORD_COLUMNS: TableColumn[] = [
  { id: 'lemma', title: 'Lemma', type: 'text', width: 160, readonly: true },
  { id: 'pos', title: 'POS', type: 'badge', width: 140, readonly: true },
  { id: 'type', title: 'Type', type: 'text', width: 110, readonly: true },
  { id: 'cefr', title: 'CEFR', type: 'text', width: 90, readonly: true },
  { id: 'frequency', title: 'Frequency', type: 'text', width: 120, readonly: true },
  { id: 'enriched', title: 'Enriched', type: 'text', width: 100, readonly: true },
  { id: 'key', title: 'Key', type: 'text', width: 220, readonly: true },
  { id: 'created_at', title: 'Created', type: 'text', width: 200, readonly: true },
]

const GRAMMAR_COLUMNS: TableColumn[] = [
  { id: 'key', title: 'Key', type: 'text', width: 240, readonly: true },
  { id: 'parent_key', title: 'Parent', type: 'text', width: 200, readonly: true },
  { id: 'title', title: 'Title', type: 'text', width: 260, readonly: true },
  { id: 'cefr_level', title: 'CEFR', type: 'text', width: 90, readonly: true },
  { id: 'enriched', title: 'Enriched', type: 'text', width: 100, readonly: true },
  { id: 'created_at', title: 'Created', type: 'text', width: 200, readonly: true },
]

const LONG_TEXT_COLUMNS = new Set(['text', 'context'])

function pad(n: number) {
  return String(n).padStart(2, '0')
}

function formatDate(v: unknown) {
  const d = new Date(v as string)
  if (Number.isNaN(d.getTime())) return String(v)
  return `${pad(d.getDate())}.${pad(d.getMonth() + 1)}.${d.getFullYear()} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function formatCell(key: string, v: unknown): string {
  if (v == null) return ''
  if (key.endsWith('_at')) return formatDate(v)
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}

function toRows(items: Row[]): Row[] {
  return items.map((e) =>
    Object.fromEntries(Object.entries(e).map(([k, v]) => [k, formatCell(k, v)])),
  ) as Row[]
}

function jsonObj(v: unknown): unknown {
  return v && typeof v === 'object' && Object.keys(v as object).length > 0 ? v : null
}

function toEventRows(items: Row[]): Row[] {
  return items.map((e) => {
    const row: Row = { id: String(e.id) }
    for (const [k, v] of Object.entries(e)) {
      row[k] = formatCell(k, v)
    }
    row.error = ''
    row.errorData = null
    row.grammarData = null
    const err = e.error
    if (err && typeof err === 'object') {
      const o = err as Record<string, unknown>
      if (e.outcome === 'ok') {
        row.grammarData = err
      } else {
        row.error = String(o.name ?? o.reason ?? o.description ?? '')
        row.errorData = err
      }
    }
    row.sourceData = jsonObj(e.source)
    row.source = row.sourceData ? '{ } json' : ''
    row.metaData = jsonObj(e.meta)
    row.meta = row.metaData ? '{ } json' : ''
    return row
  })
}

function useList(
  key: string,
  fetcher: () => Promise<unknown>,
  field: string,
  map: (items: Row[]) => Row[] = toRows,
  refetchInterval?: number,
) {
  return useQuery({
    queryKey: ['journal', key],
    queryFn: async () => map(((await fetcher()) as Record<string, Row[]>)?.[field] ?? []),
    refetchInterval,
  })
}

const DICT_COLUMNS = ['skill', 'action', 'pos', 'outcome', 'status'] as const

type Labels = Record<string, Record<string, string>>

function useDictionaries() {
  return useQuery({
    queryKey: ['journal', 'dictionaries'],
    queryFn: async () => {
      const d = ((await api.core.listCoreDictionaries()) as {
        dictionaries?: Record<string, { value: string; label: string }[]>
      }).dictionaries ?? {}
      const out: Labels = {}
      for (const [col, entries] of Object.entries(d)) {
        out[col] = Object.fromEntries(entries.map((e) => [e.value, e.label]))
      }
      return out
    },
    staleTime: Infinity,
  })
}

function labelRows(rows: Row[], labels: Labels): Row[] {
  if (!Object.keys(labels).length) return rows
  return rows.map((r) => {
    const copy = { ...r }
    for (const col of DICT_COLUMNS) {
      const v = copy[col]
      if (typeof v === 'string' && labels[col]?.[v]) copy[col] = labels[col][v]
    }
    return copy
  })
}

type ErrorData = {
  name?: string
  sentence?: string
  fragment?: string
  correction?: string
  description?: string
  reason?: string
}

function diff(removed: string, added?: string, positive?: boolean) {
  if (positive) {
    return <span className="rounded bg-emerald-100 px-1 font-medium text-emerald-700">{removed}</span>
  }
  return (
    <>
      <span className="rounded bg-red-100 px-1 text-red-700 line-through decoration-2">{removed}</span>
      {added ? <span className="ml-1 rounded bg-emerald-100 px-1 font-medium text-emerald-700">{added}</span> : null}
    </>
  )
}

function renderSentence(sentence: string, fragment?: string, correction?: string, positive?: boolean) {
  const i = fragment ? sentence.toLowerCase().indexOf(fragment.toLowerCase()) : -1
  if (!fragment || i < 0) {
    return (
      <>
        {sentence}
        {fragment ? <span className="ml-2 text-base">{diff(fragment, correction, positive)}</span> : null}
      </>
    )
  }
  return (
    <>
      {sentence.slice(0, i)}
      {diff(sentence.slice(i, i + fragment.length), correction, positive)}
      {sentence.slice(i + fragment.length)}
    </>
  )
}

function highlightJson(data: unknown): string {
  const json = JSON.stringify(data, null, 2)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  return json.replace(
    /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g,
    (match) => {
      let cls = 'text-emerald-700'
      if (/^"/.test(match)) cls = /:$/.test(match) ? 'text-sky-700' : 'text-amber-700'
      else if (/true|false/.test(match)) cls = 'text-purple-700'
      else if (/null/.test(match)) cls = 'text-neutral-400'
      return `<span class="${cls}">${match}</span>`
    },
  )
}

function JsonModal({ modal, onClose }: { modal: { title: string; data: unknown } | null; onClose: () => void }) {
  if (!modal) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className="flex max-h-[80vh] w-full max-w-2xl flex-col rounded-xl bg-white shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b border-neutral-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-neutral-900">{modal.title}</h2>
          <button onClick={onClose} className="text-2xl leading-none text-neutral-400 hover:text-neutral-700">
            ×
          </button>
        </div>
        <pre className="overflow-auto rounded-b-xl bg-neutral-50 px-6 py-4 font-mono text-sm leading-relaxed">
          <code dangerouslySetInnerHTML={{ __html: highlightJson(modal.data) }} />
        </pre>
      </div>
    </div>
  )
}

const markdownComponents: Components = {
  h1: ({ node, ...p }) => <h1 className="mb-3 mt-4 text-xl font-bold" {...p} />,
  h2: ({ node, ...p }) => <h2 className="mb-2 mt-4 text-lg font-semibold" {...p} />,
  h3: ({ node, ...p }) => <h3 className="mb-2 mt-3 text-base font-semibold" {...p} />,
  p: ({ node, ...p }) => <p className="mb-3 whitespace-pre-wrap" {...p} />,
  ul: ({ node, ...p }) => <ul className="mb-3 list-disc pl-6" {...p} />,
  ol: ({ node, ...p }) => <ol className="mb-3 list-decimal pl-6" {...p} />,
  li: ({ node, ...p }) => <li className="mb-1" {...p} />,
  a: ({ node, ...p }) => <a className="text-sky-600 underline" target="_blank" rel="noreferrer" {...p} />,
  code: ({ node, ...p }) => <code className="rounded bg-neutral-100 px-1 py-0.5 font-mono text-[0.85em]" {...p} />,
  pre: ({ node, ...p }) => <pre className="mb-3 overflow-auto rounded bg-neutral-100 p-3 font-mono text-xs" {...p} />,
  blockquote: ({ node, ...p }) => <blockquote className="mb-3 border-l-4 border-neutral-300 pl-3 italic text-neutral-600" {...p} />,
  table: ({ node, ...p }) => <table className="mb-3 border-collapse" {...p} />,
  th: ({ node, ...p }) => <th className="border border-neutral-300 px-2 py-1 text-left" {...p} />,
  td: ({ node, ...p }) => <td className="border border-neutral-300 px-2 py-1" {...p} />,
}

function TextModal({ modal, onClose }: { modal: { title: string; text: string } | null; onClose: () => void }) {
  if (!modal) return null
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className="flex max-h-[85vh] w-full max-w-3xl flex-col rounded-xl bg-white shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b border-neutral-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-neutral-900">{modal.title}</h2>
          <button onClick={onClose} className="text-2xl leading-none text-neutral-400 hover:text-neutral-700">
            ×
          </button>
        </div>
        <div className="overflow-auto px-6 py-4 text-sm leading-relaxed text-neutral-800">
          <ReactMarkdown components={markdownComponents}>{modal.text}</ReactMarkdown>
        </div>
      </div>
    </div>
  )
}

function DetailModal({ detail, onClose }: { detail: { data: ErrorData; positive: boolean } | null; onClose: () => void }) {
  if (!detail) return null
  const err = detail.data
  const box = detail.positive ? 'bg-emerald-50 text-emerald-800' : 'bg-red-50 text-red-800'
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className="w-full max-w-2xl rounded-xl bg-white shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between border-b border-neutral-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-neutral-900">{err.name || (detail.positive ? 'Construction' : 'Error')}</h2>
          <button onClick={onClose} className="text-2xl leading-none text-neutral-400 hover:text-neutral-700">
            ×
          </button>
        </div>
        <div className="flex flex-col gap-5 px-6 py-6">
          {err.sentence ? (
            <p className="text-xl leading-relaxed text-neutral-800">{renderSentence(err.sentence, err.fragment, err.correction, detail.positive)}</p>
          ) : null}
          {err.description ? (
            <div className={`rounded-lg px-4 py-3 text-sm leading-relaxed ${box}`}>{err.description}</div>
          ) : null}
          {!err.sentence && !err.description && err.reason ? (
            <div className={`rounded-lg px-4 py-3 font-mono text-sm leading-relaxed ${box}`}>{err.reason}</div>
          ) : null}
        </div>
      </div>
    </div>
  )
}

function Grid({
  title,
  columns,
  rows,
  kind,
  onReload,
  onWipe,
  onDelete,
  wiping,
  onActivate,
  height = 360,
  fill,
}: {
  title: string
  columns: TableColumn[]
  rows: Row[]
  kind: 'events' | 'raw' | 'words' | 'rules'
  onReload: () => void
  onWipe: () => void
  onDelete: (kind: string, ids: string[]) => Promise<unknown>
  wiping: boolean
  onActivate?: (columnId: string, row: Row) => void
  height?: number
  fill?: boolean
}) {
  const [selected, setSelected] = useState<Row[]>([])
  const [nonce, setNonce] = useState(0)
  const [deleting, setDeleting] = useState(false)
  const [confirming, setConfirming] = useState(false)

  const removeSelected = async () => {
    if (!selected.length) return
    setConfirming(false)
    setDeleting(true)
    try {
      await onDelete(kind, selected.map((r) => String(r.id)))
      setSelected([])
      setNonce((n) => n + 1)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="min-w-0" style={fill ? { height: '100%' } : { height }}>
      <DataTable<Row>
        key={nonce}
        data={rows}
        columns={columns}
        keyField="id"
        editMode="never"
        showControls={false}
        showRowMarkers
        showSaveControls={false}
        showValidationPanel={false}
        allowAddRows={false}
        brandColor="#059669"
        className="h-full"
        onSelectionChange={setSelected}
        title={
          <>
            {title} <span className="font-normal text-neutral-400">({rows.length})</span>
          </>
        }
        toolbar={
          <>
            {selected.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                className="h-8 gap-1.5 px-2 text-red-600 hover:bg-red-50 hover:text-red-700"
                onClick={() => setConfirming(true)}
                disabled={deleting}
                title="Delete selected"
              >
                <Trash2 className="h-4 w-4" />
                {selected.length}
              </Button>
            )}
            <Button variant="ghost" size="icon" className="h-8 w-8 text-neutral-500 hover:text-neutral-900" onClick={onReload} title="Reload">
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8 text-neutral-500 hover:bg-red-50 hover:text-red-600" onClick={onWipe} disabled={wiping} title="Wipe">
              <Trash2 className="h-4 w-4" />
            </Button>
          </>
        }
        onCellActivated={(item, columnId) => onActivate?.(columnId, item)}
      />
      {confirming && (
        <ConfirmDialog
          title="Delete rows"
          description={`Delete ${selected.length} selected row(s)?`}
          onConfirm={() => void removeSelected()}
          onClose={() => setConfirming(false)}
        />
      )}
    </div>
  )
}

const TABS = [
  { id: 'events', label: 'Events' },
  { id: 'catalog', label: 'Catalog' },
  { id: 'raw', label: 'Raw requests' },
] as const

function JournalView() {
  const [tab, setTab] = useState<(typeof TABS)[number]['id']>('events')
  const [detail, setDetail] = useState<{ data: ErrorData; positive: boolean } | null>(null)
  const [jsonModal, setJsonModal] = useState<{ title: string; data: unknown } | null>(null)
  const [textModal, setTextModal] = useState<{ title: string; text: string } | null>(null)
  const qc = useQueryClient()
  const events = useList('events', () => api.core.listCoreEvents({ params: { query: { status: 'all' } } }), 'events', toEventRows, 15000)
  const catalog = useQuery({
    queryKey: ['journal', 'catalog'],
    queryFn: async () => (await api.core.listCoreCatalog()) as { words?: Row[]; rules?: Row[] },
  })
  const labels = useDictionaries().data ?? {}

  const all = events.data ?? []
  const processed = labelRows(all.filter((e) => e.status === 'processed'), labels)
  const failed = labelRows(all.filter((e) => e.status === 'failed'), labels)
  const pending = labelRows(all.filter((e) => e.status === 'pending'), labels)
  const wordRows = labelRows(toRows(catalog.data?.words ?? []), labels)
  const ruleRows = toRows(catalog.data?.rules ?? [])
  const raw = useList('raw', () => api.core.listCoreEvents({ params: { query: { status: 'raw' } } }), 'events', toEventRows, 15000)
  const rawRows = labelRows(raw.data ?? [], labels)

  const wipe = useMutation({
    mutationFn: () => api.core.wipeCore(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['journal'] }),
  })

  const del = useMutation({
    mutationFn: (vars: { kind: string; ids: string[] }) =>
      api.core.deleteCoreRows({ body: { kind: vars.kind as 'events' | 'raw' | 'words' | 'rules', ids: vars.ids } }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['journal'] }),
  })

  const onDelete = (kind: string, ids: string[]) => del.mutateAsync({ kind, ids })

  const [confirmingWipe, setConfirmingWipe] = useState(false)
  const onWipe = () => setConfirmingWipe(true)

  const onActivate = (columnId: string, row: Row) => {
    if (columnId === 'error' && row.errorData) setDetail({ data: row.errorData as ErrorData, positive: false })
    else if (columnId === 'grammar_key' && row.grammarData) setDetail({ data: row.grammarData as ErrorData, positive: true })
    else if (columnId === 'source' && row.sourceData) setJsonModal({ title: 'Source', data: row.sourceData })
    else if (columnId === 'meta' && row.metaData) setJsonModal({ title: 'Meta', data: row.metaData })
    else if (LONG_TEXT_COLUMNS.has(columnId) && row[columnId]) {
      setTextModal({ title: columnId.charAt(0).toUpperCase() + columnId.slice(1), text: String(row[columnId]) })
    }
  }

  return (
    <div className="flex h-full w-full flex-col bg-neutral-50">
      <div className="flex shrink-0 gap-2 px-6 pb-2 pt-4">
        {TABS.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`rounded px-3 py-1.5 text-sm font-medium ${
              tab === t.id ? 'bg-emerald-600 text-white' : 'border border-neutral-300 bg-white text-neutral-600'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden px-6 pb-6 pt-2">
        {tab === 'events' && (
          <div className="flex min-h-0 flex-1 flex-col gap-4">
            <div className="min-h-0 flex-1">
              <Grid title="Processed" kind="events" columns={PROCESSED_COLUMNS} rows={processed} onReload={() => events.refetch()} onWipe={onWipe} onDelete={onDelete} wiping={wipe.isPending} onActivate={onActivate} fill />
            </div>
            <div className="grid shrink-0 grid-cols-1 gap-4 lg:grid-cols-2">
              <Grid title="Failed" kind="raw" columns={FAILED_COLUMNS} rows={failed} onReload={() => events.refetch()} onWipe={onWipe} onDelete={onDelete} wiping={wipe.isPending} onActivate={onActivate} />
              <Grid title="Pending" kind="raw" columns={PENDING_COLUMNS} rows={pending} onReload={() => events.refetch()} onWipe={onWipe} onDelete={onDelete} wiping={wipe.isPending} onActivate={onActivate} />
            </div>
          </div>
        )}
        {tab === 'catalog' && (
          <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto lg:grid lg:grid-cols-2 lg:overflow-hidden">
            <div className="h-[360px] min-w-0 lg:h-auto">
              <Grid title="Words" kind="words" columns={WORD_COLUMNS} rows={wordRows} onReload={() => catalog.refetch()} onWipe={onWipe} onDelete={onDelete} wiping={wipe.isPending} fill />
            </div>
            <div className="h-[360px] min-w-0 lg:h-auto">
              <Grid title="Grammar rules" kind="rules" columns={GRAMMAR_COLUMNS} rows={ruleRows} onReload={() => catalog.refetch()} onWipe={onWipe} onDelete={onDelete} wiping={wipe.isPending} fill />
            </div>
          </div>
        )}
        {tab === 'raw' && (
          <div className="min-h-0 flex-1">
            <Grid title="Raw requests" kind="raw" columns={RAW_COLUMNS} rows={rawRows} onReload={() => raw.refetch()} onWipe={onWipe} onDelete={onDelete} wiping={wipe.isPending} onActivate={onActivate} fill />
          </div>
        )}
      </div>

      <DetailModal detail={detail} onClose={() => setDetail(null)} />
      <JsonModal modal={jsonModal} onClose={() => setJsonModal(null)} />
      <TextModal modal={textModal} onClose={() => setTextModal(null)} />
      {confirmingWipe && (
        <ConfirmDialog
          title="Wipe journal"
          description="Wipe all events and catalog? This cannot be undone."
          confirmLabel="Wipe"
          onConfirm={() => {
            wipe.mutate()
            setConfirmingWipe(false)
          }}
          onClose={() => setConfirmingWipe(false)}
        />
      )}
    </div>
  )
}

export function JournalPage() {
  return (
    <QueryClientProvider client={queryClient}>
      <JournalView />
    </QueryClientProvider>
  )
}
