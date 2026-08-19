import {
  API_PATHS,
  getJson,
  postJson,
  requestJson,
  type ApiResult,
  type CardsResult,
  type ConfigsResult,
  type FunctionTicketAdjust,
  type LoginResult,
  type MetadataItem,
  type MetadataResult,
  type ProfileUpdate,
  type RegisterResult,
  type SiteSettingsResult,
  type StatsResult,
  type ChuniStatsResult,
  type OngekiStatsResult,
  type UsersResult,
  type Terminal,
  type TerminalsResult,
  type GameEvent,
  type GameCharge,
  type EventsResult,
  type ChargesResult,
} from "@/types"

export const api = {
  getSiteSettings: () => getJson<SiteSettingsResult>(API_PATHS.site),
  getSiteMetadata: () => getJson<MetadataResult>(API_PATHS.siteMetadata),
  getStats: (cardId?: number) =>
    getJson<StatsResult>(
      cardId ? `${API_PATHS.stats}?cardId=${cardId}` : API_PATHS.stats
    ),
  getChuniStats: (cardId?: number) =>
    getJson<ChuniStatsResult>(
      cardId ? `${API_PATHS.chuniStats}?cardId=${cardId}` : API_PATHS.chuniStats
    ),
  getOngekiStats: (cardId?: number) =>
    getJson<OngekiStatsResult>(
      cardId
        ? `${API_PATHS.ongekiStats}?cardId=${cardId}`
        : API_PATHS.ongekiStats
    ),
  updateProfile: (profile: ProfileUpdate) =>
    postJson<ApiResult>(API_PATHS.updateProfile, profile),
  adjustFunctionTicket: (ticket: FunctionTicketAdjust) =>
    postJson<ApiResult>(API_PATHS.adjustTicket, ticket),
  login: (email: string, password: string) =>
    postJson<LoginResult>(API_PATHS.login, { email, password }),
  currentUser: () => getJson<LoginResult>(API_PATHS.me),
  register: (email: string, password: string, username: string) =>
    postJson<RegisterResult>(API_PATHS.register, { email, password, username }),
  verifyEmail: (email: string, token: string) =>
    postJson<ApiResult>(API_PATHS.verify, { email, token }),
  logout: () => postJson<ApiResult>("/api/auth/logout", {}),
  getCards: () => getJson<CardsResult>(API_PATHS.cards),
  bindCard: (
    email: string,
    accessCode: string,
    cardName: string,
    gameUserId?: number
  ) =>
    postJson<ApiResult>(API_PATHS.bindCard, {
      email,
      accessCode,
      cardName,
      gameUserId,
    }),
  getUsers: () => getJson<UsersResult>(API_PATHS.users),
  getTerminals: () => getJson<TerminalsResult>(API_PATHS.terminals),
  createTerminal: (
    terminal: Pick<Terminal, "keychipId" | "name" | "gameId" | "gameVersion">
  ) => postJson<ApiResult>(API_PATHS.createTerminal, terminal),
  updateTerminal: (terminal: Terminal) =>
    postJson<ApiResult>(API_PATHS.updateTerminal, terminal),
  deleteTerminal: (id: number) =>
    postJson<ApiResult>(API_PATHS.deleteTerminal, { id }),
  getUserTerminals: () => getJson<TerminalsResult>(API_PATHS.userTerminals),
  createUserTerminal: (
    terminal: Pick<Terminal, "keychipId" | "name" | "gameId" | "gameVersion">
  ) => postJson<ApiResult>(API_PATHS.createUserTerminal, terminal),
  updateUserTerminal: (terminal: Terminal) =>
    postJson<ApiResult>(API_PATHS.updateUserTerminal, terminal),
  deleteUserTerminal: (id: number) =>
    postJson<ApiResult>(API_PATHS.deleteUserTerminal, { id }),
  getGameEvents: () => getJson<EventsResult>(API_PATHS.events),
  createGameEvent: (event: Omit<GameEvent, "ID">) =>
    postJson<ApiResult>(API_PATHS.createEvent, event),
  updateGameEvent: (event: GameEvent) =>
    postJson<ApiResult>(API_PATHS.updateEvent, event),
  deleteGameEvent: (id: number) =>
    postJson<ApiResult>(API_PATHS.deleteEvent, { id }),
  getGameCharges: () => getJson<ChargesResult>(API_PATHS.charges),
  createGameCharge: (charge: GameCharge) =>
    postJson<ApiResult>(API_PATHS.createCharge, charge),
  updateGameCharge: (charge: GameCharge) =>
    postJson<ApiResult>(API_PATHS.updateCharge, charge),
  deleteGameCharge: (chargeId: number) =>
    postJson<ApiResult>(API_PATHS.deleteCharge, { chargeId }),
  getConfigs: () => getJson<ConfigsResult>(API_PATHS.configs),
  updateConfig: (key: string, value: string) =>
    postJson<ApiResult>(API_PATHS.updateConfig, { key, value }),
  getMetadata: () => getJson<MetadataResult>(API_PATHS.metadata),
  saveMetadata: (dataName: string, items: MetadataItem[]) =>
    postJson<ApiResult>(API_PATHS.metadata, { dataName, items }),
  importMetadata: async (file: File, game = "maimai") => {
    const form = new FormData()
    form.append("file", file)
    form.append("game", game)
    return requestJson<ApiResult>(API_PATHS.metadataImport, {
      method: "POST",
      body: form,
    })
  },
}

export default api
