import { useEffect, useState } from 'react'
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import api from '@/lib/api'
import { AdminPage } from '@/pages/AdminPage'
import { AuthPage } from '@/pages/AuthPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { HomePage } from '@/pages/HomePage'
import { MaimaiPage } from '@/pages/MaimaiPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { SetupPage } from '@/pages/SetupPage'
import type { LoginResult, PageId, Stats, SystemConfig, UserAccount, UserCard } from '@/types'
import { apiErrorMessage } from '@/types'

const pagePaths: Record<PageId, string> = {
  home: '/',
  dashboard: '/dashboard',
  maimai: '/maimai',
  setup: '/setup',
  admin: '/admin',
  settings: '/settings',
}

function pageFromPath(pathname: string): PageId {
  const match = (Object.entries(pagePaths) as [PageId, string][]).find(([, path]) => path === pathname)
  return match?.[0] ?? 'home'
}

function accountFromSession(result: LoginResult): UserAccount {
  return {
    ID: 0,
    email: result.email,
    username: result.username,
    isVerified: true,
    isAdmin: Boolean(result.isAdmin),
  }
}

export default function App() {
  const [user, setUser] = useState<UserAccount | null>(null)
  const [isAuthReady, setIsAuthReady] = useState(false)
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [stats, setStats] = useState<Stats | null>(null)
  const [cards, setCards] = useState<UserCard[]>([])
  const [users, setUsers] = useState<UserAccount[]>([])
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  const location = useLocation()
  const navigate = useNavigate()
  const page = pageFromPath(location.pathname)

  useEffect(() => {
    let active = true
    void api.currentUser()
      .then((result) => {
        if (active && result.success) setUser(accountFromSession(result))
      })
      .catch(() => undefined)
      .finally(() => {
        if (active) setIsAuthReady(true)
      })
    return () => { active = false }
  }, [])

  const refreshStats = async () => {
    if (!user) return
    try {
      const result = await api.getStats()
      if (result.success) setStats(result)
    } catch (error) {
      console.error('Failed to load stats:', apiErrorMessage(error))
    }
  }

  const refreshCards = async () => {
    if (!user) return
    try {
      const result = await api.getCards()
      if (result.success) setCards(result.cards || [])
    } catch (error) {
      console.error('Failed to load cards:', apiErrorMessage(error))
    }
  }

  const refreshUsers = async () => {
    try {
      const result = await api.getUsers()
      if (result.success) setUsers(result.users || [])
    } catch (error) {
      console.error('Failed to load users:', apiErrorMessage(error))
    }
  }

  const refreshConfigs = async () => {
    try {
      const result = await api.getConfigs()
      if (result.success) setConfigs(result.configs || [])
    } catch (error) {
      console.error('Failed to load system configs:', apiErrorMessage(error))
    }
  }

  useEffect(() => {
    if (!user) return
    void refreshStats()
    void refreshCards()
    if (user.isAdmin) {
      void refreshUsers()
      void refreshConfigs()
    }
  }, [user])

  const logout = () => {
    void api.logout()
    setUser(null)
    setStats(null)
    setCards([])
    setUsers([])
    setConfigs([])
    navigate('/', { replace: true })
  }

  if (!isAuthReady) {
    return <div className="min-h-screen bg-slate-950 text-slate-400 flex items-center justify-center">正在恢复登录会话…</div>
  }

  if (!user) {
    return <AuthPage onAuthenticated={(account) => { setUser(account); navigate('/dashboard', { replace: true }) }} />
  }

  return (
    <AppShell
      user={user}
      activePage={page}
      isSidebarOpen={isSidebarOpen}
      onNavigate={(target) => navigate(pagePaths[target])}
      onToggleSidebar={() => setIsSidebarOpen((open) => !open)}
      onLogout={logout}
    >
      <Routes>
        <Route path="/" element={<HomePage stats={stats} />} />
        <Route path="/dashboard" element={<DashboardPage stats={stats} />} />
        <Route path="/maimai" element={<MaimaiPage stats={stats} />} />
        <Route path="/setup" element={<SetupPage />} />
        <Route path="/settings" element={<SettingsPage user={user} cards={cards} onCardsChanged={refreshCards} />} />
        <Route
          path="/admin"
          element={user.isAdmin
            ? <AdminPage users={users} configs={configs} onUsersChanged={refreshUsers} onConfigsChanged={refreshConfigs} />
            : <Navigate to="/" replace />}
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppShell>
  )
}
