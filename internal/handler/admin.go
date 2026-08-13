package handler

import (
	"encoding/json"
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

// HandleAdminUsers 获取系统所有用户列表（管理员专属）
func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var users []model.UserAccount
	if err := database.DB.Find(&users).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "获取用户列表失败"})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"users":   users,
	})
}

// HandleGetConfigs 获取系统全局配置（下发给机台或前端）
func HandleGetConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var configs []model.SystemConfig
	database.DB.Find(&configs)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"configs": configs,
	})
}

// HandleUpdateConfig 管理员更新系统配置
func HandleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req model.UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	var config model.SystemConfig
	if err := database.DB.Where("key = ?", req.Key).First(&config).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "配置项不存在"})
		return
	}

	config.Value = req.Value
	database.DB.Save(&config)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "配置更新成功",
	})
}

// HandleBindCard 绑定 Aime 卡片
func HandleBindCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req model.BindCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	var user model.UserAccount
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "用户不存在"})
		return
	}

	var existingCard model.UserCard
	if err := database.DB.Where("access_code = ?", req.AccessCode).First(&existingCard).Error; err == nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "该卡片已被绑定"})
		return
	}

	cardName := req.CardName
	if cardName == "" {
		cardName = "My Aime Card"
	}

	card := model.UserCard{
		UserID:     user.ID,
		AccessCode: req.AccessCode,
		CardName:   cardName,
	}

	if err := database.DB.Create(&card).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "卡片绑定失败"})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "卡片绑定成功",
		"card":    card,
	})
}

// HandleGetUserCards 获取用户的卡片列表
func HandleGetUserCards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	email := r.URL.Query().Get("email")
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "缺少 email 参数"})
		return
	}

	var user model.UserAccount
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "用户不存在"})
		return
	}

	var cards []model.UserCard
	database.DB.Where("user_id = ?", user.ID).Find(&cards)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cards":   cards,
	})
}
