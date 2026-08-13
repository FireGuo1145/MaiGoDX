import { useEffect, useState, type FormEvent } from 'react'
import { CalendarDays, CircleDollarSign, Plus, Save, Trash2 } from 'lucide-react'
import { api } from '@/lib/api'
import { apiErrorMessage, type GameCharge, type GameEvent } from '@/types'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface GameDataPanelProps {
  events: GameEvent[]
  charges: GameCharge[]
  onEventsChanged: () => Promise<void>
  onChargesChanged: () => Promise<void>
}

const blankEvent = (): Omit<GameEvent, 'ID'> => ({ type: 0, startDate: '', endDate: '', disableArea: '' })
const blankCharge = (): GameCharge => ({ chargeId: 0, orderId: 0, price: 0, startDate: '', endDate: '' })

function EventEditor({ event, onSaved, onDeleted }: { event: GameEvent; onSaved: (event: GameEvent) => Promise<void>; onDeleted: (event: GameEvent) => Promise<void> }) {
  const [draft, setDraft] = useState(event)
  const [busy, setBusy] = useState(false)

  useEffect(() => setDraft(event), [event])

  return (
    <Card className="border-slate-800 bg-slate-900">
      <CardContent className="space-y-3 p-4">
        <div className="flex items-center justify-between gap-3">
          <Badge className="border-indigo-500/25 bg-indigo-500/15 text-indigo-200">事件 #{event.ID}</Badge>
          <Button type="button" size="sm" variant="outline" className="border-rose-900/70 text-rose-300 hover:bg-rose-950/40" isDisabled={busy} onClick={() => void onDeleted(event)}><Trash2 size={15} /> 删除</Button>
        </div>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <input type="number" value={draft.type} onChange={(item) => setDraft({ ...draft, type: Number(item.target.value) })} placeholder="事件类型" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
          <input value={draft.disableArea} onChange={(item) => setDraft({ ...draft, disableArea: item.target.value })} placeholder="禁用区域（可选）" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
          <input value={draft.startDate} onChange={(item) => setDraft({ ...draft, startDate: item.target.value })} placeholder="开始时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
          <input value={draft.endDate} onChange={(item) => setDraft({ ...draft, endDate: item.target.value })} placeholder="结束时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <Button type="button" size="sm" className="bg-indigo-600 hover:bg-indigo-500" isDisabled={busy} onClick={async () => { setBusy(true); try { await onSaved(draft) } finally { setBusy(false) } }}><Save size={15} /> 保存事件</Button>
      </CardContent>
    </Card>
  )
}

function ChargeEditor({ charge, onSaved, onDeleted }: { charge: GameCharge; onSaved: (charge: GameCharge) => Promise<void>; onDeleted: (charge: GameCharge) => Promise<void> }) {
  const [draft, setDraft] = useState(charge)
  const [busy, setBusy] = useState(false)

  useEffect(() => setDraft(charge), [charge])

  return (
    <Card className="border-slate-800 bg-slate-900">
      <CardContent className="space-y-3 p-4">
        <div className="flex items-center justify-between gap-3">
          <Badge className="border-emerald-500/25 bg-emerald-500/15 text-emerald-200">收费 ID #{charge.chargeId}</Badge>
          <Button type="button" size="sm" variant="outline" className="border-rose-900/70 text-rose-300 hover:bg-rose-950/40" isDisabled={busy} onClick={() => void onDeleted(charge)}><Trash2 size={15} /> 删除</Button>
        </div>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
          <input type="number" value={draft.orderId} onChange={(item) => setDraft({ ...draft, orderId: Number(item.target.value) })} placeholder="排序 ID" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
          <input type="number" min="0" value={draft.price} onChange={(item) => setDraft({ ...draft, price: Number(item.target.value) })} placeholder="价格" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
          <input value={draft.startDate} onChange={(item) => setDraft({ ...draft, startDate: item.target.value })} placeholder="开始时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
          <input value={draft.endDate} onChange={(item) => setDraft({ ...draft, endDate: item.target.value })} placeholder="结束时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
        </div>
        <Button type="button" size="sm" className="bg-emerald-600 hover:bg-emerald-500" isDisabled={busy} onClick={async () => { setBusy(true); try { await onSaved(draft) } finally { setBusy(false) } }}><Save size={15} /> 保存收费项目</Button>
      </CardContent>
    </Card>
  )
}

