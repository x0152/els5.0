import { createContext, useContext } from 'react'

export type Viewer = (src: string, alt?: string) => void

export const ImageViewerContext = createContext<Viewer>(() => {})

export function useImageViewer(): Viewer {
  return useContext(ImageViewerContext)
}
