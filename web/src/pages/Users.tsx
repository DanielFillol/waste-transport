import { useEffect, useState, useCallback } from 'react'
import { Plus, Search, Pencil, Trash2 } from 'lucide-react'
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
import { usersApi } from '../api/users'
import type { User } from '../types'
import { fmtDate } from '../utils/formatters'

interface FormState {
  name: string
  username: string
  password: string
  role: 'admin' | 'user'
}

const EMPTY: FormState = { name: '', username: '', password: '', role: 'user' }

export function Users() {
  const toast = useToast()
  const [items, setItems] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<User | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<User | null>(null)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await usersApi.list(page, 20)
      const filtered = search
        ? res.data.filter(u => u.name.toLowerCase().includes(search.toLowerCase()) || u.username.toLowerCase().includes(search.toLowerCase()))
        : res.data
      setItems(filtered); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [page, search])

  useEffect(() => { load() }, [load])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (u: User) => {
    setEditing(u)
    setForm({ name: u.name, username: u.username, password: '', role: u.role })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      if (editing) {
        const payload: { name?: string; role?: 'admin' | 'user'; password?: string } = {
          name: form.name,
          role: form.role,
        }
        if (form.password) payload.password = form.password
        await usersApi.update(editing.id, payload)
        toast('Usuário atualizado')
      } else {
        await usersApi.create(form)
        toast('Usuário criado')
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
      await usersApi.delete(deleting.id)
      toast('Usuário excluído')
      setDeleting(null)
      load()
    } catch (e: unknown) {
      toast(e instanceof Error ? e.message : 'Erro ao excluir', 'error')
    }
  }

  return (
    <Layout
      title="Usuários"
      subtitle="Gerencie os usuários e permissões do sistema"
      actions={<Button icon={<Plus size={16} />} onClick={openCreate}>Novo Usuário</Button>}
    >
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        <div className="p-4">
          <Input
            placeholder="Buscar por nome ou usuário…"
            value={search}
            onChange={e => { setSearch(e.target.value); setPage(1) }}
            leftIcon={<Search size={15} />}
            className="max-w-sm"
          />
        </div>

        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhum usuário" description="Crie o primeiro usuário para começar." action={<Button onClick={openCreate}>Criar Usuário</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-t border-gray-100">
                  {['Nome', 'Usuário', 'Papel', 'Criado em', ''].map(h => (
                    <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(u => (
                  <tr key={u.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900">{u.name}</td>
                    <td className="px-4 py-3 text-gray-500 font-mono text-xs">{u.username}</td>
                    <td className="px-4 py-3">
                      <Badge className={u.role === 'admin' ? 'bg-purple-100 text-purple-700' : 'bg-gray-100 text-gray-500'}>
                        {u.role === 'admin' ? 'Admin' : 'Usuário'}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-gray-400 text-xs">{fmtDate(u.created_at ?? '')}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button onClick={() => openEdit(u)} className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600">
                          <Pencil size={15} />
                        </button>
                        <button onClick={() => setDeleting(u)} className="p-1.5 rounded-lg hover:bg-red-50 text-gray-400 hover:text-red-500">
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

      <Modal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title={editing ? 'Editar Usuário' : 'Novo Usuário'}
        footer={
          <>
            <Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button>
            <Button onClick={handleSave} loading={saving}>Salvar</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Input label="Nome" required value={form.name} onChange={e => setForm(p => ({ ...p, name: e.target.value }))} placeholder="Nome completo" />
          <Input label="Usuário" required value={form.username} onChange={e => setForm(p => ({ ...p, username: e.target.value }))} placeholder="nome.usuario" disabled={!!editing} />
          <Input
            label={editing ? 'Nova senha (deixe em branco para manter)' : 'Senha'}
            type="password"
            value={form.password}
            onChange={e => setForm(p => ({ ...p, password: e.target.value }))}
            placeholder="Mínimo 6 caracteres"
            required={!editing}
          />
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1.5">Papel</label>
            <select
              value={form.role}
              onChange={e => setForm(p => ({ ...p, role: e.target.value as 'admin' | 'user' }))}
              className="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-brand-500/20 focus:border-brand-400"
            >
              <option value="user">Usuário</option>
              <option value="admin">Admin</option>
            </select>
          </div>
        </div>
      </Modal>

      <ConfirmDialog
        open={!!deleting}
        onClose={() => setDeleting(null)}
        onConfirm={handleDelete}
        title="Excluir Usuário"
        message={`Tem certeza que deseja excluir "${deleting?.name}"? Esta ação não pode ser desfeita.`}
        confirmLabel="Excluir"
      />
    </Layout>
  )
}
