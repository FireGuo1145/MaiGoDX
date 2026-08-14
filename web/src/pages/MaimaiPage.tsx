import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import api from '@/lib/api'
import { apiErrorMessage, RANK_SUMMARY, type FunctionTicket, type Region, type Stats, type TravelPartner } from '@/types'

interface MaimaiPageProps {
  stats: Stats | null
  onProfileChanged: () => Promise<void>
}

export function MaimaiPage({ stats, onProfileChanged }: MaimaiPageProps) {
  const plays = stats?.recentPlays || []
  const travelPartners = stats?.travelPartners || []
  const functionTickets = stats?.functionTickets || []
  const regions = stats?.regions || []
  const [isEditingProfile, setIsEditingProfile] = useState(false)
  const [partnerID, setPartnerID] = useState('0')
  const [travelPartnerText, setTravelPartnerText] = useState('')
  const [ticketText, setTicketText] = useState('')
  const [regionText, setRegionText] = useState('')
  const [profileError, setProfileError] = useState('')
  const [isSavingProfile, setIsSavingProfile] = useState(false)

  const startProfileEdit = () => {
    setPartnerID(String(stats?.partner?.partnerId || 0))
    setTravelPartnerText(travelPartners.map((partner) => `${partner.partnerId}, ${partner.intimateLevel}, ${partner.intimateCountRewarded}`).join('\n'))
    setTicketText(functionTickets.map((ticket) => `${ticket.itemId}, ${ticket.stock}`).join('\n'))
    setRegionText(regions.map((region) => `${region.regionId}, ${region.playCount}`).join('\n'))
    setProfileError('')
    setIsEditingProfile(true)
  }

  const saveProfile = async () => {
    try {
      setIsSavingProfile(true)
      setProfileError('')
      await api.updateProfile({
        partnerId: parseInteger(partnerID, '搭档 ID'),
        travelPartners: parseTravelPartners(travelPartnerText),
        functionTickets: parseFunctionTickets(ticketText),
        regions: parseRegions(regionText),
      })
      await onProfileChanged()
      setIsEditingProfile(false)
    } catch (error) {
      setProfileError(apiErrorMessage(error))
    } finally {
      setIsSavingProfile(false)
    }
  }

  return (
    <div className="space-y-6">
      <Tabs defaultSelectedKey="recent" className="w-full">
        <TabsList className="bg-slate-900 border border-slate-800">
          <TabsTrigger id="recent" className="data-[selected]:bg-indigo-600">最近游玩</TabsTrigger>
          <TabsTrigger id="stats" className="data-[selected]:bg-indigo-600">成绩统计</TabsTrigger>
          <TabsTrigger id="profile" className="data-[selected]:bg-indigo-600">游戏档案</TabsTrigger>
        </TabsList>

        <TabsContent id="recent" className="mt-6">
          <Card className="bg-slate-900 border-slate-800 overflow-hidden">
            <Table>
              <TableHeader className="bg-slate-800/50">
                  <TableHead className="text-slate-300">乐曲</TableHead>
                  <TableHead className="text-slate-300">难度</TableHead>
                  <TableHead className="text-slate-300 text-right">达成率</TableHead>
                  <TableHead className="text-slate-300 text-right">分数</TableHead>
                  <TableHead className="text-slate-300 text-right">日期</TableHead>
              </TableHeader>
              <TableBody>
                {plays.length ? plays.map((play) => (
                  <TableRow key={play.ID} className="border-slate-800 hover:bg-slate-800/30">
                    <TableCell className="font-bold text-white">乐曲 ID：{play.musicId}</TableCell>
                    <TableCell><Badge variant="outline" className="border-indigo-500/50 text-indigo-400">LV.{play.level}</Badge></TableCell>
                    <TableCell className="text-right font-mono text-emerald-400">{(play.achievement / 10000).toFixed(4)}%</TableCell>
                    <TableCell className="text-right font-mono">{play.score.toLocaleString()}</TableCell>
                    <TableCell className="text-right text-slate-500 text-xs">{play.createDate || '—'}</TableCell>
                  </TableRow>
                )) : (
                  <TableRow><TableCell className="text-center py-12 text-slate-500">暂无游玩记录。</TableCell><TableCell /><TableCell /><TableCell /><TableCell /></TableRow>
                )}
              </TableBody>
            </Table>
          </Card>
        </TabsContent>

        <TabsContent id="stats" className="mt-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
            {RANK_SUMMARY.map((rank) => (
              <Card key={rank} className="bg-slate-900 border-slate-800 text-center p-6">
                <p className="text-2xl font-black text-indigo-400 mb-1">{rank}</p>
                <p className="text-sm text-slate-500 font-bold">{stats?.rankCounts?.[rank] ?? 0} 首</p>
              </Card>
            ))}
          </div>
        </TabsContent>

        <TabsContent id="profile" className="mt-6 space-y-6">
          <div className="flex justify-end">
            <Button variant={isEditingProfile ? 'outline' : 'default'} onPress={() => isEditingProfile ? setIsEditingProfile(false) : startProfileEdit()}>
              {isEditingProfile ? '取消编辑' : '编辑档案'}
            </Button>
          </div>
          {isEditingProfile ? (
            <Card className="space-y-4 border-indigo-500/40 bg-slate-900 p-5">
              <div>
                <label className="mb-1 block text-sm font-medium text-slate-300">当前搭档 ID</label>
                <input value={partnerID} onChange={(event) => setPartnerID(event.target.value)} inputMode="numeric" className="w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-white outline-none focus:border-indigo-500" />
              </div>
              <ProfileTextInput label="旅行伙伴" hint="每行：搭档 ID, 亲密度等级, 已领奖励次数" value={travelPartnerText} onChange={setTravelPartnerText} />
              <ProfileTextInput label="功能票" hint="每行：票种 ID, 库存" value={ticketText} onChange={setTicketText} />
              <ProfileTextInput label="区域游玩记录" hint="每行：区域 ID, 游玩次数" value={regionText} onChange={setRegionText} />
              {profileError && <p className="text-sm text-red-400">{profileError}</p>}
              <div className="flex justify-end gap-3">
                <Button variant="outline" onPress={() => setIsEditingProfile(false)}>取消</Button>
                <Button onPress={() => void saveProfile()} isDisabled={isSavingProfile}>{isSavingProfile ? '保存中…' : '保存档案'}</Button>
              </div>
            </Card>
          ) : null}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <Card className="border-slate-800 bg-slate-900 p-5">
              <p className="text-sm font-medium text-slate-400">当前搭档</p>
              <p className="mt-2 text-3xl font-black text-indigo-400">
                {stats?.partner?.partnerId ? `ID ${stats.partner.partnerId}` : '未装备'}
              </p>
              <p className="mt-2 text-xs text-slate-500">来自机台同步的 partnerId。</p>
            </Card>
            <Card className="border-slate-800 bg-slate-900 p-5">
              <p className="text-sm font-medium text-slate-400">功能票库存</p>
              <p className="mt-2 text-3xl font-black text-emerald-400">{functionTickets.reduce((total, ticket) => total + ticket.stock, 0)}</p>
              <p className="mt-2 text-xs text-slate-500">共 {functionTickets.length} 种功能票。</p>
            </Card>
          </div>

          <ProfileTable
            title="旅行伙伴"
            empty="尚无旅行伙伴数据。完成机台同步后会显示在这里。"
            headers={['搭档 ID', '亲密度等级', '已领奖励次数']}
            rows={travelPartners.map((partner) => [partner.partnerId, partner.intimateLevel, partner.intimateCountRewarded])}
          />
          <ProfileTable
            title="功能票"
            empty="尚未持有功能票。"
            headers={['票种 ID', '库存']}
            rows={functionTickets.map((ticket) => [ticket.itemId, ticket.stock])}
          />
          <ProfileTable
            title="区域游玩记录"
            empty="尚无区域游玩记录。"
            headers={['区域 ID', '游玩次数']}
            rows={regions.map((region) => [region.regionId, region.playCount])}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}

interface ProfileTextInputProps {
  label: string
  hint: string
  value: string
  onChange: (value: string) => void
}

function ProfileTextInput({ label, hint, value, onChange }: ProfileTextInputProps) {
  return <div>
    <label className="mb-1 block text-sm font-medium text-slate-300">{label}</label>
    <textarea value={value} onChange={(event) => onChange(event.target.value)} placeholder={hint} rows={3} className="w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm text-white outline-none focus:border-indigo-500" />
    <p className="mt-1 text-xs text-slate-500">{hint}。留空将清空该类数据。</p>
  </div>
}

function parseInteger(value: string, label: string) {
  const parsed = Number(value.trim())
  if (!Number.isInteger(parsed) || parsed < 0) throw new Error(`${label} 必须是非负整数`)
  return parsed
}

function parseFunctionTickets(value: string): FunctionTicket[] {
  return parsePairs(value, '功能票').map(([itemId, stock]) => ({ itemId, stock }))
}

function parseRegions(value: string): Region[] {
  return parsePairs(value, '区域').map(([regionId, playCount]) => ({ regionId, playCount }))
}

function parsePairs(value: string, label: string) {
  return parseRows(value, 2, label)
}

function parseTravelPartners(value: string): TravelPartner[] {
  return parseRows(value, 3, '旅行伙伴').map(([partnerId, intimateLevel, intimateCountRewarded]) => ({ partnerId, intimateLevel, intimateCountRewarded }))
}

function parseRows(value: string, width: number, label: string): number[][] {
  const rows = value.split('\n').map((line) => line.trim()).filter(Boolean)
  const ids = new Set<number>()
  return rows.map((line, index) => {
    const fields = line.split(',').map((field) => Number(field.trim()))
    if (fields.length !== width || fields.some((field) => !Number.isInteger(field) || field < 0) || ids.has(fields[0])) {
      throw new Error(`${label} 第 ${index + 1} 行格式无效或 ID 重复`)
    }
    ids.add(fields[0])
    return fields
  })
}

interface ProfileTableProps {
  title: string
  headers: string[]
  rows: Array<Array<number | string>>
  empty: string
}

function ProfileTable({ title, headers, rows, empty }: ProfileTableProps) {
  return (
    <Card className="overflow-hidden border-slate-800 bg-slate-900">
      <div className="border-b border-slate-800 px-5 py-4 text-sm font-bold text-white">{title}</div>
      <Table>
        <TableHeader className="bg-slate-800/50">
          {headers.map((header) => <TableHead key={header} className="text-slate-300">{header}</TableHead>)}
        </TableHeader>
        <TableBody>
          {rows.length ? rows.map((row, index) => (
            <TableRow key={`${title}-${index}`} className="border-slate-800">
              {row.map((value, column) => <TableCell key={column} className="font-mono text-slate-200">{value}</TableCell>)}
            </TableRow>
          )) : (
            <TableRow>
              <TableCell colSpan={headers.length} className="py-8 text-center text-sm text-slate-500">{empty}</TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </Card>
  )
}
