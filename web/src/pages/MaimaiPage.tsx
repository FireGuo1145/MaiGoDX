import { Textarea } from "@/components/ui/textarea"
import { Input } from "@/components/ui/input"
import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import api from "@/lib/api"
import { DashboardPage } from "@/pages/DashboardPage"
import {
  apiErrorMessage,
  RANK_SUMMARY,
  type FunctionTicket,
  type MetadataItem,
  type Region,
  type Stats,
  type TravelPartner,
} from "@/types"

interface MaimaiPageProps {
  stats: Stats | null
  metadata: Record<string, MetadataItem[]>
  onProfileChanged: () => Promise<void>
}

export function MaimaiPage({
  stats,
  metadata,
  onProfileChanged,
}: MaimaiPageProps) {
  const plays = stats?.recentPlays || []
  const travelPartners = stats?.travelPartners || []
  const functionTickets = stats?.functionTickets || []
  const regions = stats?.regions || []
  const metadataName = (kind: string, id: number) =>
    metadata[kind]?.find((item) => item.id === id)?.name || `ID ${id}`
  const [isEditingProfile, setIsEditingProfile] = useState(false)
  const [partnerID, setPartnerID] = useState("0")
  const [maimile, setMaimile] = useState("0")
  const [travelPartnerText, setTravelPartnerText] = useState("")
  const [ticketText, setTicketText] = useState("")
  const [regionText, setRegionText] = useState("")
  const [profileError, setProfileError] = useState("")
  const [isSavingProfile, setIsSavingProfile] = useState(false)
  const [ticketItemID, setTicketItemID] = useState("11001")
  const [ticketGrantError, setTicketGrantError] = useState("")
  const [isGrantingTicket, setIsGrantingTicket] = useState(false)
  const selectedCardID = stats?.selectedCardId || 0

  const startProfileEdit = () => {
    setPartnerID(String(stats?.partner?.partnerId || 0))
    setMaimile(String(stats?.user?.totalPoint || 0))
    setTravelPartnerText(
      travelPartners
        .map(
          (partner) =>
            `${partner.partnerId}, ${partner.intimateLevel}, ${partner.intimateCountRewarded}`
        )
        .join("\n")
    )
    setTicketText(
      functionTickets
        .map((ticket) => `${ticket.itemId}, ${ticket.stock}`)
        .join("\n")
    )
    setRegionText(
      regions
        .map((region) => `${region.regionId}, ${region.playCount}`)
        .join("\n")
    )
    setProfileError("")
    setIsEditingProfile(true)
  }

  const saveProfile = async () => {
    try {
      setIsSavingProfile(true)
      setProfileError("")
      await api.updateProfile({
        cardId: selectedCardID,
        partnerId: parseInteger(partnerID, "搭档 ID"),
        maimile: parseInteger(maimile, "maimile 数量"),
        travelPartners: parseTravelPartners(travelPartnerText),
        functionTickets: parseFunctionTickets(ticketText),
        regions: parseRegions(regionText),
      })
      await onProfileChanged()
      setIsEditingProfile(false)
    } catch (error) {
      setProfileError(apiErrorMessage(error))
    } finally {
      setIsSavingProfile(false)
    }
  }

  const grantTicket = async (amount: number) => {
    try {
      setIsGrantingTicket(true)
      setTicketGrantError("")
      await api.adjustFunctionTicket({
        cardId: selectedCardID,
        itemId: parseInteger(ticketItemID, "功能票 ID"),
        amount,
      })
      await onProfileChanged()
    } catch (error) {
      setTicketGrantError(apiErrorMessage(error))
    } finally {
      setIsGrantingTicket(false)
    }
  }

  return (
    <div className="space-y-6">
      <Tabs defaultSelectedKey="overview" className="w-full">
        <TabsList className="border border-border bg-card">
          <TabsTrigger id="overview" className="data-[selected]:bg-neutral-600">
            概览
          </TabsTrigger>
          <TabsTrigger id="recent" className="data-[selected]:bg-neutral-600">
            最近游玩
          </TabsTrigger>
          <TabsTrigger id="stats" className="data-[selected]:bg-neutral-600">
            成绩统计
          </TabsTrigger>
          <TabsTrigger id="profile" className="data-[selected]:bg-neutral-600">
            游戏档案
          </TabsTrigger>
        </TabsList>

        <TabsContent id="overview" className="mt-6">
          <DashboardPage stats={stats} metadata={metadata} />
        </TabsContent>

        <TabsContent id="recent" className="mt-6">
          <Card className="overflow-hidden border-border bg-card">
            <Table>
              <TableHeader className="bg-muted/50">
                <TableHead isRowHeader className="text-muted-foreground">
                  乐曲
                </TableHead>
                <TableHead className="text-muted-foreground">难度</TableHead>
                <TableHead className="text-right text-muted-foreground">
                  达成率
                </TableHead>
                <TableHead className="text-right text-muted-foreground">
                  分数
                </TableHead>
                <TableHead className="text-right text-muted-foreground">
                  日期
                </TableHead>
              </TableHeader>
              <TableBody>
                {plays.length ? (
                  plays.map((play) => (
                    <TableRow
                      key={play.ID}
                      className="border-border hover:bg-muted/30"
                    >
                      <TableCell className="font-bold text-foreground">
                        {metadataName("music", play.musicId)}{" "}
                        <span className="text-xs font-normal text-muted-foreground">
                          #{play.musicId}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant="outline"
                          className="border-neutral-500/50 text-neutral-400"
                        >
                          LV.{play.level}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right font-mono text-neutral-400">
                        {(play.achievement / 10000).toFixed(4)}%
                      </TableCell>
                      <TableCell className="text-right font-mono">
                        {play.score.toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right text-xs text-muted-foreground">
                        {play.createDate || "—"}
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell className="py-12 text-center text-muted-foreground">
                      暂无游玩记录。
                    </TableCell>
                    <TableCell />
                    <TableCell />
                    <TableCell />
                    <TableCell />
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </Card>
        </TabsContent>

        <TabsContent id="stats" className="mt-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-4">
            {RANK_SUMMARY.map((rank) => (
              <Card
                key={rank}
                className="border-border bg-card p-6 text-center"
              >
                <p className="mb-1 text-2xl font-black text-neutral-400">
                  {rank}
                </p>
                <p className="text-sm font-bold text-muted-foreground">
                  {stats?.rankCounts?.[rank] ?? 0} 首
                </p>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent id="profile" className="mt-6 space-y-6">
          <div className="flex justify-end">
            <Button
              variant={isEditingProfile ? "outline" : "default"}
              onPress={() =>
                isEditingProfile
                  ? setIsEditingProfile(false)
                  : startProfileEdit()
              }
            >
              {isEditingProfile ? "取消编辑" : "编辑档案"}
            </Button>
          </div>
          {isEditingProfile ? (
            <Card className="space-y-4 border-neutral-500/40 bg-card p-5">
              <div>
                <label className="mb-1 block text-sm font-medium text-muted-foreground">
                  当前搭档：{metadataName("partner", Number(partnerID))}（ID{" "}
                  {partnerID || "0"}）
                </label>
                <Input
                  value={partnerID}
                  onChange={(event) => setPartnerID(event.target.value)}
                  inputMode="numeric"
                  className="w-full rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-neutral-500"
                />
              </div>
              <div>
                <label className="mb-1 block text-sm font-medium text-muted-foreground">
                  maimile 数量
                </label>
                <Input
                  value={maimile}
                  onChange={(event) => setMaimile(event.target.value)}
                  inputMode="numeric"
                  className="w-full rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-neutral-500"
                />
                <p className="mt-1 text-xs text-muted-foreground">
                  对应游戏档案的 maimile（totalPoint），保存后在机台同步。
                </p>
              </div>
              <ProfileTextInput
                label="旅行伙伴"
                hint="每行：搭档 ID, 亲密度等级, 已领奖励次数"
                value={travelPartnerText}
                onChange={setTravelPartnerText}
              />
              <ProfileTextInput
                label="功能票"
                hint="每行：票种 ID, 库存；11001 = 1.5 倍区域前进票"
                value={ticketText}
                onChange={setTicketText}
              />
              <ProfileTextInput
                label="区域游玩记录"
                hint="每行：区域 ID, 游玩次数"
                value={regionText}
                onChange={setRegionText}
              />
              {profileError && (
                <p className="text-sm text-red-400">{profileError}</p>
              )}
              <div className="flex justify-end gap-3">
                <Button
                  variant="outline"
                  onPress={() => setIsEditingProfile(false)}
                >
                  取消
                </Button>
                <Button
                  onPress={() => void saveProfile()}
                  isDisabled={isSavingProfile}
                >
                  {isSavingProfile ? "保存中…" : "保存档案"}
                </Button>
              </div>
            </Card>
          ) : null}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <Card className="border-border bg-card p-5">
              <p className="text-sm font-medium text-muted-foreground">
                当前搭档
              </p>
              <p className="mt-2 text-3xl font-black text-neutral-400">
                {stats?.partner?.partnerId
                  ? `${metadataName("partner", stats.partner.partnerId)}（ID ${stats.partner.partnerId}）`
                  : "未装备"}
              </p>
              <p className="mt-2 text-xs text-muted-foreground">
                来自机台同步的 partnerId。
              </p>
            </Card>
            <Card className="border-border bg-card p-5">
              <p className="text-sm font-medium text-muted-foreground">
                功能票库存
              </p>
              <p className="mt-2 text-3xl font-black text-neutral-400">
                {functionTickets.reduce(
                  (total, ticket) => total + ticket.stock,
                  0
                )}
              </p>
              <p className="mt-2 text-xs text-muted-foreground">
                共 {functionTickets.length} 种功能票。
              </p>
            </Card>
          </div>

          <ProfileTable
            title="旅行伙伴"
            empty="尚无旅行伙伴数据。完成机台同步后会显示在这里。"
            headers={["旅行伙伴", "亲密度等级", "已领奖励次数"]}
            rows={travelPartners.map((partner) => [
              `${metadataName("chara", partner.partnerId)}（ID ${partner.partnerId}）`,
              partner.intimateLevel,
              partner.intimateCountRewarded,
            ])}
          />
          <ProfileTable
            title="功能票"
            empty="尚未持有功能票。"
            headers={["票种 ID", "名称", "库存"]}
            rows={functionTickets.map((ticket) => [
              ticket.itemId,
              metadataName("ticket", ticket.itemId),
              ticket.stock,
            ])}
          />
          <TicketGrantPanel
            itemID={ticketItemID}
            onItemIDChange={setTicketItemID}
            onGrant={grantTicket}
            error={ticketGrantError}
            isGranting={isGrantingTicket}
            metadataName={(id) => metadataName("ticket", id)}
          />
          <ProfileTable
            title="区域游玩记录"
            empty="尚无区域游玩记录。"
            headers={["区域 ID", "游玩次数"]}
            rows={regions.map((region) => [region.regionId, region.playCount])}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}

interface ProfileTextInputProps {
  label: string
  hint: string
  value: string
  onChange: (value: string) => void
}

function ProfileTextInput({
  label,
  hint,
  value,
  onChange,
}: ProfileTextInputProps) {
  return (
    <div>
      <label className="mb-1 block text-sm font-medium text-muted-foreground">
        {label}
      </label>
      <Textarea
        value={value}
        onChange={(event) => onChange(event.target.value)}
        placeholder={hint}
        rows={3}
        className="w-full rounded-md border border-border bg-background px-3 py-2 font-mono text-sm text-foreground outline-none focus:border-neutral-500"
      />
      <p className="mt-1 text-xs text-muted-foreground">
        {hint}。留空将清空该类数据。
      </p>
    </div>
  )
}

function parseInteger(value: string, label: string) {
  const parsed = Number(value.trim())
  if (!Number.isInteger(parsed) || parsed < 0)
    throw new Error(`${label} 必须是非负整数`)
  return parsed
}

function parseFunctionTickets(value: string): FunctionTicket[] {
  return parsePairs(value, "功能票").map(([itemId, stock]) => ({
    itemId,
    stock,
  }))
}

function parseRegions(value: string): Region[] {
  return parsePairs(value, "区域").map(([regionId, playCount]) => ({
    regionId,
    playCount,
  }))
}

function parsePairs(value: string, label: string) {
  return parseRows(value, 2, label)
}

function parseTravelPartners(value: string): TravelPartner[] {
  return parseRows(value, 3, "旅行伙伴").map(
    ([partnerId, intimateLevel, intimateCountRewarded]) => ({
      partnerId,
      intimateLevel,
      intimateCountRewarded,
    })
  )
}

function parseRows(value: string, width: number, label: string): number[][] {
  const rows = value
    .split("\n")
    .map((line) => line.trim())
    .filter(Boolean)
  const ids = new Set<number>()
  return rows.map((line, index) => {
    const fields = line.split(",").map((field) => Number(field.trim()))
    if (
      fields.length !== width ||
      fields.some((field) => !Number.isInteger(field) || field < 0) ||
      ids.has(fields[0])
    ) {
      throw new Error(`${label} 第 ${index + 1} 行格式无效或 ID 重复`)
    }
    ids.add(fields[0])
    return fields
  })
}

interface TicketGrantPanelProps {
  itemID: string
  onItemIDChange: (value: string) => void
  onGrant: (amount: number) => Promise<void>
  error: string
  isGranting: boolean
  metadataName: (id: number) => string
}

function TicketGrantPanel({
  itemID,
  onItemIDChange,
  onGrant,
  error,
  isGranting,
  metadataName,
}: TicketGrantPanelProps) {
  return (
    <Card className="space-y-3 border-neutral-500/30 bg-card p-5">
      <div>
        <p className="font-bold text-foreground">快速发放功能票</p>
        <p className="mt-1 text-xs text-muted-foreground">
          直接为当前选中的 Aime 档案增加指定票种库存。
        </p>
      </div>
      <div className="flex flex-wrap items-end gap-3">
        <label className="block text-sm text-muted-foreground">
          票种：{metadataName(Number(itemID))}（ID {itemID || "0"}）
          <Input
            value={itemID}
            onChange={(event) => onItemIDChange(event.target.value)}
            inputMode="numeric"
            className="mt-1 block w-28 rounded-md border border-border bg-background px-3 py-2 text-foreground outline-none focus:border-neutral-500"
          />
        </label>
        {[1, 5, 10].map((amount) => (
          <Button
            key={amount}
            variant="secondary"
            onPress={() => void onGrant(amount)}
            isDisabled={isGranting}
          >
            +{amount}
          </Button>
        ))}
      </div>
      {error && <p className="text-sm text-red-400">{error}</p>}
    </Card>
  )
}

interface ProfileTableProps {
  title: string
  headers: string[]
  rows: Array<Array<number | string>>
  empty: string
}

function ProfileTable({ title, headers, rows, empty }: ProfileTableProps) {
  return (
    <Card className="overflow-hidden border-border bg-card">
      <div className="border-b border-border px-5 py-4 text-sm font-bold text-foreground">
        {title}
      </div>
      <Table>
        <TableHeader className="bg-muted/50">
          {headers.map((header, index) => (
            <TableHead
              key={header}
              isRowHeader={index === 0}
              className="text-muted-foreground"
            >
              {header}
            </TableHead>
          ))}
        </TableHeader>
        <TableBody>
          {rows.length ? (
            rows.map((row, index) => (
              <TableRow key={`${title}-${index}`} className="border-border">
                {row.map((value, column) => (
                  <TableCell key={column} className="font-mono text-foreground">
                    {value}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell
                colSpan={headers.length}
                className="py-8 text-center text-sm text-muted-foreground"
              >
                {empty}
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </Card>
  )
}
