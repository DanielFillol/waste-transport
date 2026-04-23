import { api } from './client'
import type { Generator, PaginatedResponse } from '../types'
import type { ImportResult, DeleteResult } from '../components/ui/CsvImportModal'

export interface GeneratorFilters {
  search?: string
  active?: boolean
  include_deleted?: boolean
  page?: number
  limit?: number
}

export const generatorsApi = {
  list: (f: GeneratorFilters = {}) => {
    const q = new URLSearchParams()
    if (f.search) q.set('search', f.search)
    if (f.active !== undefined) q.set('active', String(f.active))
    if (f.include_deleted) q.set('include_deleted', 'true')
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<Generator>>(`/generators?${q}`)
  },

  get: (id: string) => api.get<Generator>(`/generators/${id}`),

  create: (data: Partial<Generator>) => api.post<Generator>('/generators', data),

  update: (id: string, data: Partial<Generator>) =>
    api.patch<Generator>(`/generators/${id}`, data),

  delete: (id: string) => api.delete<void>(`/generators/${id}`),

  import: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<ImportResult>('/generators/import', fd)
  },

  importDelete: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<DeleteResult>('/generators/import-delete', fd)
  },
}
