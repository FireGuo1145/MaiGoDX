package handler

import (
	"encoding/json"
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func HandleGetChuniStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAccount(w, r)
	if !ok {
		return
	}
	cardID, err := portalCardIDFromRequest(r)
	if err != nil {
		writePortalUpdateError(w, http.StatusBadRequest, "档案卡片参数无效")
		return
	}
	var card model.UserCard
	query := database.DB.Where("user_id = ?", account.ID).Order("id asc").Limit(1)
	if cardID != 0 {
		query = database.DB.Where("id = ? AND user_id = ?", cardID, account.ID).Limit(1)
	}
	if result := query.Find(&card); result.Error != nil || result.RowsAffected == 0 {
		writeChuniStats(w, map[string]any{"success": true, "message": "当前账户尚未绑定 Aime 卡", "recentPlays": []any{}, "musicDetails": []any{}})
		return
	}
	var profile model.ChuniUser
	if result := database.DB.Where("user_id = ?", card.GameUserID).First(&profile); result.Error != nil {
		writeChuniStats(w, map[string]any{"success": true, "selectedCardId": card.ID, "message": "该卡片尚未创建 CHUNITHM 档案", "recentPlays": []any{}, "musicDetails": []any{}})
		return
	}
	var plays []model.ChuniPlaylog
	database.DB.Where("user_id = ?", card.GameUserID).Order("id desc").Limit(100).Find(&plays)
	var details []model.ChuniMusicDetail
	database.DB.Where("user_id = ?", card.GameUserID).Order("music_id asc, level asc").Find(&details)
	writeChuniStats(w, map[string]any{
		"success": true, "selectedCardId": card.ID, "userId": card.GameUserID,
		"profile": decodeChuniJSON(profile.ProfileJSON), "recentPlays": decodeChuniRows(plays), "musicDetails": decodeChuniDetails(details),
	})
}

func writeChuniStats(w http.ResponseWriter, value map[string]any) {
	_ = json.NewEncoder(w).Encode(value)
}

func decodeChuniJSON(raw string) map[string]any {
	value := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &value)
	return value
}

func decodeChuniRows(rows []model.ChuniPlaylog) []map[string]any {
	values := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		value := decodeChuniJSON(row.PlaylogJSON)
		values = append(values, value)
	}
	return values
}

func decodeChuniDetails(rows []model.ChuniMusicDetail) []map[string]any {
	values := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		value := decodeChuniJSON(row.DetailJSON)
		values = append(values, value)
	}
	return values
}
