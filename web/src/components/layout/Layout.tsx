import type { ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

interface LayoutProps {
  title: string
  subtitle?: string
  children: ReactNode
  actions?: ReactNode
}

export function Layout({ title, subtitle, children, actions }: LayoutProps) {
  return (
    <div className="flex h-screen bg-slate-50 overflow-hidden">
      <Sidebar />
      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="flex items-center justify-between pr-6">
          <Header title={title} subtitle={subtitle} />
          {actions && (
            <div className="flex items-center gap-2 shrink-0 pl-4">
              {actions}
            </div>
          )}
        </div>
        <main className="flex-1 overflow-y-auto p-6">
          {children}
        </main>
      </div>
    </div>
  )
}
