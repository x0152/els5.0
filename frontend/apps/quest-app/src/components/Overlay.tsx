import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { Loader2, RotateCcw, X } from 'lucide-react'
import { avatarKey } from '../lib/helpers.ts'
import { useMission, useRegenerateImage } from '../store/missions.ts'
import type { Character } from '../lib/types.ts'

interface OverlayApi {
  openImage: (url: string, caption?: string) => void
  openCharacter: (character: Character, missionId: string) => void
}

const Ctx = createContext<OverlayApi | null>(null)

export function useOverlay(): OverlayApi {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('useOverlay must be used within OverlayProvider')
  return ctx
}

export function OverlayProvider({ children }: { children: ReactNode }) {
  const [image, setImage] = useState<{ url: string; caption?: string } | null>(null)
  const [character, setCharacter] = useState<{ character: Character; missionId: string } | null>(null)

  const openImage = useCallback((url: string, caption?: string) => setImage({ url, caption }), [])
  const openCharacter = useCallback((c: Character, missionId: string) => setCharacter({ character: c, missionId }), [])

  return (
    <Ctx.Provider value={{ openImage, openCharacter }}>
      {children}
      {image && <ImageLightbox url={image.url} caption={image.caption} onClose={() => setImage(null)} />}
      {character && (
        <CharacterModal
          character={character.character}
          missionId={character.missionId}
          onClose={() => setCharacter(null)}
          onAvatar={openImage}
        />
      )}
    </Ctx.Provider>
  )
}

function useEscape(onClose: () => void) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose()
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])
}

function ImageLightbox({ url, caption, onClose }: { url: string; caption?: string; onClose: () => void }) {
  const [show, setShow] = useState(false)
  useEffect(() => setShow(true), [])
  useEscape(onClose)

  return createPortal(
    <div
      onClick={onClose}
      className={`fixed inset-0 z-[100] flex flex-col items-center justify-center gap-3 bg-black/80 p-4 backdrop-blur-sm transition-opacity duration-200 ${show ? 'opacity-100' : 'opacity-0'}`}
    >
      <button
        type="button"
        onClick={onClose}
        className="absolute right-4 top-4 grid h-10 w-10 place-items-center rounded-full bg-white/10 text-white transition-colors hover:bg-white/20"
      >
        <X className="h-5 w-5" />
      </button>
      <img
        src={url}
        alt={caption ?? ''}
        onClick={(e) => e.stopPropagation()}
        className={`max-h-[85vh] max-w-full rounded-xl object-contain shadow-2xl transition-transform duration-200 ${show ? 'scale-100' : 'scale-95'}`}
      />
      {caption && (
        <p onClick={(e) => e.stopPropagation()} className="max-w-2xl text-center text-sm leading-relaxed text-white/80">
          {caption}
        </p>
      )}
    </div>,
    document.body,
  )
}

function CharacterModal({
  character,
  missionId,
  onClose,
  onAvatar,
}: {
  character: Character
  missionId: string
  onClose: () => void
  onAvatar: (url: string, caption?: string) => void
}) {
  const [show, setShow] = useState(false)
  useEffect(() => setShow(true), [])
  useEscape(onClose)

  const missionQ = useMission(missionId)
  const regen = useRegenerateImage(missionId)
  const mission = missionQ.data?.mission
  const key = avatarKey(character.name)
  const avatar = mission?.characterAvatars?.[key]
  const generating = mission?.characterAvatarStatus?.[key] === 'generating'

  return createPortal(
    <div
      onClick={onClose}
      className={`fixed inset-0 z-[100] flex items-center justify-center bg-black/70 p-4 backdrop-blur-sm transition-opacity duration-200 ${show ? 'opacity-100' : 'opacity-0'}`}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className={`relative w-full max-w-md overflow-hidden rounded-2xl bg-white shadow-2xl transition-transform duration-200 ${show ? 'scale-100' : 'scale-95'}`}
      >
        <button
          type="button"
          onClick={onClose}
          className="absolute right-3 top-3 grid h-8 w-8 place-items-center rounded-full bg-black/5 text-neutral-500 transition-colors hover:bg-black/10"
        >
          <X className="h-4 w-4" />
        </button>
        <div className="flex flex-col items-center gap-3 px-6 pt-8">
          <div className="relative h-24 w-24">
            {avatar ? (
              <img
                src={avatar}
                alt={character.name}
                onClick={() => onAvatar(avatar, character.name)}
                className="h-24 w-24 cursor-zoom-in rounded-full object-cover ring-2 ring-brand-100"
              />
            ) : (
              <div className="grid h-24 w-24 place-items-center rounded-full bg-brand-50 text-3xl font-bold text-brand-700 ring-2 ring-brand-100">
                {character.name.slice(0, 1).toUpperCase()}
              </div>
            )}
            {generating && (
              <div className="absolute inset-0 grid place-items-center rounded-full bg-black/40">
                <Loader2 className="h-6 w-6 animate-spin text-white" />
              </div>
            )}
          </div>
          <div className="text-center">
            <h2 className="text-lg font-bold text-neutral-900">{character.name}</h2>
            <p className="text-sm text-brand-700">{character.role}</p>
            {character.gender && <p className="text-xs text-neutral-400">{character.gender}</p>}
          </div>
          {!generating && (
            <button
              type="button"
              onClick={() => regen.mutate({ kind: 'avatar', key: character.name })}
              disabled={regen.isPending}
              className="inline-flex items-center gap-1.5 rounded-full bg-neutral-100 px-3 py-1.5 text-xs font-medium text-neutral-700 transition-colors hover:bg-neutral-200 disabled:opacity-60"
            >
              {regen.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <RotateCcw className="h-3.5 w-3.5" />}
              {avatar ? 'Regenerate portrait' : 'Generate portrait'}
            </button>
          )}
        </div>
        {character.appearance && (
          <div className="px-6 py-5">
            <div className="mb-1 text-[11px] font-semibold uppercase tracking-wider text-neutral-400">Appearance</div>
            <p className="text-sm leading-relaxed text-neutral-700">{character.appearance}</p>
          </div>
        )}
      </div>
    </div>,
    document.body,
  )
}
