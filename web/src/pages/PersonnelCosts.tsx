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
import { driversApi } from '../api/drivers'
import type { PersonnelCost, Driver } from '../types'
import { fmtCurrency, PERSONNEL_ROLE_LABEL } from '../utils/formatters'

interface FormState {
  driver_id: string; role: string; period_month: string
  base_salary: string; benefits: string; notes: string
}

const EMPTY: FormState = { driver_id: '', role: 'driver', period_month: '', base_salary: '', benefits: '0', notes: '' }

const thisMonth = () => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

export function PersonnelCosts() {
  const toast = useToast()
  const [items, setItems] = useState<PersonnelCost[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>({ ...EMPTY, period_month: thisMonth() })
  const [editing, setEditing] = useState<PersonnelCost | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<PersonnelCost | null>(null)
  const [saving, setSaving] = useState(false)
  const [drivers, setDrivers] = useState<Driver[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await financialApi.listPersonnelCosts({ page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [page])

  useEffect(() => { load() }, [load])
  useEffect(() => { driversApi.list({ limit: 100 }).then(r => setDrivers(r.data)) }, [])

  const openCreate = () => { setEditing(null); setForm({ ...EMPTY, period_month: thisMonth() }); setModalOpen(true) }
  const openEdit = (c: PersonnelCost) => {
    setEditing(c)
    setForm({
      driver_id: c.driver_id, role: c.role,
      period_month: c.period_month ? c.period_month.slice(0, 7) : '',
      base_salary: String(c.base_salary), benefits: String(c.benefits), notes: c.notes,
    })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = {
        driver_id: form.driver_id,
        role: form.role as PersonnelCost['role'],
        period_month: form.period_month + '-01',
        base_salary: Number(form.base_salary),
        benefits: Number(form.benefits),
        notes: form.notes,
      }
      if (editing) { await financialApi.updatePersonnelCost(editing.id, payload); toast('Custo atualizado') }
      else { await financialApi.createPersonnelCost(payload); toast('Custo registrado') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await financialApi.deletePersonnelCost(deleting.id); toast('Custo excluído'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  return (
    <Layout title="Custos de Pessoal" subtitle="Salários e benefícios"
      actions={<Button icon={<Plus size={16} />} onClick={openCreate}>Novo Custo</Button>}>
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum custo" action={<Button onClick={openCreate}>Registrar Custo</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Motorista', 'Função', 'Mês', 'Salário Base', 'Benefícios', 'Total', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(c => (
                  <tr key={c.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900">
                      {c.driver?.name ?? drivers.find(d => d.id === c.driver_id)?.name ?? '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500">{PERSONNEL_ROLE_LABEL[c.role] ?? c.role}</td>
                    <td className="px-4 py-3 text-gray-500 text-xs font-mono">{c.period_month ? c.period_month.slice(0, 7) : '—'}</td>
                    <td className="px-4 py-3 text-gray-600">{fmtCurrency(c.base_salary)}</td>
                    <td className="px-4 py-3 text-gray-600">{fmtCurrency(c.benefits)}</td>
                    <td className="px-4 py-3 font-semibold text-gray-900">{fmtCurrency(c.total_cost)}</td>
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
        title={editing ? 'Editar Custo' : 'Novo Custo de Pessoal'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="grid grid-cols-2 gap-4">
          <div className="col-span-2">
            <Select label="Motorista/Coletor *" required value={form.driver_id}
              onChange={e => setForm(p => ({ ...p, driver_id: e.target.value }))}
              options={[{ value: '', label: 'Selecionar…' }, ...drivers.map(d => ({ value: d.id, label: d.name }))]} />
          </div>
          <Select label="Função" value={form.role} onChange={e => setForm(p => ({ ...p, role: e.target.value }))}
            options={[{ value: 'driver', label: 'Motorista' }, { value: 'collector', label: 'Coletor' }]} />
          <Input label="Mês de referência" type="month" required value={form.period_month}
            onChange={e => setForm(p => ({ ...p, period_month: e.target.value }))} />
          <Input label="Salário base (R$)" type="number" required value={form.base_salary}
            onChange={e => setForm(p => ({ ...p, base_salary: e.target.value }))} />
          <Input label="Benefícios (R$)" type="number" value={form.benefits}
            onChange={e => setForm(p => ({ ...p, benefits: e.target.value }))} />
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
