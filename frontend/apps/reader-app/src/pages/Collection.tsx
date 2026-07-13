import { useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { ArrowLeft, Check, CheckCheck, FileText, FileUp, Layers, Loader2, Pencil, Plus, Trash2, Upload } from 'lucide-react'
import { Button, ConfirmDialog, FileField, Input, LoadingState, Tabs, cn, useAgentView } from '@els/ui'
import { useBooks, useDeleteBook, useDeleteCollection, useImportArticle, useMarkRead, useUploadBook } from '../lib/books.ts'
import { EditBookForm } from '../components/EditBookForm.tsx'
import type { BookSummary } from '../lib/types.ts'

export function Collection() {
  const { key = '' } = useParams()
  const navigate = useNavigate()
  const title = decodeURIComponent(key)
  const { data: books, isLoading } = useBooks()
  const deleteBook = useDeleteBook()
  const deleteCollection = useDeleteCollection()
  const [editing, setEditing] = useState<BookSummary | null>(null)
  const [showAdd, setShowAdd] = useState(false)
  const [deleting, setDeleting] = useState<BookSummary | null>(null)
  const [deletingCollection, setDeletingCollection] = useState(false)

  const articles = useMemo(
    () =>
      (books ?? [])
        .filter((b) => b.kind === 'article' && b.group_title === title)
        .sort((a, b) => a.created_at.localeCompare(b.created_at)),
    [books, title],
  )

  useAgentView({
    app: 'reader',
    screen: 'collection',
    title,
    info: 'The user is viewing articles in a collection. Book list — list_books; text — read_book_text.',
    state: { articles: articles.length },
  })

  if (isLoading) {
    return <LoadingState className="h-full items-center bg-neutral-50 py-0" />
  }

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-3xl space-y-6 p-6">
        <header className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => navigate('..')}
            className="rounded-lg p-1.5 text-neutral-500 transition-colors hover:bg-neutral-100 hover:text-neutral-900"
          >
            <ArrowLeft size={18} />
          </button>
          <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
            <Layers className="h-6 w-6 text-brand-600" />
            {title}
          </h1>
          <div className="ml-auto flex items-center gap-2">
            <Button variant="brand" onClick={() => setShowAdd((v) => !v)}>
              <Plus size={16} /> Add article
            </Button>
            {articles.length > 0 && (
              <button
                type="button"
                onClick={() => setDeletingCollection(true)}
                className="inline-flex items-center gap-1.5 rounded-lg border border-red-200 bg-white px-3 py-1.5 text-sm font-medium text-red-600 transition-colors hover:bg-red-50"
              >
                <Trash2 size={15} /> Delete collection
              </button>
            )}
          </div>
        </header>

        {showAdd && <AddArticleForm title={title} onDone={() => setShowAdd(false)} />}
        {editing && <EditBookForm book={editing} onDone={() => setEditing(null)} />}

        {articles.length === 0 ? (
          <p className="py-16 text-center text-sm text-neutral-500">No articles yet.</p>
        ) : (
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
            {articles.map((a) => (
              <ArticleTile
                key={a.id}
                article={a}
                onEdit={() => setEditing(a)}
                onDelete={() => setDeleting(a)}
              />
            ))}
          </div>
        )}

        {deletingCollection && (
          <ConfirmDialog
            title="Delete collection"
            description={`Delete collection "${title}" and all ${articles.length} articles? This cannot be undone.`}
            pending={deleteCollection.isPending}
            onConfirm={() =>
              deleteCollection.mutate(
                articles.map((a) => a.id),
                { onSuccess: () => navigate('..') },
              )
            }
            onClose={() => setDeletingCollection(false)}
          />
        )}
        {deleting && (
          <ConfirmDialog
            title="Delete article"
            description={`Delete "${deleting.title}"? This cannot be undone.`}
            onConfirm={() => {
              deleteBook.mutate(deleting.id)
              setDeleting(null)
            }}
            onClose={() => setDeleting(null)}
          />
        )}
      </div>
    </div>
  )
}

