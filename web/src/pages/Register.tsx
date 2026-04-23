import { useState, type FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Recycle, Building2, User, Lock, Info } from 'lucide-react'
import { authApi } from '../api/auth'
import { useAuth } from '../context/AuthContext'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'

export function Register() {
  const { authenticate } = useAuth()
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirm, setConfirm] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [createdSlug, setCreatedSlug] = useState('')

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    if (password !== confirm) { setError('As senhas não conferem'); return }
    setLoading(true)
    try {
      const res = await authApi.register(name, username, password)
      setCreatedSlug(res.tenant?.slug ?? '')
      authenticate(res.token, res.user!)
      setTimeout(() => navigate('/'), 2500)
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Erro ao criar conta')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-8">
      <div className="w-full max-w-sm">
        <div className="flex items-center gap-2 mb-8">
          <div className="w-8 h-8 rounded-lg bg-brand-500 flex items-center justify-center">
            <Recycle size={16} className="text-white" />
          </div>
          <span className="text-lg font-bold text-gray-900">Waste</span>
        </div>

        <h1 className="text-2xl font-bold text-gray-900 mb-1">Criar conta</h1>
        <p className="text-sm text-gray-500 mb-8">
          Crie o tenant da sua empresa para começar.
        </p>

        {createdSlug && (
          <div className="mb-4 p-3 bg-brand-50 border border-brand-200 rounded-lg flex gap-2 text-sm text-brand-700">
            <Info size={16} className="shrink-0 mt-0.5" />
            <span>
              Conta criada! Seu ID de acesso é <strong>{createdSlug}</strong>. Guarde-o para fazer login.
              Redirecionando…
            </span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Nome da empresa"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="Empresa Ambiental Ltda"
            required
            leftIcon={<Building2 size={15} />}
          />
          <Input
            label="Usuário admin"
            value={username}
            onChange={e => setUsername(e.target.value)}
            placeholder="admin"
            required
            leftIcon={<User size={15} />}
          />
          <Input
            label="Senha"
            type="password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            placeholder="Mínimo 6 caracteres"
            required
            leftIcon={<Lock size={15} />}
          />
          <Input
            label="Confirmar senha"
            type="password"
            value={confirm}
            onChange={e => setConfirm(e.target.value)}
            placeholder="••••••••"
            required
            leftIcon={<Lock size={15} />}
          />
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
              {error}
            </div>
          )}
          <Button type="submit" loading={loading} className="w-full" size="lg" disabled={!!createdSlug}>
            Criar conta
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-500">
          Já tem conta?{' '}
          <Link to="/login" className="text-brand-600 font-medium hover:underline">
            Entrar
          </Link>
        </p>
      </div>
    </div>
  )
}
