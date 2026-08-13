import type { ComponentType, ReactNode } from 'react'
import {
  BookOpen,
  Gamepad2,
  Home,
  LayoutDashboard,
  LogOut,
  Menu,
  Server,
  Settings,
  Users,
  X,
} from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { PageId, UserAccount } from '@/types'
import { initialOf } from '@/types'

interface AppShellProps {
  activePage: PageId
  isSidebarOpen: boolean
  user: UserAccount
  children: ReactNode
  onNavigate: (page: PageId) => void
  onToggleSidebar: () => void
  onLogout: () => void
}

interface NavigationItem {
  id: PageId
  label: string
  icon: ComponentType<{ size?: number }>
  adminOnly?: boolean
}

const navigation: NavigationItem[] = [
  { id: 'home', label: 'Home', icon: Home },
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'maimai', label: 'maimai DX', icon: Gamepad2 },
  { id: 'setup', label: 'Setup Guide', icon: Server },
  { id: 'admin', label: 'Admin Panel', icon: Users, adminOnly: true },
  { id: 'settings', label: 'Settings', icon: Settings },
]

export function AppShell({
  activePage,
  isSidebarOpen,
  user,
  children,
  onNavigate,
  onToggleSidebar,
  onLogout,
}: AppShellProps) {
  const pageTitle = navigation.find((item) => item.id === activePage)?.label || activePage

  return (
    <div className="min-h-screen bg-slate-950 text-slate-50 flex overflow-hidden">
      <aside
        className={`${isSidebarOpen ? 'w-64' : 'w-20'} bg-slate-900 border-r border-slate-800 flex flex-col transition-all duration-300 z-50`}
      >
        <div className="p-6 flex items-center justify-between gap-3">
          <span className={`font-black text-2xl text-indigo-500 tracking-tighter ${!isSidebarOpen ? 'hidden' : ''}`}>
            MaiGoDX
          </span>
          <button
            type="button"
            aria-label="Toggle sidebar"
            onClick={onToggleSidebar}
            className="text-slate-400 hover:text-white transition-colors"
          >
            {isSidebarOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>

        <nav aria-label="Main navigation" className="flex-1 px-3 space-y-2">
          {navigation.map(({ id, label, icon: Icon, adminOnly }) => {
            if (adminOnly && !user.isAdmin) return null

            const active = activePage === id
            return (
              <button
                key={id}
                type="button"
                onClick={() => onNavigate(id)}
                className={`flex items-center gap-3 px-4 py-3 w-full rounded-lg transition-all ${
                  active
                    ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/20'
                    : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                }`}
              >
                <Icon size={20} />
                <span className={`font-medium whitespace-nowrap ${!isSidebarOpen ? 'hidden' : ''}`}>{label}</span>
              </button>
            )
          })}
        </nav>

        <div className="p-4 border-t border-slate-800">
          <button
            type="button"
            onClick={onLogout}
            className="flex items-center gap-3 px-4 py-3 w-full text-rose-400 hover:bg-rose-500/10 rounded-lg transition-all"
          >
            <LogOut size={20} />
            <span className={`font-medium ${!isSidebarOpen ? 'hidden' : ''}`}>Logout</span>
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-y-auto relative">
        <header className="h-16 border-b border-slate-800 bg-slate-950/50 backdrop-blur-xl flex items-center justify-between px-8 sticky top-0 z-40">
          <h1 className="text-lg font-bold">{pageTitle}</h1>
          <div className="flex items-center gap-4">
            <div className="text-right hidden sm:block">
              <p className="text-sm font-bold">{user.username}</p>
              <p className="text-[10px] text-slate-500">{user.email}</p>
            </div>
            <Avatar className="h-9 w-9 border border-slate-700">
              <AvatarFallback className="bg-indigo-600 text-white font-bold">{initialOf(user.username)}</AvatarFallback>
            </Avatar>
          </div>
        </header>

        <div className="p-8">{children}</div>
      </main>
    </div>
  )
}

export const pageIcons = { Home, LayoutDashboard, Gamepad2, Server, BookOpen, Settings, Users }
