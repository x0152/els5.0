import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { BookOpen, Check, CheckCheck, FileText, FileUp, Layers, Library as LibraryIcon, Loader2, Pencil, Plus, Trash2, Upload } from 'lucide-react'
import {
  Badge,
  Button,
  ConfirmDialog,
  EmptyState,
  Field,
  FileField,
  ImageField,
  Input,
  LoadingState,
  Tabs,
  Textarea,
  useAgentView,
} from '@els/ui'
import {
  useBooks,
  useCollections,
  useDeleteBook,
  useImportArticle,
  useMarkRead,
  useUpdateCollection,
  useUploadBook,
  type CollectionMeta,
} from '../lib/books.ts'
import { EditBookForm } from '../components/EditBookForm.tsx'
import type { BookSummary } from '../lib/types.ts'

interface CollectionGroup {
  title: string
  articles: BookSummary[]
}

function groupCollections(list: BookSummary[]): CollectionGroup[] {
  const map = new Map<string, BookSummary[]>()
  for (const b of list) {
    if (b.kind !== 'article' || !b.group_title) continue
    const arr = map.get(b.group_title) ?? []
    arr.push(b)
    map.set(b.group_title, arr)
  }
  return [...map.entries()]
    .map(([title, articles]) => ({ title, articles }))
    .sort((a, b) => a.title.localeCompare(b.title))
}

