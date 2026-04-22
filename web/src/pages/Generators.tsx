import { useEffect, useState, useCallback } from 'react'
import { Plus, Search, Pencil, Trash2, ToggleLeft, ToggleRight } from 'lucide-react'
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
import { generatorsApi } from '../api/generators'
import type { Generator } from '../types'
import { fmtDate } from '../utils/formatters'

interface FormState {
  name: string; cnpj: string; address: string; external_id: string
}

const EMPTY: FormState = { name: '', cnpj: '', address: '', external_id: '' }

export function Generators() {
  const toast = useToast()
  const [items, setItems] = useState<Generator[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<Generator | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<Generator | null>(null)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await generatorsApi.list({ search, page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [search, page])

  useEffect(() => { load() }, [load])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (g: Generator) => {
    setEditing(g)
    setForm({ name: g.name, cnpj: g.cnpj, address: g.address, external_id: g.external_id })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      if (editing) {
        await generatorsApi.update(editing.id, form)
        toast('Gerador atualizado com sucesso')
      } else {
        await generatorsApi.create(form)
        toast('Gerador criado com sucesso')
      }
      setModalOpen(false)
      load()
    } catch (e: unknown) {
      toast(e instanceof Error ? e.message : 'Erro ao salvar', 'error')
    } finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try {
      await generatorsApi.delete(deleting.id)
      toast('Gerador excluído')
      setDeleting(null)
      load()
    } catch (e: unknown) {
      toast(e instanceof Error ? e.message : 'Erro ao excluir', 'error')
    }
  }

  const toggleActive = async (g: Generator) => {
    try {
      await generatorsApi.update(g.id, { active: !g.active })
      toast(g.active ? 'Gerador desativado' : 'Gerador ativado')
      load()
    } catch { toast('Erro ao atualizar', 'error') }
  }

  return (
    <Layout
      title="Geradores"
      subtitle="Empresas geradoras de resíduos"
      actions={<Button icon={<Plus size={16} />} onClick={openCreate}>Novo Gerador</Button>}
    >
      {/* Filters */}
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        <div className="p-4">
          <Input
            placeholder="Buscar por nome ou CNPJ…"
            value={search}
            onChange={e => { setSearch(e.target.value); setPage(1) }}
            leftIcon={<Search size={15} />}
            className="max-w-sm"
          />
        </div>

        {/* Table */}
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum gerador" description="Crie o primeiro gerador para começar." action={<Button onClick={openCreate}>Criar Gerador</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-t border-gray-100">
                  {['Nome', 'CNPJ', 'Endereço', 'Status', 'Criado em', ''].map(h => (
                    <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(g => (
                  <tr key={g.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900">{g.name}</td>
                    <td className="px-4 py-3 text-gray-500 font-mono text-xs">{g.cnpj || '—'}</td>
                    <td className="px-4 py-3 text-gray-500 max-w-xs truncate">{g.address || '—'}</td>
                    <td className="px-4 py-3">
                      <Badge className={g.active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}>
                        {g.active ? 'Ativo' : 'Inativo'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-gray-400 text-xs">{fmtDate(g.created_at)}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button onClick={() => toggleActive(g)} title={g.active ? 'Desativar' : 'Ativar'}
                          className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600">
                          {g.active ? <ToggleRight size={16} className="text-brand-500" /> : <ToggleLeft size={16} />}
                        </button>
                        <button onClick={() => openEdit(g)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600">
                          <Pencil size={15} />
                        </button>
                        <button onClick={() => setDeleting(g)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500">
                          <Trash2 size={15} />
                        </button>
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

      {/* Modal */}
      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title={editing ? 'Editar Gerador' : 'Novo Gerador'}
        footer={
          <>
            <Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button>
            <Button onClick={handleSave} loading={saving}>Salvar</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input label="Nome" required value={form.name} onChange={e => setForm(p => ({ ...p, name: e.target.value }))} placeholder="Empresa S/A" />
          <Input label="CNPJ" value={form.cnpj} onChange={e => setForm(p => ({ ...p, cnpj: e.target.value }))} placeholder="00.000.000/0001-00" />
          <Input label="Endereço" value={form.address} onChange={e => setForm(p => ({ ...p, address: e.target.value }))} placeholder="Rua, número, cidade" />
          <Input label="ID Externo" value={form.external_id} onChange={e => setForm(p => ({ ...p, external_id: e.target.value }))} placeholder="Código no sistema legado" />
        </div>
      </Modal>

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={handleDelete}
        title="Excluir Gerador"
        message={`Tem certeza que deseja excluir "${deleting?.name}"? Esta ação não pode ser desfeita.`}
        confirmLabel="Excluir"
      />
    </Layout>
  )
}
