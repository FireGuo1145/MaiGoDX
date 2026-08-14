import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { RANK_SUMMARY, type Stats } from '@/types'

interface MaimaiPageProps {
  stats: Stats | null
}

export function MaimaiPage({ stats }: MaimaiPageProps) {
  const plays = stats?.recentPlays || []
  const travelPartners = stats?.travelPartners || []
  const functionTickets = stats?.functionTickets || []
  const regions = stats?.regions || []

  return (
    <div className="space-y-6">
      <Tabs defaultSelectedKey="recent" className="w-full">
        <TabsList className="bg-slate-900 border border-slate-800">
          <TabsTrigger id="recent" className="data-[selected]:bg-indigo-600">最近游玩</TabsTrigger>
          <TabsTrigger id="stats" className="data-[selected]:bg-indigo-600">成绩统计</TabsTrigger>
          <TabsTrigger id="profile" className="data-[selected]:bg-indigo-600">游戏档案</TabsTrigger>
        </TabsList>

        <TabsContent id="recent" className="mt-6">
          <Card className="bg-slate-900 border-slate-800 overflow-hidden">
            <Table>
              <TableHeader className="bg-slate-800/50">
                  <TableHead className="text-slate-300">乐曲</TableHead>
                  <TableHead className="text-slate-300">难度</TableHead>
                  <TableHead className="text-slate-300 text-right">达成率</TableHead>
                  <TableHead className="text-slate-300 text-right">分数</TableHead>
                  <TableHead className="text-slate-300 text-right">日期</TableHead>
              </TableHeader>
              <TableBody>
                {plays.length ? plays.map((play) => (
                  <TableRow key={play.ID} className="border-slate-800 hover:bg-slate-800/30">
                    <TableCell className="font-bold text-white">乐曲 ID：{play.musicId}</TableCell>
                    <TableCell><Badge variant="outline" className="border-indigo-500/50 text-indigo-400">LV.{play.level}</Badge></TableCell>
                    <TableCell className="text-right font-mono text-emerald-400">{(play.achievement / 10000).toFixed(4)}%</TableCell>
                    <TableCell className="text-right font-mono">{play.score.toLocaleString()}</TableCell>
                    <TableCell className="text-right text-slate-500 text-xs">{play.createDate || '—'}</TableCell>
                  </TableRow>
                )) : (
                  <TableRow><TableCell className="text-center py-12 text-slate-500">暂无游玩记录。</TableCell><TableCell /><TableCell /><TableCell /><TableCell /></TableRow>
                )}
              </TableBody>
            </Table>
          </Card>
        </TabsContent>

        <TabsContent id="stats" className="mt-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
            {RANK_SUMMARY.map((rank) => (
              <Card key={rank} className="bg-slate-900 border-slate-800 text-center p-6">
                <p className="text-2xl font-black text-indigo-400 mb-1">{rank}</p>
                <p className="text-sm text-slate-500 font-bold">{stats?.rankCounts?.[rank] ?? 0} 首</p>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent id="profile" className="mt-6 space-y-6">
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <Card className="border-slate-800 bg-slate-900 p-5">
              <p className="text-sm font-medium text-slate-400">当前搭档</p>
              <p className="mt-2 text-3xl font-black text-indigo-400">
                {stats?.partner?.partnerId ? `ID ${stats.partner.partnerId}` : '未装备'}
              </p>
              <p className="mt-2 text-xs text-slate-500">来自机台同步的 partnerId。</p>
            </Card>
            <Card className="border-slate-800 bg-slate-900 p-5">
              <p className="text-sm font-medium text-slate-400">功能票库存</p>
              <p className="mt-2 text-3xl font-black text-emerald-400">{functionTickets.reduce((total, ticket) => total + ticket.stock, 0)}</p>
              <p className="mt-2 text-xs text-slate-500">共 {functionTickets.length} 种功能票。</p>
            </Card>
          </div>

          <ProfileTable
            title="旅行伙伴"
            empty="尚无旅行伙伴数据。完成机台同步后会显示在这里。"
            headers={['搭档 ID', '亲密度等级', '已领奖励次数']}
            rows={travelPartners.map((partner) => [partner.partnerId, partner.intimateLevel, partner.intimateCountRewarded])}
          />
          <ProfileTable
            title="功能票"
            empty="尚未持有功能票。"
            headers={['票种 ID', '库存']}
            rows={functionTickets.map((ticket) => [ticket.itemId, ticket.stock])}
          />
          <ProfileTable
            title="区域游玩记录"
            empty="尚无区域游玩记录。"
            headers={['区域 ID', '游玩次数']}
            rows={regions.map((region) => [region.regionId, region.playCount])}
          />
        </TabsContent>
      </Tabs>
    </div>
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
    <Card className="overflow-hidden border-slate-800 bg-slate-900">
      <div className="border-b border-slate-800 px-5 py-4 text-sm font-bold text-white">{title}</div>
      <Table>
        <TableHeader className="bg-slate-800/50">
          {headers.map((header) => <TableHead key={header} className="text-slate-300">{header}</TableHead>)}
        </TableHeader>
        <TableBody>
          {rows.length ? rows.map((row, index) => (
            <TableRow key={`${title}-${index}`} className="border-slate-800">
              {row.map((value, column) => <TableCell key={column} className="font-mono text-slate-200">{value}</TableCell>)}
            </TableRow>
          )) : (
            <TableRow>
              <TableCell colSpan={headers.length} className="py-8 text-center text-sm text-slate-500">{empty}</TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </Card>
  )
}
