import { Input } from "@/components/ui/input"
import { useEffect, useState, type FormEvent } from "react"
import { MonitorCog, Plus, Save, Trash2 } from "lucide-react"
import api from "@/lib/api"
import { apiErrorMessage, type Terminal } from "@/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

function TerminalEditor({
  terminal,
  onSaved,
  onDeleted,
}: {
  terminal: Terminal
  onSaved: (terminal: Terminal) => Promise<void>
  onDeleted: (terminal: Terminal) => Promise<void>
}) {
  const [draft, setDraft] = useState(terminal)
  const [busy, setBusy] = useState(false)

  useEffect(() => setDraft(terminal), [terminal])

  const save = async () => {
    setBusy(true)
    try {
      await onSaved(draft)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Card className="border-border bg-card">
      <CardContent className="space-y-3 p-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p className="font-bold text-foreground">
              {terminal.name || "未命名机台"}
            </p>
            <p className="mt-1 font-mono text-xs text-neutral-300">
              {terminal.keychipId}
            </p>
          </div>
          <Badge className="border-neutral-500/25 bg-neutral-500/15 text-neutral-300">
            已登记
          </Badge>
        </div>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <Input
            value={draft.name}
            onChange={(event) =>
              setDraft({ ...draft, name: event.target.value })
            }
            placeholder="机台名称"
            className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
          />
          <Input
            value={draft.gameId}
            onChange={(event) =>
              setDraft({ ...draft, gameId: event.target.value.toUpperCase() })
            }
            placeholder="游戏 ID"
            className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
          />
          <Input
            value={draft.gameVersion}
            onChange={(event) =>
              setDraft({ ...draft, gameVersion: event.target.value })
            }
            placeholder="游戏版本（可选）"
            className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
          />
          <p className="self-center text-xs text-muted-foreground">
            最后连接：
            {terminal.lastSeenAt
              ? new Date(terminal.lastSeenAt).toLocaleString("zh-CN")
              : "从未连接"}
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            type="button"
            size="sm"
            className="bg-neutral-600 hover:bg-neutral-500"
            isDisabled={busy}
            onClick={() => void save()}
          >
            <Save size={15} /> 保存
          </Button>
          <Button
            type="button"
            size="sm"
            variant="outline"
            className="border-neutral-900/70 text-neutral-300 hover:bg-neutral-950/40"
            isDisabled={busy}
            onClick={() => void onDeleted(terminal)}
          >
            <Trash2 size={15} /> 解除绑定
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export function PersonalTerminalPanel() {
  const [terminals, setTerminals] = useState<Terminal[]>([])
  const [keychipId, setKeychipId] = useState("")
  const [name, setName] = useState("")
  const [gameId, setGameId] = useState("SDEZ")
  const [gameVersion, setGameVersion] = useState("")
  const [notice, setNotice] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const refresh = async () => {
    try {
      const result = await api.getUserTerminals()
      if (!result.success) throw new Error(result.message || "加载我的机台失败")
      setTerminals(result.terminals || [])
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  useEffect(() => {
    void refresh()
  }, [])

  const createTerminal = async (event: FormEvent) => {
    event.preventDefault()
    setIsSubmitting(true)
    setNotice(null)
    try {
      const result = await api.createUserTerminal({
        keychipId,
        name,
        gameId,
        gameVersion,
      })
      if (!result.success) throw new Error(result.message || "机台绑定失败")
      setKeychipId("")
      setName("")
      setGameId("SDEZ")
      setGameVersion("")
      await refresh()
      setNotice(
        "机台已登记。机台通过 ALL.Net PowerOn 后即可获得受保护的游戏连接地址。"
      )
    } catch (error) {
      setNotice(apiErrorMessage(error))
    } finally {
      setIsSubmitting(false)
    }
  }

  const saveTerminal = async (terminal: Terminal) => {
    try {
      const result = await api.updateUserTerminal(terminal)
      if (!result.success) throw new Error(result.message || "保存机台失败")
      await refresh()
      setNotice("机台信息已保存")
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
      const result = await api.deleteUserTerminal(terminal.ID)
      if (!result.success) throw new Error(result.message || "解除机台绑定失败")
      await refresh()
      setNotice("机台绑定已删除")
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  return (
    <Card className="border-border bg-card">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-foreground">
          <MonitorCog className="text-neutral-400" /> 我的机台
        </CardTitle>
        <CardDescription className="text-muted-foreground">
          你只能管理当前账户登记的
          Keychip。启停、跨账户归属与全局机台审核仍由管理员控制。
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
            minLength={11}
            maxLength={32}
            value={keychipId}
            onChange={(event) => setKeychipId(event.target.value.toUpperCase())}
            placeholder="Keychip 序列号（11–32 位）"
            className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
          />
          <Input
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="机台名称（可选）"
            className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
          />
          <div className="grid grid-cols-2 gap-2">
            <Input
              value={gameId}
              onChange={(event) => setGameId(event.target.value.toUpperCase())}
              placeholder="游戏 ID"
              className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
            />
            <Input
              value={gameVersion}
              onChange={(event) => setGameVersion(event.target.value)}
              placeholder="版本"
              className="h-10 rounded-md border border-border bg-muted px-3 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <Button
            type="submit"
            className="bg-neutral-600 hover:bg-neutral-500"
            isDisabled={isSubmitting}
          >
            <Plus size={16} /> {isSubmitting ? "正在登记…" : "登记机台"}
          </Button>
        </form>
        <div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
          {terminals.length ? (
            terminals.map((terminal) => (
              <TerminalEditor
                key={terminal.ID}
                terminal={terminal}
                onSaved={saveTerminal}
                onDeleted={deleteTerminal}
              />
            ))
          ) : (
            <p className="text-sm text-muted-foreground">
              尚未登记机台。填写 Keychip 序列号后即可绑定到当前账户。
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
