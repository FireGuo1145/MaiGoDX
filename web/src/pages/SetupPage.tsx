import { BookOpen } from "lucide-react"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { SETUP_STEPS } from "@/types"

export function SetupPage() {
  return (
    <div className="max-w-3xl space-y-6">
      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-foreground">
            <BookOpen className="text-neutral-400" /> 连接与接入指南
          </CardTitle>
          <CardDescription className="text-muted-foreground">
            请按以下步骤将已授权的游戏客户端接入 MaiGoDX。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-sm leading-relaxed text-muted-foreground">
          {SETUP_STEPS.map((step, index) => (
            <section
              key={step.title}
              className="space-y-2 rounded-lg bg-muted/60 p-4"
            >
              <h3 className="font-bold text-foreground">
                {index + 1}. {step.title}
              </h3>
              <p>{step.body}</p>
            </section>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
