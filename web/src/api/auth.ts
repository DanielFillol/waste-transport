import { api } from './client'
import type { AuthResponse, User } from '../types'

export const authApi = {
  register: (name: string) =>
    api.post<AuthResponse>('/auth/tenants', { name }),

  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password }),

  me: () => api.get<User>('/me'),

  refresh: () => api.post<AuthResponse>('/auth/refresh'),
}
