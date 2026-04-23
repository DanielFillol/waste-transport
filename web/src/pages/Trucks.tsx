import { useEffect, useState, useCallback } from 'react'
import { Plus, Search, Pencil, Trash2, ToggleLeft, ToggleRight, FileUp, Trash } from 'lucide-react'
import { Layout } from '../components/layout/Layout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Badge } from '../components/ui/Badge'
import { Modal } from '../components/ui/Modal'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { Pagination } from '../components/ui/Pagination'
import { EmptyState } from '../components/ui/EmptyState'
import { PageLoader } from '../components/ui/Spinner'
import { useToast } from '../components/ui/Toast'
import { trucksApi } from '../api/trucks'
import { CsvImportModal } from '../components/ui/CsvImportModal'
import { CsvDeleteModal } from '../components/ui/CsvDeleteModal'
import { DropdownButton } from '../components/ui/DropdownButton'
import type { Truck } from '../types'

interface FormState { plate: string; model: string; year: string; capacity_kg: string; capacity_m3: string }
const EMPTY: FormState = { plate: '', model: '', year: '', capacity_kg: '', capacity_m3: '' }

export function Trucks() {
  const toast = useToast()
  const [items, setItems] = useState<Truck[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<Truck | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<Truck | null>(null)
  const [saving, setSaving] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [deleteImportOpen, setDeleteImportOpen] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await trucksApi.list({ search, page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [search, page])

  useEffect(() => { load() }, [load])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (t: Truck) => {
    setEditing(t)
    setForm({ plate: t.plate, model: t.model, year: String(t.year), capacity_kg: String(t.capacity_kg), capacity_m3: String(t.capacity_m3) })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = { plate: form.plate, model: form.model, year: Number(form.year), capacity_kg: Number(form.capacity_kg), capacity_m3: Number(form.capacity_m3) }
      if (editing) { await trucksApi.update(editing.id, payload); toast('Veículo atualizado') }
      else { await trucksApi.create(payload); toast('Veículo criado') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await trucksApi.delete(deleting.id); toast('Veículo excluído'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const toggleActive = async (t: Truck) => {
    try { await trucksApi.update(t.id, { active: !t.active }); toast(t.active ? 'Desativado' : 'Ativado'); load() }
    catch { toast('Erro', 'error') }
  }

  return (
    <Layout title="Veículos" subtitle="Frota de coleta" actions={
      <div className="flex gap-2">
        <DropdownButton
            label="Importar CSV"
            icon={<FileUp size={16} />}
            options={[
              { label: 'Criar / Atualizar', icon: <FileUp size={14} />, onClick: () => setImportOpen(true) },
              { label: 'Excluir por CSV', icon: <Trash size={14} />, onClick: () => setDeleteImportOpen(true), variant: 'danger' },
            ]}
          />
        <Button icon={<Plus size={16} />} onClick={openCreate}>Novo Veículo</Button>
      </div>
    }>
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        <div className="p-4">
          <Input placeholder="Buscar por placa ou modelo…" value={search} onChange={e => { setSearch(e.target.value); setPage(1) }}
            leftIcon={<Search size={15} />} className="max-w-sm" />
        </div>
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum veículo" action={<Button onClick={openCreate}>Criar Veículo</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Placa', 'Modelo', 'Ano', 'Cap. (kg)', 'Cap. (m³)', 'Status', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(t => (
                  <tr key={t.id} className="hover:bg-gray-50/50">
                    <td className="px-4 py-3 font-mono font-semibold text-gray-900">{t.plate}</td>
                    <td className="px-4 py-3 text-gray-700">{t.model}</td>
                    <td className="px-4 py-3 text-gray-500">{t.year}</td>
                    <td className="px-4 py-3 text-gray-500">{t.capacity_kg > 0 ? `${t.capacity_kg} kg` : '—'}</td>
                    <td className="px-4 py-3 text-gray-500">{t.capacity_m3 > 0 ? `${t.capacity_m3} m³` : '—'}</td>
                    <td className="px-4 py-3">
                      <Badge className={t.active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}>
                        {t.active ? 'Ativo' : 'Inativo'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button onClick={() => toggleActive(t)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400">
                          {t.active ? <ToggleRight size={16} className="text-brand-500" /> : <ToggleLeft size={16} />}
                        </button>
                        <button onClick={() => openEdit(t)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400"><Pencil size={15} /></button>
                        <button onClick={() => setDeleting(t)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500"><Trash2 size={15} /></button>
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

      <Modal open={modalOpen} onClose={() => setModalOpen(false)} title={editing ? 'Editar Veículo' : 'Novo Veículo'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="grid grid-cols-2 gap-4">
          <Input label="Placa" required value={form.plate} onChange={e => setForm(p => ({ ...p, plate: e.target.value }))} placeholder="ABC-1234" />
          <Input label="Modelo" required value={form.model} onChange={e => setForm(p => ({ ...p, model: e.target.value }))} placeholder="Volkswagen Delivery" />
          <Input label="Ano" type="number" required value={form.year} onChange={e => setForm(p => ({ ...p, year: e.target.value }))} />
          <Input label="Capacidade (kg)" type="number" value={form.capacity_kg} onChange={e => setForm(p => ({ ...p, capacity_kg: e.target.value }))} />
          <Input label="Capacidade (m³)" type="number" value={form.capacity_m3} onChange={e => setForm(p => ({ ...p, capacity_m3: e.target.value }))} />
        </div>
      </Modal>

      <ConfirmDialog open={!!deleting} onClose={() => setDeleting(null)} onConfirm={handleDelete}
        title="Excluir Veículo" message={`Excluir "${deleting?.plate} – ${deleting?.model}"?`} confirmLabel="Excluir" />

      <CsvImportModal
        open={importOpen}
        onClose={() => setImportOpen(false)}
        title="Importar Veículos"
        templateHeaders={['plate', 'model', 'year', 'capacity_kg', 'capacity_m3']}
        templateExample={['ABC-1234', 'Volkswagen Delivery', '2022', '5000', '20']}
        onImport={trucksApi.import}
      />

      <CsvDeleteModal
        open={deleteImportOpen}
        onClose={() => setDeleteImportOpen(false)}
        title="Excluir Veículos por CSV"
        onDelete={trucksApi.importDelete}
      />
    </Layout>
  )
}
