import { api } from './client'
import type { Alert, PaginatedResponse } from '../types'

export const alertsApi = {
  list: () => api.get<PaginatedResponse<Alert>>('/alerts?limit=50'),
  markRead: (id: string) => api.patch<Alert>(`/alerts/${id}/read`),
  markAllRead: () => api.patch<void>('/alerts/read-all'),
}
