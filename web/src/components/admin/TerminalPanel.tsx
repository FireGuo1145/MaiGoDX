import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useState } from "react"
import { MonitorCog, Plus, Power, Trash2 } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import api from "@/lib/api"
import {
  DEFAULT_TERMINAL_GAME_ID,
  TERMINAL_GAMES,
  terminalGameLabel,
} from "@/lib/terminal-games"
import type { Terminal } from "@/types"
import { apiErrorMessage } from "@/types"

interface TerminalPanelProps {
  terminals: Terminal[]
  onChanged: () => Promise<void>
}

export function TerminalPanel({ terminals, onChanged }: TerminalPanelProps) {
  const [keychipId, setKeychipId] = useState("")
  const [name, setName] = useState("")
  const [gameID, setGameID] = useState(DEFAULT_TERMINAL_GAME_ID)
  const [gameVersion, setGameVersion] = useState("")
  const [notice, setNotice] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const createTerminal = async (event: React.FormEvent) => {
    event.preventDefault()
    setIsSubmitting(true)
    setNotice(null)
    try {
      const result = await api.createTerminal({
        keychipId,
        name,
        gameId: gameID,
        gameVersion,
      })
      if (!result.success) throw new Error(result.message || "机台绑定失败")
      setKeychipId("")
      setName("")
      setGameVersion("")
      setNotice("机台绑定成功。请重启机台或重新执行 ALL.Net PowerOn。")
      await onChanged()
    } catch (error) {
      setNotice(apiErrorMessage(error))
    } finally {
      setIsSubmitting(false)
    }
  }

  const updateTerminal = async (terminal: Terminal, isEnabled: boolean) => {
    try {
      const result = await api.updateTerminal({ ...terminal, isEnabled })
      if (!result.success) throw new Error(result.message || "更新机台失败")
      await onChanged()
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  const deleteTerminal = async (terminal: Terminal) => {
    if (
      !window.confirm(
        `确认解除机台「${terminal.name || terminal.keychipId}」的绑定吗？`
      )
    )
      return
    try {
      const result = await api.deleteTerminal(terminal.ID)
      if (!result.success) throw new Error(result.message || "删除机台失败")
      await onChanged()
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="flex items-center gap-2 text-xl font-bold">
          <MonitorCog /> 机台绑定与授权
        </h2>
        <Button
          type="button"
          onClick={() => void onChanged()}
          size="sm"
          variant="outline"
          className="border-border"
        >
          刷新机台
        </Button>
      </div>
      <Card className="border-border bg-card">
        <CardHeader>
          <CardTitle className="text-sm text-foreground">
            绑定 ALL.Net Keychip
          </CardTitle>
          <CardDescription className="text-muted-foreground">
            登记格式：Axxx-xxxxxxxxxxx。系统只按横线前的前缀匹配，末四位仅作兼容信息，前缀不能重复。
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {notice && (
            <p className="rounded-md border border-neutral-500/30 bg-neutral-500/10 p-3 text-sm text-neutral-200">
              {notice}
            </p>
          )}
          <form
            onSubmit={createTerminal}
            className="grid grid-cols-1 gap-3 md:grid-cols-4"
          >
            <Input
              required
              pattern="A[A-Z0-9]{3}-[A-Z0-9]{11}"
              minLength={16}
              maxLength={16}
              value={keychipId}
              onChange={(event) =>
                setKeychipId(event.target.value.toUpperCase())
              }
              placeholder="Axxx-xxxxxxxxxxx"
              className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
            />
            <Input
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="机台名称，例如：一号机"
              className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
            />
            <div className="grid grid-cols-2 gap-2">
              <Select
                selectedKey={gameID}
                onSelectionChange={(key) => setGameID(String(key))}
              >
                <SelectTrigger className="h-10 rounded-md border border-border bg-muted px-3 text-foreground">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TERMINAL_GAMES.map((game) => (
                    <SelectItem key={game.id} id={game.id}>
                      {game.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Input
                value={gameVersion}
                onChange={(event) => setGameVersion(event.target.value)}
                placeholder="版本（可选）"
                className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <Button
              isDisabled={isSubmitting}
              type="submit"
              className="bg-neutral-600 hover:bg-neutral-500"
            >
              <Plus size={16} /> {isSubmitting ? "正在绑定…" : "绑定机台"}
            </Button>
          </form>
        </CardContent>
      </Card>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {terminals.length ? (
          terminals.map((terminal) => (
            <Card key={terminal.ID} className="border-border bg-card">
              <CardContent className="space-y-4 p-5">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="font-bold text-foreground">
                      {terminal.name || "未命名机台"}
                    </p>
                    <p className="mt-1 font-mono text-xs text-neutral-300">
                      登记 Keychip：{terminal.keychipId}
                    </p>
                    {terminal.lastSeenKeychip &&
                    terminal.lastSeenKeychip.replaceAll("-", "") !==
                      terminal.keychipId.replaceAll("-", "") ? (
                      <p className="mt-1 font-mono text-xs text-neutral-300">
                        最近上报：{terminal.lastSeenKeychip.replaceAll("-", "")}
                      </p>
                    ) : null}
                  </div>
                  <Badge
                    className={
                      terminal.isEnabled
                        ? "border-neutral-500/25 bg-neutral-500/15 text-neutral-300"
                        : "border-neutral-500/25 bg-neutral-500/15 text-neutral-300"
                    }
                  >
                    {terminal.isEnabled ? "已启用" : "已停用"}
                  </Badge>
                </div>
                <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
                  <p>
                    游戏：
                    <span className="text-foreground">
                      {terminalGameLabel(terminal.gameId)}
                    </span>
                  </p>
                  <p>
                    版本：
                    <span className="text-foreground">
                      {terminal.gameVersion || "未上报"}
                    </span>
                  </p>
                  <p className="col-span-2">
                    最后连接：
                    <span className="text-foreground">
                      {terminal.lastSeenAt
                        ? new Date(terminal.lastSeenAt).toLocaleString("zh-CN")
                        : "从未连接"}
                    </span>
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="border-border"
                    onClick={() =>
                      void updateTerminal(terminal, !terminal.isEnabled)
                    }
                  >
                    <Power size={15} /> {terminal.isEnabled ? "停用" : "启用"}
                  </Button>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    className="border-neutral-900/70 text-neutral-300 hover:bg-neutral-950/40"
                    onClick={() => void deleteTerminal(terminal)}
                  >
                    <Trash2 size={15} /> 解除绑定
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))
        ) : (
          <p className="text-sm text-muted-foreground">
            尚未绑定机台。请先登记机台 Keychip 序列号。
          </p>
        )}
      </div>
    </section>
  )
}
