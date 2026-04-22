import { useEffect, useState, useCallback } from 'react'
import { Plus, Eye, CheckCircle, Send, FileText } from 'lucide-react'
import { Layout } from '../components/layout/Layout'
import { Button } from '../components/ui/Button'
import { Select } from '../components/ui/Select'
import { Input } from '../components/ui/Input'
import { Badge } from '../components/ui/Badge'
import { Modal } from '../components/ui/Modal'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { Pagination } from '../components/ui/Pagination'
import { EmptyState } from '../components/ui/EmptyState'
import { PageLoader } from '../components/ui/Spinner'
import { useToast } from '../components/ui/Toast'
import { financialApi } from '../api/financial'
import { generatorsApi } from '../api/generators'
import type { Invoice, InvoiceStatus, Generator } from '../types'
import { fmtDate, fmtCurrency, INVOICE_STATUS_LABEL, INVOICE_STATUS_COLOR } from '../utils/formatters'

interface GenerateForm {
  generator_id: string
  period_start: string
  period_end: string
  notes: string
}

const thisMonth = () => {
  const now = new Date()
  const start = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString().split('T')[0]
  return { start, end }
}

export function Invoices() {
  const toast = useToast()
  const [items, setItems] = useState<Invoice[]>([])
  const [total, setTotal] = useState(0)
  const [pages, setPages] = useState(1)
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<InvoiceStatus | ''>('')
  const [loading, setLoading] = useState(true)
  const [generators, setGenerators] = useState<Generator[]>([])

  const [genForm, setGenForm] = useState<GenerateForm>(() => {
    const { start, end } = thisMonth()
    return { generator_id: '', period_start: start, period_end: end, notes: '' }
  })
  const [genOpen, setGenOpen] = useState(false)
  const [generating, setGenerating] = useState(false)

  const [detailInvoice, setDetailInvoice] = useState<Invoice | null>(null)
  const [issuing, setIssuing] = useState<string | null>(null)
  const [paying, setPaying] = useState<Invoice | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await financialApi.listInvoices({ status: statusFilter || undefined, page, limit: 20 })
      setItems(res.data); setTotal(res.total); setPages(res.pages)
    } finally { setLoading(false) }
  }, [statusFilter, page])

  useEffect(() => { load() }, [load])

  useEffect(() => {
    generatorsApi.list({ limit: 200 }).then(r => setGenerators(r.data))
  }, [])

  const handleGenerate = async () => {
    setGenerating(true)
    try {
      await financialApi.generateInvoice(genForm)
      toast('Nota fiscal gerada com sucesso')
      setGenOpen(false)
      load()
    } catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setGenerating(false) }
  }

  const handleIssue = async (id: string) => {
    setIssuing(id)
    try { await financialApi.issueInvoice(id); toast('Nota fiscal emitida'); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setIssuing(null) }
  }

  const handleMarkPaid = async () => {
    if (!paying) return
    try { await financialApi.markInvoicePaid(paying.id); toast('Pagamento registrado'); setPaying(null); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
  }

  const openDetail = async (inv: Invoice) => {
    try {
      const full = await financialApi.getInvoice(inv.id)
      setDetailInvoice(full)
    } catch { setDetailInvoice(inv) }
  }

  return (
    <Layout title="Notas Fiscais" subtitle="Faturamento por gerador"
      actions={<Button icon={<Plus size={16} />} onClick={() => setGenOpen(true)}>Gerar NF</Button>}>

      {/* Status filter */}
      <div className="bg-white rounded-2xl border border-gray-200 mb-4 p-4 flex gap-3 items-center">
        <Select label="Status" value={statusFilter}
          onChange={e => { setStatusFilter(e.target.value as InvoiceStatus | ''); setPage(1) }}
          options={[
            { value: '', label: 'Todos' },
            { value: 'draft', label: 'Rascunho' },
            { value: 'issued', label: 'Emitida' },
            { value: 'paid', label: 'Paga' },
          ]}
          className="w-48" />
        <Select label="Gerador" value=""
          onChange={e => e.target.value && setPage(1)}
          options={[{ value: '', label: 'Todos geradores' }, ...generators.map(g => ({ value: g.id, label: g.name }))]}
          className="w-64" />
      </div>

      <div className="bg-white rounded-2xl border border-gray-200 mb-4">
        {loading ? <PageLoader /> : items.length === 0 ? (
          <EmptyState title="Nenhuma nota fiscal" description="Gere a primeira NF para começar."
            action={<Button onClick={() => setGenOpen(true)}>Gerar NF</Button>} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-t border-gray-100">
                {['Nº NF', 'Gerador', 'Período', 'Valor Total', 'Status', 'Emitida em', 'Paga em', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr></thead>
              <tbody className="divide-y divide-gray-50">
                {items.map(inv => (
                  <tr key={inv.id} className="hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3 font-mono text-xs text-gray-700">{inv.invoice_number || '—'}</td>
                    <td className="px-4 py-3 font-medium text-gray-900 max-w-[160px] truncate">
                      {inv.generator?.name ?? generators.find(g => g.id === inv.generator_id)?.name ?? '—'}
                    </td>
                    <td className="px-4 py-3 text-gray-500 text-xs">
                      {fmtDate(inv.period_start)} – {fmtDate(inv.period_end)}
                    </td>
                    <td className="px-4 py-3 font-semibold text-gray-900">{fmtCurrency(inv.total_amount)}</td>
                    <td className="px-4 py-3">
                      <Badge className={INVOICE_STATUS_COLOR[inv.status]}>
                        {INVOICE_STATUS_LABEL[inv.status]}
                      </Badge>
                    </td>
                    <td className="px-4 py-3 text-gray-400 text-xs">{fmtDate(inv.issued_at)}</td>
                    <td className="px-4 py-3 text-gray-400 text-xs">{fmtDate(inv.paid_at)}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1 justify-end">
                        <button onClick={() => openDetail(inv)} title="Ver detalhes"
                          className="p-1.5 rounded-lg hover:bg-gray-100 text-gray-400 hover:text-gray-600">
                          <Eye size={15} />
                        </button>
                        {inv.status === 'draft' && (
                          <button onClick={() => handleIssue(inv.id)}
                            disabled={issuing === inv.id}
                            title="Emitir"
                            className="p-1.5 rounded-lg hover:bg-blue-50 text-gray-400 hover:text-blue-500 disabled:opacity-50">
                            <Send size={15} />
                          </button>
                        )}
                        {inv.status === 'issued' && (
                          <button onClick={() => setPaying(inv)} title="Marcar como paga"
                            className="p-1.5 rounded-lg hover:bg-green-50 text-gray-400 hover:text-green-500">
                            <CheckCircle size={15} />
                          </button>
                        )}
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

      {/* Generate Invoice Modal */}
      <Modal open={genOpen} onClose={() => setGenOpen(false)} title="Gerar Nota Fiscal"
        footer={
          <>
            <Button variant="secondary" onClick={() => setGenOpen(false)}>Cancelar</Button>
            <Button onClick={handleGenerate} loading={generating} disabled={!genForm.generator_id}>Gerar</Button>
          </>
        }>
        <div className="space-y-4">
          <Select label="Gerador *" required value={genForm.generator_id}
            onChange={e => setGenForm(p => ({ ...p, generator_id: e.target.value }))}
            options={[{ value: '', label: 'Selecionar…' }, ...generators.map(g => ({ value: g.id, label: g.name }))]} />
          <div className="grid grid-cols-2 gap-4">
            <Input label="Início do período" type="date" required value={genForm.period_start}
              onChange={e => setGenForm(p => ({ ...p, period_start: e.target.value }))} />
            <Input label="Fim do período" type="date" required value={genForm.period_end}
              onChange={e => setGenForm(p => ({ ...p, period_end: e.target.value }))} />
          </div>
          <Input label="Observações" value={genForm.notes} onChange={e => setGenForm(p => ({ ...p, notes: e.target.value }))} />
        </div>
      </Modal>

      {/* Invoice Detail Modal */}
      <Modal open={!!detailInvoice} onClose={() => setDetailInvoice(null)} title="Detalhes da NF" size="lg"
        footer={<Button variant="secondary" onClick={() => setDetailInvoice(null)}>Fechar</Button>}>
        {detailInvoice && (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div><p className="text-xs text-gray-400 mb-1">Gerador</p><p className="font-medium">{detailInvoice.generator?.name ?? '—'}</p></div>
              <div><p className="text-xs text-gray-400 mb-1">Status</p>
                <Badge className={INVOICE_STATUS_COLOR[detailInvoice.status]}>{INVOICE_STATUS_LABEL[detailInvoice.status]}</Badge>
              </div>
              <div><p className="text-xs text-gray-400 mb-1">Período</p><p>{fmtDate(detailInvoice.period_start)} – {fmtDate(detailInvoice.period_end)}</p></div>
              <div><p className="text-xs text-gray-400 mb-1">Valor total</p><p className="font-semibold text-gray-900">{fmtCurrency(detailInvoice.total_amount)}</p></div>
              <div><p className="text-xs text-gray-400 mb-1">Emitida em</p><p>{fmtDate(detailInvoice.issued_at)}</p></div>
              <div><p className="text-xs text-gray-400 mb-1">Paga em</p><p>{fmtDate(detailInvoice.paid_at)}</p></div>
            </div>
            {detailInvoice.items && detailInvoice.items.length > 0 && (
              <div className="border-t border-gray-100 pt-4">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3 flex items-center gap-2">
                  <FileText size={12} /> Itens
                </p>
                <table className="w-full text-xs">
                  <thead><tr className="text-gray-400">
                    <th className="text-left pb-2">Descrição</th>
                    <th className="text-right pb-2">Qtd</th>
                    <th className="text-right pb-2">Un.</th>
                    <th className="text-right pb-2">Preço unit.</th>
                    <th className="text-right pb-2">Total</th>
                  </tr></thead>
                  <tbody className="divide-y divide-gray-50">
                    {detailInvoice.items.map(item => (
                      <tr key={item.id}>
                        <td className="py-2 text-gray-700">{item.description}</td>
                        <td className="py-2 text-right text-gray-600">{item.quantity}</td>
                        <td className="py-2 text-right text-gray-600">{item.unit}</td>
                        <td className="py-2 text-right text-gray-600">{fmtCurrency(item.unit_price)}</td>
                        <td className="py-2 text-right font-semibold text-gray-900">{fmtCurrency(item.total_price)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}
      </Modal>

      <ConfirmDialog open={!!paying} onClose={() => setPaying(null)} onConfirm={handleMarkPaid}
        title="Registrar Pagamento" message={`Confirmar pagamento da NF ${paying?.invoice_number ?? ''}?`}
        confirmLabel="Confirmar Pagamento" />
    </Layout>
  )
}
