export const TERMINAL_GAMES = [
  { id: "SDHD", label: "CHUNITHM" },
  { id: "SDEZ", label: "maimai DX" },
  { id: "SDGA", label: "maimai DX (Intl)" },
  { id: "SDED", label: "Card Maker" },
  { id: "SDDT", label: "O.N.G.E.K.I." },
  { id: "SBZV", label: "Project DIVA" },
  { id: "SDFE", label: "Wacca" },
] as const

export const DEFAULT_TERMINAL_GAME_ID = "SDEZ"

export function terminalGameLabel(gameID: string) {
  return (
    TERMINAL_GAMES.find((game) => game.id === gameID)?.label || gameID
  )
}

export function isTerminalGameID(gameID: string) {
  return TERMINAL_GAMES.some((game) => game.id === gameID)
}
