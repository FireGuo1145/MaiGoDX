package aime

import "strings"

var supportedTerminalGameIDs = map[string]struct{}{
	"SDHD": {},
	"SDEZ": {},
	"SDGA": {},
	"SDED": {},
	"SDDT": {},
	"SBZV": {},
	"SDFE": {},
}

func normalizeTerminalGameID(value string) (string, bool) {
	gameID := strings.ToUpper(strings.TrimSpace(value))
	_, supported := supportedTerminalGameIDs[gameID]
	return gameID, supported
}

func NormalizeTerminalGameID(value string) (string, bool) {
	return normalizeTerminalGameID(value)
}
