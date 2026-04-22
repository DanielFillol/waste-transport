import { api } from './client'
import type { Material, Packaging, Treatment, UF, City } from '../types'

export const domainApi = {
  materials: () => api.get<Material[]>('/domain/materials'),
  packagings: () => api.get<Packaging[]>('/domain/packagings'),
  treatments: () => api.get<Treatment[]>('/domain/treatments'),
  ufs: () => api.get<UF[]>('/domain/ufs'),
  cities: (uf_id?: number) =>
    api.get<City[]>(`/domain/cities${uf_id ? `?uf_id=${uf_id}` : ''}`),
}
