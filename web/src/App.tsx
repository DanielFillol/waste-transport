import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'
import { ToastProvider } from './components/ui/Toast'

import { Login } from './pages/Login'
import { Register } from './pages/Register'
import { Dashboard } from './pages/Dashboard'
import { Generators } from './pages/Generators'
import { Receivers } from './pages/Receivers'
import { Drivers } from './pages/Drivers'
import { Trucks } from './pages/Trucks'
import { Routes as RoutesPage } from './pages/Routes'
import { Collects } from './pages/Collects'
import { Invoices } from './pages/Invoices'
import { Financial } from './pages/Financial'
import { PricingRules } from './pages/PricingRules'
import { TruckCosts } from './pages/TruckCosts'
import { PersonnelCosts } from './pages/PersonnelCosts'
import { Alerts } from './pages/Alerts'
import { AuditLogs } from './pages/AuditLogs'
import { PageLoader } from './components/ui/Spinner'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <PageLoader />
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

function GuestRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <PageLoader />
  if (user) return <Navigate to="/" replace />
  return <>{children}</>
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
      <Route path="/register" element={<GuestRoute><Register /></GuestRoute>} />

      <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      <Route path="/geradores" element={<ProtectedRoute><Generators /></ProtectedRoute>} />
      <Route path="/recebedores" element={<ProtectedRoute><Receivers /></ProtectedRoute>} />
      <Route path="/motoristas" element={<ProtectedRoute><Drivers /></ProtectedRoute>} />
      <Route path="/veiculos" element={<ProtectedRoute><Trucks /></ProtectedRoute>} />
      <Route path="/rotas" element={<ProtectedRoute><RoutesPage /></ProtectedRoute>} />
      <Route path="/coletas" element={<ProtectedRoute><Collects /></ProtectedRoute>} />
      <Route path="/notas-fiscais" element={<ProtectedRoute><Invoices /></ProtectedRoute>} />
      <Route path="/financeiro" element={<ProtectedRoute><Financial /></ProtectedRoute>} />
      <Route path="/regras-preco" element={<ProtectedRoute><PricingRules /></ProtectedRoute>} />
      <Route path="/custos-veiculos" element={<ProtectedRoute><TruckCosts /></ProtectedRoute>} />
      <Route path="/custos-pessoal" element={<ProtectedRoute><PersonnelCosts /></ProtectedRoute>} />
      <Route path="/alertas" element={<ProtectedRoute><Alerts /></ProtectedRoute>} />
      <Route path="/auditoria" element={<ProtectedRoute><AuditLogs /></ProtectedRoute>} />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <AppRoutes />
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}
