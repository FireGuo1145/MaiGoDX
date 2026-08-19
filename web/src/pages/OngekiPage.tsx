import { Activity, Coins, Gauge, Music2, UserRound } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import type { MetadataItem, OngekiStats } from "@/types"

type OngekiRecord = Record<string, unknown>

export function OngekiPage({
  stats,
  metadata,
}: {
  stats: OngekiStats | null
  metadata: Record<string, MetadataItem[]>
}) {
  const profile = stats?.profile || {}
  const plays = stats?.recentPlays || []
  const details = stats?.musicDetails || []
  const musicName = (record: OngekiRecord) => {
    const musicID = Number(rawValue(record, "musicId"))
    return (
      metadata.ongeki_music?.find((item) => item.id === musicID)?.name ||
      String(musicID)
    )
  }

  return (
    <div className="space-y-6">
      {stats?.message && (
        <p className="rounded-md border border-border bg-card p-4 text-sm text-muted-foreground">
          {stats.message}
        </p>
      )}

      <Tabs defaultSelectedKey="overview" className="w-full">
        <TabsList className="w-full justify-start overflow-x-auto rounded-xl border border-border bg-card p-1">
          <TabsTrigger
            id="overview"
            className="min-w-fit rounded-lg px-3 data-selected:bg-neutral-600"
          >
            概览
          </TabsTrigger>
          <TabsTrigger
            id="recent"
            className="min-w-fit rounded-lg px-3 data-selected:bg-neutral-600"
          >
            最近游玩
          </TabsTrigger>
          <TabsTrigger
            id="scores"
            className="min-w-fit rounded-lg px-3 data-selected:bg-neutral-600"
          >
            成绩明细
          </TabsTrigger>
        </TabsList>

        <TabsContent id="overview" className="mt-6 space-y-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-5">
            <SummaryCard
              icon={<UserRound />}
              label="玩家名称"
              value={displayValue(profile, "userName")}
            />
            <SummaryCard
              icon={<Gauge />}
              label="等级 · level"
              value={displayNumber(profile, "level")}
            />
            <SummaryCard
              icon={<Activity />}
              label="Rating · newPlayerRating"
              value={displayPreferredNumber(profile, [
                "newPlayerRating",
                "playerRating",
              ])}
            />
            <SummaryCard
              icon={<Music2 />}
              label="游玩次数 · playCount"
              value={displayNumber(profile, "playCount")}
            />
            <SummaryCard
              icon={<Coins />}
              label="当前点数 · point"
              value={displayNumber(profile, "point")}
            />
          </div>

          {stats?.profile && (
            <Card className="border-border bg-card">
              <CardHeader>
                <CardTitle>档案数据</CardTitle>
              </CardHeader>
              <CardContent className="grid grid-cols-1 gap-x-8 gap-y-4 sm:grid-cols-2 lg:grid-cols-4">
                <ProfileField
                  label="经验值"
                  field="exp"
                  value={displayNumber(profile, "exp")}
                />
                <ProfileField
                  label="当前点数"
                  field="point"
                  value={displayNumber(profile, "point")}
                />
                <ProfileField
                  label="累计点数"
                  field="totalPoint"
                  value={displayNumber(profile, "totalPoint")}
                />
                <ProfileField
                  label="宝石"
                  field="jewelCount"
                  value={displayNumber(profile, "jewelCount")}
                />
                <ProfileField
                  label="累计宝石"
                  field="totalJewelCount"
                  value={displayNumber(profile, "totalJewelCount")}
                />
                <ProfileField
                  label="Battle Point"
                  field="battlePoint"
                  value={displayNumber(profile, "battlePoint")}
                />
                <ProfileField
                  label="最高 Battle Point"
                  field="bestBattlePoint"
                  value={displayNumber(profile, "bestBattlePoint")}
                />
                <ProfileField
                  label="轮回次数"
                  field="reincarnationNum"
                  value={displayNumber(profile, "reincarnationNum")}
                />
                <ProfileField
                  label="历史最高 Rating"
                  field="newHighestRating"
                  value={displayPreferredNumber(profile, [
                    "newHighestRating",
                    "highestRating",
                  ])}
                />
                <ProfileField
                  label="累计 TECH 分"
                  field="sumTechHighScore"
                  value={displayNumber(profile, "sumTechHighScore")}
                />
                <ProfileField
                  label="累计 Platinum Star"
                  field="sumPlatinumScoreStar"
                  value={displayNumber(profile, "sumPlatinumScoreStar")}
                />
                <ProfileField
                  label="最后游玩地点"
                  field="lastPlaceName"
                  value={displayValue(profile, "lastPlaceName")}
                />
                <ProfileField
                  label="最后游玩"
                  field="lastPlayDate"
                  value={displayValue(profile, "lastPlayDate")}
                />
                <ProfileField
                  label="ROM 版本"
                  field="lastRomVersion"
                  value={displayValue(profile, "lastRomVersion")}
                />
                <ProfileField
                  label="数据版本"
                  field="lastDataVersion"
                  value={displayValue(profile, "lastDataVersion")}
                />
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent id="recent" className="mt-6">
          <Card className="overflow-hidden border-border bg-card">
            <CardHeader>
              <CardTitle>最近游玩</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableHead isRowHeader>乐曲 ID</TableHead>
                  <TableHead>难度</TableHead>
                  <TableHead>TECH SCORE</TableHead>
                  <TableHead>TECH RANK</TableHead>
                  <TableHead>BATTLE SCORE</TableHead>
                  <TableHead>PLATINUM</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>游玩时间</TableHead>
                </TableHeader>
                <TableBody>
                  {plays.length ? (
                    plays.map((play, index) => (
                      <TableRow
                        key={`${rawValue(play, "musicId")}-${rawValue(play, "userPlayDate")}-${index}`}
                      >
                        <TableCell>
                          <span className="font-medium">{musicName(play)}</span>
                          <span className="ml-2 font-mono text-xs text-muted-foreground">
                            #{displayValue(play, "musicId")}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {displayValue(play, "level")}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(play, "techScore")}
                        </TableCell>
                        <TableCell>
                          {displayValue(play, "techScoreRank")}
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(play, "battleScore")}
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(play, "platinumScore")}
                        </TableCell>
                        <TableCell>{playStatus(play)}</TableCell>
                        <TableCell className="text-muted-foreground">
                          {displayPreferredValue(play, [
                            "userPlayDate",
                            "playDate",
                          ])}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <EmptyRow columns={8} text="暂无 Ongeki 游玩记录。" />
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent id="scores" className="mt-6">
          <Card className="overflow-hidden border-border bg-card">
            <CardHeader>
              <CardTitle>成绩明细</CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              <Table>
                <TableHeader>
                  <TableHead isRowHeader>乐曲 ID</TableHead>
                  <TableHead>难度</TableHead>
                  <TableHead>最高 TECH SCORE</TableHead>
                  <TableHead>TECH RANK</TableHead>
                  <TableHead>最高 BATTLE SCORE</TableHead>
                  <TableHead>最高 PLATINUM</TableHead>
                  <TableHead>游玩次数</TableHead>
                  <TableHead>最大连击</TableHead>
                </TableHeader>
                <TableBody>
                  {details.length ? (
                    details.map((detail, index) => (
                      <TableRow
                        key={`${rawValue(detail, "musicId")}-${rawValue(detail, "level")}-${index}`}
                      >
                        <TableCell>
                          <span className="font-medium">
                            {musicName(detail)}
                          </span>
                          <span className="ml-2 font-mono text-xs text-muted-foreground">
                            #{displayValue(detail, "musicId")}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {displayValue(detail, "level")}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(detail, "techScoreMax")}
                        </TableCell>
                        <TableCell>
                          {displayValue(detail, "techScoreRank")}
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(detail, "battleScoreMax")}
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(detail, "platinumScoreMax")}
                        </TableCell>
                        <TableCell>
                          {displayNumber(detail, "playCount")}
                        </TableCell>
                        <TableCell>
                          {displayNumber(detail, "maxComboCount")}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <EmptyRow columns={8} text="暂无 Ongeki 成绩明细。" />
                  )}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

function rawValue(source: OngekiRecord, key: string) {
  return source[key]
}

function displayValue(source: OngekiRecord, key: string) {
  const value = rawValue(source, key)
  return value === undefined || value === null || value === ""
    ? "—"
    : String(value)
}

function displayPreferredValue(source: OngekiRecord, keys: string[]) {
  for (const key of keys) {
    const value = rawValue(source, key)
    if (value !== undefined && value !== null && value !== "")
      return String(value)
  }
  return "—"
}

function displayNumber(source: OngekiRecord, key: string) {
  const value = rawValue(source, key)
  if (typeof value === "number") return value.toLocaleString()
  if (typeof value !== "string" || !/^-?\d+$/.test(value))
    return displayValue(source, key)
  try {
    return BigInt(value).toLocaleString()
  } catch {
    return value
  }
}

function displayPreferredNumber(source: OngekiRecord, keys: string[]) {
  for (const key of keys) {
    const value = rawValue(source, key)
    if (value !== undefined && value !== null && value !== "")
      return displayNumber(source, key)
  }
  return "—"
}

function playStatus(play: OngekiRecord) {
  if (isTruthy(rawValue(play, "isAllBreak"))) return "ALL BREAK"
  if (isTruthy(rawValue(play, "isFullCombo"))) return "FULL COMBO"
  const clearStatus = rawValue(play, "clearStatus")
  return clearStatus === undefined || clearStatus === null
    ? "—"
    : String(clearStatus)
}

function isTruthy(value: unknown) {
  return value === true || value === 1 || value === "1" || value === "true"
}

function SummaryCard({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode
  label: string
  value: string
}) {
  return (
    <Card>
      <CardContent className="p-5">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span className="text-neutral-400">{icon}</span>
          {label}
        </div>
        <p className="mt-3 truncate text-2xl font-black" title={value}>
          {value}
        </p>
      </CardContent>
    </Card>
  )
}

function ProfileField({
  label,
  field,
  value,
}: {
  label: string
  field: string
  value: string
}) {
  return (
    <div className="min-w-0">
      <p className="text-xs text-muted-foreground">
        {label} · {field}
      </p>
      <p className="mt-1 truncate font-mono text-sm font-medium" title={value}>
        {value}
      </p>
    </div>
  )
}

function EmptyRow({ columns, text }: { columns: number; text: string }) {
  return (
    <TableRow>
      <TableCell className="py-10 text-center text-muted-foreground">
        {text}
      </TableCell>
      {Array.from({ length: columns - 1 }, (_, index) => (
        <TableCell key={index} />
      ))}
    </TableRow>
  )
}
