import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  BookOpenText,
  Brain,
  Briefcase,
  CheckCircle2,
  ChevronLeft,
  ChevronsUpDown,
  Clock,
  Coffee,
  Folder,
  Headphones,
  Loader2,
  Mic,
  PenLine,
  Plane,
  Plus,
  Shapes,
  Trash2,
} from 'lucide-react'
import { AppInfoButton, cn, useAgentView } from '@els/ui'
import { api } from './lib/api.ts'
import { isDue } from './lib/types.ts'
import { MainCard } from './components/MainCard.tsx'
import { ListeningPanel } from './components/ListeningPanel.tsx'
import { SpeakingPanel } from './components/SpeakingPanel.tsx'
import { UseItPanel } from './components/UseItPanel.tsx'
import { RecallPanel } from './components/RecallPanel.tsx'

const areaIcons: Record<string, typeof Mic> = {
  coffee: Coffee,
  briefcase: Briefcase,
  plane: Plane,
  'book-open': BookOpenText,
  'pen-line': PenLine,
}

function AreaIcon({ name }: { name?: string }) {
  const Icon = areaIcons[name ?? ''] ?? Folder
  return <Icon className="h-3.5 w-3.5 shrink-0 text-neutral-400" />
}

function SkillChip({ done, icon: Icon }: { done: boolean; icon: typeof Mic }) {
  return (
    <span
      className={cn(
        'flex h-5 w-5 items-center justify-center rounded-full',
        done ? 'bg-emerald-100 text-emerald-600' : 'bg-neutral-100 text-neutral-400',
      )}
    >
      <Icon className="h-3 w-3" />
    </span>
  )
}

