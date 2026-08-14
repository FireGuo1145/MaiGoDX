package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

type portalSong struct {
	MusicID     int `json:"musicId"`
	Level       int `json:"level"`
	Achievement int `json:"achievement"`
	Score       int `json:"score"`
	ScoreRank   int `json:"scoreRank"`
}

type portalTrendPoint struct {
	Date   string `json:"date"`
	Rating int    `json:"rating"`
}

type portalPartner struct {
	PartnerID int `json:"partnerId"`
}

type portalTravelPartner struct {
	PartnerID             int `json:"partnerId"`
	IntimateLevel         int `json:"intimateLevel"`
	IntimateCountRewarded int `json:"intimateCountRewarded"`
}

type portalFunctionTicket struct {
	ItemID int `json:"itemId"`
	Stock  int `json:"stock"`
}

type portalRegion struct {
	RegionID  int `json:"regionId"`
	PlayCount int `json:"playCount"`
}

type portalProfileUpdateRequest struct {
	PartnerID       int                    `json:"partnerId"`
	TravelPartners  []portalTravelPartner  `json:"travelPartners"`
	FunctionTickets []portalFunctionTicket `json:"functionTickets"`
	Regions         []portalRegion         `json:"regions"`
}

// HandleGetStats returns only persisted game data. It deliberately does not invent a
// profile, rating, trend, or ranking when no maimai record is associated with the account.
func HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var accountCount, playCount int64
	database.DB.Model(&model.UserAccount{}).Count(&accountCount)
	database.DB.Model(&model.UserPlaylog{}).Count(&playCount)

	response := map[string]interface{}{
		"success":     true,
		"totalUsers":  accountCount,
		"totalPlays":  playCount,
		"recentPlays": []model.UserPlaylog{},
		"trend":       []portalTrendPoint{},
		"rankCounts":  map[string]int{"SSS+": 0, "SSS": 0, "SS": 0, "S": 0},
		"ratingComposition": map[string]interface{}{
			"bests":    []portalSong{},
			"newBests": []portalSong{},
		},
		"travelPartners":  []portalTravelPartner{},
		"functionTickets": []portalFunctionTicket{},
		"regions":         []portalRegion{},
	}

	account, ok := requireAccount(w, r)
	if !ok {
		return
	}

	var card model.UserCard
	cardLookup := database.DB.Where("user_id = ? AND game_user_id > 0", account.ID).Order("id asc").Limit(1).Find(&card)
	if cardLookup.Error != nil || cardLookup.RowsAffected == 0 {
		response["message"] = "当前账户尚未关联 maimai 游戏档案"
		_ = json.NewEncoder(w).Encode(response)
		return
	}
	var detail model.UserDetail
	detailLookup := database.DB.Where("user_id = ?", card.GameUserID).Limit(1).Find(&detail)
	if detailLookup.Error != nil || detailLookup.RowsAffected == 0 {
		response["message"] = "已关联的卡片尚未创建 maimai 游戏档案"
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	var plays []model.UserPlaylog
	database.DB.Where("user_id = ?", detail.UserID).Order("id desc").Limit(100).Find(&plays)

	response["user"] = detail
	response["recentPlays"] = plays
	response["trend"] = makeTrend(plays)
	response["rankCounts"] = rankCounts(plays)
	response["ratingComposition"] = makeRatingComposition(detail.UserID)
	response["partner"] = portalPartner{PartnerID: detail.PartnerID}
	response["travelPartners"] = portalTravelPartners(detail.UserID)
	response["functionTickets"] = portalFunctionTickets(detail.UserID)
	response["regions"] = portalRegions(detail.UserID)

	_ = json.NewEncoder(w).Encode(response)
}

// HandleUpdatePortalProfile lets an authenticated owner update the parts of
// their maimai profile exposed by the portal. The game user ID is resolved
// from the account's bound Aime card; callers cannot edit another profile.
func HandleUpdatePortalProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	account, ok := requireAccount(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request portalProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writePortalUpdateError(w, http.StatusBadRequest, "请求参数错误")
		return
	}
	detail, err := portalDetailForAccount(account.ID)
	if err != nil {
		writePortalUpdateError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := savePortalProfile(detail, request); err != nil {
		writePortalUpdateError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "游戏档案已保存"})
}

func writePortalUpdateError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": message})
}

func portalDetailForAccount(accountID uint) (model.UserDetail, error) {
	var card model.UserCard
	lookup := database.DB.Where("user_id = ? AND game_user_id > 0", accountID).Order("id asc").Limit(1).Find(&card)
	if lookup.Error != nil || lookup.RowsAffected == 0 {
		return model.UserDetail{}, errors.New("当前账户尚未关联 maimai 游戏档案")
	}
	var detail model.UserDetail
	lookup = database.DB.Where("user_id = ?", card.GameUserID).Limit(1).Find(&detail)
	if lookup.Error != nil || lookup.RowsAffected == 0 {
		return model.UserDetail{}, errors.New("已关联的卡片尚未创建 maimai 游戏档案")
	}
	return detail, nil
}

