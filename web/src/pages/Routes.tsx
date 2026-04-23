import { useEffect, useState, useCallback } from 'react'
import { Plus, Search, Pencil, Trash2, CalendarPlus, FileUp, Trash } from 'lucide-react'
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
import { routesApi } from '../api/routes'
import { CsvImportModal } from '../components/ui/CsvImportModal'
import { CsvDeleteModal } from '../components/ui/CsvDeleteModal'
import { DropdownButton } from '../components/ui/DropdownButton'
import { driversApi } from '../api/drivers'
import { generatorsApi } from '../api/generators'
import { receiversApi } from '../api/receivers'
import type { Route, Driver, Generator, Receiver } from '../types'
import { WEEK_DAYS } from '../utils/formatters'

interface FormState {
  name: string
  week_day: string
  week_number: string
  driver_ids: string[]
}

const EMPTY: FormState = { name: '', week_day: '1', week_number: '1', driver_ids: [] }

interface GenerateForm {
  target_date: string
  generator_ids: string[]
  receiver_id: string
}

const WEEK_NUMBER_OPTS = [
  { value: '1', label: '1ª semana' },
  { value: '2', label: '2ª semana' },
  { value: '3', label: '3ª semana' },
  { value: '4', label: '4ª semana' },
  { value: '0', label: 'Toda semana' },
]

