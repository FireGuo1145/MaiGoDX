import { Award, TrendingUp } from 'lucide-react'
import { Line, LineChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { DEFAULT_MAX_RATING, DEFAULT_PLAY_COUNT, DEFAULT_USER_RATING, RATING_TREND, type SongComp, type Stats } from '@/types'

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
          <div key={`${song.title}-${index}`} className="flex items-center justify-between p-3 bg-slate-800/60 rounded-lg gap-4">
            <div className="min-w-0">
              <p className="font-bold text-white text-sm truncate">{song.title}</p>
              <p className="text-xs text-slate-400">Level: {song.level} | Score: {song.score.toLocaleString()}</p>
            </div>
            <Badge className={`${badge} text-white font-mono shrink-0`}>+{song.rating}</Badge>
          </div>
        )) : <p className="text-sm text-slate-500">{emptyMessage}</p>}
      </CardContent>
    </Card>
  )
}

export function DashboardPage({ stats }: DashboardPageProps) {
  const rating = stats?.user?.Rating || DEFAULT_USER_RATING
  const maxRating = stats?.user?.MaxRating || DEFAULT_MAX_RATING
  const playCount = stats?.totalPlays || DEFAULT_PLAY_COUNT

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2"><CardDescription>Global Rating</CardDescription><CardTitle className="text-3xl text-white">{rating.toLocaleString()}</CardTitle></CardHeader>
          <CardContent><p className="text-xs text-emerald-400 flex items-center gap-1"><TrendingUp size={12} /> Max: {maxRating.toLocaleString()}</p></CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2"><CardDescription>Play Count</CardDescription><CardTitle className="text-3xl text-white">{playCount.toLocaleString()}</CardTitle></CardHeader>
          <CardContent><p className="text-xs text-slate-500">Total recorded plays</p></CardContent>
        </Card>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader className="pb-2"><CardDescription>Server Status</CardDescription><CardTitle className="text-3xl text-emerald-400">ONLINE</CardTitle></CardHeader>
          <CardContent><p className="text-xs text-slate-500">All services operational</p></CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <RatingList title="Best Bests (Top Rating)" songs={stats?.ratingComposition?.bests} tone="indigo" emptyMessage="No composition data." />
        <RatingList title="New Bests" songs={stats?.ratingComposition?.newBests} tone="emerald" emptyMessage="No new best data." />
      </div>

      <Card className="bg-slate-900 border-slate-800">
        <CardHeader><CardTitle className="text-white">Rating Trend</CardTitle></CardHeader>
        <CardContent className="h-[300px]">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={RATING_TREND}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
              <XAxis dataKey="name" stroke="#64748b" fontSize={12} />
              <YAxis stroke="#64748b" fontSize={12} domain={['dataMin - 100', 'dataMax + 100']} />
              <Tooltip contentStyle={{ backgroundColor: '#0f172a', border: '1px solid #1e293b' }} />
              <Line type="monotone" dataKey="rating" stroke="#6366f1" strokeWidth={3} dot={{ r: 4, fill: '#6366f1' }} />
            </LineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    </div>
  )
}
