import { useEffect, useState } from 'react'
import { TrendingUp, TrendingDown, DollarSign, Car, Users, BarChart3 } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts'
import { Layout } from '../components/layout/Layout'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { PageLoader } from '../components/ui/Spinner'
import { financialApi } from '../api/financial'
import type { FinancialSummary } from '../types'
import { fmtCurrency } from '../utils/formatters'

export function Financial() {
  const [summary, setSummary] = useState<FinancialSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [from, setFrom] = useState(() => {
    const now = new Date()
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`
  })
  const [to, setTo] = useState(() => {
    const now = new Date()
    return new Date(now.getFullYear(), now.getMonth() + 1, 0).toISOString().split('T')[0]
  })

  const load = async () => {
    setLoading(true)
    try {
      const s = await financialApi.summary(from, to)
      setSummary(s)
    } catch (e) { console.error(e) }
    finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])

  const barData = summary ? [
    { name: 'Receita', value: summary.revenue, fill: '#10b981' },
    { name: 'Veículos', value: summary.truck_costs, fill: '#f59e0b' },
    { name: 'Pessoal', value: summary.personnel_costs, fill: '#8b5cf6' },
    { name: 'Margem', value: summary.gross_margin, fill: summary.gross_margin >= 0 ? '#3b82f6' : '#ef4444' },
  ] : []

  const marginPct = summary && summary.revenue > 0
    ? ((summary.gross_margin / summary.revenue) * 100).toFixed(1)
    : null

  return (
    <Layout title="Resumo Financeiro" subtitle="Visão geral do período">
      {/* Period picker */}
      <div className="bg-white rounded-2xl border border-gray-200 mb-6 p-4 flex items-end gap-4">
        <Input label="De" type="date" value={from} onChange={e => setFrom(e.target.value)} className="w-44" />
        <Input label="Até" type="date" value={to} onChange={e => setTo(e.target.value)} className="w-44" />
        <Button onClick={load} className="mb-0.5">Atualizar</Button>
      </div>

      {loading ? <PageLoader /> : summary && (
        <>
          {/* KPI cards */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <KpiCard
              icon={<DollarSign size={20} />}
              label="Receita"
              value={fmtCurrency(summary.revenue)}
              color="green"
              trend="up"
            />
            <KpiCard
              icon={<Car size={20} />}
              label="Custos de Veículos"
              value={fmtCurrency(summary.truck_costs)}
              color="amber"
              trend="down"
            />
            <KpiCard
              icon={<Users size={20} />}
              label="Custos de Pessoal"
              value={fmtCurrency(summary.personnel_costs)}
              color="purple"
              trend="down"
            />
            <KpiCard
              icon={<BarChart3 size={20} />}
              label="Margem Bruta"
              value={fmtCurrency(summary.gross_margin)}
              sub={marginPct ? `${marginPct}% da receita` : undefined}
              color={summary.gross_margin >= 0 ? 'blue' : 'red'}
              trend={summary.gross_margin >= 0 ? 'up' : 'down'}
            />
          </div>

          {/* Chart */}
          <div className="bg-white rounded-2xl border border-gray-200 p-6">
            <div className="flex items-center gap-2 mb-6">
              <BarChart3 size={16} className="text-gray-400" />
              <h3 className="text-sm font-semibold text-gray-900">Comparativo do Período</h3>
            </div>
            <ResponsiveContainer width="100%" height={240}>
              <BarChart data={barData} barSize={48}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis dataKey="name" tick={{ fontSize: 12, fill: '#94a3b8' }} axisLine={false} tickLine={false} />
                <YAxis tickFormatter={v => fmtCurrency(v).replace('R$ ', 'R$ ')} tick={{ fontSize: 11, fill: '#94a3b8' }} axisLine={false} tickLine={false} width={90} />
                <Tooltip formatter={(v: number) => [fmtCurrency(v), '']} />
                <Bar dataKey="value" radius={[6, 6, 0, 0]}>
                  {barData.map((d, i) => <Cell key={i} fill={d.fill} />)}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Breakdown */}
          <div className="mt-6 bg-white rounded-2xl border border-gray-200 p-6">
            <h3 className="text-sm font-semibold text-gray-900 mb-4">Detalhamento</h3>
            <div className="space-y-3">
              <BreakdownRow label="Receita total" value={summary.revenue} positive total={summary.revenue} />
              <BreakdownRow label="Custos de veículos" value={summary.truck_costs} total={summary.revenue} />
              <BreakdownRow label="Custos de pessoal" value={summary.personnel_costs} total={summary.revenue} />
              <div className="border-t border-gray-100 pt-3">
                <BreakdownRow label="Margem bruta" value={summary.gross_margin} positive={summary.gross_margin >= 0} total={summary.revenue} bold />
              </div>
            </div>
          </div>
        </>
      )}
    </Layout>
  )
}

function KpiCard({ icon, label, value, sub, color, trend }: {
  icon: React.ReactNode
  label: string
  value: string
  sub?: string
  color: string
  trend: 'up' | 'down'
}) {
  const colors: Record<string, string> = {
    green: 'bg-green-50 text-green-600',
    amber: 'bg-amber-50 text-amber-600',
    purple: 'bg-purple-50 text-purple-600',
    blue: 'bg-blue-50 text-blue-600',
    red: 'bg-red-50 text-red-600',
  }
  return (
    <div className="bg-white rounded-2xl border border-gray-200 p-5">
      <div className={`w-10 h-10 rounded-xl flex items-center justify-center mb-3 ${colors[color]}`}>
        {icon}
      </div>
      <p className="text-2xl font-bold text-gray-900 truncate">{value}</p>
      <p className="text-sm text-gray-500 mt-0.5">{label}</p>
      {sub && <p className="text-xs text-gray-400 mt-1">{sub}</p>}
    </div>
  )
}

function BreakdownRow({ label, value, positive, total, bold }: {
  label: string
  value: number
  positive?: boolean
  total: number
  bold?: boolean
}) {
  const pct = total > 0 ? Math.abs((value / total) * 100) : 0
  return (
    <div className={`flex items-center gap-4 ${bold ? 'font-semibold' : ''}`}>
      <span className={`text-sm w-44 ${bold ? 'text-gray-900' : 'text-gray-600'}`}>{label}</span>
      <div className="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden">
        <div className="h-full rounded-full transition-all"
          style={{ width: `${Math.min(pct, 100)}%`, background: positive ? '#10b981' : '#f59e0b' }} />
      </div>
      <span className={`text-sm w-28 text-right ${positive ? 'text-green-600' : bold && value < 0 ? 'text-red-500' : 'text-gray-600'}`}>
        {positive && value >= 0 && <TrendingUp size={12} className="inline mr-1" />}
        {!positive && <TrendingDown size={12} className="inline mr-1" />}
        {fmtCurrency(value)}
      </span>
    </div>
  )
}
