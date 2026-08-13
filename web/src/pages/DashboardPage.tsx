import { Award, TrendingUp } from 'lucide-react'
import { Line, LineChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { type SongComp, type Stats } from '@/types'

interface DashboardPageProps {
  stats: Stats | null
}

interface RatingListProps {
  title: string
  songs: SongComp[] | undefined
  tone: 'indigo' | 'emerald'
  emptyMessage: string
}

function RatingList({ title, songs, tone, emptyMessage }: RatingListProps) {
  const color = tone === 'indigo' ? 'text-indigo-400' : 'text-emerald-400'
  const badge = tone === 'indigo' ? 'bg-indigo-600' : 'bg-emerald-600'

  return (
    <Card className="bg-slate-900 border-slate-800">
      <CardHeader><CardTitle className={`text-white flex items-center gap-2 ${color}`}><Award size={20} />{title}</CardTitle></CardHeader>
      <CardContent className="space-y-3">
        {songs?.length ? songs.map((song, index) => (
          <div key={`${song.musicId}-${song.level}-${index}`} className="flex items-center justify-between p-3 bg-slate-800/60 rounded-lg gap-4">
            <div className="min-w-0">
              <p className="font-bold text-white text-sm truncate">乐曲 ID：{song.musicId}</p>
              <p className="text-xs text-slate-400">难度：{song.level} | 达成率：{(song.achievement / 10000).toFixed(4)}% | DX 分数：{song.score.toLocaleString()}</p>
            </div>
            <Badge className={`${badge} text-white font-mono shrink-0`}>评级 {song.scoreRank || '—'}</Badge>
          </div>
        )) : <p className="text-sm text-slate-500">{emptyMessage}</p>}
      </CardContent>
    </Card>
  )
}

export function DashboardPage({ stats }: DashboardPageProps) {
  const rating = stats?.user?.rating
  const maxRating = stats?.user?.maxRating
  const playCount = stats?.totalPlays
  const trend = stats?.trend ?? []

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2"><CardDescription>当前 Rating</CardDescription><CardTitle className="text-3xl text-white">{rating?.toLocaleString() ?? '—'}</CardTitle></CardHeader>
          <CardContent><p className="text-xs text-emerald-400 flex items-center gap-1"><TrendingUp size={12} /> 历史最高：{maxRating?.toLocaleString() ?? '—'}</p></CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2"><CardDescription>游玩次数</CardDescription><CardTitle className="text-3xl text-white">{playCount?.toLocaleString() ?? '—'}</CardTitle></CardHeader>
          <CardContent><p className="text-xs text-slate-500">已关联档案的成绩记录数</p></CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2"><CardDescription>档案状态</CardDescription><CardTitle className="text-3xl text-white">{stats?.user?.userName ?? '未关联'}</CardTitle></CardHeader>
          <CardContent><p className="text-xs text-slate-500">{stats?.message ?? '已关联 maimai 档案的实时数据'}</p></CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <RatingList title="最佳成绩（旧曲）" songs={stats?.ratingComposition?.bests} tone="indigo" emptyMessage="暂无评级构成数据。" />
        <RatingList title="最佳成绩（新曲）" songs={stats?.ratingComposition?.newBests} tone="emerald" emptyMessage="暂无新曲最佳成绩。" />
      </div>

      <Card className="bg-slate-900 border-slate-800">
        <CardHeader><CardTitle className="text-white">Rating 趋势</CardTitle></CardHeader>
        <CardContent className="h-[300px]">
          {trend.length ? (
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trend}>
                <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
                <XAxis dataKey="date" stroke="#64748b" fontSize={12} />
                <YAxis stroke="#64748b" fontSize={12} domain={['dataMin - 100', 'dataMax + 100']} />
                <Tooltip contentStyle={{ backgroundColor: '#0f172a', border: '1px solid #1e293b' }} />
                <Line type="monotone" dataKey="rating" stroke="#6366f1" strokeWidth={3} dot={{ r: 4, fill: '#6366f1' }} />
              </LineChart>
            </ResponsiveContainer>
          ) : <p className="text-sm text-slate-500">该档案暂无已持久化的 Rating 历史记录。</p>}
        </CardContent>
      </Card>
    </div>
  )
}
