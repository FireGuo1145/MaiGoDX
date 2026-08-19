import { useState, type ComponentType, type ReactNode } from "react"
import {
  BookOpen,
  CreditCard,
  Gamepad2,
  Home,
  LogOut,
  Menu,
  Server,
  Settings,
  Users,
  X,
} from "lucide-react"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import type { PageId, UserAccount, UserCard } from "@/types"
import { initialOf } from "@/types"

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
  { id: "home", label: "主页", icon: Home },
  { id: "maimai", label: "maimai DX", icon: Gamepad2 },
  { id: "chunithm", label: "CHUNITHM", icon: Gamepad2 },
  { id: "setup", label: "接入指南", icon: Server },
  { id: "admin", label: "管理后台", icon: Users, adminOnly: true },
  { id: "settings", label: "设置", icon: Settings },
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
  const pageTitle =
    navigation.find((item) => item.id === activePage)?.label || activePage
  const [isProfilePickerOpen, setIsProfilePickerOpen] = useState(false)
  const profileCards = cards.filter((card) => card.gameUserId > 0)
  const selectedCard =
    profileCards.find((card) => card.ID === selectedProfileCardID) ||
    profileCards[0]

  const chooseProfileCard = (cardID: number) => {
    onProfileCardSelected(cardID)
    setIsProfilePickerOpen(false)
  }

  return (
    <SidebarProvider
      open={isSidebarOpen}
      onOpenChange={(open) => {
        if (open !== isSidebarOpen) onToggleSidebar()
      }}
    >
      <Sidebar collapsible="icon">
        <SidebarHeader>
          <div className="flex items-center justify-between gap-2 px-2 py-1">
            <span className="truncate text-lg font-bold group-data-[collapsible=icon]:hidden">
              MaiGoDX
            </span>
            <SidebarTrigger aria-label="切换侧边栏">
              <Menu />
            </SidebarTrigger>
          </div>
        </SidebarHeader>

        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent>
              <SidebarMenu>
                {navigation.map(({ id, label, icon: Icon, adminOnly }) => {
                  if (adminOnly && !user.isAdmin) return null
                  return (
                    <SidebarMenuItem key={id}>
                      <SidebarMenuButton
                        isActive={activePage === id}
                        tooltip={label}
                        onClick={() => onNavigate(id)}
                      >
                        <Icon />
                        <span>{label}</span>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>

        <SidebarFooter>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton
                isDisabled={!profileCards.length}
                tooltip={
                  selectedCard
                    ? `${selectedCard.cardName || "未命名卡片"} · 档案 #${selectedCard.gameUserId}`
                    : "暂无 maimai 档案"
                }
                onClick={() => setIsProfilePickerOpen(true)}
              >
                <CreditCard />
                <span>{selectedCard?.cardName || "选择档案"}</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <SidebarMenuButton tooltip="退出登录" onClick={onLogout}>
                <LogOut />
                <span>退出登录</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarFooter>
      </Sidebar>

      <SidebarInset className="min-h-svh bg-background">
        <header className="sticky top-0 z-40 flex h-16 items-center justify-between border-b bg-background/80 px-4 backdrop-blur-xl md:px-8">
          <div className="flex items-center gap-3">
            <SidebarTrigger className="md:hidden" aria-label="打开侧边栏">
              <Menu />
            </SidebarTrigger>
            <h1 className="text-lg font-bold">{pageTitle}</h1>
          </div>
          <div className="flex items-center gap-4">
            <div className="hidden text-right sm:block">
              <p className="text-sm font-bold">{user.username}</p>
              <p className="text-[10px] text-muted-foreground">{user.email}</p>
            </div>
            <Avatar className="h-9 w-9">
              <AvatarFallback>{initialOf(user.username)}</AvatarFallback>
            </Avatar>
          </div>
        </header>
        <main className="min-w-0 flex-1 p-8">{children}</main>
      </SidebarInset>

      {isProfilePickerOpen && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center p-4"
          role="dialog"
          aria-modal="true"
          aria-label="选择 maimai 档案"
        >
          <button
            type="button"
            aria-label="关闭档案选择"
            onClick={() => setIsProfilePickerOpen(false)}
            className="absolute inset-0 bg-black/70"
          />
          <section className="relative z-10 w-full max-w-md rounded-xl border bg-card p-5 shadow-2xl">
            <div className="mb-4 flex items-center justify-between gap-4">
              <div>
                <h2 className="font-bold">选择 Aime 档案</h2>
                <p className="mt-1 text-xs text-muted-foreground">
                  每张卡对应独立的 maimai 存档。
                </p>
              </div>
              <button
                type="button"
                onClick={() => setIsProfilePickerOpen(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X size={20} />
              </button>
            </div>
            <div className="space-y-2">
              {profileCards.map((card) => (
                <button
                  key={card.ID}
                  type="button"
                  onClick={() => chooseProfileCard(card.ID)}
                  className={`w-full rounded-lg border p-4 text-left transition-colors ${card.ID === selectedCard?.ID ? "border-primary bg-muted" : "hover:bg-muted"}`}
                >
                  <p className="font-medium">{card.cardName || "未命名卡片"}</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    maimai 档案 #{card.gameUserId}
                  </p>
                </button>
              ))}
            </div>
          </section>
        </div>
      )}
    </SidebarProvider>
  )
}

export const pageIcons = {
  Home,
  Gamepad2,
  Server,
  BookOpen,
  Settings,
  Users,
}
