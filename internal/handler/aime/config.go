package aime

import (
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func maimaiConfigValue(key, fallback string) string {
	var config model.SystemConfig
	if err := database.DB.Where(&model.SystemConfig{Key: key}).First(&config).Error; err != nil {
		return fallback
	}
	value := strings.TrimSpace(config.Value)
	if value == "" {
		return fallback
	}
	return value
}
