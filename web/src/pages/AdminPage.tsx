import { useState } from 'react'
import { Sliders, Users } from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import api from '@/lib/api'
import type { SystemConfig, UserAccount } from '@/types'
import { apiErrorMessage, initialOf, isTruthyConfig } from '@/types'

interface AdminPageProps {
  users: UserAccount[]
  configs: SystemConfig[]
  onUsersChanged: () => Promise<void>
  onConfigsChanged: () => Promise<void>
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
      setStatus('Saved')
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
            {isTruthyConfig(value) ? 'Enabled' : 'Disabled'}
          </Badge>
        )}
      </div>
      <div className="flex gap-3">
        <input value={value} onChange={(event) => setValue(event.target.value)} onBlur={save} className="flex-1 h-9 px-3 bg-slate-900 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        <Button type="button" size="sm" onClick={save} className="bg-indigo-600 hover:bg-indigo-500">Save</Button>
      </div>
      {status && <p className={`text-xs ${status === 'Saved' ? 'text-emerald-400' : 'text-rose-400'}`}>{status}</p>}
    </div>
  )
}

export function AdminPage({ users, configs, onUsersChanged, onConfigsChanged }: AdminPageProps) {
  return (
    <div className="space-y-8">
      <section className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-xl font-bold flex items-center gap-2"><Users /> User Management</h2>
          <Button type="button" onClick={onUsersChanged} size="sm" variant="outline" className="border-slate-700">Refresh List</Button>
        </div>
        <Card className="bg-slate-900 border-slate-800 overflow-hidden">
          <Table>
            <TableHeader className="bg-slate-800/50">
                <TableHead className="text-slate-300">User</TableHead>
                <TableHead className="text-slate-300">Status</TableHead>
                <TableHead className="text-slate-300">Role</TableHead>
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
                  <TableCell>{account.isVerified ? <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20">Verified</Badge> : <Badge variant="outline" className="text-slate-500 border-slate-800">Pending</Badge>}</TableCell>
                  <TableCell>{account.isAdmin ? <Badge className="bg-amber-500/10 text-amber-500 border-amber-500/20">Admin</Badge> : <Badge variant="outline" className="text-slate-500 border-slate-800">User</Badge>}</TableCell>
                </TableRow>
              )) : <TableRow><TableCell className="text-center py-10 text-slate-500">No users found.</TableCell><TableCell /><TableCell /></TableRow>}
            </TableBody>
          </Table>
        </Card>
      </section>

      <section className="space-y-6">
        <div className="flex justify-between items-center">
          <h2 className="text-xl font-bold flex items-center gap-2"><Sliders /> Server Configuration Management</h2>
          <Button type="button" onClick={onConfigsChanged} size="sm" variant="outline" className="border-slate-700">Refresh Configs</Button>
        </div>
        <Card className="bg-slate-900 border-slate-800">
          <CardHeader><CardTitle className="text-white text-sm">Global System Settings & Flags</CardTitle></CardHeader>
          <CardContent className="space-y-4">
            {configs.length ? configs.map((config) => <ConfigRow key={config.ID} config={config} onSaved={onConfigsChanged} />) : <p className="text-sm text-slate-500">No system configuration entries found.</p>}
          </CardContent>
        </Card>
      </section>
    </div>
  )
}
