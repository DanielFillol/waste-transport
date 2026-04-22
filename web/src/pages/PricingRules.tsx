import { useEffect, useState } from 'react'
import { Plus, Pencil, Trash2, ToggleLeft, ToggleRight } from 'lucide-react'
import { Layout } from '../components/layout/Layout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Select } from '../components/ui/Select'
import { Modal } from '../components/ui/Modal'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { EmptyState } from '../components/ui/EmptyState'
import { PageLoader } from '../components/ui/Spinner'
import { useToast } from '../components/ui/Toast'
import { financialApi } from '../api/financial'
import type { PricingRule } from '../types'
import { fmtCurrency } from '../utils/formatters'

interface FormState {
  collect_type: string; price_per_unit: string; unit: string; active: boolean
}

const EMPTY: FormState = { collect_type: '', price_per_unit: '', unit: 'KG', active: true }

const UNIT_OPTIONS = [
  { value: 'KG', label: 'Kg' },
  { value: 'L', label: 'Litros' },
  { value: 'M3', label: 'm³' },
  { value: 'UN', label: 'Unidades' },
]

export function PricingRules() {
  const toast = useToast()
  const [items, setItems] = useState<PricingRule[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<PricingRule | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<PricingRule | null>(null)
  const [saving, setSaving] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const res = await financialApi.listPricingRules()
      setItems(res.data)
    } finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (r: PricingRule) => {
    setEditing(r)
    setForm({
      collect_type: r.collect_type ?? '', price_per_unit: String(r.price_per_unit),
      unit: r.unit, active: r.active,
    })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = {
        collect_type: form.collect_type as PricingRule['collect_type'] || null,
        price_per_unit: Number(form.price_per_unit),
        unit: form.unit as PricingRule['unit'],
        active: form.active,
      }
      if (editing) { await financialApi.updatePricingRule(editing.id, payload); toast('Regra atualizada') }
      else { await financialApi.createPricingRule(payload); toast('Regra criada') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await financialApi.deletePricingRule(deleting.id); toast('Regra excluída'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const toggleActive = async (r: PricingRule) => {
    try { await financialApi.updatePricingRule(r.id, { active: !r.active }); toast(r.active ? 'Desativada' : 'Ativada'); load() }
    catch { toast('Erro', 'error') }
  }

  const typeLabel = (t: string | null) => {
    if (!t) return 'Todas'
    const map: Record<string, string> = { normal: 'Normal', emergency: 'Emergencial', scheduled: 'Agendada' }
    return map[t] ?? t
  }

  return (
    <Layout title="Regras de Precificação" subtitle="Preços por tipo de coleta e material"
      actions={<Button icon={<Plus size={16} />} onClick={openCreate}>Nova Regra</Button>}>
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhuma regra" description="Crie regras de precificação para gerar faturas automaticamente."
            action={<Button onClick={openCreate}>Criar Regra</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Tipo de Coleta', 'Material', 'Embalagem', 'Preço / Unidade', 'Unidade', 'Status', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(r => (
                  <tr key={r.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 text-gray-700">{typeLabel(r.collect_type)}</td>
                    <td className="px-4 py-3 text-gray-500">{r.material?.name ?? '—'}</td>
                    <td className="px-4 py-3 text-gray-500">{r.packaging?.name ?? '—'}</td>
                    <td className="px-4 py-3 font-semibold text-gray-900">{fmtCurrency(r.price_per_unit)}</td>
                    <td className="px-4 py-3 text-gray-500">{r.unit}</td>
                    <td className="px-4 py-3">
                      <span className={`text-xs font-medium ${r.active ? 'text-green-600' : 'text-gray-400'}`}>
                        {r.active ? 'Ativa' : 'Inativa'}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button onClick={() => toggleActive(r)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400">
                          {r.active ? <ToggleRight size={16} className="text-brand-500" /> : <ToggleLeft size={16} />}
                        </button>
                        <button onClick={() => openEdit(r)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600"><Pencil size={15} /></button>
                        <button onClick={() => setDeleting(r)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500"><Trash2 size={15} /></button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <Modal open={modalOpen} onClose={() => setModalOpen(false)}
        title={editing ? 'Editar Regra' : 'Nova Regra de Precificação'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="grid grid-cols-2 gap-4">
          <Select label="Tipo de coleta" value={form.collect_type}
            onChange={e => setForm(p => ({ ...p, collect_type: e.target.value }))}
            options={[
              { value: '', label: 'Todas' },
              { value: 'normal', label: 'Normal' },
              { value: 'emergency', label: 'Emergencial' },
              { value: 'scheduled', label: 'Agendada' },
            ]} />
          <Select label="Unidade de medida" value={form.unit}
            onChange={e => setForm(p => ({ ...p, unit: e.target.value }))}
            options={UNIT_OPTIONS} />
          <div className="col-span-2">
            <Input label="Preço por unidade (R$)" type="number" required value={form.price_per_unit}
              onChange={e => setForm(p => ({ ...p, price_per_unit: e.target.value }))} />
          </div>
          <div className="col-span-2 flex items-center gap-3">
            <input type="checkbox" id="rule-active" checked={form.active}
              onChange={e => setForm(p => ({ ...p, active: e.target.checked }))}
              className="rounded border-gray-300 text-brand-500 focus:ring-brand-400" />
            <label htmlFor="rule-active" className="text-sm text-gray-700">Regra ativa</label>
          </div>
        </div>
      </Modal>

      <ConfirmDialog open={!!deleting} onClose={() => setDeleting(null)} onConfirm={handleDelete}
        title="Excluir Regra" message="Tem certeza que deseja excluir esta regra?" confirmLabel="Excluir" />
    </Layout>
  )
}
