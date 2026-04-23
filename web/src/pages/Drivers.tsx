import { useEffect, useState, useCallback } from 'react'
import { Plus, Search, Pencil, Trash2, ToggleLeft, ToggleRight, AlertTriangle, FileUp, Trash } from 'lucide-react'
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
import { driversApi } from '../api/drivers'
import { CsvImportModal } from '../components/ui/CsvImportModal'
import { CsvDeleteModal } from '../components/ui/CsvDeleteModal'
import { DropdownButton } from '../components/ui/DropdownButton'
import type { Driver } from '../types'
import { fmtDate } from '../utils/formatters'

interface FormState {
  name: string; email: string; phone: string; cpf: string
  cnh_number: string; cnh_category: string; cnh_expiry_date: string
}

const EMPTY: FormState = { name: '', email: '', phone: '', cpf: '', cnh_number: '', cnh_category: '', cnh_expiry_date: '' }

function isExpiringOrExpired(date: string | null): boolean {
  if (!date) return false
  const expiry = new Date(date)
  const soon = new Date()
  soon.setDate(soon.getDate() + 30)
  return expiry <= soon
}

export function Drivers() {
  const toast = useToast()
  const [items, setItems] = useState<Driver[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<Driver | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<Driver | null>(null)
  const [saving, setSaving] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [deleteImportOpen, setDeleteImportOpen] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await driversApi.list({ search, page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [search, page])

  useEffect(() => { load() }, [load])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (d: Driver) => {
    setEditing(d)
    setForm({
      name: d.name, email: d.email, phone: d.phone, cpf: d.cpf,
      cnh_number: d.cnh_number, cnh_category: d.cnh_category,
      cnh_expiry_date: d.cnh_expiry_date ? d.cnh_expiry_date.split('T')[0] : '',
    })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = { ...form, cnh_expiry_date: form.cnh_expiry_date || undefined }
      if (editing) { await driversApi.update(editing.id, payload); toast('Motorista atualizado') }
      else { await driversApi.create(payload); toast('Motorista criado') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await driversApi.delete(deleting.id); toast('Motorista excluído'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const toggleActive = async (d: Driver) => {
    try { await driversApi.update(d.id, { active: !d.active }); toast(d.active ? 'Desativado' : 'Ativado'); load() }
    catch { toast('Erro ao atualizar', 'error') }
  }

  return (
    <Layout title="Motoristas" subtitle="Cadastro de motoristas" actions={
      <div className="flex gap-2">
        <DropdownButton
            label="Importar CSV"
            icon={<FileUp size={16} />}
            options={[
              { label: 'Criar / Atualizar', icon: <FileUp size={14} />, onClick: () => setImportOpen(true) },
              { label: 'Excluir por CSV', icon: <Trash size={14} />, onClick: () => setDeleteImportOpen(true), variant: 'danger' },
            ]}
          />
        <Button icon={<Plus size={16} />} onClick={openCreate}>Novo Motorista</Button>
      </div>
    }>
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        <div className="p-4">
          <Input placeholder="Buscar por nome ou CPF…" value={search}
            onChange={e => { setSearch(e.target.value); setPage(1) }}
            leftIcon={<Search size={15} />} className="max-w-sm" />
        </div>
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum motorista" action={<Button onClick={openCreate}>Criar Motorista</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Nome', 'CPF', 'CNH', 'Validade CNH', 'Status', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(d => {
                  const warn = isExpiringOrExpired(d.cnh_expiry_date)
                  return (
                    <tr key={d.id} className="hover:bg-gray-50/50">
                      <td className="px-4 py-3 font-medium text-gray-900">{d.name}</td>
                      <td className="px-4 py-3 text-gray-500 font-mono text-xs">{d.cpf || '—'}</td>
                      <td className="px-4 py-3 text-gray-500">{d.cnh_number || '—'} {d.cnh_category && <span className="ml-1 text-xs bg-gray-100 px-1 rounded">{d.cnh_category}</span>}</td>
                      <td className="px-4 py-3">
                        {d.cnh_expiry_date ? (
                          <span className={`flex items-center gap-1 text-xs ${warn ? 'text-red-500 font-medium' : 'text-gray-500'}`}>
                            {warn && <AlertTriangle size={12} />}
                            {fmtDate(d.cnh_expiry_date)}
                          </span>
                        ) : '—'}
                      </td>
                      <td className="px-4 py-3">
                        <Badge className={d.active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}>
                          {d.active ? 'Ativo' : 'Inativo'}
                        </Badge>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-1 justify-end">
                          <button onClick={() => toggleActive(d)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400">
                            {d.active ? <ToggleRight size={16} className="text-brand-500" /> : <ToggleLeft size={16} />}
                          </button>
                          <button onClick={() => openEdit(d)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600"><Pencil size={15} /></button>
                          <button onClick={() => setDeleting(d)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500"><Trash2 size={15} /></button>
                        </div>
                      </td>
                    </tr>
                  )
                })}
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
        title={editing ? 'Editar Motorista' : 'Novo Motorista'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="grid grid-cols-2 gap-4">
          <div className="col-span-2"><Input label="Nome" required value={form.name} onChange={e => setForm(p => ({ ...p, name: e.target.value }))} /></div>
          <Input label="E-mail" type="email" value={form.email} onChange={e => setForm(p => ({ ...p, email: e.target.value }))} />
          <Input label="Telefone" value={form.phone} onChange={e => setForm(p => ({ ...p, phone: e.target.value }))} />
          <Input label="CPF" value={form.cpf} onChange={e => setForm(p => ({ ...p, cpf: e.target.value }))} placeholder="000.000.000-00" />
          <Input label="Nº CNH" value={form.cnh_number} onChange={e => setForm(p => ({ ...p, cnh_number: e.target.value }))} />
          <Input label="Categoria CNH" value={form.cnh_category} onChange={e => setForm(p => ({ ...p, cnh_category: e.target.value }))} placeholder="AB, C, D..." />
          <Input label="Validade CNH" type="date" value={form.cnh_expiry_date} onChange={e => setForm(p => ({ ...p, cnh_expiry_date: e.target.value }))} />
        </div>
      </Modal>

      <ConfirmDialog open={!!deleting} onClose={() => setDeleting(null)} onConfirm={handleDelete}
        title="Excluir Motorista" message={`Excluir "${deleting?.name}"?`} confirmLabel="Excluir" />

      <CsvImportModal
        open={importOpen}
        onClose={() => setImportOpen(false)}
        title="Importar Motoristas"
        templateHeaders={['name', 'external_id', 'email', 'phone', 'cpf', 'cnh_number', 'cnh_category', 'cnh_expiry_date']}
        templateExample={['João Silva', 'EXT001', 'joao@empresa.com', '11999990000', '000.000.000-00', '00000000000', 'AB', '2027-12-31']}
        onImport={driversApi.import}
      />

      <CsvDeleteModal
        open={deleteImportOpen}
        onClose={() => setDeleteImportOpen(false)}
        title="Excluir Motoristas por CSV"
        onDelete={driversApi.importDelete}
      />
    </Layout>
  )
}
