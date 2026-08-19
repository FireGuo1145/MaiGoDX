import { Award, Coins, Gamepad2, TrendingUp } from "lucide-react"
import {
  Line,
  LineChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { type MetadataItem, type SongComp, type Stats } from "@/types"

interface DashboardPageProps {
  stats: Stats | null
  metadata: Record<string, MetadataItem[]>
}

interface RatingListProps {
  title: string
  songs: SongComp[] | undefined
  tone: "indigo" | "emerald"
  emptyMessage: string
}

function RatingList({
  title,
  songs,
  tone,
  emptyMessage,
  metadata,
}: RatingListProps & { metadata: Record<string, MetadataItem[]> }) {
  const color =
    tone === "indigo" ? "text-muted-foreground" : "text-muted-foreground"
  const songName = (id: number) =>
    metadata.music?.find((item) => item.id === id)?.name || `ID ${id}`

  return (
    <Card>
      <CardHeader>
        <CardTitle
          className={`flex items-center gap-2 text-foreground ${color}`}
        >
          <Award size={20} />
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {songs?.length ? (
          songs.map((song, index) => (
            <div
              key={`${song.musicId}-${song.level}-${index}`}
              className="flex items-center justify-between gap-4 rounded-lg bg-muted p-3"
            >
              <div className="min-w-0">
                <p className="truncate text-sm font-bold">
                  {songName(song.musicId)}{" "}
                  <span className="text-xs font-normal text-muted-foreground">
                    #{song.musicId}
                  </span>
                </p>
                <p className="text-xs text-muted-foreground">
                  难度：{song.level} | 达成率：
                  {(song.achievement / 10000).toFixed(4)}% | DX 分数：
                  {song.score.toLocaleString()}
                </p>
              </div>
              <Badge className="shrink-0 font-mono">
                评级 {song.scoreRank || "—"}
              </Badge>
            </div>
          ))
        ) : (
          <p className="text-sm text-muted-foreground">{emptyMessage}</p>
        )}
      </CardContent>
    </Card>
  )
}

export function DashboardPage({ stats, metadata }: DashboardPageProps) {
  const rating = stats?.user?.playerRating
  const maxRating = stats?.user?.highestRating
  const playCount = stats?.user?.playCount
  const trend = [...(stats?.trend ?? [])].sort((left, right) =>
    left.date.localeCompare(right.date)
  )

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 xl:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>当前 Rating</CardDescription>
            <CardTitle className="text-3xl">
              {rating?.toLocaleString() ?? "—"}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="flex items-center gap-1 text-xs text-muted-foreground">
              <TrendingUp size={12} /> 历史最高：
              {maxRating?.toLocaleString() ?? "—"}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>游玩次数</CardDescription>
            <CardTitle className="text-3xl">
              {playCount?.toLocaleString() ?? "—"}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">
              机台档案记录的累计游玩次数
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>当前 maimile</CardDescription>
            <CardTitle className="flex items-center gap-2 text-3xl">
              <Coins size={24} />
              {stats?.user?.point?.toLocaleString() ?? "—"}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">
              累计：{stats?.user?.totalPoint?.toLocaleString() ?? "—"}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>档案状态</CardDescription>
            <CardTitle className="text-3xl">
              {stats?.user?.userName ?? "未关联"}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">
              {stats?.message ?? "已关联 maimai 档案的实时数据"}
            </p>
          </CardContent>
        </Card>
      </div>

      {stats?.user && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Gamepad2 size={20} />
              档案明细
            </CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-x-8 gap-y-4 md:grid-cols-4">
            <Detail label="段位 Rating" value={stats.user.gradeRating} />
            <Detail label="乐曲 Rating" value={stats.user.musicRating} />
            <Detail label="段位等级" value={stats.user.gradeRank} />
            <Detail label="Class Rank" value={stats.user.classRank} />
            <Detail label="累计 DX 分数" value={stats.user.totalDeluxscore} />
            <Detail label="累计达成率" value={stats.user.totalAchievement} />
            <Detail
              label="最后游玩地点"
              value={stats.user.lastPlaceName || "—"}
            />
            <Detail
              label="最后游玩时间"
              value={stats.user.lastPlayDate || "—"}
            />
          </CardContent>
        </Card>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <RatingList
          title="最佳成绩（旧曲）"
          songs={stats?.ratingComposition?.bests}
          tone="indigo"
          emptyMessage="暂无评级构成数据。"
          metadata={metadata}
        />
        <RatingList
          title="最佳成绩（新曲）"
          songs={stats?.ratingComposition?.newBests}
          tone="emerald"
          emptyMessage="暂无新曲最佳成绩。"
          metadata={metadata}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Rating 趋势</CardTitle>
        </CardHeader>
        <CardContent className="h-[300px]">
          {trend.length ? (
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trend}>
                <CartesianGrid strokeDasharray="3 3" stroke="currentColor" />
                <XAxis dataKey="date" stroke="currentColor" fontSize={12} />
                <YAxis
                  stroke="currentColor"
                  fontSize={12}
                  domain={["dataMin - 100", "dataMax + 100"]}
                />
                <Tooltip />
                <Line
                  type="monotone"
                  dataKey="rating"
                  stroke="currentColor"
                  strokeWidth={3}
                  dot={{ r: 4, fill: "currentColor" }}
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-muted-foreground">
              该档案暂无已持久化的 Rating 历史记录。
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function Detail({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="min-w-0">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="mt-1 truncate font-mono font-medium" title={String(value)}>
        {typeof value === "number" ? value.toLocaleString() : value}
      </p>
    </div>
  )
}