func savePortalProfile(detail model.UserDetail, request portalProfileUpdateRequest) error {
	if request.PartnerID < 0 || !validPortalProfileCollections(request) {
		return errors.New("档案数据包含无效的负数或重复 ID")
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserDetail{}).Where("user_id = ?", detail.UserID).Update("partner_id", request.PartnerID).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", detail.UserID).Delete(&model.UserIntimate{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND item_kind = ?", detail.UserID, 12).Delete(&model.UserItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", detail.UserID).Delete(&model.UserRegion{}).Error; err != nil {
			return err
		}
		for _, value := range request.TravelPartners {
			if err := tx.Create(&model.UserIntimate{UserID: detail.UserID, PartnerID: value.PartnerID, IntimateLevel: value.IntimateLevel, IntimateCountRewarded: value.IntimateCountRewarded}).Error; err != nil {
				return err
			}
		}
		for _, value := range request.FunctionTickets {
			if err := tx.Create(&model.UserItem{UserID: detail.UserID, ItemKind: 12, ItemID: value.ItemID, Stock: value.Stock, IsValid: true}).Error; err != nil {
				return err
			}
		}
		for _, value := range request.Regions {
			if err := tx.Create(&model.UserRegion{UserID: detail.UserID, RegionID: value.RegionID, PlayCount: value.PlayCount}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func validPortalProfileCollections(request portalProfileUpdateRequest) bool {
	partners, tickets, regions := map[int]bool{}, map[int]bool{}, map[int]bool{}
	for _, value := range request.TravelPartners {
		if value.PartnerID < 0 || value.IntimateLevel < 0 || value.IntimateCountRewarded < 0 || partners[value.PartnerID] {
			return false
		}
		partners[value.PartnerID] = true
	}
	for _, value := range request.FunctionTickets {
		if value.ItemID < 0 || value.Stock < 0 || tickets[value.ItemID] {
			return false
		}
		tickets[value.ItemID] = true
	}
	for _, value := range request.Regions {
		if value.RegionID < 0 || value.PlayCount < 0 || regions[value.RegionID] {
			return false
		}
		regions[value.RegionID] = true
	}
	return true
}

func portalTravelPartners(userID int64) []portalTravelPartner {
	var values []model.UserIntimate
	database.DB.Where("user_id = ?", userID).Order("partner_id asc").Find(&values)
	result := make([]portalTravelPartner, 0, len(values))
	for _, value := range values {
		result = append(result, portalTravelPartner{
			PartnerID: value.PartnerID, IntimateLevel: value.IntimateLevel, IntimateCountRewarded: value.IntimateCountRewarded,
		})
	}
	return result
}

func portalFunctionTickets(userID int64) []portalFunctionTicket {
	var values []model.UserItem
	// Item kind 12 is the maimai DX function-ticket inventory.
	database.DB.Where("user_id = ? AND item_kind = ?", userID, 12).Order("item_id asc").Find(&values)
	result := make([]portalFunctionTicket, 0, len(values))
	for _, value := range values {
		result = append(result, portalFunctionTicket{ItemID: value.ItemID, Stock: value.Stock})
	}
	return result
}

func portalRegions(userID int64) []portalRegion {
	var values []model.UserRegion
	database.DB.Where("user_id = ?", userID).Order("region_id asc").Find(&values)
	result := make([]portalRegion, 0, len(values))
	for _, value := range values {
		result = append(result, portalRegion{RegionID: value.RegionID, PlayCount: value.PlayCount})
	}
	return result
}

func makeTrend(plays []model.UserPlaylog) []portalTrendPoint {
	points := make([]portalTrendPoint, 0, len(plays))
	for _, play := range plays {
		if play.AfterRating == 0 || play.CreateDate == "" {
			continue
		}
		points = append(points, portalTrendPoint{Date: portalDate(play.CreateDate), Rating: play.AfterRating})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Date < points[j].Date })
	return points
}

func rankCounts(plays []model.UserPlaylog) map[string]int {
	counts := map[string]int{"SSS+": 0, "SSS": 0, "SS": 0, "S": 0}
	for _, play := range plays {
		switch {
		case play.Achievement >= 1005000:
			counts["SSS+"]++
		case play.Achievement >= 1000000:
			counts["SSS"]++
		case play.Achievement >= 990000:
			counts["SS"]++
		case play.Achievement >= 970000:
			counts["S"]++
		}
	}
	return counts
}

func makeRatingComposition(userID int64) map[string]interface{} {
	return map[string]interface{}{
		"bests":    ratingSongs(userID, ratingKeyCurrent),
		"newBests": ratingSongs(userID, ratingKeyNew),
	}
}

func ratingSongs(userID int64, key string) []portalSong {
	rates := loadRateData(userID, key)
	songs := make([]portalSong, 0, len(rates))
	for _, rate := range rates {
		song := portalSong{MusicID: rate.MusicID, Level: rate.Level, Achievement: rate.Achievement}
		var detail model.UserMusicDetail
		if err := database.DB.Where("user_id = ? AND music_id = ? AND level = ?", userID, rate.MusicID, rate.Level).First(&detail).Error; err == nil {
			song.Score = detail.DeluxScoreMax
			song.ScoreRank = detail.ScoreRank
		}
		songs = append(songs, song)
	}
	return songs
}

func portalDate(value string) string {
	for _, layout := range []string{"20060102150405", "2006-01-02 15:04:05", time.RFC3339, "2006/01/02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}
