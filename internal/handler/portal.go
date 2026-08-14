package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
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
