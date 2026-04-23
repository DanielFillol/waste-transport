import { render, type RenderOptions } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { ToastProvider } from '../components/ui/Toast'
import type { ReactElement } from 'react'

export function renderWithRouter(
  ui: ReactElement,
  options?: RenderOptions & { initialEntries?: string[] },
) {
  const { initialEntries, ...rest } = options ?? {}
  return render(
    <MemoryRouter initialEntries={initialEntries ?? ['/']}>
      <ToastProvider>{ui}</ToastProvider>
    </MemoryRouter>,
    rest,
  )
}

export const mockUser = {
  id: 'user-1',
  tenant_id: 'tenant-1',
  name: 'Test User',
  email: 'test@test.com',
  role: 'admin' as const,
}

export function paginated<T>(data: T[], total?: number) {
  return { data, total: total ?? data.length, page: 1, limit: 20, pages: 1 }
}
