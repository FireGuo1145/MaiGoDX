import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Stats } from '@/types'

interface HomePageProps {
  stats: Stats | null
}

export function HomePage({ stats }: HomePageProps) {
  return (
    <div className="max-w-4xl space-y-8">
      <div className="space-y-2">
        <h2 className="text-4xl font-black tracking-tight">欢迎使用 MaiGoDX 管理门户</h2>
        <p className="text-slate-400 text-lg">面向街机游戏服务器的高性能管理系统，参考 AquaDX 的功能设计。</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader><CardTitle className="text-white">已支持游戏</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between p-3 bg-slate-800 rounded-lg">
              <span className="font-bold text-white">maimai DX</span>
              <Badge className="bg-emerald-500">已启用</Badge>
            </div>
            <div className="flex items-center justify-between p-3 bg-slate-800 rounded-lg opacity-50">
              <span className="font-bold text-white">CHUNITHM</span>
              <Badge variant="outline" className="text-slate-400 border-slate-700">暂未支持</Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-slate-900 border-slate-800">
          <CardHeader><CardTitle className="text-white">快速统计</CardTitle></CardHeader>
          <CardContent className="grid grid-cols-2 gap-4 text-center">
            <div className="p-4 bg-indigo-500/10 rounded-xl border border-indigo-500/20">
              <p className="text-2xl font-black text-indigo-400">{stats?.totalUsers || 0}</p>
              <p className="text-[10px] uppercase tracking-wider text-slate-500">用户数</p>
            </div>
            <div className="p-4 bg-emerald-500/10 rounded-xl border border-emerald-500/20">
              <p className="text-2xl font-black text-emerald-400">{stats?.totalPlays || 0}</p>
              <p className="text-[10px] uppercase tracking-wider text-slate-500">总游玩次数</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
