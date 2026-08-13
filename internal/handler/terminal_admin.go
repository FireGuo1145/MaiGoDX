package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

type terminalRequest struct {
	ID             uint   `json:"id"`
	KeychipID      string `json:"keychipId"`
	Name           string `json:"name"`
	GameID         string `json:"gameId"`
	GameVersion    string `json:"gameVersion"`
	OwnerAccountID uint   `json:"ownerAccountId"`
	IsEnabled      bool   `json:"isEnabled"`
}

func HandleAdminTerminals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var terminals []model.Terminal
	database.DB.Order("id asc").Find(&terminals)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "terminals": terminals})
}

func HandleCreateTerminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var request terminalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeTerminalError(w, http.StatusBadRequest, "机台参数无效")
		return
	}
	keychip := normalizeKeychip(request.KeychipID)
	if len(keychip) < 11 || len(keychip) > 32 {
		writeTerminalError(w, http.StatusBadRequest, "Keychip 序列号长度必须为 11 至 32 位")
		return
	}
	gameID := strings.ToUpper(strings.TrimSpace(request.GameID))
	if gameID == "" {
		gameID = "SDEZ"
	}
	owner := request.OwnerAccountID
	if owner == 0 {
		owner = account.ID
	}
	terminal := model.Terminal{KeychipID: keychip, Name: strings.TrimSpace(request.Name), GameID: gameID, GameVersion: strings.TrimSpace(request.GameVersion), OwnerAccountID: owner, IsEnabled: true}
	// Older revisions used GORM soft deletion. Restore such an old Keychip row
	// in place so an administrator can bind the same physical cabinet again.
	var deleted model.Terminal
	if err := database.DB.Unscoped().Where("keychip_id = ?", keychip).First(&deleted).Error; err == nil && deleted.DeletedAt.Valid {
		terminal.ID = deleted.ID
		if err := database.DB.Unscoped().Model(&deleted).Updates(map[string]interface{}{
			"deleted_at": nil, "name": terminal.Name, "game_id": terminal.GameID,
			"game_version": terminal.GameVersion, "owner_account_id": terminal.OwnerAccountID,
			"is_enabled": true,
		}).Error; err != nil {
			writeTerminalError(w, http.StatusInternalServerError, "恢复机台绑定失败")
			return
		}
		_ = database.DB.First(&terminal, deleted.ID).Error
	} else if err := database.DB.Create(&terminal).Error; err != nil {
		writeTerminalError(w, http.StatusConflict, "该 Keychip 已被绑定")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "terminal": terminal, "message": "机台绑定成功"})
}

func HandleUpdateTerminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var request terminalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ID == 0 {
		writeTerminalError(w, http.StatusBadRequest, "机台参数无效")
		return
	}
	var terminal model.Terminal
	if err := database.DB.First(&terminal, request.ID).Error; err != nil {
		writeTerminalError(w, http.StatusNotFound, "机台不存在")
		return
	}
	terminal.Name = strings.TrimSpace(request.Name)
	terminal.GameID = strings.ToUpper(strings.TrimSpace(request.GameID))
	terminal.GameVersion = strings.TrimSpace(request.GameVersion)
	terminal.OwnerAccountID = request.OwnerAccountID
	terminal.IsEnabled = request.IsEnabled
	if terminal.GameID == "" {
		terminal.GameID = "SDEZ"
	}
	if err := database.DB.Save(&terminal).Error; err != nil {
		writeTerminalError(w, http.StatusInternalServerError, "保存机台失败")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "terminal": terminal, "message": "机台已更新"})
}

func HandleDeleteTerminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var request terminalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ID == 0 {
		writeTerminalError(w, http.StatusBadRequest, "机台参数无效")
		return
	}
	database.DB.Where("terminal_id = ?", request.ID).Delete(&model.TerminalSession{})
	if result := database.DB.Unscoped().Delete(&model.Terminal{}, request.ID); result.Error != nil || result.RowsAffected == 0 {
		writeTerminalError(w, http.StatusNotFound, "机台不存在")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "机台绑定已删除"})
}

func writeTerminalError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": message})
}
