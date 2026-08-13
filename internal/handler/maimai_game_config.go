package handler

import (
	"strconv"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func gameEventPayload() map[string]interface{} {
	var events []model.GameEvent
	database.DB.Order("id asc").Find(&events)
	return map[string]interface{}{"type": 1, "gameEventList": events}
}

func gameChargePayload() map[string]interface{} {
	var charges []model.GameCharge
	database.DB.Order("order_id asc, charge_id asc").Find(&charges)
	return map[string]interface{}{"length": len(charges), "gameChargeList": charges}
}

func recommendedMusicPayload(userID int64, configKey, responseKey string) map[string]interface{} {
	return map[string]interface{}{
		"userId":    userID,
		responseKey: configuredMusicIDs(configKey),
	}
}

// configuredMusicIDs parses a comma-separated list managed through SystemConfig.
// Invalid, duplicate, and non-positive IDs are excluded instead of being sent to cabs.
func configuredMusicIDs(configKey string) []int {
	var config model.SystemConfig
	if err := database.DB.Where("key = ?", configKey).First(&config).Error; err != nil {
		return []int{}
	}

	seen := make(map[int]struct{})
	musicIDs := make([]int, 0)
	for _, raw := range strings.Split(config.Value, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		musicIDs = append(musicIDs, id)
	}
	return musicIDs
}