export function Routes() {
  const toast = useToast()
  const [items, setItems] = useState<Route[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState>(EMPTY)
  const [editing, setEditing] = useState<Route | null>(null)
  const [modalOpen, setModalOpen] = useState(false)
  const [deleting, setDeleting] = useState<Route | null>(null)
  const [saving, setSaving] = useState(false)

  // Generate collects modal
  const [genRoute, setGenRoute] = useState<Route | null>(null)
  const [genForm, setGenForm] = useState<GenerateForm>({ target_date: '', generator_ids: [], receiver_id: '' })
  const [generating, setGenerating] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [deleteImportOpen, setDeleteImportOpen] = useState(false)
  const [generators, setGenerators] = useState<Generator[]>([])
  const [receivers, setReceivers] = useState<Receiver[]>([])

  const [drivers, setDrivers] = useState<Driver[]>([])

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await routesApi.list({ search, page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [search, page])

  useEffect(() => { load() }, [load])

  useEffect(() => {
    driversApi.list({ limit: 100 }).then(r => setDrivers(r.data))
    generatorsApi.list({ limit: 100 }).then(r => setGenerators(r.data))
    receiversApi.list({ limit: 100 }).then(r => setReceivers(r.data))
  }, [])

  const openCreate = () => { setEditing(null); setForm(EMPTY); setModalOpen(true) }
  const openEdit = (r: Route) => {
    setEditing(r)
    setForm({
      name: r.name,
      week_day: String(r.week_day),
      week_number: String(r.week_number),
      driver_ids: r.drivers?.map(d => d.id) ?? [],
    })
    setModalOpen(true)
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      const payload = {
        name: form.name,
        week_day: Number(form.week_day),
        week_number: Number(form.week_number),
        driver_ids: form.driver_ids,
      }
      if (editing) { await routesApi.update(editing.id, payload); toast('Rota atualizada') }
      else { await routesApi.create(payload); toast('Rota criada') }
      setModalOpen(false); load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setSaving(false) }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try { await routesApi.delete(deleting.id); toast('Rota excluída'); setDeleting(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const openGenerate = (r: Route) => {
    setGenRoute(r)
    setGenForm({ target_date: new Date().toISOString().split('T')[0], generator_ids: [], receiver_id: '' })
  }

  const handleGenerate = async () => {
    if (!genRoute) return
    setGenerating(true)
    try {
      const res = await routesApi.generateCollects(genRoute.id, genForm)
      toast(`${res.created} coleta(s) criada(s)`)
      setGenRoute(null)
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setGenerating(false) }
  }

  const toggleDriver = (id: string) => {
    setForm(p => ({
      ...p,
      driver_ids: p.driver_ids.includes(id) ? p.driver_ids.filter(d => d !== id) : [...p.driver_ids, id],
    }))
  }

  const toggleGenerator = (id: string) => {
    setGenForm(p => ({
      ...p,
      generator_ids: p.generator_ids.includes(id) ? p.generator_ids.filter(g => g !== id) : [...p.generator_ids, id],
    }))
  }

  return (
    <Layout title="Rotas" subtitle="Roteiros de coleta"
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
          <Button icon={<Plus size={16} />} onClick={openCreate}>Nova Rota</Button>
        </div>
      }>
      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        <div className="p-4">
          <Input placeholder="Buscar rota…" value={search}
            onChange={e => { setSearch(e.target.value); setPage(1) }}
            leftIcon={<Search size={15} />} className="max-w-sm" />
        </div>
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhuma rota" description="Crie a primeira rota para começar."
            action={<Button onClick={openCreate}>Criar Rota</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Nome', 'Dia', 'Semana', 'Motoristas', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(r => (
                  <tr key={r.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 font-medium text-gray-900">{r.name}</td>
                    <td className="px-4 py-3 text-gray-500">{WEEK_DAYS[r.week_day] ?? '—'}</td>
                    <td className="px-4 py-3 text-gray-500">{r.week_number === 0 ? 'Toda semana' : `${r.week_number}ª`}</td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {r.drivers?.length ? r.drivers.map(d => (
                          <span key={d.id} className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full">{d.name.split(' ')[0]}</span>
                        )) : <span className="text-gray-400 text-xs">—</span>}
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button onClick={() => openGenerate(r)} title="Gerar coletas"
                          className="p-1.5 rounded-lg hover:bg-brand-50 text-gray-400 hover:text-brand-500">
                          <CalendarPlus size={15} />
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
        {!loading && items.length > 0 && (
          <div className="px-4 border-t border-gray-100">
            <Pagination page={page} pages={pages} total={total} limit={20} onChange={setPage} />
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)}
        title={editing ? 'Editar Rota' : 'Nova Rota'}
        footer={<><Button variant="secondary" onClick={() => setModalOpen(false)}>Cancelar</Button><Button onClick={handleSave} loading={saving}>Salvar</Button></>}>
        <div className="space-y-4">
          <Input label="Nome da rota" required value={form.name} onChange={e => setForm(p => ({ ...p, name: e.target.value }))} placeholder="Rota Centro" />
          <div className="grid grid-cols-2 gap-4">
            <Select label="Dia da semana" value={form.week_day} onChange={e => setForm(p => ({ ...p, week_day: e.target.value }))}
              options={Object.entries(WEEK_DAYS).map(([v, l]) => ({ value: v, label: l }))} />
            <Select label="Frequência" value={form.week_number} onChange={e => setForm(p => ({ ...p, week_number: e.target.value }))}
              options={WEEK_NUMBER_OPTS} />
          </div>
          <div>
            <p className="text-xs font-medium text-gray-700 mb-2">Motoristas</p>
            <div className="border border-gray-200 rounded-xl max-h-40 overflow-y-auto divide-y divide-gray-50">
              {drivers.length === 0 ? (
                <p className="text-xs text-gray-400 p-3">Nenhum motorista cadastrado</p>
              ) : drivers.map(d => (
                <label key={d.id} className="flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-gray-50">
                  <input type="checkbox" checked={form.driver_ids.includes(d.id)} onChange={() => toggleDriver(d.id)}
                    className="rounded border-gray-300 text-brand-500 focus:ring-brand-400" />
                  <span className="text-sm text-gray-700">{d.name}</span>
                </label>
              ))}
            </div>
          </div>
        </div>
      </Modal>

      {/* Generate Collects Modal */}
      <Modal open={!!genRoute} onClose={() => setGenRoute(null)} title="Gerar Coletas" size="lg"
        footer={
          <>
            <Button variant="secondary" onClick={() => setGenRoute(null)}>Cancelar</Button>
            <Button onClick={handleGenerate} loading={generating}
              disabled={!genForm.target_date || genForm.generator_ids.length === 0 || !genForm.receiver_id}>
              Gerar {genForm.generator_ids.length > 0 ? `(${genForm.generator_ids.length})` : ''}
            </Button>
          </>
        }>
        <div className="space-y-4">
          <Input label="Data alvo" type="date" required value={genForm.target_date}
            onChange={e => setGenForm(p => ({ ...p, target_date: e.target.value }))} />
          <Select label="Recebedor" required value={genForm.receiver_id}
            onChange={e => setGenForm(p => ({ ...p, receiver_id: e.target.value }))}
            options={[{ value: '', label: 'Selecionar…' }, ...receivers.map(r => ({ value: r.id, label: r.name }))]} />
          <div>
            <p className="text-xs font-medium text-gray-700 mb-2">Geradores <span className="text-gray-400 font-normal">(selecione os que participam desta coleta)</span></p>
            <div className="border border-gray-200 rounded-xl max-h-48 overflow-y-auto divide-y divide-gray-50">
              {generators.map(g => (
                <label key={g.id} className="flex items-center gap-3 px-3 py-2 cursor-pointer hover:bg-gray-50">
                  <input type="checkbox" checked={genForm.generator_ids.includes(g.id)} onChange={() => toggleGenerator(g.id)}
                    className="rounded border-gray-300 text-brand-500 focus:ring-brand-400" />
                  <span className="text-sm text-gray-700">{g.name}</span>
                </label>
              ))}
            </div>
          </div>
        </div>
      </Modal>

      <ConfirmDialog open={!!deleting} onClose={() => setDeleting(null)} onConfirm={handleDelete}
        title="Excluir Rota" message={`Excluir rota "${deleting?.name}"?`} confirmLabel="Excluir" />

      <CsvImportModal
        open={importOpen}
        onClose={() => setImportOpen(false)}
        title="Importar Rotas"
        templateHeaders={['name', 'week_day', 'week_number']}
        templateExample={['Rota Norte', '1', '2']}
        onImport={routesApi.import}
      />

      <CsvDeleteModal
        open={deleteImportOpen}
        onClose={() => setDeleteImportOpen(false)}
        title="Excluir Rotas por CSV"
        onDelete={routesApi.importDelete}
      />
    </Layout>
  )
}
