export type PageId = 'home' | 'dashboard' | 'maimai' | 'setup' | 'admin' | 'settings'
export type AuthMode = 'login' | 'register' | 'verify'

export interface UserAccount {
  ID: number
  email: string
  username: string
  isVerified: boolean
  isAdmin: boolean
}

export interface UserCard {
  ID: number
  accessCode: string
  cardName: string
  gameUserId: number
}

export interface SystemConfig {
  ID: number
  key: string
  value: string
  desc: string
}

export interface Terminal {
  ID: number
  keychipId: string
  name: string
  gameId: string
  gameVersion: string
  ownerAccountId: number
  isEnabled: boolean
  lastSeenAt: string
  lastSeenIp: string
}

export interface GameEvent {
  ID: number
  type: number
  startDate: string
  endDate: string
  disableArea: string
}
export interface GameCharge {
  chargeId: number
  orderId: number
  price: number
  startDate: string
  endDate: string
}export interface Playlog {
  ID: number
  musicId: number
  level: number
  achievement: number
  score: number
  createDate: string
  beforeRating: number
  afterRating: number
}

export interface UserDetail {
  userId: number
  userName: string
  playerRating: number
  highestRating: number
  totalPoint: number
}

export interface Partner {
  partnerId: number
}

export interface TravelPartner {
  partnerId: number
  intimateLevel: number
  intimateCountRewarded: number
}

export interface FunctionTicket {
  itemId: number
  stock: number
}

export interface Region {
  regionId: number
  playCount: number
}

export interface ProfileUpdate {
  cardId: number
  partnerId: number
  travelPartners: TravelPartner[]
  functionTickets: FunctionTicket[]
  regions: Region[]
}

export interface FunctionTicketAdjust {
  cardId: number
  itemId: number
  amount: number
}

export interface SongComp {
  musicId: number
  level: number
  achievement: number
  score: number
  scoreRank: number
}

export interface TrendPoint {
  date: string
  rating: number
}

export interface RatingComposition {
  bests: SongComp[]
  newBests: SongComp[]
}

export interface Stats {
  totalUsers: number
  totalPlays: number
  user?: UserDetail
  selectedCardId?: number
  partner?: Partner
  travelPartners: TravelPartner[]
  functionTickets: FunctionTicket[]
  regions: Region[]
  recentPlays: Playlog[]
  trend: TrendPoint[]
  rankCounts: Record<'SSS+' | 'SSS' | 'SS' | 'S', number>
  ratingComposition: RatingComposition
  message?: string
}

export interface ApiResult {
  success: boolean
  message?: string
}

export interface LoginResult extends ApiResult {
  email: string
  username: string
  isAdmin?: boolean
}

export interface RegisterResult extends ApiResult {
  verifyToken?: string
}

export interface CardsResult extends ApiResult {
  cards: UserCard[]
}

export interface UsersResult extends ApiResult {
  users: UserAccount[]
}

export interface ConfigsResult extends ApiResult {
  configs: SystemConfig[]
}

export interface TerminalsResult extends ApiResult {
  terminals: Terminal[]
}
export interface EventsResult extends ApiResult {
  events: GameEvent[]
}
export interface ChargesResult extends ApiResult {
  charges: GameCharge[]
}
export interface StatsResult extends ApiResult, Stats {}

export interface AuthNotice {
  text: string
  developmentToken?: string
}

export const API_PATHS = {
  stats: '/api/stats',
  updateProfile: '/api/maimai/profile/update',
  adjustTicket: '/api/maimai/profile/ticket/adjust',
  login: '/api/auth/login',
  me: '/api/auth/me',
  register: '/api/auth/register',
  verify: '/api/auth/verify',
  cards: '/api/card/list',
  bindCard: '/api/card/bind',
  users: '/api/admin/users',
  terminals: '/api/admin/terminals',
  createTerminal: '/api/admin/terminal/create',
  updateTerminal: '/api/admin/terminal/update',
  deleteTerminal: '/api/admin/terminal/delete',
  userTerminals: '/api/terminal/list',
  createUserTerminal: '/api/terminal/create',
  updateUserTerminal: '/api/terminal/update',
  deleteUserTerminal: '/api/terminal/delete',
  events: '/api/admin/events',
  createEvent: '/api/admin/event/create',
  updateEvent: '/api/admin/event/update',
  deleteEvent: '/api/admin/event/delete',
  charges: '/api/admin/charges',
  createCharge: '/api/admin/charge/create',
  updateCharge: '/api/admin/charge/update',
  deleteCharge: '/api/admin/charge/delete',
  configs: '/api/admin/config/get',
  updateConfig: '/api/admin/config/update',
} as const

export const DEFAULT_ADMIN_EMAIL = 'admin@maigodx.local'
export const DEFAULT_CARD_NAME = 'My Aime Card'
export const RANK_SUMMARY = ['SSS+', 'SSS', 'SS', 'S'] as const

export const SETUP_STEPS = [
  { title: '配置 Hosts / DNS', body: '将 ALL.Net 域名请求指向 MaiGoDX 服务器 IP 地址。本地调试时可使用回环地址。' },
  { title: '启动游戏客户端', body: '启动游戏客户端，并确认授权流程已在本地服务器实例上完成。' },
  { title: '进入管理门户', body: '登录后可查看成绩数据、评级构成、绑定 Aime 卡片并管理账户设置。' },
] as const

export const cardListPath = (email: string) => `${API_PATHS.cards}?email=${encodeURIComponent(email)}`

export const requestJson = async <T>(path: string, init?: RequestInit): Promise<T> => {
  const response = await fetch(path, init)
  const payload = (await response.json()) as T & ApiResult
  if (!response.ok) throw new Error(payload.message || '请求失败')
  return payload
}

export const getJson = <T>(path: string) => requestJson<T>(path)
export const postJson = <T>(path: string, body: unknown) => requestJson<T>(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
export const initialOf = (username?: string) => username?.trim().charAt(0).toUpperCase() || 'U'
export const formatAchievement = (value: number) => `${(value / 10000).toFixed(4)}%`
export const normalizeAccessCode = (value: string) => value.replace(/\D/g, '')
export const isAccessCodeValid = (value: string) => /^\d{20}$/.test(value)
export const cardPreview = (value: string) => value.length > 8 ? `${value.slice(0, 4)} •••• •••• ${value.slice(-4)}` : value
export const isTruthyConfig = (value: string) => value.trim().toLowerCase() === 'true'
export const apiErrorMessage = (error: unknown) => error instanceof Error ? error.message : '网络错误，请稍后重试。'
