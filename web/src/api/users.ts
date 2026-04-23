import { api } from './client'
import type { User, PaginatedResponse } from '../types'

export interface CreateUserPayload {
  name: string
  username: string
  password: string
  role: 'admin' | 'user'
}

export interface UpdateUserPayload {
  name?: string
  role?: 'admin' | 'user'
  password?: string
}

export const usersApi = {
  list: (page = 1, limit = 20) =>
    api.get<PaginatedResponse<User>>(`/users?page=${page}&limit=${limit}`),

  create: (data: CreateUserPayload) => api.post<User>('/users', data),

  update: (id: string, data: UpdateUserPayload) => api.put<User>(`/users/${id}`, data),

  delete: (id: string) => api.delete<void>(`/users/${id}`),
}
