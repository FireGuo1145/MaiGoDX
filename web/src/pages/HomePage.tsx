import { Award, CircleUserRound, TrendingUp } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import type { Stats } from "@/types"

interface HomePageProps {
  stats: Stats | null
}

export function HomePage({ stats }: HomePageProps) {
  const profile = stats?.user

  return (
    <div className="max-w-4xl space-y-8">
      <div className="space-y-2">
        <h2 className="text-4xl font-black tracking-tight">
          {profile?.userName ? `欢迎回来，${profile.userName}` : "欢迎回来"}
        </h2>
        <p className="text-lg text-muted-foreground">
          查看你的 maimai DX 档案、Rating 与最近游玩数据。
        </p>
      </div>

      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <CircleUserRound className="text-neutral-400" /> 我的 maimai DX 档案
          </CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <ProfileStat
            icon={<TrendingUp />}
            label="当前 Rating"
            value={profile?.playerRating}
          />
          <ProfileStat
            icon={<Award />}
            label="历史最高 Rating"
            value={profile?.highestRating}
          />
          <ProfileStat
            icon={<CircleUserRound />}
            label="maimile"
            value={profile?.totalPoint}
          />
        </CardContent>
      </Card>
    </div>
  )
}

function ProfileStat({
  icon,
  label,
  value,
}: {
  icon: React.ReactNode
  label: string
  value: number | undefined
}) {
  return (
    <div className="rounded-xl border border-neutral-500/20 bg-neutral-500/10 p-4">
      <div className="flex items-center gap-2 text-xs font-medium text-muted-foreground">
        <span className="text-neutral-400">{icon}</span>
        {label}
      </div>
      <p className="mt-3 text-2xl font-black text-foreground">
        {value?.toLocaleString() ?? "—"}
      </p>
    </div>
  )
}
