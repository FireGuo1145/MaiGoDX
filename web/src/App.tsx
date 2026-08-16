import { useEffect, useState } from "react"
import {
  Navigate,
  Route,
  Routes,
  useLocation,
  useNavigate,
} from "react-router-dom"
import { AppShell } from "@/components/layout/AppShell"
import api from "@/lib/api"
import { AdminPage } from "@/pages/AdminPage"
import { AuthPage } from "@/pages/AuthPage"
import { DashboardPage } from "@/pages/DashboardPage"
import { HomePage } from "@/pages/HomePage"
import { MaimaiPage } from "@/pages/MaimaiPage"
import { SettingsPage } from "@/pages/SettingsPage"
import { SetupPage } from "@/pages/SetupPage"
import type {
  GameCharge,
  GameEvent,
  LoginResult,
  MetadataItem,
  PageId,
  Stats,
  SystemConfig,
  Terminal,
  UserAccount,
  UserCard,
} from "@/types"
import { apiErrorMessage } from "@/types"

const pageTitles: Record<PageId, string> = {
  home: "主页",
  dashboard: "概览",
  maimai: "maimai DX",
  setup: "接入指南",
  admin: "管理后台",
  settings: "设置",
}

const pagePaths: Record<PageId, string> = {
  home: "/",
  dashboard: "/dashboard",
  maimai: "/maimai",
  setup: "/setup",
  admin: "/admin",
  settings: "/settings",
}

