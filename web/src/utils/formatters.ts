export function fmtDate(d: string | null | undefined): string {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('pt-BR')
}

export function fmtDateTime(d: string | null | undefined): string {
  if (!d) return '—'
  return new Date(d).toLocaleString('pt-BR')
}

export function fmtCurrency(v: number): string {
  return v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })
}

export function fmtNumber(v: number): string {
  return v.toLocaleString('pt-BR')
}

export const COLLECT_STATUS_LABEL: Record<number, string> = {
  1: 'Planejada',
  2: 'Coletada',
  3: 'Cancelada',
}

export const COLLECT_STATUS_COLOR: Record<number, string> = {
  1: 'bg-blue-100 text-blue-700',
  2: 'bg-green-100 text-green-700',
  3: 'bg-red-100 text-red-700',
}

export const INVOICE_STATUS_LABEL: Record<string, string> = {
  draft: 'Rascunho',
  issued: 'Emitida',
  paid: 'Paga',
}

export const INVOICE_STATUS_COLOR: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-700',
  issued: 'bg-blue-100 text-blue-700',
  paid: 'bg-green-100 text-green-700',
}

export const WEEK_DAYS: Record<number, string> = {
  1: 'Segunda',
  2: 'Terça',
  3: 'Quarta',
  4: 'Quinta',
  5: 'Sexta',
  6: 'Sábado',
  7: 'Domingo',
}

export const TRUCK_COST_TYPE_LABEL: Record<string, string> = {
  fuel: 'Combustível',
  maintenance: 'Manutenção',
  other: 'Outros',
}

export const PERSONNEL_ROLE_LABEL: Record<string, string> = {
  driver: 'Motorista',
  collector: 'Coletor',
}

export const ACTION_LABEL: Record<string, string> = {
  create: 'Criação',
  update: 'Atualização',
  delete: 'Exclusão',
  import: 'Importação',
  generate: 'Geração',
  generate_collects: 'Gerar Coletas',
  bulk_status: 'Status em Lote',
  bulk_cancel: 'Cancelamento em Lote',
  bulk_assign_route: 'Atribuição de Rota',
  issue: 'Emissão',
  mark_paid: 'Marcação como Pago',
  mark_read: 'Leitura',
  mark_all_read: 'Leitura de Todos',
}

export const ENTITY_TYPE_LABEL: Record<string, string> = {
  generator: 'Gerador',
  receiver: 'Recebedor',
  driver: 'Motorista',
  truck: 'Veículo',
  route: 'Rota',
  collect: 'Coleta',
  invoice: 'Nota Fiscal',
  pricing_rule: 'Regra de Preço',
  truck_cost: 'Custo de Veículo',
  personnel_cost: 'Custo de Pessoal',
  alert: 'Alerta',
  user: 'Usuário',
}
