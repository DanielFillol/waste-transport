import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard, Truck, MapPin, Building2, Users, Car,
  FileText, DollarSign, Bell, ClipboardList, Tag, Wrench,
  UserCog, History, ChevronRight, Recycle,
} from 'lucide-react'

interface NavItem {
  label: string
  to: string
  icon: React.ReactNode
}

interface NavSection {
  title: string
  items: NavItem[]
}

const sections: NavSection[] = [
  {
    title: 'Geral',
    items: [
      { label: 'Dashboard', to: '/', icon: <LayoutDashboard size={18} /> },
    ],
  },
  {
    title: 'Operações',
    items: [
      { label: 'Coletas', to: '/coletas', icon: <ClipboardList size={18} /> },
      { label: 'Rotas', to: '/rotas', icon: <MapPin size={18} /> },
    ],
  },
  {
    title: 'Cadastros',
    items: [
      { label: 'Geradores', to: '/geradores', icon: <Building2 size={18} /> },
      { label: 'Recebedores', to: '/recebedores', icon: <Building2 size={18} /> },
      { label: 'Motoristas', to: '/motoristas', icon: <Users size={18} /> },
      { label: 'Veículos', to: '/veiculos', icon: <Car size={18} /> },
    ],
  },
  {
    title: 'Financeiro',
    items: [
      { label: 'Visão Geral', to: '/financeiro', icon: <DollarSign size={18} /> },
      { label: 'Notas Fiscais', to: '/notas-fiscais', icon: <FileText size={18} /> },
      { label: 'Regras de Preço', to: '/regras-preco', icon: <Tag size={18} /> },
      { label: 'Custos de Veículos', to: '/custos-veiculos', icon: <Wrench size={18} /> },
      { label: 'Custos de Pessoal', to: '/custos-pessoal', icon: <UserCog size={18} /> },
    ],
  },
  {
    title: 'Sistema',
    items: [
      { label: 'Alertas', to: '/alertas', icon: <Bell size={18} /> },
      { label: 'Auditoria', to: '/auditoria', icon: <History size={18} /> },
    ],
  },
]

export function Sidebar() {
  return (
    <aside className="w-60 bg-slate-900 flex flex-col h-screen sticky top-0 shrink-0">
      {/* Logo */}
      <div className="px-5 py-5 border-b border-slate-800">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-brand-500 flex items-center justify-center">
            <Recycle size={16} className="text-white" />
          </div>
          <div>
            <p className="text-sm font-bold text-white tracking-tight">Waste</p>
            <p className="text-[10px] text-slate-500 uppercase tracking-widest">Gestão de Resíduos</p>
          </div>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-5">
        {sections.map((section) => (
          <div key={section.title}>
            <p className="px-2 mb-1.5 text-[10px] font-semibold text-slate-500 uppercase tracking-widest">
              {section.title}
            </p>
            <div className="space-y-0.5">
              {section.items.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.to === '/'}
                  className={({ isActive }) =>
                    `flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-sm font-medium transition-all duration-100
                    ${isActive
                      ? 'bg-brand-500/15 text-brand-400 shadow-sm'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'
                    }`
                  }
                >
                  {({ isActive }) => (
                    <>
                      <span className={isActive ? 'text-brand-400' : 'text-slate-500'}>
                        {item.icon}
                      </span>
                      <span className="flex-1">{item.label}</span>
                      {isActive && <ChevronRight size={12} className="text-brand-400" />}
                    </>
                  )}
                </NavLink>
              ))}
            </div>
          </div>
        ))}
      </nav>

      {/* Bottom truck decoration */}
      <div className="px-5 py-4 border-t border-slate-800">
        <div className="flex items-center gap-2 text-slate-600">
          <Truck size={14} />
          <span className="text-xs">v1.0.0</span>
        </div>
      </div>
    </aside>
  )
}
