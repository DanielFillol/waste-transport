import { api } from './client'
import type { Route, PaginatedResponse } from '../types'
import type { ImportResult, DeleteResult } from '../components/ui/CsvImportModal'

export const routesApi = {
  list: (f: { search?: string; page?: number; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (f.search) q.set('search', f.search)
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<Route>>(`/routes?${q}`)
  },
  get: (id: string) => api.get<Route>(`/routes/${id}`),
  create: (data: {
    name: string
    material_id?: number
    packaging_id?: number
    treatment_id?: number
    week_day: number
    week_number: number
    driver_ids?: string[]
  }) => api.post<Route>('/routes', data),
  update: (id: string, data: Partial<{
    name: string
    material_id: number
    packaging_id: number
    treatment_id: number
    week_day: number
    week_number: number
    driver_ids: string[]
  }>) => api.patch<Route>(`/routes/${id}`, data),
  delete: (id: string) => api.delete<void>(`/routes/${id}`),
  generateCollects: (id: string, data: {
    target_date: string
    generator_ids: string[]
    receiver_id: string
  }) => api.post<{ created: number }>(`/routes/${id}/generate-collects`, data),

  import: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<ImportResult>('/routes/import', fd)
  },

  importDelete: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<DeleteResult>('/routes/import-delete', fd)
  },
}
