import { useState } from 'react'
import { CreditCard } from 'lucide-react'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import api from '@/lib/api'
import { PersonalTerminalPanel } from '@/components/PersonalTerminalPanel'
import type { UserAccount, UserCard } from '@/types'
import { apiErrorMessage, cardPreview, DEFAULT_CARD_NAME, initialOf, isAccessCodeValid, normalizeAccessCode } from '@/types'

interface SettingsPageProps {
  user: UserAccount
  cards: UserCard[]
  onCardsChanged: () => Promise<void>
}

export function SettingsPage({ user, cards, onCardsChanged }: SettingsPageProps) {
  const [accessCode, setAccessCode] = useState('')
  const [cardName, setCardName] = useState('')
  const [gameUserId, setGameUserId] = useState('')
  const [notice, setNotice] = useState<string | null>(null)
  const [isBinding, setIsBinding] = useState(false)

  const bindCard = async (event: React.FormEvent) => {
    event.preventDefault()
    const normalizedCode = normalizeAccessCode(accessCode)

    if (!isAccessCodeValid(normalizedCode)) {
      setNotice('请输入有效的 20 位 Aime Access Code。')
      return
    }

    const parsedGameUserId = gameUserId.trim() === '' ? undefined : Number(gameUserId)
    if (parsedGameUserId !== undefined && (!Number.isSafeInteger(parsedGameUserId) || parsedGameUserId <= 0)) {
      setNotice('游戏用户 ID 必须是正整数；若尚未创建 maimai 档案可留空。')
      return
    }

    setIsBinding(true)
    setNotice(null)
    try {
      const result = await api.bindCard(user.email, normalizedCode, cardName.trim() || DEFAULT_CARD_NAME, parsedGameUserId)
      if (!result.success) throw new Error(result.message || '卡片绑定失败')
      setAccessCode('')
      setCardName('')
      setGameUserId('')
      setNotice('卡片绑定成功。')
      await onCardsChanged()
    } catch (error) {
      setNotice(apiErrorMessage(error))
    } finally {
      setIsBinding(false)
    }
  }

  return (
    <div className="max-w-2xl space-y-8">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader><CardTitle className="text-white">账户设置</CardTitle></CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center gap-6">
            <Avatar className="h-20 w-20 border-2 border-slate-800">
              <AvatarFallback className="bg-indigo-600 text-2xl font-black">{initialOf(user.username)}</AvatarFallback>
            </Avatar>
            <Button isDisabled variant="outline" className="border-slate-700">修改头像</Button>
          </div>
          <Separator className="bg-slate-800" />
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <label className="space-y-2">
              <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">用户名</span>
              <input disabled value={user.username} className="w-full h-10 px-3 bg-slate-800/50 border-slate-700 rounded-md text-slate-400 cursor-not-allowed" />
            </label>
            <label className="space-y-2">
              <span className="text-xs font-bold text-slate-500 uppercase tracking-wider">邮箱</span>
              <input disabled value={user.email} className="w-full h-10 px-3 bg-slate-800/50 border-slate-700 rounded-md text-slate-400 cursor-not-allowed" />
            </label>
          </div>
        </CardContent>
      </Card>
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader>
          <CardTitle className="text-white flex items-center gap-2"><CreditCard className="text-indigo-400" /> Aime 卡片绑定</CardTitle>
          <CardDescription className="text-slate-400">将已授权的 20 位 Aime Access Code 关联到当前账户。</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {notice && <div className="p-3 bg-indigo-500/10 border border-indigo-500/20 rounded-lg text-sm text-indigo-300">{notice}</div>}
          <form onSubmit={bindCard} className="space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <label className="space-y-2">
                <span className="text-xs font-medium text-slate-300">Access Code（20 位数字）</span>
                <input required maxLength={20} value={accessCode} onChange={(event) => setAccessCode(normalizeAccessCode(event.target.value))} placeholder="01234567890123456789" className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </label>
              <label className="space-y-2">
                <span className="text-xs font-medium text-slate-300">卡片别名</span>
                <input value={cardName} onChange={(event) => setCardName(event.target.value)} placeholder="我的主卡" className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </label>
              <label className="space-y-2 sm:col-span-2">
                <span className="text-xs font-medium text-slate-300">maimai 游戏用户 ID（可选）</span>
                <input inputMode="numeric" value={gameUserId} onChange={(event) => setGameUserId(event.target.value.replace(/\D/g, ''))} placeholder="在游戏创建档案后填写外部用户 ID" className="w-full h-10 px-3 bg-slate-800 border border-slate-700 rounded-md text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                <span className="block text-xs text-slate-500">游戏创建档案前可留空；之后补充该 ID，即可在门户显示该档案的真实成绩。</span>
              </label>
            </div>
            <Button isDisabled={isBinding} type="submit" className="bg-indigo-600 hover:bg-indigo-500 text-white font-bold">{isBinding ? '正在绑定…' : '绑定卡片'}</Button>
          </form>

          <Separator className="bg-slate-800" />
          <section className="space-y-3">
            <h3 className="font-bold text-sm text-white">已绑定的卡片</h3>
            {cards.length ? cards.map((card) => (
              <div key={card.ID} className="flex items-center justify-between p-3 bg-slate-800/60 rounded-lg gap-4">
                <div>
                  <p className="font-bold text-white text-sm">{card.cardName || DEFAULT_CARD_NAME}</p>
                  <p className="font-mono text-xs text-indigo-400">{cardPreview(card.accessCode)}</p>
                  <p className="font-mono text-xs text-slate-500">游戏用户 ID：{card.gameUserId || '未关联'}</p>
                </div>
                <Badge className="bg-emerald-500/10 text-emerald-500 border-emerald-500/20">已绑定</Badge>
              </div>
            )) : <p className="text-sm text-slate-500">暂未绑定任何卡片。</p>}
          </section>
        </CardContent>
      </Card>
      <PersonalTerminalPanel />
    </div>
  )
}