function AddArticleForm({ title, onDone }: { title: string; onDone: () => void }) {
  const upload = useUploadBook()
  const importArticle = useImportArticle()
  const [source, setSource] = useState<'file' | 'url'>('url')
  const [file, setFile] = useState<File | null>(null)
  const [url, setUrl] = useState('')
  const pending = upload.isPending || importArticle.isPending
  const failed = upload.isError || importArticle.isError

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (source === 'url') {
      if (!url.trim()) return
      importArticle.mutate({ url: url.trim(), groupTitle: title }, { onSuccess: onDone })
      return
    }
    if (!file) return
    upload.mutate({ file, kind: 'article', groupTitle: title }, { onSuccess: onDone })
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-brand-200">
      <Tabs
        value={source}
        onChange={setSource}
        options={[
          { value: 'url', label: 'URL' },
          { value: 'file', label: 'File' },
        ]}
      />
      {source === 'url' ? (
        <Input
          type="url"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="https://example.com/article"
        />
      ) : (
        <FileField
          value={file}
          onChange={setFile}
          accept=".fb2,.epub,.html,.htm,.docx,.odt,.rtf,.md,.markdown,.txt,application/epub+zip"
          placeholder="Choose a file or drop it here"
          icon={<FileUp className="h-4 w-4" />}
        />
      )}
      {failed && <p className="text-sm text-red-600">Could not add the article.</p>}
      <div className="flex items-center gap-3">
        <Button type="submit" variant="brand" disabled={(source === 'url' ? !url.trim() : !file) || pending}>
          {pending ? (
            <>
              <Loader2 size={16} className="animate-spin" /> Adding…
            </>
          ) : (
            <>
              <Upload size={16} /> Add article
            </>
          )}
        </Button>
        <button type="button" onClick={onDone} className="text-sm text-neutral-500 hover:text-neutral-700">
          Cancel
        </button>
      </div>
    </form>
  )
}

function ArticleTile({ article: a, onEdit, onDelete }: { article: BookSummary; onEdit: () => void; onDelete: () => void }) {
  const markRead = useMarkRead()
  const ready = a.status === 'ready'
  const done = ready && a.percent >= 95
  const content = (
    <>
      <div className="relative flex aspect-[3/4] items-center justify-center overflow-hidden bg-brand-50">
        {a.cover_url ? (
          <img src={a.cover_url} alt={a.title} className="h-full w-full object-cover" />
        ) : (
          <FileText className="h-8 w-8 text-brand-300" />
        )}
        {done && (
          <span className="absolute right-2 top-2 inline-flex items-center rounded-full bg-emerald-600 p-1 text-white">
            <Check size={12} />
          </span>
        )}
        {ready && !done && a.percent > 0 && (
          <div className="absolute inset-x-0 bottom-0 h-1 bg-black/40">
            <div className="h-full bg-brand-500" style={{ width: `${a.percent}%` }} />
          </div>
        )}
      </div>
      <div className="p-3">
        <p className="truncate text-sm font-semibold text-neutral-900">{a.title}</p>
        {a.status === 'processing' ? (
          <span className="mt-1 inline-flex items-center gap-1 text-xs text-amber-600">
            <Loader2 size={12} className="animate-spin" /> Converting…
          </span>
        ) : a.status === 'failed' ? (
          <span className="mt-1 inline-block text-xs text-red-600">Failed</span>
        ) : null}
      </div>
    </>
  )

  return (
    <div className={cn('group relative overflow-hidden rounded-2xl bg-white ring-1 transition-colors', ready ? 'ring-neutral-200 hover:ring-brand-300' : 'opacity-60 ring-neutral-200')}>
      {ready ? (
        <Link to={`../${a.id}`} className="block">
          {content}
        </Link>
      ) : (
        content
      )}
      <div className="absolute right-2 top-2 z-10 flex gap-1.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
        {ready && !done && (
          <button
            type="button"
            onClick={() => markRead.mutate(a)}
            disabled={markRead.isPending}
            title="Mark as read"
            className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-emerald-600 hover:text-white"
          >
            {markRead.isPending ? <Loader2 size={14} className="animate-spin" /> : <CheckCheck size={14} />}
          </button>
        )}
        <button
          type="button"
          onClick={onEdit}
          title="Edit article"
          className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-brand-600 hover:text-white"
        >
          <Pencil size={14} />
        </button>
        <button
          type="button"
          onClick={onDelete}
          title="Delete article"
          className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-red-600 hover:text-white"
        >
          <Trash2 size={14} />
        </button>
      </div>
    </div>
  )
}
