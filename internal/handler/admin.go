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
