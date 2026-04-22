import React, { useEffect, useState, useCallback } from 'react'
import { Shield, Filter } from 'lucide-react'
import { Layout } from '../components/layout/Layout'
import { Input } from '../components/ui/Input'
import { Select } from '../components/ui/Select'
import { Button } from '../components/ui/Button'
import { Pagination } from '../components/ui/Pagination'
import { EmptyState } from '../components/ui/EmptyState'
import { PageLoader } from '../components/ui/Spinner'
import { auditApi } from '../api/audit'
import type { AuditLog } from '../types'
import { fmtDateTime, ACTION_LABEL, ENTITY_TYPE_LABEL } from '../utils/formatters'

const ENTITY_TYPES = [
  '', 'generator', 'receiver', 'driver', 'truck', 'route', 'collect',
  'invoice', 'pricing_rule', 'truck_cost', 'personnel_cost', 'alert',
]

const ACTIONS = ['', 'create', 'update', 'delete', 'generate', 'issue', 'mark_paid', 'bulk_status', 'bulk_cancel']

const ACTION_COLOR: Record<string, string> = {
  create: 'bg-green-100 text-green-700',
  update: 'bg-blue-100 text-blue-700',
  delete: 'bg-red-100 text-red-700',
  generate: 'bg-purple-100 text-purple-700',
  issue: 'bg-indigo-100 text-indigo-700',
  mark_paid: 'bg-emerald-100 text-emerald-700',
  bulk_status: 'bg-orange-100 text-orange-700',
  bulk_cancel: 'bg-red-100 text-red-700',
}

export function AuditLogs() {
  const [items, setItems] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [expanded, setExpanded] = useState<string | null>(null)

  const [entityType, setEntityType] = useState('')
  const [action, setAction] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [entityId, setEntityId] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await auditApi.list({
        entity_type: entityType || undefined,
        action: action || undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
        entity_id: entityId || undefined,
        page, limit: 30,
      })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [entityType, action, dateFrom, dateTo, entityId, page])

  useEffect(() => { load() }, [load])

  const tryParsePayload = (payload: string) => {
    try { return JSON.stringify(JSON.parse(payload), null, 2) }
    catch { return payload }
  }

  return (
    <Layout title="Histórico de Alterações" subtitle="Audit log de todas as operações do sistema">

      {/* Filters */}
      <div className="bg-white rounded-2xl border border-gray-200 mb-4 p-4">
        <div className="flex flex-wrap gap-3 items-end">
          <Select label="Entidade" value={entityType} onChange={e => { setEntityType(e.target.value); setPage(1) }}
            options={ENTITY_TYPES.map(v => ({ value: v, label: v ? (ENTITY_TYPE_LABEL[v] ?? v) : 'Todas' }))}
            className="w-48" />
          <Select label="Ação" value={action} onChange={e => { setAction(e.target.value); setPage(1) }}
            options={ACTIONS.map(v => ({ value: v, label: v ? (ACTION_LABEL[v] ?? v) : 'Todas' }))}
            className="w-44" />
          <Input label="De" type="date" value={dateFrom} onChange={e => { setDateFrom(e.target.value); setPage(1) }} className="w-40" />
          <Input label="Até" type="date" value={dateTo} onChange={e => { setDateTo(e.target.value); setPage(1) }} className="w-40" />
          <Input label="ID da entidade" value={entityId} onChange={e => { setEntityId(e.target.value); setPage(1) }}
            placeholder="UUID…" className="w-64" />
          <Button variant="secondary" icon={<Filter size={14} />} onClick={() => setPage(1)} className="mb-0.5">Filtrar</Button>
        </div>
      </div>

      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum registro" description="Nenhuma operação encontrada com os filtros selecionados."
            icon={<Shield size={32} className="text-gray-300" />} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Data/Hora', 'Entidade', 'ID', 'Ação', 'Usuário', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(log => (
                  <React.Fragment key={log.id}>
                    <tr className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-4 py-3 text-gray-500 text-xs whitespace-nowrap">{fmtDateTime(log.created_at)}</td>
                      <td className="px-4 py-3 text-gray-700">{ENTITY_TYPE_LABEL[log.entity_type] ?? log.entity_type}</td>
                      <td className="px-4 py-3 font-mono text-xs text-gray-400">{log.entity_id ? log.entity_id.slice(0, 8) + '…' : '—'}</td>
                      <td className="px-4 py-3">
                        <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${ACTION_COLOR[log.action] ?? 'bg-gray-100 text-gray-700'}`}>
                          {ACTION_LABEL[log.action] ?? log.action}
                        </span>
                      </td>
                      <td className="px-4 py-3 font-mono text-xs text-gray-400">{log.actor_id ? log.actor_id.slice(0, 8) + '…' : 'Sistema'}</td>
                      <td className="px-4 py-3">
                        {log.payload && (
                          <button onClick={() => setExpanded(expanded === log.id ? null : log.id)}
                            className="text-xs text-brand-500 hover:text-brand-700 hover:underline">
                            {expanded === log.id ? 'Ocultar' : 'Ver payload'}
                          </button>
                        )}
                      </td>
                    </tr>
                    {expanded === log.id && log.payload && (
                      <tr key={`${log.id}-payload`}>
                        <td colSpan={6} className="px-4 py-3 bg-gray-50">
                          <pre className="text-xs text-gray-600 overflow-x-auto whitespace-pre-wrap break-all font-mono bg-gray-900 text-green-400 p-4 rounded-xl max-h-64 overflow-y-auto">
                            {tryParsePayload(log.payload)}
                          </pre>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {!loading && items.length > 0 && (
          <div className="px-4 border-t border-gray-100">
            <Pagination page={page} pages={pages} total={total} limit={30} onChange={setPage} />
          </div>
        )}
      </div>
    </Layout>
  )
}
