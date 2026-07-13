import { GridView } from './components/GridView.tsx'
import { ToastProvider } from './components/Toasts.tsx'

export function AdminAppPage() {
  return (
    <ToastProvider>
      <div className="h-full flex flex-col bg-white overflow-hidden">
        <header className="px-6 pt-6 pb-4 border-b border-neutral-200 bg-white shrink-0">
          <h1 className="text-2xl font-semibold tracking-tight">Users</h1>
          <p className="text-sm text-neutral-500 mt-1">Manage platform users.</p>
        </header>

        <div className="flex-1 flex min-h-0 bg-neutral-50">
          <GridView basePath="/api/v1/administrators" />
        </div>
      </div>
    </ToastProvider>
  )
}
