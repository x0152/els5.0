import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
  componentStack: string | null
}

/**
 * Dev-only error boundary used by `<DevShell/>`.
 *
 * The goal here is NOT to gracefully degrade — main-app handles that
 * in production. The goal is the opposite: turn a white screen into
 * something that immediately tells the developer what blew up, with
 * stack + component stack + a "reload" button.
 */
export class DevErrorBoundary extends Component<Props, State> {
  override state: State = { error: null, componentStack: null }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { error }
  }

  override componentDidCatch(error: Error, info: ErrorInfo): void {
    this.setState({ componentStack: info.componentStack ?? null })
    console.error('[dev-harness] uncaught error', error, info)
  }

  private reset = (): void => {
    this.setState({ error: null, componentStack: null })
  }

  override render(): ReactNode {
    const { error, componentStack } = this.state
    if (!error) return this.props.children

    return (
      <div className="h-full w-full overflow-auto bg-red-50 text-red-900 p-6 font-mono text-xs">
        <div className="max-w-4xl mx-auto space-y-4">
          <div className="flex items-center justify-between">
            <h1 className="text-lg font-bold">💥 Render error in feature</h1>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={this.reset}
                className="px-3 py-1 bg-white border border-red-200 rounded hover:bg-red-100"
              >
                retry render
              </button>
              <button
                type="button"
                onClick={() => window.location.reload()}
                className="px-3 py-1 bg-white border border-red-200 rounded hover:bg-red-100"
              >
                reload page
              </button>
            </div>
          </div>

          <section>
            <div className="font-semibold mb-1">{error.name}: {error.message}</div>
            {error.stack && (
              <pre className="whitespace-pre-wrap bg-white/70 p-3 rounded border border-red-200">
                {error.stack}
              </pre>
            )}
          </section>

          {componentStack && (
            <section>
              <div className="font-semibold mb-1">Component stack</div>
              <pre className="whitespace-pre-wrap bg-white/70 p-3 rounded border border-red-200">
                {componentStack}
              </pre>
            </section>
          )}
        </div>
      </div>
    )
  }
}
