import { useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Button, Field, ImageField, Input, Textarea } from '@els/ui'
import { useUpdateBook } from '../lib/books.ts'
import type { BookSummary } from '../lib/types.ts'

export function EditBookForm({ book, onDone }: { book: BookSummary; onDone: () => void }) {
  const update = useUpdateBook()
  const [title, setTitle] = useState(book.title)
  const [author, setAuthor] = useState(book.author ?? '')
  const [description, setDescription] = useState(book.description ?? '')
  const [cover, setCover] = useState<File | null>(null)

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    update.mutate({ id: book.id, title, author, description, cover: cover ?? undefined }, { onSuccess: onDone })
  }

  return (
    <form onSubmit={submit} className="space-y-4 rounded-2xl bg-white p-5 ring-1 ring-brand-200">
      <p className="text-sm font-semibold text-neutral-900">Edit "{book.title}"</p>
      <div className="grid grid-cols-2 gap-3">
        <Field label="Title">
          <Input value={title} onChange={(e) => setTitle(e.target.value)} />
        </Field>
        <Field label="Author">
          <Input value={author} onChange={(e) => setAuthor(e.target.value)} />
        </Field>
      </div>
      <Field label="Description">
        <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} />
      </Field>
      <Field label="Cover (optional)">
        <ImageField value={cover} onChange={setCover} initialUrl={book.cover_url} aspect="aspect-[3/4]" placeholder="Add cover" className="w-28" />
      </Field>
      {update.isError && <p className="text-sm text-red-600">Failed to save. Please try again.</p>}
      <div className="flex items-center gap-3">
        <Button type="submit" variant="brand" disabled={!title.trim() || update.isPending}>
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
