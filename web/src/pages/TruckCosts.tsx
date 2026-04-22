import { useEffect, useState, useCallback } from 'react'
import { Plus, Pencil, Trash2 } from 'lucide-react'
import { Layout } from '../components/layout/Layout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Select } from '../components/ui/Select'
import { Modal } from '../components/ui/Modal'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { Pagination } from '../components/ui/Pagination'
import { EmptyState } from '../components/ui/EmptyState'
import { PageLoader } from '../components/ui/Spinner'
import { useToast } from '../components/ui/Toast'
import { financialApi } from '../api/financial'
import { trucksApi } from '../api/trucks'
import type { TruckCost, Truck } from '../types'
import { fmtDate, fmtCurrency, TRUCK_COST_TYPE_LABEL } from '../utils/formatters'

interface FormState {
  truck_id: string; type: string; period_start: string; period_end: string
  total_amount: string; total_km: string; notes: string
}

const EMPTY: FormState = {
  truck_id: '', type: 'fuel', period_start: '', period_end: '',
  total_amount: '', total_km: '', notes: '',
}

export function TruckCosts() {
  const toast = useToast()
  const [items, setItems] = useState<TruckCost[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<TruckCost | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<TruckCost | null>(null)
  const [saving, setSaving] = useState(false)
  const [trucks, setTrucks] = useState<Truck[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await financialApi.listTruckCosts({ page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [page])

  useEffect(() => { load() }, [load])
  useEffect(() => { trucksApi.list({ limit: 100 }).then(r => setTrucks(r.data)) }, [])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (c: TruckCost) => {
    setEditing(c)
    setForm({
      truck_id: c.truck_id, type: c.type,
      period_start: c.period_start?.split('T')[0] ?? '',
      period_end: c.period_end?.split('T')[0] ?? '',
      total_amount: String(c.total_amount), total_km: String(c.total_km), notes: c.notes,
    })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = {
        truck_id: form.truck_id, type: form.type as TruckCost['type'],
        period_start: form.period_start, period_end: form.period_end,
        total_amount: Number(form.total_amount), total_km: Number(form.total_km), notes: form.notes,
      }
      if (editing) { await financialApi.updateTruckCost(editing.id, payload); toast('Custo atualizado') }
      else { await financialApi.createTruckCost(payload); toast('Custo registrado') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await financialApi.deleteTruckCost(deleting.id); toast('Custo excluído'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  return (
    <Layout title="Custos de Veículos" subtitle="Combustível, manutenção e outros"
      actions={<Button icon={<Plus size={16} />} onClick={openCreate}>Novo Custo</Button>}>
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum custo" action={<Button onClick={openCreate}>Registrar Custo</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Veículo', 'Tipo', 'Período', 'Total', 'km', 'R$/km', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(c => (
                  <tr key={c.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900">
                      {c.truck ? `${c.truck.plate} — ${c.truck.model}` : trucks.find(t => t.id === c.truck_id)?.plate ?? '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500">{TRUCK_COST_TYPE_LABEL[c.type] ?? c.type}</td>
                    <td className="px-4 py-3 text-gray-500 text-xs">{fmtDate(c.period_start)} – {fmtDate(c.period_end)}</td>
                    <td className="px-4 py-3 font-semibold text-gray-900">{fmtCurrency(c.total_amount)}</td>
                    <td className="px-4 py-3 text-gray-500">{c.total_km > 0 ? `${c.total_km} km` : '—'}</td>
                    <td className="px-4 py-3 text-gray-500">{c.cost_per_km > 0 ? fmtCurrency(c.cost_per_km) : '—'}</td>
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

      <Modal open={modalOpen} onClose={() => setModalOpen(false)}
        title={editing ? 'Editar Custo' : 'Novo Custo'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="grid grid-cols-2 gap-4">
          <div className="col-span-2">
            <Select label="Veículo *" required value={form.truck_id}
              onChange={e => setForm(p => ({ ...p, truck_id: e.target.value }))}
              options={[{ value: '', label: 'Selecionar…' }, ...trucks.map(t => ({ value: t.id, label: `${t.plate} — ${t.model}` }))]} />
          </div>
          <Select label="Tipo" value={form.type} onChange={e => setForm(p => ({ ...p, type: e.target.value }))}
            options={[{ value: 'fuel', label: 'Combustível' }, { value: 'maintenance', label: 'Manutenção' }, { value: 'other', label: 'Outros' }]} />
          <Input label="Valor total (R$)" type="number" required value={form.total_amount}
            onChange={e => setForm(p => ({ ...p, total_amount: e.target.value }))} />
          <Input label="Início" type="date" value={form.period_start} onChange={e => setForm(p => ({ ...p, period_start: e.target.value }))} />
          <Input label="Fim" type="date" value={form.period_end} onChange={e => setForm(p => ({ ...p, period_end: e.target.value }))} />
          <div className="col-span-2">
            <Input label="KM total (opcional)" type="number" value={form.total_km}
              onChange={e => setForm(p => ({ ...p, total_km: e.target.value }))} />
          </div>
          <div className="col-span-2">
            <Input label="Observações" value={form.notes} onChange={e => setForm(p => ({ ...p, notes: e.target.value }))} />
          </div>
        </div>
      </Modal>

      <ConfirmDialog open={!!deleting} onClose={() => setDeleting(null)} onConfirm={handleDelete}
        title="Excluir Custo" message="Tem certeza que deseja excluir este registro de custo?" confirmLabel="Excluir" />
    </Layout>
  )
}
