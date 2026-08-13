package handler

import (
	"encoding/json"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

// maimaiCompatibilityPayload contains API responses which are defined by the
// maimai2 protocol but do not need player-record persistence. Editable values
// live in SystemConfig so operators never need to rebuild the server to change
// machine-delivered settings.
func maimaiCompatibilityPayload(apiName string, userID int64, body []byte) (interface{}, bool) {
	switch apiName {
	case "CreateToken":
		return map[string]interface{}{"Bearer": maimaiConfigValue("maimai_bearer_token", "")}, true
	case "CMUpsertUserPrintlog":
		orderID := requestString(body, "orderId")
		if orderID == "" {
			orderID = "0"
		}
		serialID, err := nextPrintSerial()
		if err != nil {
			return nil, true
		}
		return map[string]interface{}{
			"returnCode": 1,
			"orderId":    orderID,
			"serialId":   serialID,
		}, true
	case "CMGetSellingCard":
		var cards []model.GameSellingCard
		if err := database.DB.Order("card_id asc").Find(&cards).Error; err != nil {
			return nil, true
		}
		return map[string]interface{}{"length": len(cards), "sellingCardList": cards}, true
	case "GetGameNationalData":
		return map[string]interface{}{"nextIndex": 0, "nationalDataList": []interface{}{}}, true
	case "GetGameTournamentInfo":
		return map[string]interface{}{"length": 0, "gameTournamentInfoList": []interface{}{}}, true
	case "GetGameKaleidxScope":
		return map[string]interface{}{"gameKaleidxScopeList": []map[string]int{
			{"gateId": 1, "phaseId": 6}, {"gateId": 2, "phaseId": 6}, {"gateId": 3, "phaseId": 6},
			{"gateId": 4, "phaseId": 6}, {"gateId": 5, "phaseId": 6}, {"gateId": 6, "phaseId": 6},
			{"gateId": 7, "phaseId": 6}, {"gateId": 8, "phaseId": 6}, {"gateId": 9, "phaseId": 6},
			{"gateId": 10, "phaseId": 13},
		}}, true
	case "GetUserFriendBonus":
		return map[string]interface{}{"userId": userID, "returnCode": 0, "getMiles": 0}, true
	case "GetTransferFriend":
		return map[string]interface{}{"userId": userID, "transferFriendList": []interface{}{}}, true
	case "GetUserNewItem":
		return map[string]interface{}{"userId": userID, "itemKind": 0, "itemId": 0}, true
	case "GetUserNewItemList":
		return map[string]interface{}{"userId": userID, "userItemList": []interface{}{}}, true
	case "GetUserFriendCheck":
		return map[string]interface{}{"returnCode": 0}, true
	case "UserFriendRegist":
		return map[string]interface{}{"returnCode1": 0, "returnCode2": 0}, true
	case "GetUserShopStock":
		stocks := make([]map[string]interface{}, 0)
		for _, shopItemID := range requestIntSlice(body, "shopItemIdList") {
			stocks = append(stocks, map[string]interface{}{"shopItemId": shopItemID, "tradeCount": 0})
		}
		return map[string]interface{}{"userId": userID, "userShopStockList": stocks}, true
	case "GetUserRivalData":
		rivalID := requestInt64(body, "rivalId")
		var rival model.UserDetail
		_ = database.DB.Where("id = ?", rivalID).First(&rival).Error
		return map[string]interface{}{
			"userId":        userID,
			"userRivalData": map[string]interface{}{"rivalId": rivalID, "rivalName": rival.UserName},
		}, true
	case "GetUserRivalMusic":
		rivalID := requestInt64(body, "rivalId")
		var details []model.UserMusicDetail
		database.DB.Where("user_id = ?", rivalID).Order("music_id asc, level asc").Find(&details)
		grouped := map[int][]map[string]int{}
		for _, detail := range details {
			grouped[detail.MusicID] = append(grouped[detail.MusicID], map[string]int{
				"level": detail.Level, "achievement": detail.Achievement, "deluxscoreMax": detail.DeluxScoreMax,
			})
		}
		musicIDs := make([]int, 0, len(grouped))
		for musicID := range grouped {
			musicIDs = append(musicIDs, musicID)
		}
		sort.Ints(musicIDs)
		music := make([]map[string]interface{}, 0, len(musicIDs))
		for _, musicID := range musicIDs {
			music = append(music, map[string]interface{}{"musicId": musicID, "userRivalMusicDetailList": grouped[musicID]})
		}
		return map[string]interface{}{"userId": userID, "rivalId": rivalID, "nextIndex": 0, "userRivalMusicList": music}, true
	case "GetUserRegion":
		var regions []model.UserRegion
		database.DB.Where("user_id = ?", userID).Order("region_id asc").Find(&regions)
		payload := make([]map[string]int, 0, len(regions))
		for _, region := range regions {
			payload = append(payload, map[string]int{"regionId": region.RegionID, "playCount": region.PlayCount})
		}
		return map[string]interface{}{"userId": userID, "length": len(payload), "userRegionList": payload}, true
	case "UserLogin":
		regionID := requestInt(body, "regionId")
		if userID > 0 && regionID > 0 {
			var detail model.UserDetail
			if err := database.DB.Where("user_id = ?", userID).First(&detail).Error; err == nil {
				region := model.UserRegion{UserID: userID, RegionID: regionID}
				lookup := database.DB.Where(&model.UserRegion{UserID: userID, RegionID: regionID}).FirstOrCreate(&region)
				if lookup.Error == nil && lookup.RowsAffected == 0 {
					region.PlayCount++
					_ = database.DB.Save(&region).Error
				}
			}
		}
		return map[string]interface{}{
			"returnCode": 1, "loginCount": 1, "lastLoginDate": "2020-01-01 00:00:00.0",
			"consecutiveLoginCount": 0, "loginId": 1,
			"Bearer": maimaiConfigValue("maimai_bearer_token", ""), "bearer": maimaiConfigValue("maimai_bearer_token", ""),
		}, true
	case "GetGameWeeklyData":
		return map[string]interface{}{
			"gameWeeklyData": map[string]interface{}{
				"missionCategory": maimaiConfigInt("maimai_weekly_mission_category", 0),
				"updateDate":      maimaiConfigValue("maimai_weekly_update_date", "2024-01-01 00:00:00.0"),
				"beforeDate":      maimaiConfigValue("maimai_weekly_before_date", "2077-01-01 00:00:00.0"),
			},
		}, true
	case "GetUserMissionData":
		return map[string]interface{}{
			"userId": userID,
			"userWeeklyData": map[string]interface{}{
				"lastLoginWeek":   "",
				"beforeLoginWeek": "",
				"friendBonusFlag": false,
			},
			"userMissionDataList": []interface{}{},
		}, true
	case "GetGameMusicScore":
		var request struct {
			MusicID interface{} `json:"musicId"`
			Level   interface{} `json:"level"`
			Type    interface{} `json:"type"`
		}
		_ = json.Unmarshal(body, &request)
		return map[string]interface{}{
			"gameMusicScore": map[string]interface{}{
				"musicId":   request.MusicID,
				"level":     request.Level,
				"type":      request.Type,
				"scoreData": "",
			},
		}, true
	case "GetGameFesta":
		jackingSide := maimaiConfigInt("maimai_festa_jacking_side_id", -1)
		if jackingSide < 0 || jackingSide > 2 {
			jackingSide = rand.Intn(3)
		}
		return map[string]interface{}{
			"eventId":                maimaiConfigInt("maimai_festa_event_id", 0),
			"isRallyPeriod":          maimaiConfigBool("maimai_festa_rally_period", false),
			"isCircleJoinNotAllowed": maimaiConfigBool("maimai_festa_circle_join_not_allowed", false),
			"jackingFestaSideId":     jackingSide,
			"festaSideDataList":      []interface{}{},
		}, true
	case "GetPlaceCircleData":
		return map[string]interface{}{"returnCode": 0, "circleId": 0, "aggrDate": ""}, true
	case "GetUserCircleData":
		return map[string]interface{}{
			"circleId":               0,
			"circleName":             maimaiConfigValue("maimai_circle_name", "一緒に歌おう！"),
			"isPlace":                false,
			"circleClass":            0,
			"lastLoginDate":          "",
			"circlePointRankingList": []interface{}{},
		}, true
	case "GetUserCircleChallenge":
		return map[string]interface{}{
			"userId":                userID,
			"userCircleChallenge":   nil,
			"circleCircleChallenge": nil,
			"achievement":           0,
		}, true
	case "GetUserCirclePointData":
		return map[string]interface{}{"userId": userID, "aggrDate": "", "userCirclePointDataList": []interface{}{}}, true
	case "GetUserCirclePointRanking":
		return map[string]interface{}{
			"circleId":            0,
			"circleName":          maimaiConfigValue("maimai_circle_name", "一緒に歌おう！"),
			"aggrDate":            "",
			"lastMonthCircleRank": 0,
			"lastMonthPoint":      0,
		}, true
	case "GetUserFesta":
		return map[string]interface{}{
			"userFestaData": map[string]interface{}{
				"eventId":                maimaiConfigInt("maimai_festa_event_id", 0),
				"circleId":               0,
				"festaSideId":            0,
				"circleTotalFestaPoint":  0,
				"currentTotalFestaPoint": 0,
				"circleRankInFestaSide":  0,
				"circleRecordDate":       "",
				"isDailyBonus":           false,
				"participationRewardGet": false,
				"receivedRewardBorder":   0,
			},
			"userResultFestaData": map[string]interface{}{
				"eventId":               maimaiConfigInt("maimai_festa_event_id", 0),
				"circleId":              0,
				"circleName":            maimaiConfigValue("maimai_circle_name", "一緒に歌おう！"),
				"festaSideId":           0,
				"circleRankInFestaSide": 0,
				"receivedRewardBorder":  0,
				"circleTotalFestaPoint": 0,
				"resultRewardGet":       false,
			},
		}, true
	}
	return nil, false
}

func requestString(body []byte, key string) string {
	var request map[string]json.RawMessage
	if err := json.Unmarshal(body, &request); err != nil {
		return ""
	}
	var value string
	_ = json.Unmarshal(request[key], &value)
	return value
}

func requestInt64(body []byte, key string) int64 {
	var request map[string]json.RawMessage
	if err := json.Unmarshal(body, &request); err != nil {
		return 0
	}
	var value int64
	_ = json.Unmarshal(request[key], &value)
	return value
}

func requestIntSlice(body []byte, key string) []int {
	var request map[string]json.RawMessage
	if err := json.Unmarshal(body, &request); err != nil {
		return []int{}
	}
	var values []int
	_ = json.Unmarshal(request[key], &values)
	return values
}

func maimaiConfigValue(key, fallback string) string {
	var config model.SystemConfig
	if err := database.DB.Where(&model.SystemConfig{Key: key}).First(&config).Error; err != nil {
		return fallback
	}
	return strings.TrimSpace(config.Value)
}

func maimaiConfigInt(key string, fallback int) int {
	value, err := strconv.Atoi(maimaiConfigValue(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func maimaiConfigBool(key string, fallback bool) bool {
	value := strings.ToLower(maimaiConfigValue(key, strconv.FormatBool(fallback)))
	return value == "true" || value == "1" || value == "yes"
}
