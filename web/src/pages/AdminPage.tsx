import { useState } from 'react'
import { Sliders, Users } from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import api from '@/lib/api'
import type { GameCharge, GameEvent, SystemConfig, Terminal, UserAccount } from '@/types'
import { TerminalPanel } from '@/components/admin/TerminalPanel'
import { GameDataPanel } from '@/components/admin/GameDataPanel'
import { apiErrorMessage, initialOf, isTruthyConfig } from '@/types'

interface AdminPageProps {
  users: UserAccount[]
  configs: SystemConfig[]
  terminals: Terminal[]
  events: GameEvent[]
  charges: GameCharge[]
  onUsersChanged: () => Promise<void>
  onConfigsChanged: () => Promise<void>
  onTerminalsChanged: () => Promise<void>
  onEventsChanged: () => Promise<void>
  onChargesChanged: () => Promise<void>
}

function ConfigRow({ config, onSaved }: { config: SystemConfig; onSaved: () => Promise<void> }) {
  const [value, setValue] = useState(config.value)
  const [status, setStatus] = useState<string | null>(null)

  const save = async () => {
    if (value === config.value) return
    setStatus(null)
    try {
      const result = await api.updateConfig(config.key, value)
      if (!result.success) throw new Error(result.message || '配置保存失败')
      setStatus('已保存')
      await onSaved()
    } catch (error) {
      setStatus(apiErrorMessage(error))
    }
  }

  return (
    <div className="p-4 bg-slate-800/50 rounded-lg space-y-3">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="font-mono font-bold text-indigo-400 text-sm">{config.key}</p>
          <p className="text-xs text-slate-400 mt-1">{config.desc}</p>
        </div>
        {config.key.endsWith('_mode') && (
          <Badge className={isTruthyConfig(value) ? 'bg-amber-500/15 text-amber-300 border-amber-500/25' : 'bg-emerald-500/15 text-emerald-300 border-emerald-500/25'}>
            {isTruthyConfig(value) ? '已启用' : '已禁用'}
          </Badge>
        )}
      </div>
      <div className="flex gap-3">
        <input value={value} onChange={(event) => setValue(event.target.value)} onBlur={save} className="flex-1 h-9 px-3 bg-slate-900 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        <Button type="button" size="sm" onClick={save} className="bg-indigo-600 hover:bg-indigo-500">保存</Button>
      </div>
      {status && <p className={`text-xs ${status === '已保存' ? 'text-emerald-400' : 'text-rose-400'}`}>{status}</p>}
    </div>
  )
}

export function AdminPage({ users, configs, terminals, events, charges, onUsersChanged, onConfigsChanged, onTerminalsChanged, onEventsChanged, onChargesChanged }: AdminPageProps) {
  return (
    <div className="space-y-8">
      <section className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-xl font-bold flex items-center gap-2"><Users /> 用户管理</h2>
          <Button type="button" onClick={onUsersChanged} size="sm" variant="outline" className="border-slate-700">刷新列表</Button>
        </div>
        <Card className="bg-slate-900 border-slate-800 overflow-hidden">
          <Table>
            <TableHeader className="bg-slate-800/50">
                <TableHead isRowHeader className="text-slate-300">用户</TableHead>
                <TableHead className="text-slate-300">状态</TableHead>
                <TableHead className="text-slate-300">角色</TableHead>
            </TableHeader>
            <TableBody>
              {users.length ? users.map((account) => (
                <TableRow key={account.ID} className="border-slate-800 hover:bg-slate-800/30">
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar className="h-8 w-8"><AvatarFallback>{initialOf(account.username)}</AvatarFallback></Avatar>
                      <div><p className="font-bold text-white text-sm">{account.username}</p><p className="text-[10px] text-slate-500">{account.email}</p></div>
                    </div>
                  </TableCell>
                  <TableCell>{account.isVerified ? <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20">已验证</Badge> : <Badge variant="outline" className="text-slate-500 border-slate-800">待验证</Badge>}</TableCell>
                  <TableCell>{account.isAdmin ? <Badge className="bg-amber-500/10 text-amber-500 border-amber-500/20">管理员</Badge> : <Badge variant="outline" className="text-slate-500 border-slate-800">用户</Badge>}</TableCell>
                </TableRow>
              )) : <TableRow><TableCell className="text-center py-10 text-slate-500">未找到用户。</TableCell><TableCell /><TableCell /></TableRow>}
            </TableBody>
          </Table>
        </Card>
      </section>

      <TerminalPanel terminals={terminals} onChanged={onTerminalsChanged} />
      <GameDataPanel events={events} charges={charges} onEventsChanged={onEventsChanged} onChargesChanged={onChargesChanged} />
      <section className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-xl font-bold flex items-center gap-2"><Sliders /> 服务器配置管理</h2>
          <Button type="button" onClick={onConfigsChanged} size="sm" variant="outline" className="border-slate-700">刷新配置</Button>
        </div>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader><CardTitle className="text-white text-sm">全局系统设置与开关</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            {configs.length ? configs.map((config) => <ConfigRow key={config.ID} config={config} onSaved={onConfigsChanged} />) : <p className="text-sm text-slate-500">暂无系统配置项。</p>}
          </CardContent>
        </Card>
      </section>
    </div>
  )
}
