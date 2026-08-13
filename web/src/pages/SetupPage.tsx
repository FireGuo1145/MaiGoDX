import { BookOpen } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { SETUP_STEPS } from '@/types'

export function SetupPage() {
  return (
    <div className="max-w-3xl space-y-6">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-white flex items-center gap-2"><BookOpen className="text-indigo-400" /> 连接与接入指南</CardTitle>
          <CardDescription className="text-slate-400">请按以下步骤将已授权的游戏客户端接入 MaiGoDX。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-sm text-slate-300 leading-relaxed">
          {SETUP_STEPS.map((step, index) => (
            <section key={step.title} className="p-4 bg-slate-800/60 rounded-lg space-y-2">
              <h3 className="font-bold text-white">{index + 1}. {step.title}</h3>
              <p>{step.body}</p>
            </section>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
