import { api } from './client'
import type { Tenant, User } from '../types'

export const authApi = {
  register: (name: string, username: string, password: string) =>
    api.post<{ token: string; user: User; tenant: Tenant }>('/auth/tenants', { name, username, password }),

  login: (slug: string, username: string, password: string) =>
    api.post<{ token: string; user: User }>('/auth/login', { slug, username, password }),

  me: () => api.get<User>('/me'),

  refresh: () => api.post<{ token: string; user: User }>('/auth/refresh'),
}
