package handler

import (
	"encoding/json"
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

type recordIDRequest struct {
	ID uint `json:"id"`
}

type chargeIDRequest struct {
	ChargeID int64 `json:"chargeId"`
}

func HandleAdminEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var events []model.GameEvent
	if err := database.DB.Order("id asc").Find(&events).Error; err != nil {
		writeGameAdminError(w, http.StatusInternalServerError, "获取游戏事件失败")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "events": events})
}

func HandleCreateGameEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var event model.GameEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		writeGameAdminError(w, http.StatusBadRequest, "事件参数无效")
		return
	}
	event.ID = 0
	if err := database.DB.Create(&event).Error; err != nil {
		writeGameAdminError(w, http.StatusInternalServerError, "创建游戏事件失败")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "event": event, "message": "游戏事件已创建"})
}

func HandleUpdateGameEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var event model.GameEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil || event.ID == 0 {
		writeGameAdminError(w, http.StatusBadRequest, "事件参数无效")
		return
	}
	updates := map[string]interface{}{
		"type": event.Type, "start_date": event.StartDate, "end_date": event.EndDate, "disable_area": event.DisableArea,
	}
	result := database.DB.Model(&model.GameEvent{}).Where("id = ?", event.ID).Updates(updates)
	if result.Error != nil {
		writeGameAdminError(w, http.StatusInternalServerError, "更新游戏事件失败")
		return
	}
	if result.RowsAffected == 0 {
		writeGameAdminError(w, http.StatusNotFound, "游戏事件不存在")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "游戏事件已更新"})
}

func HandleDeleteGameEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request recordIDRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ID == 0 {
		writeGameAdminError(w, http.StatusBadRequest, "事件 ID 无效")
		return
	}
	result := database.DB.Delete(&model.GameEvent{}, request.ID)
	if result.Error != nil {
		writeGameAdminError(w, http.StatusInternalServerError, "删除游戏事件失败")
		return
	}
	if result.RowsAffected == 0 {
		writeGameAdminError(w, http.StatusNotFound, "游戏事件不存在")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "游戏事件已删除"})
}

func HandleAdminCharges(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var charges []model.GameCharge
	if err := database.DB.Order("order_id asc, charge_id asc").Find(&charges).Error; err != nil {
		writeGameAdminError(w, http.StatusInternalServerError, "获取收费项目失败")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "charges": charges})
}

func HandleCreateGameCharge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var charge model.GameCharge
	if err := json.NewDecoder(r.Body).Decode(&charge); err != nil || charge.ChargeID <= 0 {
		writeGameAdminError(w, http.StatusBadRequest, "收费项目参数无效")
		return
	}
	if err := database.DB.Create(&charge).Error; err != nil {
		writeGameAdminError(w, http.StatusConflict, "收费项目 ID 已存在或无法创建")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "charge": charge, "message": "收费项目已创建"})
}

func HandleUpdateGameCharge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var charge model.GameCharge
	if err := json.NewDecoder(r.Body).Decode(&charge); err != nil || charge.ChargeID <= 0 {
		writeGameAdminError(w, http.StatusBadRequest, "收费项目参数无效")
		return
	}
	updates := map[string]interface{}{
		"order_id": charge.OrderID, "price": charge.Price, "start_date": charge.StartDate, "end_date": charge.EndDate,
	}
	result := database.DB.Model(&model.GameCharge{}).Where("charge_id = ?", charge.ChargeID).Updates(updates)
	if result.Error != nil {
		writeGameAdminError(w, http.StatusInternalServerError, "更新收费项目失败")
		return
	}
	if result.RowsAffected == 0 {
		writeGameAdminError(w, http.StatusNotFound, "收费项目不存在")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "收费项目已更新"})
}

func HandleDeleteGameCharge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request chargeIDRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.ChargeID <= 0 {
		writeGameAdminError(w, http.StatusBadRequest, "收费项目 ID 无效")
		return
	}
	result := database.DB.Delete(&model.GameCharge{}, request.ChargeID)
	if result.Error != nil {
		writeGameAdminError(w, http.StatusInternalServerError, "删除收费项目失败")
		return
	}
	if result.RowsAffected == 0 {
		writeGameAdminError(w, http.StatusNotFound, "收费项目不存在")
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "收费项目已删除"})
}

func writeGameAdminError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": message})
}
