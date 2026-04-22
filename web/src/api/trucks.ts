import { api } from './client'
import type { Truck, PaginatedResponse } from '../types'

export const trucksApi = {
  list: (f: { search?: string; active?: boolean; page?: number; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (f.search) q.set('search', f.search)
    if (f.active !== undefined) q.set('active', String(f.active))
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<Truck>>(`/trucks?${q}`)
  },
  get: (id: string) => api.get<Truck>(`/trucks/${id}`),
  create: (data: Partial<Truck>) => api.post<Truck>('/trucks', data),
  update: (id: string, data: Partial<Truck>) => api.patch<Truck>(`/trucks/${id}`, data),
  delete: (id: string) => api.delete<void>(`/trucks/${id}`),
}
