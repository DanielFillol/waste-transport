import { api } from './client'
import type { Collect, CollectStatus, PaginatedResponse } from '../types'
import type { ImportResult, DeleteResult } from '../components/ui/CsvImportModal'

export interface CollectFilters {
  generator_id?: string
  receiver_id?: string
  route_id?: string
  truck_id?: string
  material_id?: number
  packaging_id?: number
  status?: CollectStatus
  collect_type?: string
  date_from?: string
  date_to?: string
  include_deleted?: boolean
  page?: number
  limit?: number
}

export const collectsApi = {
  list: (f: CollectFilters = {}) => {
    const q = new URLSearchParams()
    if (f.generator_id) q.set('generator_id', f.generator_id)
    if (f.receiver_id) q.set('receiver_id', f.receiver_id)
    if (f.route_id) q.set('route_id', f.route_id)
    if (f.truck_id) q.set('truck_id', f.truck_id)
    if (f.material_id) q.set('material_id', String(f.material_id))
    if (f.packaging_id) q.set('packaging_id', String(f.packaging_id))
    if (f.status) q.set('status', String(f.status))
    if (f.collect_type) q.set('collect_type', f.collect_type)
    if (f.date_from) q.set('date_from', f.date_from)
    if (f.date_to) q.set('date_to', f.date_to)
    if (f.include_deleted) q.set('include_deleted', 'true')
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<Collect>>(`/collects?${q}`)
  },
  get: (id: string) => api.get<Collect>(`/collects/${id}`),
  create: (data: Partial<Collect> & { generator_id: string; receiver_id: string; planned_date: string }) =>
    api.post<Collect>('/collects', data),
  update: (id: string, data: Partial<Collect>) =>
    api.patch<Collect>(`/collects/${id}`, data),
  delete: (id: string) => api.delete<void>(`/collects/${id}`),
  bulkStatus: (ids: string[], status: CollectStatus) =>
    api.post('/collects/bulk-status', { ids, status }),
  bulkCancel: (ids: string[]) =>
    api.post('/collects/bulk-cancel', { ids }),
  bulkAssignRoute: (ids: string[], route_id: string) =>
    api.post('/collects/bulk-assign-route', { ids, route_id }),

  import: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<ImportResult>('/collects/import', fd)
  },

  importDelete: (file: File) => {
    const fd = new FormData(); fd.append('file', file)
    return api.postForm<DeleteResult>('/collects/import-delete', fd)
  },
}
