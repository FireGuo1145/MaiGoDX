package handler

import (
	"encoding/json"
	"math/rand"
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

func maimaiConfigValue(key, fallback string) string {
	var config model.SystemConfig
	if err := database.DB.Where("key = ?", key).First(&config).Error; err != nil {
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
