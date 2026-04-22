import { useState, type FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { Recycle, Building2 } from 'lucide-react'
import { authApi } from '../api/auth'
import { setToken } from '../api/client'
import { Input } from '../components/ui/Input'
import { Button } from '../components/ui/Button'

export function Register() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await authApi.register(name)
      setToken(res.token)
      navigate('/login')
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

        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Nome da empresa"
            value={name}
            onChange={e => setName(e.target.value)}
            placeholder="Empresa Ambiental Ltda"
            required
            leftIcon={<Building2 size={15} />}
          />
          {error && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
              {error}
            </div>
          )}
          <Button type="submit" loading={loading} className="w-full" size="lg">
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