export function Library() {
  const { data: books, isLoading } = useBooks()
  const { data: collectionsMeta } = useCollections()
  const deleteBook = useDeleteBook()
  const [showUpload, setShowUpload] = useState(false)
  const [editing, setEditing] = useState<BookSummary | null>(null)
  const [editingCollection, setEditingCollection] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<BookSummary | null>(null)

  const list = books ?? []
  const booksList = list.filter((b) => b.kind !== 'article')
  const articles = list.filter((b) => b.kind === 'article' && !b.group_title)
  const collections = groupCollections(list)
  const metaByTitle = useMemo(() => new Map((collectionsMeta ?? []).map((m) => [m.title, m])), [collectionsMeta])
  const isEmpty = booksList.length === 0 && articles.length === 0 && collections.length === 0

  useAgentView({
    app: 'reader',
    screen: 'library',
    info: 'The user is in the book library. Book list — list_books; book text — read_book_text.',
  })

  return (
    <div className="h-full min-h-0 w-full overflow-y-auto bg-neutral-50">
      <div className="mx-auto max-w-4xl space-y-6 p-6">
        <header className="flex items-end justify-between gap-4">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold text-neutral-900">
              <LibraryIcon className="h-6 w-6 text-brand-600" />
              Reader
            </h1>
            <p className="mt-1 text-sm text-neutral-500">Upload books and articles and read with saved position</p>
          </div>
          <Button variant="brand" onClick={() => setShowUpload((v) => !v)}>
            <Plus size={16} /> Upload
          </Button>
        </header>

        {showUpload && <UploadForm collections={collections.map((c) => c.title)} onDone={() => setShowUpload(false)} />}
        {editing && <EditBookForm book={editing} onDone={() => setEditing(null)} />}
        {editingCollection !== null && (
          <CollectionEditForm
            title={editingCollection}
            meta={metaByTitle.get(editingCollection)}
            onDone={() => setEditingCollection(null)}
          />
        )}

        {isLoading ? (
          <LoadingState />
        ) : isEmpty ? (
          <EmptyState
            icon={<LibraryIcon className="h-8 w-8" />}
            title="No books yet"
            description="Upload a book or an article to start reading."
          />
        ) : (
          <div className="space-y-8">
            {collections.length > 0 && (
              <section>
                <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">Collections</h2>
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                  {collections.map((g) => (
                    <CollectionCard
                      key={g.title}
                      group={g}
                      meta={metaByTitle.get(g.title)}
                      onEdit={() => setEditingCollection(g.title)}
                    />
                  ))}
                </div>
              </section>
            )}
            {booksList.length > 0 && (
              <section>
                <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">Books</h2>
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                  {booksList.map((b) => (
                    <BookCard
                      key={b.id}
                      book={b}
                      onEdit={() => setEditing(b)}
                      onDelete={() => setDeleting(b)}
                    />
                  ))}
                </div>
              </section>
            )}
            {articles.length > 0 && (
              <section>
                <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-neutral-500">Articles</h2>
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3">
                  {articles.map((b) => (
                    <BookCard
                      key={b.id}
                      book={b}
                      onEdit={() => setEditing(b)}
                      onDelete={() => setDeleting(b)}
                    />
                  ))}
                </div>
              </section>
            )}
          </div>
        )}

        {deleting && (
          <ConfirmDialog
            title="Delete"
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

function CollectionCard({ group, meta, onEdit }: { group: CollectionGroup; meta?: CollectionMeta; onEdit: () => void }) {
  const processing = group.articles.some((a) => a.status === 'processing')
  const total = group.articles.length
  return (
    <div className="group relative overflow-hidden rounded-2xl bg-white ring-1 ring-neutral-200 transition-colors hover:ring-brand-300">
      <Link to={`collection/${encodeURIComponent(group.title)}`} className="block">
        <div className="relative flex aspect-[3/4] items-center justify-center overflow-hidden bg-brand-50">
          {meta?.cover_url ? (
            <img src={meta.cover_url} alt={group.title} className="h-full w-full object-cover" />
          ) : (
            <Layers className="h-10 w-10 text-brand-300" />
          )}
          <span className="absolute left-2 top-2 inline-flex items-center gap-1 rounded-full bg-black/60 px-2 py-0.5 text-xs font-medium text-white">
            <Layers size={11} /> Collection
          </span>
        </div>
        <div className="p-4">
          <p className="truncate text-sm font-semibold text-neutral-900">{group.title}</p>
          <span className="mt-1 inline-flex items-center gap-1 rounded-full bg-brand-50 px-2.5 py-0.5 text-xs font-medium text-brand-700 ring-1 ring-brand-100">
            {processing && <Loader2 size={12} className="animate-spin" />}
            {total} {total === 1 ? 'article' : 'articles'}
          </span>
        </div>
      </Link>
      <button
        type="button"
        onClick={onEdit}
        title="Edit collection"
        className="absolute right-2 top-2 rounded-lg bg-white/90 p-1.5 text-neutral-600 opacity-100 ring-1 ring-neutral-200 transition-opacity hover:bg-brand-600 hover:text-white sm:opacity-0 sm:group-hover:opacity-100"
      >
        <Pencil size={14} />
      </button>
    </div>
  )
}

function CollectionEditForm({ title, meta, onDone }: { title: string; meta?: CollectionMeta; onDone: () => void }) {
  const update = useUpdateCollection()
  const [newTitle, setNewTitle] = useState(title)
  const [description, setDescription] = useState(meta?.description ?? '')
  const [cover, setCover] = useState<File | null>(null)

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTitle.trim()) return
    update.mutate({ title, newTitle: newTitle.trim(), description, cover: cover ?? undefined }, { onSuccess: onDone })
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-brand-200">
      <p className="text-sm font-semibold text-neutral-900">Edit collection "{title}"</p>
      <Field label="Title">
        <Input value={newTitle} onChange={(e) => setNewTitle(e.target.value)} />
      </Field>
      <Field label="Description">
        <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} />
      </Field>
      <Field label="Cover (optional)">
        <ImageField value={cover} onChange={setCover} initialUrl={meta?.cover_url} aspect="aspect-[3/4]" placeholder="Add cover" className="w-28" />
      </Field>
      {update.isError && <p className="text-sm text-red-600">Failed to save. Please try again.</p>}
      <div className="flex items-center gap-3">
        <Button type="submit" variant="brand" disabled={!newTitle.trim() || update.isPending}>
          {update.isPending ? (
            <>
              <Loader2 size={16} className="animate-spin" /> Saving…
            </>
          ) : (
            'Save'
          )}
        </Button>
        <button type="button" onClick={onDone} className="text-sm text-neutral-500 hover:text-neutral-700">
          Cancel
        </button>
      </div>
    </form>
  )
}

function BookCard({ book, onEdit, onDelete }: { book: BookSummary; onEdit: () => void; onDelete: () => void }) {
  const markRead = useMarkRead()
  const readable = book.status === 'ready'
  const done = readable && book.percent >= 95
  const inner = (
    <>
      <div className="relative flex aspect-[3/4] items-center justify-center overflow-hidden bg-brand-50">
        {book.cover_url ? (
          <img src={book.cover_url} alt={book.title} className="h-full w-full object-cover" />
        ) : book.kind === 'article' ? (
          <FileText className="h-10 w-10 text-brand-300" />
        ) : (
          <BookOpen className="h-10 w-10 text-brand-300" />
        )}
        {done && (
          <span className="absolute left-2 top-2 inline-flex items-center gap-1 rounded-full bg-emerald-600 px-2 py-0.5 text-xs font-medium text-white">
            <Check size={11} /> Finished
          </span>
        )}
      </div>
      <div className="p-4">
        <p className="truncate text-sm font-semibold text-neutral-900">{book.title}</p>
        {book.author && <p className="truncate text-xs text-neutral-500">{book.author}</p>}
        {book.status === 'processing' ? (
          <Badge tone="warning" className="mt-2">
            <Loader2 size={12} className="animate-spin" /> Converting…
          </Badge>
        ) : book.status === 'failed' ? (
          <Badge tone="danger" className="mt-2">
            Failed
          </Badge>
        ) : (
          <div className="mt-2 space-y-1">
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-neutral-100">
              <div
                className={`h-full rounded-full ${done ? 'bg-emerald-500' : 'bg-brand-500'}`}
                style={{ width: `${book.percent}%` }}
              />
            </div>
            <p className="text-xs text-neutral-400">{done ? 'Finished' : `${book.percent}% read`}</p>
          </div>
        )}
      </div>
    </>
  )

  return (
    <div className="group relative overflow-hidden rounded-2xl bg-white ring-1 ring-neutral-200 transition-colors hover:ring-brand-300">
      {readable ? (
        <Link to={book.id} className="block">
          {inner}
        </Link>
      ) : (
        <div className="block">{inner}</div>
      )}
      <div className="absolute right-2 top-2 flex gap-1.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
        {readable && !done && (
          <button
            type="button"
            onClick={() => markRead.mutate(book)}
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
          className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-brand-600 hover:text-white"
        >
          <Pencil size={14} />
        </button>
        <button
          type="button"
          onClick={onDelete}
          className="rounded-lg bg-white/90 p-1.5 text-neutral-600 ring-1 ring-neutral-200 hover:bg-red-600 hover:text-white"
        >
          <Trash2 size={14} />
        </button>
      </div>
    </div>
  )
}

function UploadForm({ collections, onDone }: { collections: string[]; onDone: () => void }) {
  const upload = useUploadBook()
  const importArticle = useImportArticle()
  const [mode, setMode] = useState<'book' | 'article-url' | 'article-file'>('book')
  const [url, setUrl] = useState('')
  const [groupTitle, setGroupTitle] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [title, setTitle] = useState('')
  const [author, setAuthor] = useState('')
  const [description, setDescription] = useState('')
  const [cover, setCover] = useState<File | null>(null)

  const kind = mode === 'book' ? 'book' : 'article'
  const fromUrl = mode === 'article-url'
  const pending = upload.isPending || importArticle.isPending
  const failed = upload.isError || importArticle.isError

  const reset = () => {
    setFile(null)
    setUrl('')
    setTitle('')
    setAuthor('')
    setDescription('')
    setCover(null)
    setGroupTitle('')
    onDone()
  }

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (fromUrl) {
      if (!url.trim()) return
      importArticle.mutate({ url: url.trim(), groupTitle: groupTitle.trim() || undefined }, { onSuccess: reset })
      return
    }
    if (!file) return
    upload.mutate(
      {
        file,
        kind,
        groupTitle: kind === 'article' ? groupTitle.trim() || undefined : undefined,
        title: title.trim() || undefined,
        author: author.trim() || undefined,
        description: description.trim() || undefined,
        cover: cover ?? undefined,
      },
      { onSuccess: reset },
    )
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-neutral-200">
      <Tabs
        value={mode}
        onChange={setMode}
        options={[
          { value: 'book', label: 'Book' },
          { value: 'article-url', label: 'Article from URL' },
          { value: 'article-file', label: 'Article from file' },
        ]}
      />
      {kind === 'article' && (
        <>
          <Field label="Collection (optional)">
            <Input
              value={groupTitle}
              onChange={(e) => setGroupTitle(e.target.value)}
              placeholder="Group articles under one collection"
              list="reader-collections"
            />
            <datalist id="reader-collections">
              {collections.map((c) => (
                <option key={c} value={c} />
              ))}
            </datalist>
          </Field>
        </>
      )}
      {fromUrl ? (
        <Field
          label="Article URL *"
          hint="Readable text and images are extracted automatically; the status updates on its own."
        >
          <Input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/article"
          />
        </Field>
      ) : (
        <>
          <Field
            label={`${kind === 'article' ? 'Article' : 'Book'} file (FB2 / EPUB / HTML / DOCX / ODT / RTF / MD / TXT) *`}
            hint="The file is converted to HTML automatically; the status updates on its own."
          >
            <FileField
              value={file}
              onChange={setFile}
              accept=".fb2,.epub,.html,.htm,.docx,.odt,.rtf,.md,.markdown,.txt,application/epub+zip"
              placeholder="Choose a file or drop it here"
              icon={<FileUp className="h-4 w-4" />}
            />
          </Field>
          <div className="flex gap-4">
            <div className="shrink-0">
              <Field label="Cover">
                <ImageField value={cover} onChange={setCover} aspect="aspect-[3/4]" placeholder="Add cover" className="w-28" />
              </Field>
            </div>
            <div className="flex-1 space-y-3">
              <Field label="Title">
                <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Defaults to the file name" />
              </Field>
              <Field label="Author">
                <Input value={author} onChange={(e) => setAuthor(e.target.value)} />
              </Field>
              <Field label="Description">
                <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} />
              </Field>
            </div>
          </div>
        </>
      )}
      {failed && (
        <p className="text-sm text-red-600">
          {fromUrl ? 'Could not import the article from this URL.' : 'Upload failed. Check the file and try again.'}
        </p>
      )}
      <div className="flex items-center gap-3">
        <Button type="submit" variant="brand" disabled={(fromUrl ? !url.trim() : !file) || pending}>
          {pending ? (
            <>
              <Loader2 size={16} className="animate-spin" /> {fromUrl ? 'Importing…' : 'Uploading…'}
            </>
          ) : (
            <>
              <Upload size={16} /> {fromUrl ? 'Import article' : `Upload ${kind}`}
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

