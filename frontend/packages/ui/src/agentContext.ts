import { useEffect, useRef } from 'react'

export type AgentView = {
  app: string
  screen: string
  title?: string
  info?: string
  ids?: Record<string, string>
  state?: Record<string, string | number>
}

const stack: { current: AgentView | null }[] = []

export function readAgentContext(): { view?: AgentView } {
  for (let i = stack.length - 1; i >= 0; i--) {
    const v = stack[i]?.current
    if (v) return { view: v }
  }
  return {}
}

export function useAgentView(view: AgentView | null) {
  const ref = useRef<AgentView | null>(view)
  ref.current = view
  useEffect(() => {
    stack.push(ref)
    return () => {
      const i = stack.indexOf(ref)
      if (i >= 0) stack.splice(i, 1)
    }
  }, [])
}
