import { api } from './client'
import type { AuditLog, PaginatedResponse } from '../types'

export const auditApi = {
  list: (f: {
    entity_type?: string
    entity_id?: string
    actor_id?: string
    action?: string
    date_from?: string
    date_to?: string
    page?: number
    limit?: number
  } = {}) => {
    const q = new URLSearchParams()
    if (f.entity_type) q.set('entity_type', f.entity_type)
    if (f.entity_id) q.set('entity_id', f.entity_id)
    if (f.actor_id) q.set('actor_id', f.actor_id)
    if (f.action) q.set('action', f.action)
    if (f.date_from) q.set('date_from', f.date_from)
    if (f.date_to) q.set('date_to', f.date_to)
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<AuditLog>>(`/audit-logs?${q}`)
  },
}
