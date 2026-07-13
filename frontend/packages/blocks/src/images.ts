import { createContext, useContext } from 'react'

export type IllustrationStatus = {
  id: string
  status: 'pending' | 'generating' | 'ready' | 'error'
  url?: string
  error?: string
}

export type ImageAspect = 'square' | 'landscape' | 'portrait'

export type ImageApi = (prompt: string, trigger: boolean, aspect: ImageAspect) => Promise<IllustrationStatus>

export const ImageApiCtx = createContext<ImageApi | null>(null)

export const useImageApi = () => useContext(ImageApiCtx)
