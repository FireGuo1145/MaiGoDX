import { Activity, Gauge, Music2, UserRound } from "lucide-react"
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
import type { ChuniStats } from "@/types"

type ChuniRecord = Record<string, unknown>

export function ChuniPage({ stats }: { stats: ChuniStats | null }) {
  const profile = stats?.profile || {}
  const plays = stats?.recentPlays || []
  const details = stats?.musicDetails || []

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
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <SummaryCard
              icon={<UserRound />}
              label="玩家名称"
              value={displayValue(profile, "userName")}
            />
            <SummaryCard
              icon={<Gauge />}
              label="等级（level）"
              value={displayNumber(profile, "level")}
            />
            <SummaryCard
              icon={<Activity />}
              label="玩家评分（playerRating）"
              value={displayNumber(profile, "playerRating")}
            />
            <SummaryCard
              icon={<Music2 />}
              label="游玩次数（playCount）"
              value={displayNumber(profile, "playCount")}
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
                  label="最高评分"
                  field="highestRating"
                  value={displayNumber(profile, "highestRating")}
                />
                <ProfileField
                  label="OVER POWER"
                  field="overPowerPoint"
                  value={displayNumber(profile, "overPowerPoint")}
                />
                <ProfileField
                  label="OVER POWER 比率"
                  field="overPowerRate"
                  value={displayNumber(profile, "overPowerRate")}
                />
                <ProfileField
                  label="持有点数"
                  field="point"
                  value={displayNumber(profile, "point")}
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
                  <TableHead>难度 ID</TableHead>
                  <TableHead>分数</TableHead>
                  <TableHead>Rank</TableHead>
                  <TableHead>结果</TableHead>
                  <TableHead>游玩时评分</TableHead>
                  <TableHead>游玩时间</TableHead>
                </TableHeader>
                <TableBody>
                  {plays.length ? (
                    plays.map((play, index) => (
                      <TableRow
                        key={`${rawValue(play, "musicId")}-${rawValue(play, "userPlayDate")}-${index}`}
                      >
                        <TableCell className="font-mono">
                          {displayValue(play, "musicId")}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {displayValue(play, "level")}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(play, "score")}
                        </TableCell>
                        <TableCell>{displayValue(play, "rank")}</TableCell>
                        <TableCell>
                          {isTruthy(rawValue(play, "isClear"))
                            ? "CLEAR"
                            : "未通关"}
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(play, "playerRating")}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {displayValue(play, "userPlayDate")}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <EmptyRow columns={7} text="暂无 CHUNITHM 游玩记录。" />
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
                  <TableHead>难度 ID</TableHead>
                  <TableHead>最高分</TableHead>
                  <TableHead>Score Rank</TableHead>
                  <TableHead>游玩次数</TableHead>
                  <TableHead>最大连击</TableHead>
                  <TableHead>MISS</TableHead>
                  <TableHead>达成状态</TableHead>
                </TableHeader>
                <TableBody>
                  {details.length ? (
                    details.map((detail, index) => (
                      <TableRow
                        key={`${rawValue(detail, "musicId")}-${rawValue(detail, "level")}-${index}`}
                      >
                        <TableCell className="font-mono">
                          {displayValue(detail, "musicId")}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {displayValue(detail, "level")}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-mono">
                          {displayNumber(detail, "scoreMax")}
                        </TableCell>
                        <TableCell>
                          {displayValue(detail, "scoreRank")}
                        </TableCell>
                        <TableCell>
                          {displayNumber(detail, "playCount")}
                        </TableCell>
                        <TableCell>
                          {displayNumber(detail, "maxComboCount")}
                        </TableCell>
                        <TableCell>
                          {displayNumber(detail, "missCount")}
                        </TableCell>
                        <TableCell>
                          {isTruthy(rawValue(detail, "isSuccess"))
                            ? "已达成"
                            : "未达成"}
                        </TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <EmptyRow columns={8} text="暂无 CHUNITHM 成绩明细。" />
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

function rawValue(source: ChuniRecord, key: string) {
  return source[key]
}

function displayValue(source: ChuniRecord, key: string) {
  const value = rawValue(source, key)
  return value === undefined || value === null || value === ""
    ? "—"
    : String(value)
}

function displayNumber(source: ChuniRecord, key: string) {
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
        <p className="mt-3 text-2xl font-black">{value}</p>
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
