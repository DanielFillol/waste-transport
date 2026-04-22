import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'
import { CheckCircle, XCircle, X, AlertCircle } from 'lucide-react'

type ToastType = 'success' | 'error' | 'info'

interface Toast {
  id: number
  type: ToastType
  message: string
}

interface ToastContextValue {
  toast: (message: string, type?: ToastType) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

let nextId = 0

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const toast = useCallback((message: string, type: ToastType = 'success') => {
    const id = nextId++
    setToasts(prev => [...prev, { id, type, message }])
    setTimeout(() => setToasts(prev => prev.filter(t => t.id !== id)), 4000)
  }, [])

  const remove = (id: number) => setToasts(prev => prev.filter(t => t.id !== id))

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 w-80">
        {toasts.map(t => (
          <div
            key={t.id}
            className={`flex items-start gap-3 p-4 rounded-xl shadow-lg border backdrop-blur-sm animate-in slide-in-from-right-5
              ${t.type === 'success' ? 'bg-green-50 border-green-200 text-green-800' : ''}
              ${t.type === 'error' ? 'bg-red-50 border-red-200 text-red-800' : ''}
              ${t.type === 'info' ? 'bg-blue-50 border-blue-200 text-blue-800' : ''}
            `}
          >
            {t.type === 'success' && <CheckCircle size={18} className="text-green-500 shrink-0 mt-0.5" />}
            {t.type === 'error' && <XCircle size={18} className="text-red-500 shrink-0 mt-0.5" />}
            {t.type === 'info' && <AlertCircle size={18} className="text-blue-500 shrink-0 mt-0.5" />}
            <p className="text-sm flex-1">{t.message}</p>
            <button onClick={() => remove(t.id)} className="text-current opacity-50 hover:opacity-100">
              <X size={14} />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx.toast
}
