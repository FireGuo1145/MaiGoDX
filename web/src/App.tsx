import { useEffect, useState } from 'react'
import { AppShell } from '@/components/layout/AppShell'
import api from '@/lib/api'
import { AdminPage } from '@/pages/AdminPage'
import { AuthPage } from '@/pages/AuthPage'
import { DashboardPage } from '@/pages/DashboardPage'
import { HomePage } from '@/pages/HomePage'
import { MaimaiPage } from '@/pages/MaimaiPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { SetupPage } from '@/pages/SetupPage'
import type { PageId, Stats, SystemConfig, UserAccount, UserCard } from '@/types'
import { apiErrorMessage } from '@/types'

export default function App() {
  const [user, setUser] = useState<UserAccount | null>(null)
  const [page, setPage] = useState<PageId>('home')
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [stats, setStats] = useState<Stats | null>(null)
  const [cards, setCards] = useState<UserCard[]>([])
  const [users, setUsers] = useState<UserAccount[]>([])
  const [configs, setConfigs] = useState<SystemConfig[]>([])

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
    setPage('home')
    setStats(null)
    setCards([])
    setUsers([])
    setConfigs([])
  }

  if (!user) {
    return <AuthPage onAuthenticated={(account) => { setUser(account); setPage('dashboard') }} />
  }

  const content = (() => {
    switch (page) {
      case 'home':
        return <HomePage stats={stats} />
      case 'dashboard':
        return <DashboardPage stats={stats} />
      case 'maimai':
        return <MaimaiPage stats={stats} />
      case 'setup':
        return <SetupPage />
      case 'admin':
        return user.isAdmin
          ? <AdminPage users={users} configs={configs} onUsersChanged={refreshUsers} onConfigsChanged={refreshConfigs} />
          : <HomePage stats={stats} />
      case 'settings':
        return <SettingsPage user={user} cards={cards} onCardsChanged={refreshCards} />
      default:
        return <HomePage stats={stats} />
    }
  })()

  return (
    <AppShell
      user={user}
      activePage={page}
      isSidebarOpen={isSidebarOpen}
      onNavigate={setPage}
      onToggleSidebar={() => setIsSidebarOpen((open) => !open)}
      onLogout={logout}
    >
      {content}
    </AppShell>
  )
}