export function StudioAppPage() {
  const queryClient = useQueryClient()
  const [areaId, setAreaId] = useState<string | null>(null)
  const [selected, setSelected] = useState<{ areaId: string; itemId: string } | null>(null)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [areaMenuOpen, setAreaMenuOpen] = useState(false)
  const [newArea, setNewArea] = useState('')
  const [newText, setNewText] = useState('')
  const [recallHidden, setRecallHidden] = useState(false)

  const areasQ = useQuery({
    queryKey: ['studio', 'areas'],
    queryFn: () => api.studio.studioListAreas(),
  })
  const areas = areasQ.data?.items ?? []
  const area = areas.find((a) => a.id === areaId) ?? areas[0]

  const itemsQ = useQuery({
    queryKey: ['studio', 'items', area?.id],
    queryFn: () => api.studio.studioListItems({ params: { path: { id: area!.id } } }),
    enabled: !!area,
  })
  const items = itemsQ.data?.items ?? []
  const item =
    (selected?.areaId === area?.id ? items.find((i) => i.id === selected?.itemId) : undefined) ?? items[0]

  useEffect(() => setRecallHidden(false), [item?.id])

  useAgentView({
    app: 'studio',
    screen: 'workbench',
    info: 'The user studies their own phrases across listening, speaking and writing on one screen.',
    state: { area: area?.title ?? '', phrase: item?.text ?? '' },
  })

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['studio'] })

  const createAreaM = useMutation({
    mutationFn: (title: string) => api.studio.studioCreateArea({ body: { title } }),
    onSuccess: (data) => {
      setNewArea('')
      setAreaMenuOpen(false)
      if (data) setAreaId(data.id)
      invalidate()
    },
  })

  const deleteAreaM = useMutation({
    mutationFn: (id: string) => api.studio.studioDeleteArea({ params: { path: { id } } }),
    onSuccess: () => {
      setAreaId(null)
      invalidate()
    },
  })

  const addItemM = useMutation({
    mutationFn: (text: string) =>
      api.studio.studioAddItem({ params: { path: { id: area!.id } }, body: { text } }),
    onSuccess: (data) => {
      setNewText('')
      if (data) setSelected({ areaId: data.area_id, itemId: data.id })
      invalidate()
    },
  })

  const deleteItemM = useMutation({
    mutationFn: (id: string) => api.studio.studioDeleteItem({ params: { path: { id } } }),
    onSuccess: invalidate,
  })

  const markSkillM = useMutation({
    mutationFn: ({ id, skill }: { id: string; skill: 'listened' | 'spoken' | 'written' | 'recalled' }) =>
      api.studio.studioMarkSkill({ params: { path: { id } }, body: { skill } }),
    onSuccess: invalidate,
  })

  const reviewM = useMutation({
    mutationFn: (id: string) => api.studio.studioPassReview({ params: { path: { id } } }),
    onSuccess: invalidate,
  })

  const complete = (skill: 'listened' | 'spoken' | 'written' | 'recalled') => {
    if (!item) return
    if (!item[skill]) {
      if (skill !== 'written') markSkillM.mutate({ id: item.id, skill })
    } else if (isDue(item)) {
      reviewM.mutate(item.id)
    }
  }

  const allDone = (i: (typeof items)[number]) => i.listened && i.spoken && i.written && i.recalled
  const mastered = items.filter(allDone).length
  const started = items.filter((i) => (i.listened || i.spoken || i.written || i.recalled) && !allDone(i)).length
  const untouched = items.length - mastered - started
  const dueCount = items.filter(isDue).length

  return (
    <div className="flex h-full min-h-0 w-full bg-neutral-50 text-neutral-900">
      <aside
        className={cn(
          'w-full flex-col bg-white lg:flex lg:w-80 lg:shrink-0 lg:border-r lg:border-neutral-200',
          mobileOpen ? 'hidden' : 'flex',
        )}
      >
        <div className="relative border-b border-neutral-100 px-3 py-3">
          <div className="flex items-center gap-2.5 px-1">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
              <Shapes className="h-5 w-5" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="flex items-center gap-1 text-sm font-bold leading-tight">
                Studio <AppInfoButton className="p-0.5" />
              </p>
              <p className="text-xs text-neutral-500">Your phrase collections</p>
            </div>
            {area && (
              <span className="text-xs text-neutral-400">
                {mastered}/{items.length}
              </span>
            )}
          </div>
          <button
            onClick={() => setAreaMenuOpen((v) => !v)}
            className="mt-2.5 flex w-full items-center gap-2 rounded-xl border border-neutral-200 bg-neutral-50 px-3 py-2 text-sm font-medium text-neutral-800 transition-colors hover:border-brand-300 hover:bg-white"
          >
            {area && <AreaIcon name={area.icon} />}
            <span className="flex-1 truncate text-left">{area ? area.title : 'Create a collection'}</span>
            {areas.length > 1 && (
              <span className="text-xs font-normal text-neutral-400">{areas.length}</span>
            )}
            <ChevronsUpDown className="h-4 w-4 shrink-0 text-neutral-400" />
          </button>
          {areaMenuOpen && (
            <div className="absolute left-3 right-3 top-full z-10 mt-1 rounded-xl border border-neutral-200 bg-white p-1.5 shadow-lg">
              {areas.map((a) => (
                <div key={a.id} className="group flex items-center">
                  <button
                    onClick={() => {
                      setAreaId(a.id)
                      setAreaMenuOpen(false)
                    }}
                    className={cn(
                      'flex flex-1 items-center gap-2 truncate rounded-lg px-2.5 py-1.5 text-left text-sm hover:bg-neutral-50',
                      a.id === area?.id && 'font-semibold text-brand-700',
                    )}
                  >
                    <AreaIcon name={a.icon} />
                    <span className="truncate">{a.title}</span>
                    <span className="text-xs font-normal text-neutral-400">
                      {a.done}/{a.total}
                    </span>
                    {a.due > 0 && (
                      <span className="flex items-center gap-0.5 text-xs font-medium text-amber-600">
                        <Clock className="h-3 w-3" /> {a.due}
                      </span>
                    )}
                  </button>
                  <button
                    onClick={() => deleteAreaM.mutate(a.id)}
                    className="rounded-md p-1 text-neutral-300 opacity-0 hover:bg-red-50 hover:text-red-600 group-hover:opacity-100"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              ))}
              <form
                className="mt-1 flex items-center gap-1.5 border-t border-neutral-100 pt-1.5"
                onSubmit={(e) => {
                  e.preventDefault()
                  if (newArea.trim()) createAreaM.mutate(newArea.trim())
                }}
              >
                <input
                  value={newArea}
                  onChange={(e) => setNewArea(e.target.value)}
                  placeholder="New collection…"
                  className="w-full rounded-lg border border-neutral-200 px-2.5 py-1.5 text-sm placeholder:text-neutral-400 focus:border-brand-400 focus:outline-none"
                />
                <button
                  type="submit"
                  disabled={!newArea.trim() || createAreaM.isPending}
                  className="shrink-0 rounded-lg bg-brand-600 p-1.5 text-white hover:bg-brand-700 disabled:opacity-50"
                >
                  {createAreaM.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
                </button>
              </form>
            </div>
          )}
        </div>

        {area && (
          <div className="border-b border-neutral-100 p-3">
            <form
              className="flex items-center gap-2 rounded-xl border border-neutral-200 bg-neutral-50 px-3 py-2 focus-within:border-brand-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-brand-100"
              onSubmit={(e) => {
                e.preventDefault()
                if (newText.trim() && !addItemM.isPending) addItemM.mutate(newText.trim())
              }}
            >
              {addItemM.isPending ? (
                <Loader2 className="h-4 w-4 shrink-0 animate-spin text-brand-500" />
              ) : (
                <Plus className="h-4 w-4 shrink-0 text-neutral-400" />
              )}
              <input
                value={newText}
                onChange={(e) => setNewText(e.target.value)}
                disabled={addItemM.isPending}
                placeholder={addItemM.isPending ? 'AI is preparing the card…' : 'Add a phrase or word…'}
                className="w-full bg-transparent text-sm outline-none placeholder:text-neutral-400"
              />
            </form>
          </div>
        )}

        <div className="flex-1 overflow-y-auto p-2">
          {dueCount > 0 && (
            <p className="mx-1 mb-1 flex items-center gap-1.5 rounded-xl bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-800 ring-1 ring-amber-200">
              <Clock className="h-3.5 w-3.5" /> Review today: {dueCount}
            </p>
          )}
          {items.map((i) => (
            <div key={i.id} className="group relative">
              <button
                onClick={() => {
                  setSelected({ areaId: area!.id, itemId: i.id })
                  setMobileOpen(true)
                }}
                className={cn(
                  'mt-1 w-full rounded-xl px-3 py-2.5 text-left first:mt-0',
                  i.id === item?.id
                    ? 'border border-brand-300 bg-brand-50/60 ring-1 ring-brand-200'
                    : 'hover:bg-neutral-50',
                )}
              >
                <div className="flex items-center justify-between gap-2">
                  <p className="text-sm font-medium leading-snug">{i.text}</p>
                  {isDue(i) ? (
                    <Clock className="h-4 w-4 shrink-0 text-amber-500" />
                  ) : (
                    allDone(i) && <CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-500" />
                  )}
                </div>
                <div className="mt-1.5 flex items-center gap-1">
                  <SkillChip done={i.listened} icon={Headphones} />
                  <SkillChip done={i.spoken} icon={Mic} />
                  <SkillChip done={i.written} icon={PenLine} />
                  <SkillChip done={i.recalled} icon={Brain} />
                </div>
              </button>
              <button
                onClick={() => deleteItemM.mutate(i.id)}
                className="absolute bottom-2 right-2 rounded-md p-1 text-neutral-300 opacity-0 hover:bg-red-50 hover:text-red-600 group-hover:opacity-100"
              >
                <Trash2 className="h-3.5 w-3.5" />
              </button>
            </div>
          ))}
          {area && !items.length && !itemsQ.isLoading && (
            <p className="px-3 py-6 text-center text-sm text-neutral-400">
              Add your first phrase above — AI will prepare the study card.
            </p>
          )}
        </div>

        {items.length > 0 && (
          <div className="border-t border-neutral-100 px-4 py-3">
            <div className="flex h-1.5 items-center gap-1 overflow-hidden rounded-full">
              {mastered > 0 && <div className="h-full bg-emerald-500" style={{ flex: mastered }} />}
              {started > 0 && <div className="h-full bg-brand-300" style={{ flex: started }} />}
              {untouched > 0 && <div className="h-full bg-neutral-200" style={{ flex: untouched }} />}
            </div>
            <p className="mt-2 text-xs text-neutral-400">
              {mastered} mastered · {started} in progress · {untouched} untouched
            </p>
          </div>
        )}
      </aside>

      <main
        className={cn(
          'min-h-0 flex-1 overflow-y-auto p-4 lg:block lg:overflow-hidden',
          mobileOpen ? 'block' : 'hidden',
        )}
      >
        <button
          onClick={() => setMobileOpen(false)}
          className="mb-3 inline-flex items-center gap-1.5 text-sm font-semibold text-neutral-600 lg:hidden"
        >
          <ChevronLeft className="h-4 w-4" /> All phrases
        </button>
        {item ? (
          <div className="grid gap-4 lg:h-full lg:grid-cols-3 lg:grid-rows-[minmax(0,1fr)_minmax(0,1fr)_auto]">
            <div className="order-1 flex min-h-0 flex-col lg:col-span-2 lg:row-span-2">
              <MainCard key={`main-${item.id}`} item={item} hidden={recallHidden} />
            </div>
            <div className="order-3 flex min-h-0 flex-col lg:col-start-3 lg:row-start-1">
              <SpeakingPanel key={`speak-${item.id}`} item={item} onDone={() => complete('spoken')} />
            </div>
            <div className="order-4 flex min-h-0 flex-col lg:col-start-3 lg:row-start-2">
              <RecallPanel
                key={`recall-${item.id}`}
                item={item}
                hidden={recallHidden}
                onHiddenChange={setRecallHidden}
                onDone={() => complete('recalled')}
              />
            </div>
            <div className="order-5 flex min-h-0 flex-col lg:col-span-2 lg:col-start-1 lg:row-start-3">
              <UseItPanel key={`useit-${item.id}`} item={item} onDone={() => complete('written')} />
            </div>
            <div className="order-2 flex min-h-0 flex-col lg:col-start-3 lg:row-start-3">
              <ListeningPanel key={`listen-${item.id}`} item={item} onDone={() => complete('listened')} />
            </div>
          </div>
        ) : (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-sm">
                <Shapes className="h-7 w-7" />
              </div>
              <p className="mt-4 text-lg font-bold">Your study workbench</p>
              <p className="mx-auto mt-1 max-w-sm text-sm text-neutral-500">
                {area
                  ? 'Add a phrase or word on the left — you will train it with listening, speaking and writing here.'
                  : 'Create your first collection via the switcher in the top-left corner.'}
              </p>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}