import { api } from './client'
import type {
  PricingRule, Invoice, InvoiceStatus,
  TruckCost, PersonnelCost, FinancialSummary,
  PaginatedResponse,
} from '../types'

export const financialApi = {
  // Pricing Rules
  listPricingRules: (active?: boolean) =>
    api.get<PaginatedResponse<PricingRule>>(
      `/financial/pricing-rules?limit=100${active !== undefined ? `&active=${active}` : ''}`
    ),
  createPricingRule: (data: Partial<PricingRule>) =>
    api.post<PricingRule>('/financial/pricing-rules', data),
  updatePricingRule: (id: string, data: Partial<PricingRule>) =>
    api.put<PricingRule>(`/financial/pricing-rules/${id}`, data),
  deletePricingRule: (id: string) =>
    api.delete<void>(`/financial/pricing-rules/${id}`),

  // Invoices
  listInvoices: (f: { generator_id?: string; status?: InvoiceStatus; page?: number; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (f.generator_id) q.set('generator_id', f.generator_id)
    if (f.status) q.set('status', f.status)
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<Invoice>>(`/financial/invoices?${q}`)
  },
  getInvoice: (id: string) => api.get<Invoice>(`/financial/invoices/${id}`),
  generateInvoice: (data: { generator_id: string; period_start: string; period_end: string; notes?: string }) =>
    api.post<Invoice>('/financial/invoices/generate', data),
  issueInvoice: (id: string, due_days?: number) =>
    api.patch<Invoice>(`/financial/invoices/${id}/issue`, due_days ? { due_days } : undefined),
  markInvoicePaid: (id: string) =>
    api.patch<Invoice>(`/financial/invoices/${id}/paid`),

  // Truck Costs
  listTruckCosts: (f: { truck_id?: string; from?: string; to?: string; page?: number; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (f.truck_id) q.set('truck_id', f.truck_id)
    if (f.from) q.set('from', f.from)
    if (f.to) q.set('to', f.to)
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<TruckCost>>(`/financial/truck-costs?${q}`)
  },
  createTruckCost: (data: Partial<TruckCost> & { truck_id: string }) =>
    api.post<TruckCost>('/financial/truck-costs', data),
  updateTruckCost: (id: string, data: Partial<TruckCost>) =>
    api.put<TruckCost>(`/financial/truck-costs/${id}`, data),
  deleteTruckCost: (id: string) => api.delete<void>(`/financial/truck-costs/${id}`),

  // Personnel Costs
  listPersonnelCosts: (f: { driver_id?: string; month?: string; page?: number; limit?: number } = {}) => {
    const q = new URLSearchParams()
    if (f.driver_id) q.set('driver_id', f.driver_id)
    if (f.month) q.set('month', f.month)
    if (f.page) q.set('page', String(f.page))
    if (f.limit) q.set('limit', String(f.limit))
    return api.get<PaginatedResponse<PersonnelCost>>(`/financial/personnel-costs?${q}`)
  },
  createPersonnelCost: (data: Partial<PersonnelCost> & { driver_id: string }) =>
    api.post<PersonnelCost>('/financial/personnel-costs', data),
  updatePersonnelCost: (id: string, data: Partial<PersonnelCost>) =>
    api.put<PersonnelCost>(`/financial/personnel-costs/${id}`, data),
  deletePersonnelCost: (id: string) => api.delete<void>(`/financial/personnel-costs/${id}`),

  // Summary
  summary: (period_start: string, period_end: string) =>
    api.get<FinancialSummary>(`/financial/summary?period_start=${period_start}&period_end=${period_end}`),
}
