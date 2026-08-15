import { useState } from 'react'
import { MonitorCog, Plus, Power, Trash2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import api from '@/lib/api'
import type { Terminal } from '@/types'
import { apiErrorMessage } from '@/types'

interface TerminalPanelProps {
  terminals: Terminal[]
  onChanged: () => Promise<void>
}

const defaultGameID = 'SDEZ'

export function TerminalPanel({ terminals, onChanged }: TerminalPanelProps) {
  const [keychipId, setKeychipId] = useState('')
  const [name, setName] = useState('')
  const [gameID, setGameID] = useState(defaultGameID)
  const [gameVersion, setGameVersion] = useState('')
  const [notice, setNotice] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const createTerminal = async (event: React.FormEvent) => {
    event.preventDefault()
    setIsSubmitting(true)
    setNotice(null)
    try {
      const result = await api.createTerminal({ keychipId, name, gameId: gameID, gameVersion })
      if (!result.success) throw new Error(result.message || '机台绑定失败')
      setKeychipId('')
      setName('')
      setGameVersion('')
      setNotice('机台绑定成功。请重启机台或重新执行 ALL.Net PowerOn。')
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
      if (!result.success) throw new Error(result.message || '更新机台失败')
      await onChanged()
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  const deleteTerminal = async (terminal: Terminal) => {
    if (!window.confirm(`确认解除机台「${terminal.name || terminal.keychipId}」的绑定吗？`)) return
    try {
      const result = await api.deleteTerminal(terminal.ID)
      if (!result.success) throw new Error(result.message || '删除机台失败')
      await onChanged()
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  return (
    <section className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-bold flex items-center gap-2"><MonitorCog /> 机台绑定与授权</h2>
        <Button type="button" onClick={() => void onChanged()} size="sm" variant="outline" className="border-slate-700">刷新机台</Button>
      </div>
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-white text-sm">绑定 ALL.Net Keychip</CardTitle>
          <CardDescription className="text-slate-400">登记格式：Axxx-xxxxxxxxxxx。系统只按横线前的前缀匹配，末四位仅作兼容信息，前缀不能重复。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {notice && <p className="rounded-md border border-indigo-500/30 bg-indigo-500/10 p-3 text-sm text-indigo-200">{notice}</p>}
          <form onSubmit={createTerminal} className="grid grid-cols-1 gap-3 md:grid-cols-4">
            <input required pattern="A[A-Z0-9]{3}-[A-Z0-9]{11}" minLength={16} maxLength={16} value={keychipId} onChange={(event) => setKeychipId(event.target.value.toUpperCase())} placeholder="Axxx-xxxxxxxxxxx" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
            <input value={name} onChange={(event) => setName(event.target.value)} placeholder="机台名称，例如：一号机" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
            <div className="grid grid-cols-2 gap-2">
              <input value={gameID} onChange={(event) => setGameID(event.target.value.toUpperCase())} placeholder="游戏 ID" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
              <input value={gameVersion} onChange={(event) => setGameVersion(event.target.value)} placeholder="版本（可选）" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <Button isDisabled={isSubmitting} type="submit" className="bg-indigo-600 hover:bg-indigo-500"><Plus size={16} /> {isSubmitting ? '正在绑定…' : '绑定机台'}</Button>
          </form>
        </CardContent>
      </Card>
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {terminals.length ? terminals.map((terminal) => (
          <Card key={terminal.ID} className="bg-slate-900 border-slate-800">
            <CardContent className="space-y-4 p-5">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-bold text-white">{terminal.name || '未命名机台'}</p>
                  <p className="mt-1 font-mono text-xs text-indigo-300">登记 Keychip：{terminal.keychipId}</p>{terminal.lastSeenKeychip && terminal.lastSeenKeychip.replaceAll('-', '') !== terminal.keychipId.replaceAll('-', '') ? <p className="mt-1 font-mono text-xs text-amber-300">最近上报：{terminal.lastSeenKeychip.replaceAll('-', '')}</p> : null}
                </div>
                <Badge className={terminal.isEnabled ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/25' : 'bg-rose-500/15 text-rose-300 border-rose-500/25'}>{terminal.isEnabled ? '已启用' : '已停用'}</Badge>
              </div>
              <div className="grid grid-cols-2 gap-2 text-xs text-slate-400">
                <p>游戏：<span className="text-slate-200">{terminal.gameId}</span></p>
                <p>版本：<span className="text-slate-200">{terminal.gameVersion || '未上报'}</span></p>
                <p className="col-span-2">最后连接：<span className="text-slate-200">{terminal.lastSeenAt ? new Date(terminal.lastSeenAt).toLocaleString('zh-CN') : '从未连接'}</span></p>
              </div>
              <div className="flex gap-2">
                <Button type="button" size="sm" variant="outline" className="border-slate-700" onClick={() => void updateTerminal(terminal, !terminal.isEnabled)}><Power size={15} /> {terminal.isEnabled ? '停用' : '启用'}</Button>
                <Button type="button" size="sm" variant="outline" className="border-rose-900/70 text-rose-300 hover:bg-rose-950/40" onClick={() => void deleteTerminal(terminal)}><Trash2 size={15} /> 解除绑定</Button>
              </div>
            </CardContent>
          </Card>
        )) : <p className="text-sm text-slate-500">尚未绑定机台。请先登记机台 Keychip 序列号。</p>}
      </div>
    </section>
  )
}
