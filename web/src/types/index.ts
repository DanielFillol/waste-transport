export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  pages: number
}

export interface Tenant {
  id: string
  name: string
  slug: string
}

export interface User {
  id: string
  tenant_id: string
  name: string
  email: string
  role: 'admin' | 'user'
}

export interface AuthResponse {
  token: string
  tenant?: Tenant
  user?: User
}

export interface City {
  id: number
  name: string
  uf_id: number
}

export interface UF {
  id: number
  name: string
  code: string
}

export interface Material {
  id: number
  name: string
  description: string
}

export interface Packaging {
  id: number
  name: string
  type: string
  volume: number
}

export interface Treatment {
  id: number
  name: string
  description: string
}

export interface Generator {
  id: string
  tenant_id: string
  external_id: string
  name: string
  cnpj: string
  address: string
  zipcode: string
  city_id: number | null
  city?: City
  latitude: number | null
  longitude: number | null
  active: boolean
  created_at: string
  updated_at: string
  deleted_at: string | null
}

export interface Receiver {
  id: string
  tenant_id: string
  external_id: string
  name: string
  cnpj: string
  address: string
  zipcode: string
  city_id: number | null
  city?: City
  latitude: number | null
  longitude: number | null
  license_number: string
  license_expiry: string | null
  active: boolean
  created_at: string
  updated_at: string
  deleted_at: string | null
}

export interface Driver {
  id: string
  tenant_id: string
  external_id: string
  name: string
  email: string
  phone: string
  cpf: string
  cnh_number: string
  cnh_category: string
  cnh_expiry_date: string | null
  active: boolean
  created_at: string
  updated_at: string
  deleted_at: string | null
}

export interface Truck {
  id: string
  tenant_id: string
  plate: string
  model: string
  year: number
  capacity_kg: number
  capacity_m3: number
  active: boolean
  created_at: string
  updated_at: string
  deleted_at: string | null
}

export interface Route {
  id: string
  tenant_id: string
  name: string
  material_id: number | null
  material?: Material
  packaging_id: number | null
  packaging?: Packaging
  treatment_id: number | null
  treatment?: Treatment
  week_day: number
  week_number: number
  drivers?: Driver[]
  created_at: string
  updated_at: string
  deleted_at: string | null
}

export type CollectStatus = 1 | 2 | 3
export type CollectType = 'normal' | 'emergency' | 'scheduled'
export type MeasurementUnit = 'KG' | 'L' | 'M3' | 'UN'

export interface Collect {
  id: string
  tenant_id: string
  generator_id: string
  generator?: Generator
  receiver_id: string
  receiver?: Receiver
  route_id: string | null
  route?: Route
  truck_id: string | null
  truck?: Truck
  material_id: number | null
  material?: Material
  packaging_id: number | null
  packaging?: Packaging
  treatment_id: number | null
  treatment?: Treatment
  external_id: string
  collect_type: CollectType
  planned_date: string
  status: CollectStatus
  collected_quantity: number | null
  collected_unit: MeasurementUnit | null
  collected_weight: number | null
  collected_at: string | null
  notes: string
  created_at: string
  updated_at: string
  deleted_at: string | null
}

export type InvoiceStatus = 'draft' | 'issued' | 'paid'

export interface InvoiceItem {
  id: string
  invoice_id: string
  collect_id: string
  description: string
  quantity: number
  unit: MeasurementUnit
  unit_price: number
  total_price: number
}

export interface Invoice {
  id: string
  tenant_id: string
  generator_id: string
  generator?: Generator
  invoice_number: string
  period_start: string
  period_end: string
  total_amount: number
  status: InvoiceStatus
  issued_at: string | null
  due_date: string | null
  paid_at: string | null
  notes: string
  items?: InvoiceItem[]
  created_at: string
  updated_at: string
}

export type TruckCostType = 'fuel' | 'maintenance' | 'other'

export interface TruckCost {
  id: string
  tenant_id: string
  truck_id: string
  truck?: Truck
  type: TruckCostType
  period_start: string
  period_end: string
  total_amount: number
  total_km: number
  cost_per_km: number
  notes: string
  created_at: string
  updated_at: string
}

export type PersonnelCostRole = 'driver' | 'collector'

export interface PersonnelCost {
  id: string
  tenant_id: string
  driver_id: string
  driver?: Driver
  role: PersonnelCostRole
  period_month: string
  base_salary: number
  benefits: number
  total_cost: number
  notes: string
  created_at: string
  updated_at: string
}

export interface PricingRule {
  id: string
  tenant_id: string
  collect_type: CollectType | null
  material_id: number | null
  material?: Material
  packaging_id: number | null
  packaging?: Packaging
  price_per_unit: number
  unit: MeasurementUnit
  active: boolean
  created_at: string
  updated_at: string
}

export interface FinancialSummary {
  revenue: number
  truck_costs: number
  personnel_costs: number
  gross_margin: number
}

export type AlertType = 'cnh_expiring' | 'license_expiring'

export interface Alert {
  id: string
  tenant_id: string
  type: AlertType
  title: string
  message: string
  read: boolean
  created_at: string
  updated_at: string
}

export interface AuditLog {
  id: string
  tenant_id: string
  actor_id: string | null
  entity_type: string
  entity_id: string
  action: string
  payload: string
  created_at: string
}

export interface DashboardData {
  collect_counts: Record<string, number>
  generator_count: number
  receiver_count: number
  driver_count: number
  truck_count: number
  unread_alerts: number
  financial_summary: FinancialSummary
}
