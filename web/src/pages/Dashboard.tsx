import { useEffect, useState } from 'react'
import { Building2, Users, Car, ClipboardList, TrendingUp, TrendingDown, Bell } from 'lucide-react'
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid } from 'recharts'
import { Layout } from '../components/layout/Layout'
import { PageLoader } from '../components/ui/Spinner'
import { dashboardApi } from '../api/dashboard'
import { financialApi } from '../api/financial'
import type { DashboardData, FinancialSummary } from '../types'
import { fmtCurrency } from '../utils/formatters'

const STATUS_COLORS = ['#3b82f6', '#10b981', '#ef4444']
const STATUS_LABELS = ['Planejadas', 'Coletadas', 'Canceladas']

export function Dashboard() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [summary, setSummary] = useState<FinancialSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const now = new Date()
    const start = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`
    const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)
      .toISOString().split('T')[0]

    Promise.all([
      dashboardApi.get(),
      financialApi.summary(start, end),
    ])
      .then(([d, s]) => { setData(d); setSummary(s) })
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <Layout title="Dashboard"><PageLoader /></Layout>

  const collectCounts = data?.collect_counts ?? {}
  const pieData = [
    { name: 'Planejadas', value: collectCounts['1'] ?? 0 },
    { name: 'Coletadas', value: collectCounts['2'] ?? 0 },
    { name: 'Canceladas', value: collectCounts['3'] ?? 0 },
  ]

  const barData = summary
    ? [
        { name: 'Receita', value: summary.revenue, fill: '#10b981' },
        { name: 'Veículos', value: summary.truck_costs, fill: '#f59e0b' },
        { name: 'Pessoal', value: summary.personnel_costs, fill: '#8b5cf6' },
        { name: 'Margem', value: summary.gross_margin, fill: '#3b82f6' },
      ]
    : []

  return (
    <Layout title="Dashboard" subtitle={`Resumo do mês de ${new Date().toLocaleDateString('pt-BR', { month: 'long', year: 'numeric' })}`}>
      {/* Stats grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard icon={<Building2 size={20} />} label="Geradores" value={data?.generator_count ?? 0} color="blue" />
        <StatCard icon={<Building2 size={20} />} label="Recebedores" value={data?.receiver_count ?? 0} color="purple" />
        <StatCard icon={<Users size={20} />} label="Motoristas" value={data?.driver_count ?? 0} color="amber" />
        <StatCard icon={<Car size={20} />} label="Veículos" value={data?.truck_count ?? 0} color="green" />
      </div>

      {/* Alerts banner */}
      {(data?.unread_alerts ?? 0) > 0 && (
        <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-xl flex items-center gap-3">
          <Bell size={18} className="text-amber-500 shrink-0" />
          <p className="text-sm text-amber-800">
            Você tem <strong>{data?.unread_alerts}</strong> alerta{data!.unread_alerts > 1 ? 's' : ''} não lido{data!.unread_alerts > 1 ? 's' : ''}.
          </p>
          <a href="/alertas" className="ml-auto text-sm font-medium text-amber-700 hover:underline">
            Ver alertas
          </a>
        </div>
      )}

      <div className="grid lg:grid-cols-2 gap-6 mb-6">
        {/* Collect distribution */}
        <div className="bg-white rounded-2xl border border-gray-200 p-6">
          <div className="flex items-center gap-2 mb-1">
            <ClipboardList size={16} className="text-gray-400" />
            <h3 className="text-sm font-semibold text-gray-900">Coletas por Status</h3>
          </div>
          <p className="text-xs text-gray-400 mb-4">Total: {Object.values(collectCounts).reduce((a, b) => a + b, 0)}</p>
          <div className="flex items-center gap-6">
            <ResponsiveContainer width={140} height={140}>
              <PieChart>
                <Pie data={pieData} cx="50%" cy="50%" innerRadius={40} outerRadius={65} dataKey="value" paddingAngle={3}>
                  {pieData.map((_, i) => <Cell key={i} fill={STATUS_COLORS[i]} />)}
                </Pie>
                <Tooltip formatter={(v: number) => [v, '']} />
              </PieChart>
            </ResponsiveContainer>
            <div className="space-y-2">
              {pieData.map((d, i) => (
                <div key={i} className="flex items-center gap-2">
                  <div className="w-2.5 h-2.5 rounded-full" style={{ background: STATUS_COLORS[i] }} />
                  <span className="text-xs text-gray-500">{STATUS_LABELS[i]}</span>
                  <span className="text-xs font-semibold text-gray-900 ml-auto pl-4">{d.value}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Financial summary */}
        <div className="bg-white rounded-2xl border border-gray-200 p-6">
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp size={16} className="text-gray-400" />
            <h3 className="text-sm font-semibold text-gray-900">Resumo Financeiro do Mês</h3>
          </div>
          {summary && (
            <div className="space-y-3 mb-4">
              <FinLine label="Receita" value={summary.revenue} positive />
              <FinLine label="Custos de veículos" value={summary.truck_costs} />
              <FinLine label="Custos de pessoal" value={summary.personnel_costs} />
              <div className="pt-2 border-t border-gray-100">
                <FinLine
                  label="Margem bruta"
                  value={summary.gross_margin}
                  positive={summary.gross_margin >= 0}
                  bold
                />
              </div>
            </div>
          )}
          <ResponsiveContainer width="100%" height={100}>
            <BarChart data={barData} barSize={28}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
              <XAxis dataKey="name" tick={{ fontSize: 10, fill: '#94a3b8' }} axisLine={false} tickLine={false} />
              <YAxis hide />
              <Tooltip formatter={(v: number) => [fmtCurrency(v), '']} />
              <Bar dataKey="value" radius={[4, 4, 0, 0]}>
                {barData.map((d, i) => <Cell key={i} fill={d.fill} />)}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </Layout>
  )
}

function StatCard({ icon, label, value, color }: {
  icon: React.ReactNode
  label: string
  value: number
  color: string
}) {
  const colors: Record<string, string> = {
    blue: 'bg-blue-50 text-blue-500',
    purple: 'bg-purple-50 text-purple-500',
    amber: 'bg-amber-50 text-amber-500',
    green: 'bg-green-50 text-green-500',
  }
  return (
    <div className="bg-white rounded-2xl border border-gray-200 p-5">
      <div className={`w-10 h-10 rounded-xl flex items-center justify-center mb-3 ${colors[color]}`}>
        {icon}
      </div>
      <p className="text-2xl font-bold text-gray-900">{value}</p>
      <p className="text-sm text-gray-500 mt-0.5">{label}</p>
    </div>
  )
}

function FinLine({ label, value, positive, bold }: {
  label: string
  value: number
  positive?: boolean
  bold?: boolean
}) {
  return (
    <div className={`flex justify-between items-center ${bold ? 'font-semibold' : ''}`}>
      <span className={`text-sm ${bold ? 'text-gray-900' : 'text-gray-500'}`}>{label}</span>
      <span className={`text-sm ${positive ? 'text-green-600' : 'text-red-500'}`}>
        {positive ? <TrendingUp size={12} className="inline mr-1" /> : <TrendingDown size={12} className="inline mr-1" />}
        {fmtCurrency(value)}
      </span>
    </div>
  )
}
