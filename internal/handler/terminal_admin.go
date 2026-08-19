package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/handler/aime"
	"github.com/FireGuo1145/MaiGoDX/internal/model"

	"gorm.io/gorm"
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
	keychipInput := strings.TrimSpace(request.KeychipID)
	if !aime.IsKeychipRegistrationFormat(keychipInput) {
		writeTerminalError(w, http.StatusBadRequest, "Keychip 格式必须为 Axxx-xxxxxxxxxxx")
		return
	}
	keychip := aime.FormatKeychip(keychipInput)
	gameID, supported := aime.NormalizeTerminalGameID(request.GameID)
	if !supported {
		writeTerminalError(w, http.StatusBadRequest, "不支持的游戏 ID")
		return
	}
	owner := request.OwnerAccountID
	if owner == 0 {
		owner = account.ID
	}
	terminal := model.Terminal{KeychipID: keychip, Name: strings.TrimSpace(request.Name), GameID: gameID, GameVersion: strings.TrimSpace(request.GameVersion), OwnerAccountID: owner, IsEnabled: true}
	existing, found, lookupErr := aime.FindStoredTerminalByKeychipPrefix(keychip)
	if lookupErr != nil {
		writeTerminalError(w, http.StatusInternalServerError, "检查 Keychip 失败")
		return
	}
	if found {
		if !existing.DeletedAt.Valid {
			writeTerminalError(w, http.StatusConflict, "该 Keychip 前缀已被绑定")
			return
		}
		terminal.ID = existing.ID
		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("terminal_id = ?", existing.ID).Delete(&model.TerminalSession{}).Error; err != nil {
				return err
			}
			return tx.Unscoped().Model(&existing).Updates(map[string]interface{}{
				"deleted_at": nil, "keychip_id": keychip, "name": terminal.Name, "game_id": terminal.GameID,
				"game_version": terminal.GameVersion, "owner_account_id": terminal.OwnerAccountID,
				"is_enabled": true, "last_seen_keychip": "", "last_seen_at": nil, "last_seen_ip": "",
			}).Error
		}); err != nil {
			writeTerminalError(w, http.StatusInternalServerError, "恢复机台绑定失败")
			return
		}
		lookup := database.DB.Where("id = ?", existing.ID).Find(&terminal)
		if lookup.Error != nil || lookup.RowsAffected == 0 {
			writeTerminalError(w, http.StatusInternalServerError, "读取恢复后的机台失败")
			return
		}
	} else if err := database.DB.Create(&terminal).Error; err != nil {
		writeTerminalError(w, http.StatusConflict, "该 Keychip 前缀已被绑定")
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
	if lookup := database.DB.Where("id = ?", request.ID).Find(&terminal); lookup.Error != nil || lookup.RowsAffected == 0 {
		writeTerminalError(w, http.StatusNotFound, "机台不存在")
		return
	}
	terminal.Name = strings.TrimSpace(request.Name)
	gameID, supported := aime.NormalizeTerminalGameID(request.GameID)
	if !supported {
		writeTerminalError(w, http.StatusBadRequest, "不支持的游戏 ID")
		return
	}
	terminal.GameID = gameID
	terminal.GameVersion = strings.TrimSpace(request.GameVersion)
	terminal.OwnerAccountID = request.OwnerAccountID
	terminal.IsEnabled = request.IsEnabled
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
