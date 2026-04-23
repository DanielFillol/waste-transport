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
import { receiversApi } from '../api/receivers'
import { CsvImportModal } from '../components/ui/CsvImportModal'
import { CsvDeleteModal } from '../components/ui/CsvDeleteModal'
import { DropdownButton } from '../components/ui/DropdownButton'
import type { Receiver } from '../types'
import { fmtDate } from '../utils/formatters'

interface FormState {
  name: string; cnpj: string; address: string; external_id: string
  license_number: string; license_expiry: string
}

const EMPTY: FormState = { name: '', cnpj: '', address: '', external_id: '', license_number: '', license_expiry: '' }

function isExpiringOrExpired(date: string | null): boolean {
  if (!date) return false
  const expiry = new Date(date)
  const soon = new Date()
  soon.setDate(soon.getDate() + 30)
  return expiry <= soon
}

export function Receivers() {
  const toast = useToast()
  const [items, setItems] = useState<Receiver[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<Receiver | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<Receiver | null>(null)
  const [saving, setSaving] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [deleteImportOpen, setDeleteImportOpen] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await receiversApi.list({ search, page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [search, page])

  useEffect(() => { load() }, [load])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (r: Receiver) => {
    setEditing(r)
    setForm({
      name: r.name, cnpj: r.cnpj, address: r.address, external_id: r.external_id,
      license_number: r.license_number,
      license_expiry: r.license_expiry ? r.license_expiry.split('T')[0] : '',
    })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = { ...form, license_expiry: form.license_expiry || undefined }
      if (editing) { await receiversApi.update(editing.id, payload); toast('Receptor atualizado') }
      else { await receiversApi.create(payload); toast('Receptor criado') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await receiversApi.delete(deleting.id); toast('Receptor excluído'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const toggleActive = async (r: Receiver) => {
    try { await receiversApi.update(r.id, { active: !r.active }); toast(r.active ? 'Desativado' : 'Ativado'); load() }
    catch { toast('Erro ao atualizar', 'error') }
  }

  return (
    <Layout title="Receptores" subtitle="Empresas receptoras de resíduos"
      actions={
        <div className="flex gap-2">
          <DropdownButton
            label="Importar CSV"
            icon={<FileUp size={16} />}
            options={[
              { label: 'Criar / Atualizar', icon: <FileUp size={14} />, onClick: () => setImportOpen(true) },
              { label: 'Excluir por CSV', icon: <Trash size={14} />, onClick: () => setDeleteImportOpen(true), variant: 'danger' },
            ]}
          />
          <Button icon={<Plus size={16} />} onClick={openCreate}>Novo Receptor</Button>
        </div>
      }>
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        <div className="p-4">
          <Input placeholder="Buscar por nome ou CNPJ…" value={search}
            onChange={e => { setSearch(e.target.value); setPage(1) }}
            leftIcon={<Search size={15} />} className="max-w-sm" />
        </div>
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum receptor" description="Crie o primeiro receptor para começar."
            action={<Button onClick={openCreate}>Criar Receptor</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Nome', 'CNPJ', 'Licença', 'Validade Lic.', 'Status', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(r => {
                  const warn = isExpiringOrExpired(r.license_expiry)
                  return (
                    <tr key={r.id} className="hover:bg-gray-50/50 transition-colors">
                      <td className="px-4 py-3 font-medium text-gray-900">{r.name}</td>
                      <td className="px-4 py-3 text-gray-500 font-mono text-xs">{r.cnpj || '—'}</td>
                      <td className="px-4 py-3 text-gray-500 text-xs">{r.license_number || '—'}</td>
                      <td className="px-4 py-3">
                        {r.license_expiry ? (
                          <span className={`flex items-center gap-1 text-xs ${warn ? 'text-red-500 font-medium' : 'text-gray-500'}`}>
                            {warn && <AlertTriangle size={12} />}
                            {fmtDate(r.license_expiry)}
                          </span>
                        ) : '—'}
                      </td>
                      <td className="px-4 py-3">
                        <Badge className={r.active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}>
                          {r.active ? 'Ativo' : 'Inativo'}
                        </Badge>
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
        title={editing ? 'Editar Receptor' : 'Novo Receptor'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="space-y-4">
          <Input label="Nome" required value={form.name} onChange={e => setForm(p => ({ ...p, name: e.target.value }))} placeholder="Empresa Receptora S/A" />
          <div className="grid grid-cols-2 gap-4">
            <Input label="CNPJ" value={form.cnpj} onChange={e => setForm(p => ({ ...p, cnpj: e.target.value }))} placeholder="00.000.000/0001-00" />
            <Input label="ID Externo" value={form.external_id} onChange={e => setForm(p => ({ ...p, external_id: e.target.value }))} placeholder="Código legado" />
          </div>
          <Input label="Endereço" value={form.address} onChange={e => setForm(p => ({ ...p, address: e.target.value }))} placeholder="Rua, número, cidade" />
          <div className="grid grid-cols-2 gap-4">
            <Input label="Nº Licença Ambiental" value={form.license_number} onChange={e => setForm(p => ({ ...p, license_number: e.target.value }))} />
            <Input label="Validade Licença" type="date" value={form.license_expiry} onChange={e => setForm(p => ({ ...p, license_expiry: e.target.value }))} />
          </div>
        </div>
      </Modal>

      <ConfirmDialog open={!!deleting} onClose={() => setDeleting(null)} onConfirm={handleDelete}
        title="Excluir Receptor" message={`Tem certeza que deseja excluir "${deleting?.name}"? Esta ação não pode ser desfeita.`}
        confirmLabel="Excluir" />

      <CsvImportModal
        open={importOpen}
        onClose={() => setImportOpen(false)}
        title="Importar Receptores"
        templateHeaders={['name', 'external_id', 'cnpj', 'address', 'license_number', 'license_expiry']}
        templateExample={['Receptor S/A', 'EXT001', '00.000.000/0001-00', 'Rua Exemplo 123', 'LIC-001', '2026-12-31']}
        onImport={receiversApi.import}
      />

      <CsvDeleteModal
        open={deleteImportOpen}
        onClose={() => setDeleteImportOpen(false)}
        title="Excluir Receptores por CSV"
        onDelete={receiversApi.importDelete}
      />
    </Layout>
  )
}
