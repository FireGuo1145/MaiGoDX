import { Input } from "@/components/ui/input"
import { useState } from "react"
import { CreditCard } from "lucide-react"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import api from "@/lib/api"
import { PersonalTerminalPanel } from "@/components/PersonalTerminalPanel"
import type { UserAccount, UserCard } from "@/types"
import {
  apiErrorMessage,
  cardPreview,
  DEFAULT_CARD_NAME,
  initialOf,
  isAccessCodeValid,
  normalizeAccessCode,
} from "@/types"

interface SettingsPageProps {
  user: UserAccount
  cards: UserCard[]
  onCardsChanged: () => Promise<void>
}

export function SettingsPage({
  user,
  cards,
  onCardsChanged,
}: SettingsPageProps) {
  const [accessCode, setAccessCode] = useState("")
  const [cardName, setCardName] = useState("")
  const [gameUserId, setGameUserId] = useState("")
  const [notice, setNotice] = useState<string | null>(null)
  const [isBinding, setIsBinding] = useState(false)

  const bindCard = async (event: React.FormEvent) => {
    event.preventDefault()
    const normalizedCode = normalizeAccessCode(accessCode)

    if (!isAccessCodeValid(normalizedCode)) {
      setNotice("请输入有效的 20 位 Aime Access Code。")
      return
    }

    const parsedGameUserId =
      gameUserId.trim() === "" ? undefined : Number(gameUserId)
    if (
      parsedGameUserId !== undefined &&
      (!Number.isSafeInteger(parsedGameUserId) || parsedGameUserId <= 0)
    ) {
      setNotice("游戏用户 ID 必须是正整数；若尚未创建 maimai 档案可留空。")
      return
    }

    setIsBinding(true)
    setNotice(null)
    try {
      const result = await api.bindCard(
        user.email,
        normalizedCode,
        cardName.trim() || DEFAULT_CARD_NAME,
        parsedGameUserId
      )
      if (!result.success) throw new Error(result.message || "卡片绑定失败")
      setAccessCode("")
      setCardName("")
      setGameUserId("")
      setNotice("卡片绑定成功。")
      await onCardsChanged()
    } catch (error) {
      setNotice(apiErrorMessage(error))
    } finally {
      setIsBinding(false)
    }
  }

  return (
    <div className="w-full">
      <Tabs defaultSelectedKey="account" className="space-y-6">
        <TabsList className="w-full justify-start overflow-x-auto rounded-xl border border-border bg-card p-1">
          <TabsTrigger
            id="account"
            className="min-w-fit rounded-lg px-3 text-muted-foreground data-selected:text-foreground"
          >
            账户资料
          </TabsTrigger>
          <TabsTrigger
            id="cards"
            className="min-w-fit rounded-lg px-3 text-muted-foreground data-selected:text-foreground"
          >
            Aime 卡片
          </TabsTrigger>
          <TabsTrigger
            id="terminals"
            className="min-w-fit rounded-lg px-3 text-muted-foreground data-selected:text-foreground"
          >
            我的机台
          </TabsTrigger>
        </TabsList>

        <TabsContent id="account">
          <Card className="border-border bg-card">
            <CardHeader>
              <CardTitle className="text-foreground">账户设置</CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center gap-6">
                <Avatar className="h-20 w-20 border-2 border-border">
                  <AvatarFallback className="bg-neutral-600 text-2xl font-black">
                    {initialOf(user.username)}
                  </AvatarFallback>
                </Avatar>
                <Button isDisabled variant="outline" className="border-border">
                  修改头像
                </Button>
              </div>
              <Separator className="bg-muted" />
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <label className="space-y-2">
                  <span className="text-xs font-bold tracking-wider text-muted-foreground uppercase">
                    用户名
                  </span>
                  <Input
                    disabled
                    value={user.username}
                    className="h-10 w-full cursor-not-allowed rounded-md border-border bg-muted/50 px-3 text-muted-foreground"
                  />
                </label>
                <label className="space-y-2">
                  <span className="text-xs font-bold tracking-wider text-muted-foreground uppercase">
                    邮箱
                  </span>
                  <Input
                    disabled
                    value={user.email}
                    className="h-10 w-full cursor-not-allowed rounded-md border-border bg-muted/50 px-3 text-muted-foreground"
                  />
                </label>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent id="cards">
          <Card className="border-border bg-card">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-foreground">
                <CreditCard className="text-neutral-400" /> Aime 卡片绑定
              </CardTitle>
              <CardDescription className="text-muted-foreground">
                将已授权的 20 位 Aime Access Code 关联到当前账户。
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {notice && (
                <div className="rounded-lg border border-neutral-500/20 bg-neutral-500/10 p-3 text-sm text-neutral-300">
                  {notice}
                </div>
              )}
              <form onSubmit={bindCard} className="space-y-4">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <label className="space-y-2">
                    <span className="text-xs font-medium text-muted-foreground">
                      Access Code（20 位数字）
                    </span>
                    <Input
                      required
                      maxLength={20}
                      value={accessCode}
                      onChange={(event) =>
                        setAccessCode(normalizeAccessCode(event.target.value))
                      }
                      placeholder="01234567890123456789"
                      className="h-10 w-full rounded-md border border-border bg-muted px-3 text-sm text-foreground focus:ring-2 focus:ring-ring focus:outline-none"
                    />
                  </label>
                  <label className="space-y-2">
                    <span className="text-xs font-medium text-muted-foreground">
                      卡片别名
                    </span>
                    <Input
                      value={cardName}
                      onChange={(event) => setCardName(event.target.value)}
                      placeholder="我的主卡"
                      className="h-10 w-full rounded-md border border-border bg-muted px-3 text-sm text-foreground focus:ring-2 focus:ring-ring focus:outline-none"
                    />
                  </label>
                  <label className="space-y-2 sm:col-span-2">
                    <span className="text-xs font-medium text-muted-foreground">
                      maimai 游戏用户 ID（可选）
                    </span>
                    <Input
                      inputMode="numeric"
                      value={gameUserId}
                      onChange={(event) =>
                        setGameUserId(event.target.value.replace(/\D/g, ""))
                      }
                      placeholder="在游戏创建档案后填写外部用户 ID"
                      className="h-10 w-full rounded-md border border-border bg-muted px-3 text-sm text-foreground focus:ring-2 focus:ring-ring focus:outline-none"
                    />
                    <span className="block text-xs text-muted-foreground">
                      游戏创建档案前可留空；之后补充该
                      ID，即可在门户显示该档案的真实成绩。
                    </span>
                  </label>
                </div>
                <Button
                  isDisabled={isBinding}
                  type="submit"
                  className="bg-neutral-600 font-bold text-foreground hover:bg-neutral-500"
                >
                  {isBinding ? "正在绑定…" : "绑定卡片"}
                </Button>
              </form>

              <Separator className="bg-muted" />
              <section className="space-y-3">
                <h3 className="text-sm font-bold text-foreground">
                  已绑定的卡片
                </h3>
                {cards.length ? (
                  cards.map((card) => (
                    <div
                      key={card.ID}
                      className="flex items-center justify-between gap-4 rounded-lg bg-muted/60 p-3"
                    >
                      <div>
                        <p className="text-sm font-bold text-foreground">
                          {card.cardName || DEFAULT_CARD_NAME}
                        </p>
                        <p className="font-mono text-xs text-neutral-400">
                          {cardPreview(card.accessCode)}
                        </p>
                        <p className="font-mono text-xs text-muted-foreground">
                          游戏用户 ID：{card.gameUserId || "未关联"}
                        </p>
                      </div>
                      <Badge className="border-neutral-500/20 bg-neutral-500/10 text-neutral-500">
                        已绑定
                      </Badge>
                    </div>
                  ))
                ) : (
                  <p className="text-sm text-muted-foreground">
                    暂未绑定任何卡片。
                  </p>
                )}
              </section>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent id="terminals">
          <PersonalTerminalPanel />
        </TabsContent>
      </Tabs>
    </div>
  )
}
