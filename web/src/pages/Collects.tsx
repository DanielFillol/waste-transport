import { useEffect, useState, useCallback } from 'react'
import { Plus, Search, Filter, CheckSquare, XSquare, Truck, CalendarDays, Pencil, Trash2 } from 'lucide-react'
import { Layout } from '../components/layout/Layout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Select } from '../components/ui/Select'
import { Badge } from '../components/ui/Badge'
import { Modal } from '../components/ui/Modal'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { Pagination } from '../components/ui/Pagination'
import { EmptyState } from '../components/ui/EmptyState'
import { PageLoader } from '../components/ui/Spinner'
import { useToast } from '../components/ui/Toast'
import { collectsApi, type CollectFilters } from '../api/collects'
import { generatorsApi } from '../api/generators'
import { receiversApi } from '../api/receivers'
import { trucksApi } from '../api/trucks'
import { routesApi } from '../api/routes'
import type { Collect, CollectStatus, Generator, Receiver, Truck, Route } from '../types'
import { fmtDate, fmtDateTime, COLLECT_STATUS_LABEL, COLLECT_STATUS_COLOR } from '../utils/formatters'

interface FormState {
  generator_id: string
  receiver_id: string
  truck_id: string
  route_id: string
  planned_date: string
  collect_type: string
  notes: string
  collected_quantity: string
  collected_unit: string
  collected_weight: string
}

const EMPTY: FormState = {
  generator_id: '', receiver_id: '', truck_id: '', route_id: '',
  planned_date: new Date().toISOString().split('T')[0],
  collect_type: 'normal', notes: '',
  collected_quantity: '', collected_unit: 'KG', collected_weight: '',
}

const STATUS_OPTIONS = [
  { value: '', label: 'Todos status' },
  { value: '1', label: 'Planejada' },
  { value: '2', label: 'Coletada' },
  { value: '3', label: 'Cancelada' },
]

const TYPE_OPTIONS = [
  { value: '', label: 'Todos tipos' },
  { value: 'normal', label: 'Normal' },
  { value: 'emergency', label: 'Emergencial' },
  { value: 'scheduled', label: 'Agendada' },
]

const UNIT_OPTIONS = [
  { value: 'KG', label: 'Kg' },
  { value: 'L', label: 'Litros' },
  { value: 'M3', label: 'm³' },
  { value: 'UN', label: 'Unidades' },
]

