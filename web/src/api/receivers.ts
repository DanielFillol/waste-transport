import { api } from './client'
import type { Receiver, PaginatedResponse } from '../types'
import type { ImportResult, DeleteResult } from '../components/ui/CsvImportModal'

export const receiversApi = {
  list: (f: { search?: string; active?: boolean; page?: number; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (f.search) q.set('search', f.search)
    if (f.active !== undefined) q.set('active', String(f.active))
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<Receiver>>(`/receivers?${q}`)
  },
  get: (id: string) => api.get<Receiver>(`/receivers/${id}`),
  create: (data: Partial<Receiver>) => api.post<Receiver>('/receivers', data),
  update: (id: string, data: Partial<Receiver>) => api.patch<Receiver>(`/receivers/${id}`, data),
  delete: (id: string) => api.delete<void>(`/receivers/${id}`),

  import: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<ImportResult>('/receivers/import', fd)
  },

  importDelete: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<DeleteResult>('/receivers/import-delete', fd)
  },
}
