import { useState, type ComponentType, type ReactNode } from 'react'
import {
  BookOpen,
  Gamepad2,
  Home,
  LayoutDashboard,
  LogOut,
  Menu,
  CreditCard,
  Server,
  Settings,
  Users,
  X,
} from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import type { PageId, UserAccount, UserCard } from '@/types'
import { initialOf } from '@/types'

interface AppShellProps {
  activePage: PageId
  isSidebarOpen: boolean
  user: UserAccount
  children: ReactNode
  onNavigate: (page: PageId) => void
  onToggleSidebar: () => void
  onLogout: () => void
  cards: UserCard[]
  selectedProfileCardID?: number
  onProfileCardSelected: (cardID: number) => void
}

interface NavigationItem {
  id: PageId
  label: string
  icon: ComponentType<{ size?: number }>
  adminOnly?: boolean
}

const navigation: NavigationItem[] = [
  { id: 'home', label: '主页', icon: Home },
  { id: 'dashboard', label: '概览', icon: LayoutDashboard },
  { id: 'maimai', label: 'maimai DX', icon: Gamepad2 },
  { id: 'setup', label: '接入指南', icon: Server },
  { id: 'admin', label: '管理后台', icon: Users, adminOnly: true },
  { id: 'settings', label: '设置', icon: Settings },
]

export function AppShell({
  activePage,
  isSidebarOpen,
  user,
  children,
  onNavigate,
  onToggleSidebar,
  onLogout,
  cards,
  selectedProfileCardID,
  onProfileCardSelected,
}: AppShellProps) {
  const pageTitle = navigation.find((item) => item.id === activePage)?.label || activePage
  const [isProfilePickerOpen, setIsProfilePickerOpen] = useState(false)
  const profileCards = cards.filter((card) => card.gameUserId > 0)
  const selectedCard = profileCards.find((card) => card.ID === selectedProfileCardID) || profileCards[0]

  const chooseProfileCard = (cardID: number) => {
    onProfileCardSelected(cardID)
    setIsProfilePickerOpen(false)
  }

  return (
    <div className="h-[100dvh] overflow-hidden bg-slate-950 text-slate-50 md:flex">
      {isSidebarOpen && <button type="button" aria-label="关闭侧边栏" onClick={onToggleSidebar} className="fixed inset-0 z-40 bg-black/60 md:hidden" />}
      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-slate-800 bg-slate-900 transition-transform duration-300 ${isSidebarOpen ? 'translate-x-0 md:w-64' : '-translate-x-full md:w-20 md:translate-x-0'} md:sticky md:top-0 md:h-[100dvh] md:shrink-0 md:transition-[width]`}
      >
        <div className="p-6 flex items-center justify-between gap-3">
          <span className={`font-black text-2xl text-indigo-500 tracking-tighter ${!isSidebarOpen ? 'hidden' : ''}`}>
            MaiGoDX
          </span>
          <button
            type="button"
            aria-label="切换侧边栏"
            onClick={onToggleSidebar}
            className="text-slate-400 hover:text-white transition-colors"
          >
            {isSidebarOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>

        <nav aria-label="Main navigation" className="flex-1 space-y-2 overflow-y-auto px-3">
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

        <div className="space-y-2 border-t border-slate-800 p-4">
          <button
            type="button"
            disabled={!profileCards.length}
            onClick={() => setIsProfilePickerOpen(true)}
            title={selectedCard ? `${selectedCard.cardName || '未命名卡片'} · 档案 #${selectedCard.gameUserId}` : '暂无 maimai 档案'}
            className="flex w-full items-center gap-3 rounded-lg px-4 py-3 text-left text-cyan-300 transition-all hover:bg-cyan-500/10 disabled:cursor-not-allowed disabled:opacity-40"
          >
            <CreditCard size={20} />
            <span className={`min-w-0 font-medium ${!isSidebarOpen ? 'hidden' : ''}`}>
              <span className="block truncate">{selectedCard?.cardName || '选择档案'}</span>
              {selectedCard && <span className="block text-xs text-slate-500">档案 #{selectedCard.gameUserId}</span>}
            </span>
          </button>
          <button
            type="button"
            onClick={onLogout}
            className="flex items-center gap-3 px-4 py-3 w-full text-rose-400 hover:bg-rose-500/10 rounded-lg transition-all"
          >
            <LogOut size={20} />
            <span className={`font-medium ${!isSidebarOpen ? 'hidden' : ''}`}>退出登录</span>
          </button>
        </div>
      </aside>

      <main className="h-[100dvh] min-w-0 flex-1 overflow-y-auto">
        <header className="sticky top-0 z-40 flex h-16 items-center justify-between border-b border-slate-800 bg-slate-950/50 px-4 backdrop-blur-xl md:px-8">
          <div className="flex items-center gap-3">
            <button type="button" aria-label="打开侧边栏" onClick={onToggleSidebar} className="text-slate-400 hover:text-white md:hidden"><Menu size={22} /></button>
            <h1 className="text-lg font-bold">{pageTitle}</h1>
          </div>
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

      {isProfilePickerOpen && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center p-4" role="dialog" aria-modal="true" aria-label="选择 maimai 档案">
          <button type="button" aria-label="关闭档案选择" onClick={() => setIsProfilePickerOpen(false)} className="absolute inset-0 bg-black/70" />
          <section className="relative z-10 w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-5 shadow-2xl">
            <div className="mb-4 flex items-center justify-between gap-4">
              <div><h2 className="font-bold text-white">选择 Aime 档案</h2><p className="mt-1 text-xs text-slate-400">每张卡对应独立的 maimai 存档。</p></div>
              <button type="button" onClick={() => setIsProfilePickerOpen(false)} className="text-slate-400 hover:text-white"><X size={20} /></button>
            </div>
            <div className="space-y-2">
              {profileCards.map((card) => (
                <button key={card.ID} type="button" onClick={() => chooseProfileCard(card.ID)} className={`w-full rounded-lg border p-4 text-left transition-colors ${card.ID === selectedCard?.ID ? 'border-indigo-500 bg-indigo-500/15' : 'border-slate-700 hover:border-slate-500 hover:bg-slate-800'}`}>
                  <p className="font-medium text-white">{card.cardName || '未命名卡片'}</p>
                  <p className="mt-1 text-xs text-slate-400">maimai 档案 #{card.gameUserId}</p>
                </button>
              ))}
            </div>
          </section>
        </div>
      )}
    </div>
  )
}

export const pageIcons = { Home, LayoutDashboard, Gamepad2, Server, BookOpen, Settings, Users }
