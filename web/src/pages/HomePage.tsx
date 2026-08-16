import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { Stats } from "@/types"

interface HomePageProps {
  stats: Stats | null
}

export function HomePage({ stats }: HomePageProps) {
  return (
    <div className="max-w-4xl space-y-8">
      <div className="space-y-2">
        <h2 className="text-4xl font-black tracking-tight">
          欢迎使用 MaiGoDX 管理门户
        </h2>
        <p className="text-lg text-muted-foreground">
          面向街机游戏服务器的高性能管理系统，参考 AquaDX 的功能设计。
        </p>
      </div>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="text-foreground">已支持游戏</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between rounded-lg bg-muted p-3">
              <span className="font-bold text-foreground">maimai DX</span>
              <Badge className="bg-neutral-500">已启用</Badge>
            </div>
            <div className="flex items-center justify-between rounded-lg bg-muted p-3 opacity-50">
              <span className="font-bold text-foreground">CHUNITHM</span>
              <Badge
                variant="outline"
                className="border-border text-muted-foreground"
              >
                暂未支持
              </Badge>
            </div>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader>
            <CardTitle className="text-foreground">快速统计</CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-4 text-center">
            <div className="rounded-xl border border-neutral-500/20 bg-neutral-500/10 p-4">
              <p className="text-2xl font-black text-neutral-400">
                {stats?.totalUsers || 0}
              </p>
              <p className="text-[10px] tracking-wider text-muted-foreground uppercase">
                用户数
              </p>
            </div>
            <div className="rounded-xl border border-neutral-500/20 bg-neutral-500/10 p-4">
              <p className="text-2xl font-black text-neutral-400">
                {stats?.totalPlays || 0}
              </p>
              <p className="text-[10px] tracking-wider text-muted-foreground uppercase">
                总游玩次数
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