export function Collects() {
  const toast = useToast()
  const [items, setItems] = useState<Collect[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState<CollectFilters>({})
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [selected, setSelected] = useState<Set<string>>(new Set())

  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<Collect | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<Collect | null>(null)
  const [saving, setSaving] = useState(false)

  // Assign route modal
  const [assignRouteOpen, setAssignRouteOpen] = useState(false)
  const [assignRouteId, setAssignRouteId] = useState('')

  // Reference data
  const [generators, setGenerators] = useState<Generator[]>([])
  const [receivers, setReceivers] = useState<Receiver[]>([])
  const [trucks, setTrucks] = useState<Truck[]>([])
  const [routes, setRoutes] = useState<Route[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await collectsApi.list({ ...filters, page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
      setSelected(new Set())
    } finally { setLoading(false) }
  }, [filters, page])

  useEffect(() => { load() }, [load])

  useEffect(() => {
    generatorsApi.list({ limit: 200 }).then(r => setGenerators(r.data))
    receiversApi.list({ limit: 200 }).then(r => setReceivers(r.data))
    trucksApi.list({ limit: 100 }).then(r => setTrucks(r.data))
    routesApi.list({ limit: 100 }).then(r => setRoutes(r.data))
  }, [])

  const applySearch = () => setFilters(p => ({ ...p }))

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (c: Collect) => {
    setEditing(c)
    setForm({
      generator_id: c.generator_id,
      receiver_id: c.receiver_id,
      truck_id: c.truck_id ?? '',
      route_id: c.route_id ?? '',
      planned_date: c.planned_date?.split('T')[0] ?? '',
      collect_type: c.collect_type,
      notes: c.notes,
      collected_quantity: c.collected_quantity != null ? String(c.collected_quantity) : '',
      collected_unit: c.collected_unit ?? 'KG',
      collected_weight: c.collected_weight != null ? String(c.collected_weight) : '',
    })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload: Partial<Collect> & { generator_id: string; receiver_id: string; planned_date: string } = {
        generator_id: form.generator_id,
        receiver_id: form.receiver_id,
        truck_id: form.truck_id || undefined,
        route_id: form.route_id || undefined,
        planned_date: form.planned_date,
        collect_type: form.collect_type as Collect['collect_type'],
        notes: form.notes,
        collected_quantity: form.collected_quantity ? Number(form.collected_quantity) : undefined,
        collected_unit: form.collected_unit as Collect['collected_unit'],
        collected_weight: form.collected_weight ? Number(form.collected_weight) : undefined,
      }
      if (editing) { await collectsApi.update(editing.id, payload); toast('Coleta atualizada') }
      else { await collectsApi.create(payload); toast('Coleta criada') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await collectsApi.delete(deleting.id); toast('Coleta excluída'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const bulkMarkCollected = async () => {
    if (selected.size === 0) return
    try { await collectsApi.bulkStatus([...selected], 2); toast(`${selected.size} coleta(s) marcada(s) como coletada`); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const bulkCancel = async () => {
    if (selected.size === 0) return
    try { await collectsApi.bulkCancel([...selected]); toast(`${selected.size} coleta(s) cancelada(s)`); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const bulkAssignRoute = async () => {
    if (!assignRouteId || selected.size === 0) return
    try {
      await collectsApi.bulkAssignRoute([...selected], assignRouteId)
      toast(`Rota atribuída a ${selected.size} coleta(s)`)
      setAssignRouteOpen(false)
      load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const toggleSelect = (id: string) => {
    setSelected(p => {
      const s = new Set(p)
      s.has(id) ? s.delete(id) : s.add(id)
      return s
    })
  }

  const toggleAll = () => {
    if (selected.size === items.length) setSelected(new Set())
    else setSelected(new Set(items.map(i => i.id)))
  }

  const genName = (id: string) => generators.find(g => g.id === id)?.name ?? id.slice(0, 8)
  const recName = (id: string) => receivers.find(r => r.id === id)?.name ?? id.slice(0, 8)

  return (
    <Layout title="Coletas" subtitle="Gerenciamento de coletas"
      actions={<Button icon={<Plus size={16} />} onClick={openCreate}>Nova Coleta</Button>}>

      {/* Filters */}
      <div className="bg-white rounded-2xl border border-gray-200 mb-4 p-4">
        <div className="flex flex-wrap gap-3 items-end">
          <Select label="Status" value={String(filters.status ?? '')}
            onChange={e => { const v = e.target.value; setFilters(p => ({ ...p, status: v ? Number(v) as CollectStatus : undefined })); setPage(1) }}
            options={STATUS_OPTIONS} className="w-40" />
          <Select label="Tipo" value={filters.collect_type ?? ''}
            onChange={e => { setFilters(p => ({ ...p, collect_type: e.target.value || undefined })); setPage(1) }}
            options={TYPE_OPTIONS} className="w-40" />
          <Input label="De" type="date" value={filters.date_from ?? ''}
            onChange={e => { setFilters(p => ({ ...p, date_from: e.target.value || undefined })); setPage(1) }}
            className="w-40" />
          <Input label="Até" type="date" value={filters.date_to ?? ''}
            onChange={e => { setFilters(p => ({ ...p, date_to: e.target.value || undefined })); setPage(1) }}
            className="w-40" />
          <Button variant="secondary" icon={<Filter size={14} />} onClick={applySearch} className="mb-0.5">Filtrar</Button>
        </div>
      </div>

      {/* Bulk actions */}
      {selected.size > 0 && (
        <div className="mb-4 px-4 py-3 bg-brand-50 border border-brand-200 rounded-xl flex items-center gap-3">
          <span className="text-sm font-medium text-brand-700">{selected.size} selecionada(s)</span>
          <div className="flex gap-2 ml-auto">
            <Button size="sm" variant="secondary" icon={<CheckSquare size={14} />} onClick={bulkMarkCollected}>Marcar Coletada</Button>
            <Button size="sm" variant="secondary" icon={<Truck size={14} />} onClick={() => setAssignRouteOpen(true)}>Atribuir Rota</Button>
            <Button size="sm" variant="danger" icon={<XSquare size={14} />} onClick={bulkCancel}>Cancelar</Button>
          </div>
        </div>
      )}

      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        <div className="p-4">
          <Input placeholder="Buscar por gerador…" value={search}
            onChange={e => { setSearch(e.target.value); setFilters(p => ({ ...p })); setPage(1) }}
            leftIcon={<Search size={15} />} className="max-w-sm" />
        </div>
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhuma coleta" action={<Button onClick={openCreate}>Criar Coleta</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                <th className="px-4 py-3 w-10">
                  <input type="checkbox" checked={selected.size === items.length && items.length > 0}
                    onChange={toggleAll} className="rounded border-gray-300 text-brand-500 focus:ring-brand-400" />
                </th>
                {['Gerador', 'Recebedor', 'Data Planejada', 'Tipo', 'Status', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(c => (
                  <tr key={c.id} className={`hover:bg-gray-50/50 transition-colors ${selected.has(c.id) ? 'bg-brand-50/30' : ''}`}>
                    <td className="px-4 py-3">
                      <input type="checkbox" checked={selected.has(c.id)} onChange={() => toggleSelect(c.id)}
                        className="rounded border-gray-300 text-brand-500 focus:ring-brand-400" />
                    </td>
                    <td className="px-4 py-3 font-medium text-gray-900 max-w-[160px] truncate">{c.generator?.name ?? genName(c.generator_id)}</td>
                    <td className="px-4 py-3 text-gray-500 max-w-[140px] truncate">{c.receiver?.name ?? recName(c.receiver_id)}</td>
                    <td className="px-4 py-3 text-gray-500">
                      <span className="flex items-center gap-1">
                        <CalendarDays size={12} className="text-gray-400" />
                        {fmtDate(c.planned_date)}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="text-xs text-gray-500 capitalize">{c.collect_type}</span>
                    </td>
                    <td className="px-4 py-3">
                      <Badge className={COLLECT_STATUS_COLOR[c.status]}>
                        {COLLECT_STATUS_LABEL[c.status]}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button onClick={() => openEdit(c)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600"><Pencil size={15} /></button>
                        <button onClick={() => setDeleting(c)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500"><Trash2 size={15} /></button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {!loading && items.length > 0 && (
          <div className="px-4 border-t border-gray-100">
            <Pagination page={page} pages={pages} total={total} limit={20} onChange={setPage} />
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} size="lg"
        title={editing ? 'Editar Coleta' : 'Nova Coleta'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Select label="Gerador *" required value={form.generator_id}
              onChange={e => setForm(p => ({ ...p, generator_id: e.target.value }))}
              options={[{ value: '', label: 'Selecionar…' }, ...generators.map(g => ({ value: g.id, label: g.name }))]} />
            <Select label="Recebedor *" required value={form.receiver_id}
              onChange={e => setForm(p => ({ ...p, receiver_id: e.target.value }))}
              options={[{ value: '', label: 'Selecionar…' }, ...receivers.map(r => ({ value: r.id, label: r.name }))]} />
            <Select label="Veículo" value={form.truck_id}
              onChange={e => setForm(p => ({ ...p, truck_id: e.target.value }))}
              options={[{ value: '', label: 'Nenhum' }, ...trucks.map(t => ({ value: t.id, label: `${t.plate} — ${t.model}` }))]} />
            <Select label="Rota" value={form.route_id}
              onChange={e => setForm(p => ({ ...p, route_id: e.target.value }))}
              options={[{ value: '', label: 'Nenhuma' }, ...routes.map(r => ({ value: r.id, label: r.name }))]} />
            <Input label="Data planejada" type="date" required value={form.planned_date}
              onChange={e => setForm(p => ({ ...p, planned_date: e.target.value }))} />
            <Select label="Tipo" value={form.collect_type}
              onChange={e => setForm(p => ({ ...p, collect_type: e.target.value }))}
              options={[{ value: 'normal', label: 'Normal' }, { value: 'emergency', label: 'Emergencial' }, { value: 'scheduled', label: 'Agendada' }]} />
          </div>
          {editing && (
            <div className="border-t border-gray-100 pt-4">
              <p className="text-xs font-medium text-gray-500 mb-3 uppercase tracking-wider">Dados da coleta realizada</p>
              <div className="grid grid-cols-3 gap-4">
                <Input label="Quantidade coletada" type="number" value={form.collected_quantity}
                  onChange={e => setForm(p => ({ ...p, collected_quantity: e.target.value }))} />
                <Select label="Unidade" value={form.collected_unit}
                  onChange={e => setForm(p => ({ ...p, collected_unit: e.target.value }))}
                  options={UNIT_OPTIONS} />
                <Input label="Peso (kg)" type="number" value={form.collected_weight}
                  onChange={e => setForm(p => ({ ...p, collected_weight: e.target.value }))} />
              </div>
            </div>
          )}
          <Input label="Observações" value={form.notes} onChange={e => setForm(p => ({ ...p, notes: e.target.value }))} placeholder="Notas adicionais…" />
        </div>
      </Modal>

      {/* Assign Route Modal */}
      <Modal open={assignRouteOpen} onClose={() => setAssignRouteOpen(false)} title="Atribuir Rota"
        footer={<><Button variant="secondary" onClick={() => setAssignRouteOpen(false)}>Cancelar</Button><Button onClick={bulkAssignRoute} disabled={!assignRouteId}>Atribuir</Button></>}>
        <Select label="Rota" value={assignRouteId} onChange={e => setAssignRouteId(e.target.value)}
          options={[{ value: '', label: 'Selecionar rota…' }, ...routes.map(r => ({ value: r.id, label: r.name }))]} />
      </Modal>

      <ConfirmDialog open={!!deleting} onClose={() => setDeleting(null)} onConfirm={handleDelete}
        title="Excluir Coleta" message="Tem certeza que deseja excluir esta coleta?" confirmLabel="Excluir" />
    </Layout>
  )
}
