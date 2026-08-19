package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

// HandleGetOngekiStats exposes only records persisted by the SDDT protocol.
// It does not synthesize a profile for a card that has never completed an
// Ongeki user-data upload.
func HandleGetOngekiStats(w http.ResponseWriter, r *http.Request) {
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
	card, found, err := ongekiPortalCard(account.ID, cardID)
	if err != nil || !found {
		writeOngekiStats(w, map[string]any{"success": true, "message": "当前账户尚未绑定 Aime 卡", "recentPlays": []any{}, "musicDetails": []any{}})
		return
	}
	var profile model.OngekiUser
	if result := database.DB.Where("user_id = ?", card.GameUserID).First(&profile); result.Error != nil {
		writeOngekiStats(w, map[string]any{"success": true, "selectedCardId": card.ID, "message": "该卡片尚未创建 Ongeki 档案", "recentPlays": []any{}, "musicDetails": []any{}})
		return
	}
	var plays []model.OngekiPlaylog
	database.DB.Where("user_id = ?", card.GameUserID).Order("id desc").Limit(100).Find(&plays)
	var details []model.OngekiMusicDetail
	database.DB.Where("user_id = ?", card.GameUserID).Order("music_id asc, level asc").Find(&details)
	writeOngekiStats(w, map[string]any{
		"success": true, "selectedCardId": card.ID, "userId": card.GameUserID,
		"profile": decodeOngekiJSON(profile.ProfileJSON), "recentPlays": decodeOngekiPlaylogs(plays), "musicDetails": decodeOngekiMusicDetails(details),
	})
}

func ongekiPortalCard(accountID uint, cardID uint) (model.UserCard, bool, error) {
	var card model.UserCard
	if cardID != 0 {
		result := database.DB.Where("id = ? AND user_id = ?", cardID, accountID).Limit(1).Find(&card)
		return card, result.RowsAffected > 0, result.Error
	}
	result := database.DB.Model(&model.UserCard{}).
		Select("user_cards.*").
		Joins("JOIN ongeki_users ON ongeki_users.user_id = user_cards.game_user_id AND ongeki_users.deleted_at IS NULL").
		Where("user_cards.user_id = ?", accountID).
		Order("user_cards.id asc").Limit(1).Find(&card)
	if result.Error != nil || result.RowsAffected > 0 {
		return card, result.RowsAffected > 0, result.Error
	}
	result = database.DB.Where("user_id = ?", accountID).Order("id asc").Limit(1).Find(&card)
	return card, result.RowsAffected > 0, result.Error
}

func writeOngekiStats(w http.ResponseWriter, value map[string]any) {
	_ = json.NewEncoder(w).Encode(value)
}

func decodeOngekiJSON(raw string) map[string]any {
	value := map[string]any{}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func decodeOngekiPlaylogs(rows []model.OngekiPlaylog) []map[string]any {
	values := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		values = append(values, decodeOngekiJSON(row.PlaylogJSON))
	}
	return values
}

func decodeOngekiMusicDetails(rows []model.OngekiMusicDetail) []map[string]any {
	values := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		values = append(values, decodeOngekiJSON(row.DetailJSON))
	}
	return values
}
