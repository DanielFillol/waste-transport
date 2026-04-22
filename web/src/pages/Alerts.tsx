import { useEffect, useState } from 'react'
import { Bell, BellOff, CheckCheck } from 'lucide-react'
import { Layout } from '../components/layout/Layout'
import { Button } from '../components/ui/Button'
import { PageLoader } from '../components/ui/Spinner'
import { EmptyState } from '../components/ui/EmptyState'
import { useToast } from '../components/ui/Toast'
import { alertsApi } from '../api/alerts'
import type { Alert } from '../types'
import { fmtDateTime } from '../utils/formatters'

const ALERT_TYPE_LABEL: Record<string, string> = {
  cnh_expiring: 'CNH Vencendo',
  license_expiring: 'Licença Vencendo',
}

const ALERT_TYPE_COLOR: Record<string, string> = {
  cnh_expiring: 'bg-amber-50 border-amber-200',
  license_expiring: 'bg-red-50 border-red-200',
}

export function Alerts() {
  const toast = useToast()
  const [items, setItems] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [markingAll, setMarkingAll] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const res = await alertsApi.list()
      setItems(res.data)
    } finally { setLoading(false) }
  }

  useEffect(() => { load() }, [])

  const markRead = async (a: Alert) => {
    if (a.read) return
    try { await alertsApi.markRead(a.id); load() }
    catch { toast('Erro', 'error') }
  }

  const markAllRead = async () => {
    setMarkingAll(true)
    try { await alertsApi.markAllRead(); toast('Todos os alertas marcados como lidos'); load() }
    catch (e: unknown) { toast(e instanceof Error ? e.message : 'Erro', 'error') }
    finally { setMarkingAll(false) }
  }

  const unreadCount = items.filter(a => !a.read).length

  return (
    <Layout title="Alertas" subtitle="Notificações do sistema"
      actions={
        unreadCount > 0 ? (
          <Button variant="secondary" icon={<CheckCheck size={16} />} onClick={markAllRead} loading={markingAll}>
            Marcar todos como lidos
          </Button>
        ) : undefined
      }>

      {loading ? <PageLoader /> : items.length === 0 ? (
        <EmptyState
          title="Nenhum alerta"
          description="Você está em dia! Não há alertas pendentes."
          icon={<BellOff size={32} className="text-gray-300" />}
        />
      ) : (
        <div className="space-y-3">
          {unreadCount > 0 && (
            <p className="text-sm text-gray-500 mb-4">
              <span className="font-semibold text-amber-600">{unreadCount}</span> alerta{unreadCount > 1 ? 's' : ''} não lido{unreadCount > 1 ? 's' : ''}
            </p>
          )}
          {items.map(a => (
            <div
              key={a.id}
              onClick={() => markRead(a)}
              className={`flex items-start gap-4 p-4 rounded-2xl border transition-all cursor-pointer ${
                a.read
                  ? 'bg-white border-gray-100 opacity-60'
                  : ALERT_TYPE_COLOR[a.type] ?? 'bg-amber-50 border-amber-200'
              }`}
            >
              <div className={`mt-0.5 shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${
                a.read ? 'bg-gray-100' : 'bg-amber-100'
              }`}>
                <Bell size={15} className={a.read ? 'text-gray-400' : 'text-amber-500'} />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-semibold text-gray-500 uppercase tracking-wider">
                    {ALERT_TYPE_LABEL[a.type] ?? a.type}
                  </span>
                  {!a.read && (
                    <span className="w-1.5 h-1.5 rounded-full bg-amber-500 shrink-0" />
                  )}
                </div>
                <p className="text-sm font-medium text-gray-900">{a.title}</p>
                <p className="text-sm text-gray-600 mt-0.5">{a.message}</p>
                <p className="text-xs text-gray-400 mt-2">{fmtDateTime(a.created_at)}</p>
              </div>
              {a.read && (
                <CheckCheck size={14} className="text-gray-300 shrink-0 mt-1" />
              )}
            </div>
          ))}
        </div>
      )}
    </Layout>
  )
}
