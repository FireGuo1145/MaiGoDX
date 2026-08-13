import {
  API_PATHS,
  cardListPath,
  getJson,
  postJson,
  type ApiResult,
  type CardsResult,
  type ConfigsResult,
  type LoginResult,
  type RegisterResult,
  type StatsResult,
  type UsersResult,
} from '@/types'

export const api = {
  getStats: () => getJson<StatsResult>(API_PATHS.stats),
  login: (email: string, password: string) => postJson<LoginResult>(API_PATHS.login, { email, password }),
  register: (email: string, password: string, username: string) =>
    postJson<RegisterResult>(API_PATHS.register, { email, password, username }),
  verifyEmail: (email: string, token: string) => postJson<ApiResult>(API_PATHS.verify, { email, token }),
  getCards: (email: string) => getJson<CardsResult>(cardListPath(email)),
  bindCard: (email: string, accessCode: string, cardName: string) =>
    postJson<ApiResult>(API_PATHS.bindCard, { email, accessCode, cardName }),
  getUsers: () => getJson<UsersResult>(API_PATHS.users),
  getConfigs: () => getJson<ConfigsResult>(API_PATHS.configs),
  updateConfig: (key: string, value: string) => postJson<ApiResult>(API_PATHS.updateConfig, { key, value }),
}

export default api
