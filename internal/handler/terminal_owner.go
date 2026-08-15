package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

// HandleUserTerminals returns only the cabinets that belong to the authenticated account.
func HandleUserTerminals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAccount(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var terminals []model.Terminal
	if err := database.DB.Where("owner_account_id = ?", account.ID).Order("id asc").Find(&terminals).Error; err != nil {
		writeTerminalError(w, http.StatusInternalServerError, "获取我的机台失败")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "terminals": terminals})
}

// HandleCreateUserTerminal lets a signed-in account register a Keychip it owns.
// Ownership is always derived from the server-side session and never trusted from the request body.
func HandleCreateUserTerminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAccount(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request terminalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeTerminalError(w, http.StatusBadRequest, "机台参数无效")
		return
	}
	terminal, err := createOwnedTerminal(request, account.ID)
	if err != nil {
		writeOwnedTerminalError(w, err)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "terminal": terminal, "message": "机台绑定成功"})
}

// HandleUpdateUserTerminal changes presentation and connection settings of a cabinet
// only when the requester's account owns that cabinet. Ownership itself is immutable here.
func HandleUpdateUserTerminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAccount(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request terminalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ID == 0 {
		writeTerminalError(w, http.StatusBadRequest, "机台参数无效")
		return
	}
	var terminal model.Terminal
	result := database.DB.Where("id = ? AND owner_account_id = ?", request.ID, account.ID).Find(&terminal)
	if result.Error != nil || result.RowsAffected == 0 {
		writeTerminalError(w, http.StatusNotFound, "机台不存在或不属于当前账户")
		return
	}
	terminal.Name = strings.TrimSpace(request.Name)
	terminal.GameID = strings.ToUpper(strings.TrimSpace(request.GameID))
	terminal.GameVersion = strings.TrimSpace(request.GameVersion)
	if terminal.GameID == "" {
		terminal.GameID = "SDEZ"
	}
	if err := database.DB.Save(&terminal).Error; err != nil {
		writeTerminalError(w, http.StatusInternalServerError, "保存机台失败")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "terminal": terminal, "message": "机台已更新"})
}

// HandleDeleteUserTerminal permanently unbinds an owned cabinet and invalidates
// every protected ALL.Net game session associated with it.
func HandleDeleteUserTerminal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAccount(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request terminalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ID == 0 {
		writeTerminalError(w, http.StatusBadRequest, "机台参数无效")
		return
	}
	var terminal model.Terminal
	result := database.DB.Where("id = ? AND owner_account_id = ?", request.ID, account.ID).Find(&terminal)
	if result.Error != nil || result.RowsAffected == 0 {
		writeTerminalError(w, http.StatusNotFound, "机台不存在或不属于当前账户")
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("terminal_id = ?", terminal.ID).Delete(&model.TerminalSession{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&terminal).Error
	}); err != nil {
		writeTerminalError(w, http.StatusInternalServerError, "解除机台绑定失败")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "机台绑定已删除"})
}

func createOwnedTerminal(request terminalRequest, ownerID uint) (model.Terminal, error) {
	keychipInput := strings.TrimSpace(request.KeychipID)
	if !isKeychipRegistrationFormat(keychipInput) {
		return model.Terminal{}, terminalOwnershipError{status: http.StatusBadRequest, message: "Keychip 格式必须为 Axxx-xxxxxxxxxxx"}
	}
	keychip := formatKeychip(keychipInput)
	gameID := strings.ToUpper(strings.TrimSpace(request.GameID))
	if gameID == "" {
		gameID = "SDEZ"
	}
	terminal := model.Terminal{KeychipID: keychip, Name: strings.TrimSpace(request.Name), GameID: gameID, GameVersion: strings.TrimSpace(request.GameVersion), OwnerAccountID: ownerID, IsEnabled: true}
	existing, found, lookupErr := findStoredTerminalByKeychipPrefix(keychip)
	if lookupErr != nil {
		return model.Terminal{}, terminalOwnershipError{status: http.StatusInternalServerError, message: "检查 Keychip 失败"}
	}
	if found {
		if !existing.DeletedAt.Valid {
			return model.Terminal{}, terminalOwnershipError{status: http.StatusConflict, message: "该 Keychip 前缀已被绑定"}
		}
		if err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("terminal_id = ?", existing.ID).Delete(&model.TerminalSession{}).Error; err != nil {
				return err
			}
			return tx.Unscoped().Model(&existing).Updates(map[string]interface{}{
				"deleted_at": nil, "keychip_id": keychip, "name": terminal.Name, "game_id": terminal.GameID,
				"game_version": terminal.GameVersion, "owner_account_id": ownerID, "is_enabled": true,
				"last_seen_keychip": "", "last_seen_at": nil, "last_seen_ip": "",
			}).Error
		}); err != nil {
			return model.Terminal{}, terminalOwnershipError{status: http.StatusInternalServerError, message: "恢复机台绑定失败"}
		}
		lookup := database.DB.Where("id = ?", existing.ID).Find(&terminal)
		if lookup.Error != nil || lookup.RowsAffected == 0 {
			return model.Terminal{}, terminalOwnershipError{status: http.StatusInternalServerError, message: "读取机台绑定失败"}
		}
		return terminal, nil
	}
	if err := database.DB.Create(&terminal).Error; err != nil {
		return model.Terminal{}, terminalOwnershipError{status: http.StatusInternalServerError, message: "创建机台绑定失败"}
	}
	return terminal, nil
}

type terminalOwnershipError struct {
	status  int
	message string
}

func (err terminalOwnershipError) Error() string { return err.message }

func writeOwnedTerminalError(w http.ResponseWriter, err error) {
	if ownership, ok := err.(terminalOwnershipError); ok {
		writeTerminalError(w, ownership.status, ownership.message)
		return
	}
	writeTerminalError(w, http.StatusInternalServerError, "机台操作失败")
}