function pageFromPath(pathname: string): PageId {
  const match = (Object.entries(pagePaths) as [PageId, string][]).find(
    ([, path]) => path === pathname
  )
  return match?.[0] ?? "home"
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
  const [isSidebarOpen, setIsSidebarOpen] = useState(
    () => typeof window !== "undefined" && window.innerWidth >= 768
  )
  const [stats, setStats] = useState<Stats | null>(null)
  const [selectedProfileCardID, setSelectedProfileCardID] = useState<
    number | undefined
  >()
  const [cards, setCards] = useState<UserCard[]>([])
  const [users, setUsers] = useState<UserAccount[]>([])
  const [configs, setConfigs] = useState<SystemConfig[]>([])
  const [siteName, setSiteName] = useState("MaiGoDX")
  const [metadata, setMetadata] = useState<Record<string, MetadataItem[]>>({})
  const [terminals, setTerminals] = useState<Terminal[]>([])
  const [events, setEvents] = useState<GameEvent[]>([])
  const [charges, setCharges] = useState<GameCharge[]>([])
  const location = useLocation()
  const navigate = useNavigate()
  const page = pageFromPath(location.pathname)

  const refreshMetadata = async () => {
    try {
      const result = await api.getSiteMetadata()
      if (result.success) setMetadata(result.metadata || {})
    } catch (error) {
      console.error("Failed to refresh site metadata:", apiErrorMessage(error))
    }
  }
  const refreshSiteSettings = async () => {
    try {
      const result = await api.getSiteSettings()
      if (result.success) setSiteName(result.siteName?.trim() || "MaiGoDX")
    } catch (error) {
      console.error(
        "Failed to load public site settings:",
        apiErrorMessage(error)
      )
    }
  }

  useEffect(() => {
    void refreshSiteSettings()
    void refreshMetadata()
    let active = true
    void api
      .currentUser()
      .then((result) => {
        if (active && result.success) setUser(accountFromSession(result))
      })
      .catch(() => undefined)
      .finally(() => {
        if (active) setIsAuthReady(true)
      })
    return () => {
      active = false
    }
  }, [])

  useEffect(() => {
    document.title = `${user ? pageTitles[page] : "登录"} - ${siteName}`
  }, [page, siteName, user])

  const refreshStats = async (cardID = selectedProfileCardID) => {
    if (!user) return
    try {
      const result = await api.getStats(cardID)
      if (result.success) {
        setStats(result)
        if (!cardID && result.selectedCardId)
          setSelectedProfileCardID(result.selectedCardId)
      }
    } catch (error) {
      console.error("Failed to load stats:", apiErrorMessage(error))
    }
  }

  const refreshCards = async () => {
    if (!user) return
    try {
      const result = await api.getCards()
      if (result.success) setCards(result.cards || [])
    } catch (error) {
      console.error("Failed to load cards:", apiErrorMessage(error))
    }
  }

  const refreshUsers = async () => {
    try {
      const result = await api.getUsers()
      if (result.success) setUsers(result.users || [])
    } catch (error) {
      console.error("Failed to load users:", apiErrorMessage(error))
    }
  }

  const refreshTerminals = async () => {
    try {
      const result = await api.getTerminals()
      if (result.success) setTerminals(result.terminals || [])
    } catch (error) {
      console.error("加载机台列表失败：", apiErrorMessage(error))
    }
  }

  const refreshEvents = async () => {
    try {
      const result = await api.getGameEvents()
      if (result.success) setEvents(result.events || [])
    } catch (error) {
      console.error("加载游戏事件失败：", apiErrorMessage(error))
    }
  }
  const refreshCharges = async () => {
    try {
      const result = await api.getGameCharges()
      if (result.success) setCharges(result.charges || [])
    } catch (error) {
      console.error("加载收费项目失败：", apiErrorMessage(error))
    }
  }
  const refreshConfigs = async () => {
    try {
      const result = await api.getConfigs()
      if (result.success) setConfigs(result.configs || [])
    } catch (error) {
      console.error("Failed to load system configs:", apiErrorMessage(error))
    }
  }

  const refreshAdminConfigs = async () => {
    await Promise.all([refreshConfigs(), refreshSiteSettings()])
  }

  useEffect(() => {
    if (!user) return
    void refreshStats()
    void refreshCards()
    if (user.isAdmin) {
      void refreshUsers()
      void refreshConfigs()
      void refreshTerminals()
      void refreshEvents()
      void refreshCharges()
    }
  }, [user])

  const logout = () => {
    void api.logout()
    setUser(null)
    setStats(null)
    setSelectedProfileCardID(undefined)
    setCards([])
    setUsers([])
    setConfigs([])
    setTerminals([])
    setEvents([])
    setCharges([])
    navigate("/", { replace: true })
  }

  if (!isAuthReady) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background text-muted-foreground">
        正在恢复登录会话…
      </div>
    )
  }

  if (!user) {
    return (
      <AuthPage
        siteName={siteName}
        onAuthenticated={(account) => {
          setUser(account)
          navigate("/dashboard", { replace: true })
        }}
      />
    )
  }

  return (
    <AppShell
      user={user}
      activePage={page}
      isSidebarOpen={isSidebarOpen}
      onNavigate={(target) => navigate(pagePaths[target])}
      onToggleSidebar={() => setIsSidebarOpen((open) => !open)}
      onLogout={logout}
      cards={cards}
      selectedProfileCardID={selectedProfileCardID || stats?.selectedCardId}
      onProfileCardSelected={(cardID) => {
        setSelectedProfileCardID(cardID)
        void refreshStats(cardID)
      }}
    >
      <Routes>
        <Route path="/" element={<HomePage stats={stats} />} />
        <Route
          path="/dashboard"
          element={<DashboardPage stats={stats} metadata={metadata} />}
        />
        <Route
          path="/maimai"
          element={
            <MaimaiPage
              stats={stats}
              metadata={metadata}
              onProfileChanged={() => refreshStats(selectedProfileCardID)}
            />
          }
        />
        <Route path="/setup" element={<SetupPage />} />
        <Route
          path="/settings"
          element={
            <SettingsPage
              user={user}
              cards={cards}
              onCardsChanged={refreshCards}
            />
          }
        />
        <Route
          path="/admin"
          element={
            user.isAdmin ? (
              <AdminPage
                users={users}
                configs={configs}
                terminals={terminals}
                events={events}
                charges={charges}
                onUsersChanged={refreshUsers}
                onMetadataChanged={refreshMetadata}
                onConfigsChanged={refreshAdminConfigs}
                onTerminalsChanged={refreshTerminals}
                onEventsChanged={refreshEvents}
                onChargesChanged={refreshCharges}
              />
            ) : (
              <Navigate to="/" replace />
            )
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppShell>
  )
}