export function GameDataPanel({ events, charges, onEventsChanged, onChargesChanged }: GameDataPanelProps) {
  const [eventDraft, setEventDraft] = useState(blankEvent)
  const [chargeDraft, setChargeDraft] = useState(blankCharge)
  const [notice, setNotice] = useState<string | null>(null)
  const [creatingEvent, setCreatingEvent] = useState(false)
  const [creatingCharge, setCreatingCharge] = useState(false)

  const createEvent = async (form: FormEvent) => {
    form.preventDefault()
    setCreatingEvent(true)
    setNotice(null)
    try {
      const result = await api.createGameEvent(eventDraft)
      if (!result.success) throw new Error(result.message || '创建游戏事件失败')
      setEventDraft(blankEvent())
      await onEventsChanged()
      setNotice('游戏事件已创建')
    } catch (error) {
      setNotice(apiErrorMessage(error))
    } finally {
      setCreatingEvent(false)
    }
  }

  const createCharge = async (form: FormEvent) => {
    form.preventDefault()
    setCreatingCharge(true)
    setNotice(null)
    try {
      const result = await api.createGameCharge(chargeDraft)
      if (!result.success) throw new Error(result.message || '创建收费项目失败')
      setChargeDraft(blankCharge())
      await onChargesChanged()
      setNotice('收费项目已创建')
    } catch (error) {
      setNotice(apiErrorMessage(error))
    } finally {
      setCreatingCharge(false)
    }
  }

  const saveEvent = async (event: GameEvent) => {
    try {
      const result = await api.updateGameEvent(event)
      if (!result.success) throw new Error(result.message || '更新游戏事件失败')
      await onEventsChanged()
      setNotice('游戏事件已保存')
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  const deleteEvent = async (event: GameEvent) => {
    if (!window.confirm(`确认删除游戏事件 #${event.ID} 吗？`)) return
    try {
      const result = await api.deleteGameEvent(event.ID)
      if (!result.success) throw new Error(result.message || '删除游戏事件失败')
      await onEventsChanged()
      setNotice('游戏事件已删除')
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  const saveCharge = async (charge: GameCharge) => {
    try {
      const result = await api.updateGameCharge(charge)
      if (!result.success) throw new Error(result.message || '更新收费项目失败')
      await onChargesChanged()
      setNotice('收费项目已保存')
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  const deleteCharge = async (charge: GameCharge) => {
    if (!window.confirm(`确认删除收费项目 #${charge.chargeId} 吗？`)) return
    try {
      const result = await api.deleteGameCharge(charge.chargeId)
      if (!result.success) throw new Error(result.message || '删除收费项目失败')
      await onChargesChanged()
      setNotice('收费项目已删除')
    } catch (error) {
      setNotice(apiErrorMessage(error))
    }
  }

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="flex items-center gap-2 text-xl font-bold"><CalendarDays /> 游戏事件与收费下发</h2>
        <div className="flex gap-2">
          <Button type="button" size="sm" variant="outline" className="border-slate-700" onClick={() => void onEventsChanged()}>刷新事件</Button>
          <Button type="button" size="sm" variant="outline" className="border-slate-700" onClick={() => void onChargesChanged()}>刷新收费</Button>
        </div>
      </div>
      {notice && <p className="rounded-md border border-indigo-500/30 bg-indigo-500/10 p-3 text-sm text-indigo-200">{notice}</p>}
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-2">
        <Card className="border-slate-800 bg-slate-900">
          <CardHeader><CardTitle className="flex items-center gap-2 text-white"><CalendarDays size={18} /> 游戏事件</CardTitle><CardDescription className="text-slate-400">配置 GetGameEventApi 下发的事件记录。</CardDescription></CardHeader>
          <CardContent className="space-y-4">
            <form onSubmit={createEvent} className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <input required type="number" value={eventDraft.type} onChange={(item) => setEventDraft({ ...eventDraft, type: Number(item.target.value) })} placeholder="事件类型" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
              <input value={eventDraft.disableArea} onChange={(item) => setEventDraft({ ...eventDraft, disableArea: item.target.value })} placeholder="禁用区域（可选）" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
              <input value={eventDraft.startDate} onChange={(item) => setEventDraft({ ...eventDraft, startDate: item.target.value })} placeholder="开始时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
              <input value={eventDraft.endDate} onChange={(item) => setEventDraft({ ...eventDraft, endDate: item.target.value })} placeholder="结束时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-indigo-500" />
              <Button type="submit" isDisabled={creatingEvent} className="sm:col-span-2 bg-indigo-600 hover:bg-indigo-500"><Plus size={16} /> {creatingEvent ? '正在创建…' : '新增游戏事件'}</Button>
            </form>
            <div className="space-y-3">{events.length ? events.map((event) => <EventEditor key={event.ID} event={event} onSaved={saveEvent} onDeleted={deleteEvent} />) : <p className="text-sm text-slate-500">暂无游戏事件。创建后将由 GetGameEventApi 下发。</p>}</div>
          </CardContent>
        </Card>
        <Card className="border-slate-800 bg-slate-900">
          <CardHeader><CardTitle className="flex items-center gap-2 text-white"><CircleDollarSign size={18} /> 收费项目</CardTitle><CardDescription className="text-slate-400">配置 GetGameChargeApi 下发的收费项目。</CardDescription></CardHeader>
          <CardContent className="space-y-4">
            <form onSubmit={createCharge} className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              <input required min="1" type="number" value={chargeDraft.chargeId || ''} onChange={(item) => setChargeDraft({ ...chargeDraft, chargeId: Number(item.target.value) })} placeholder="收费 ID" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
              <input required type="number" value={chargeDraft.orderId} onChange={(item) => setChargeDraft({ ...chargeDraft, orderId: Number(item.target.value) })} placeholder="排序 ID" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
              <input required min="0" type="number" value={chargeDraft.price} onChange={(item) => setChargeDraft({ ...chargeDraft, price: Number(item.target.value) })} placeholder="价格" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
              <div />
              <input value={chargeDraft.startDate} onChange={(item) => setChargeDraft({ ...chargeDraft, startDate: item.target.value })} placeholder="开始时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
              <input value={chargeDraft.endDate} onChange={(item) => setChargeDraft({ ...chargeDraft, endDate: item.target.value })} placeholder="结束时间" className="h-10 rounded-md border border-slate-700 bg-slate-800 px-3 text-sm text-white outline-none focus:ring-2 focus:ring-emerald-500" />
              <Button type="submit" isDisabled={creatingCharge} className="sm:col-span-2 bg-emerald-600 hover:bg-emerald-500"><Plus size={16} /> {creatingCharge ? '正在创建…' : '新增收费项目'}</Button>
            </form>
            <div className="space-y-3">{charges.length ? charges.map((charge) => <ChargeEditor key={charge.chargeId} charge={charge} onSaved={saveCharge} onDeleted={deleteCharge} />) : <p className="text-sm text-slate-500">暂无收费项目。创建后将由 GetGameChargeApi 下发。</p>}</div>
          </CardContent>
        </Card>
      </div>
    </section>
  )
}
