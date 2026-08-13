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
}

export interface SystemConfig {
  ID: number
  key: string
  value: string
  desc: string
}

export interface Playlog {
  ID: number
  musicId: number
  level: number
  achievement: number
  score: number
  createDate: string
}

export interface UserDetail {
  UserID: number
  UserName: string
  Rating: number
  MaxRating: number
  TotalPoint: number
}

export interface SongComp {
  title: string
  level: string
  score: number
  rating: number
}

export interface Stats {
  totalUsers: number
  totalPlays: number
  recentPlays: Playlog[]
  user: UserDetail
  ratingComposition: {
    bests: SongComp[]
    newBests: SongComp[]
  }
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

export interface StatsResult extends ApiResult, Stats {}

export interface AuthNotice {
  text: string
  developmentToken?: string
}

export const API_PATHS = {
  stats: '/api/stats',
  login: '/api/auth/login',
  register: '/api/auth/register',
  verify: '/api/auth/verify',
  cards: '/api/card/list',
  bindCard: '/api/card/bind',
  users: '/api/admin/users',
  configs: '/api/admin/config/get',
  updateConfig: '/api/admin/config/update',
} as const

export const DEFAULT_ADMIN_EMAIL = 'admin@maigodx.local'
export const DEFAULT_CARD_NAME = 'My Aime Card'
export const DEFAULT_USER_RATING = 15420
export const DEFAULT_MAX_RATING = 16000
export const DEFAULT_PLAY_COUNT = 1248

export const RATING_TREND = [
  { name: 'Mon', rating: 15100 },
  { name: 'Tue', rating: 15150 },
  { name: 'Wed', rating: 15120 },
  { name: 'Thu', rating: 15200 },
  { name: 'Fri', rating: 15350 },
  { name: 'Sat', rating: 15400 },
  { name: 'Sun', rating: 15420 },
]

export const RANK_SUMMARY = ['SSS+', 'SSS', 'SS', 'S']

export const SETUP_STEPS = [
  {
    title: 'Configure Hosts / DNS',
    body: 'Redirect ALL.Net domain requests to your MaiGoDX server IP address. For a local setup, this can be your loopback address.',
  },
  {
    title: 'Boot Game Client',
    body: 'Start the game client and confirm that authorization completes against the local server instance.',
  },
  {
    title: 'Access Portal',
    body: 'Sign in to review score data, inspect rating composition, bind Aime cards, and manage account settings.',
  },
] as const

export const cardListPath = (email: string) => `${API_PATHS.cards}?email=${encodeURIComponent(email)}`

export const requestJson = async <T>(path: string, init?: RequestInit): Promise<T> => {
  const response = await fetch(path, init)
  const payload = (await response.json()) as T & ApiResult

  if (!response.ok) {
    throw new Error(payload.message || '请求失败')
  }
  return payload
}

export const getJson = <T>(path: string) => requestJson<T>(path)

export const postJson = <T>(path: string, body: unknown) =>
  requestJson<T>(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

export const initialOf = (username?: string) => username?.trim().charAt(0).toUpperCase() || 'U'
export const formatAchievement = (value: number) => `${(value / 10000).toFixed(4)}%`
export const normalizeAccessCode = (value: string) => value.replace(/\D/g, '')
export const isAccessCodeValid = (value: string) => /^\d{20}$/.test(value)
export const cardPreview = (value: string) => value.length > 8 ? `${value.slice(0, 4)} •••• •••• ${value.slice(-4)}` : value
export const isTruthyConfig = (value: string) => value.trim().toLowerCase() === 'true'
export const apiErrorMessage = (error: unknown) => error instanceof Error ? error.message : '网络错误，请稍后重试。'
