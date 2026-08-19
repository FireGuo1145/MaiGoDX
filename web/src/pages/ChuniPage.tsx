import { Activity, Music2, UserRound } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import type { ChuniStats } from "@/types"

export function ChuniPage({ stats }: { stats: ChuniStats | null }) {
  const profile = stats?.profile || {}
  const plays = stats?.recentPlays || []
  const details = stats?.musicDetails || []
  const value = (source: Record<string, unknown>, key: string) => source[key] as string | number | undefined

  return (
    <div className="space-y-6">
      {stats?.message && <p className="rounded-md border border-border bg-card p-4 text-sm text-muted-foreground">{stats.message}</p>}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <SummaryCard icon={<UserRound />} label="玩家名称" value={value(profile, "userName") || "—"} />
        <SummaryCard icon={<Activity />} label="当前 Rating" value={value(profile, "playerRating")?.toLocaleString() || "—"} />
        <SummaryCard icon={<Music2 />} label="已保存成绩" value={details.length.toLocaleString()} />
      </div>

      <Card className="overflow-hidden border-border bg-card">
        <CardHeader><CardTitle>最近游玩</CardTitle></CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader><TableHead isRowHeader>乐曲</TableHead><TableHead>难度</TableHead><TableHead>分数</TableHead><TableHead>游玩日期</TableHead></TableHeader>
            <TableBody>
              {plays.length ? plays.map((play, index) => (
                <TableRow key={`${value(play, "musicId")}-${index}`}>
                  <TableCell className="font-mono">{value(play, "musicId") ?? "—"}</TableCell>
                  <TableCell><Badge variant="outline">LV.{value(play, "level") ?? "—"}</Badge></TableCell>
                  <TableCell className="font-mono">{value(play, "score")?.toLocaleString() ?? "—"}</TableCell>
                  <TableCell className="text-muted-foreground">{value(play, "playDate") || value(play, "userPlayDate") || "—"}</TableCell>
                </TableRow>
              )) : <TableRow><TableCell className="py-10 text-center text-muted-foreground">暂无 CHUNITHM 游玩记录。</TableCell><TableCell /><TableCell /><TableCell /></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card className="overflow-hidden border-border bg-card">
        <CardHeader><CardTitle>成绩明细</CardTitle></CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader><TableHead isRowHeader>乐曲</TableHead><TableHead>难度</TableHead><TableHead>分数</TableHead><TableHead>达成状态</TableHead></TableHeader>
            <TableBody>
              {details.length ? details.map((detail, index) => (
                <TableRow key={`${value(detail, "musicId")}-${value(detail, "level")}-${index}`}>
                  <TableCell className="font-mono">{value(detail, "musicId") ?? "—"}</TableCell>
                  <TableCell><Badge variant="outline">LV.{value(detail, "level") ?? "—"}</Badge></TableCell>
                  <TableCell className="font-mono">{value(detail, "score")?.toLocaleString() ?? "—"}</TableCell>
                  <TableCell>{value(detail, "isSuccess") === true ? "已通关" : "—"}</TableCell>
                </TableRow>
              )) : <TableRow><TableCell className="py-10 text-center text-muted-foreground">暂无 CHUNITHM 成绩明细。</TableCell><TableCell /><TableCell /><TableCell /></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}

function SummaryCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return <Card><CardContent className="p-5"><div className="flex items-center gap-2 text-sm text-muted-foreground"><span className="text-neutral-400">{icon}</span>{label}</div><p className="mt-3 text-2xl font-black">{value}</p></CardContent></Card>
}
