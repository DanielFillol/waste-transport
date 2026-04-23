import { api } from './client'
import type { Driver, PaginatedResponse } from '../types'
import type { ImportResult, DeleteResult } from '../components/ui/CsvImportModal'

export const driversApi = {
  list: (f: { search?: string; active?: boolean; page?: number; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (f.search) q.set('search', f.search)
    if (f.active !== undefined) q.set('active', String(f.active))
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<Driver>>(`/drivers?${q}`)
  },
  get: (id: string) => api.get<Driver>(`/drivers/${id}`),
  create: (data: Partial<Driver>) => api.post<Driver>('/drivers', data),
  update: (id: string, data: Partial<Driver>) => api.patch<Driver>(`/drivers/${id}`, data),
  delete: (id: string) => api.delete<void>(`/drivers/${id}`),

  import: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<ImportResult>('/drivers/import', fd)
  },

  importDelete: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<DeleteResult>('/drivers/import-delete', fd)
  },
}
