package chunithm

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

// Handler implements the JSON servlet used by CHUNITHM cabinets on SDHD and
// SDGS routes. It deliberately accepts unknown optional APIs as no-ops, as
// AquaDX does, so mixed client revisions can complete their boot sequence.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	endpoint := strings.TrimSuffix(pathTail(r.URL.Path), "Api")
	var request map[string]any
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeResponse(w, endpoint, 0, nil, "invalid JSON payload")
		return
	}
	userID := requestInt64(request, "userId")

	switch endpoint {
	case "UpsertUserAll":
		handleUpsertUserAll(w, request, userID)
	case "GetUserData":
		handleGetUserData(w, userID)
	case "GetUserPreview":
		handleGetUserPreview(w, userID)
	case "GetUserMusic":
		handleGetUserMusic(w, userID)
	case "GetGameSetting":
		writePayload(w, gameSetting(r, request))
	case "GetUserOption":
		writePayload(w, map[string]any{"userId": userID, "userGameOption": map[string]any{}})
	case "GetUserItem", "GetUserCharacter", "GetUserMapArea", "GetUserCourse", "GetUserCharge", "GetUserDuel", "GetUserGacha", "GetUserActivity", "GetUserRecentRating", "GetUserMate", "GetUserRegion", "GetUserFavoriteItem", "GetUserFavoriteCollection", "GetUserCMission", "GetUserCMissionList", "GetUserLV", "GetUserUC", "GetUserCardPrintError", "GetUserLoginBonus":
		writePayload(w, pagedResponse(endpoint, userID))
	case "GetGameEvent", "GetGameCharge", "GetGameGacha", "GetGameRanking", "GetGameCourseLevel", "GetGameUCCondition", "GetGameLVConditionOpen", "GetGameLVConditionUnlock", "GetGameMapAreaCondition", "GetGameIdlist", "GetUserRecMusic", "GetUserRecRating", "GetUserTeam", "GetUserNetBattleData", "GetUserNetBattleRankingInfo", "GetUserRivalData", "GetUserRivalMusic", "GetUserPrintedCard", "GetUserCtoCPlay", "GetTeamCourseSetting", "GetTeamCourseRule":
		writePayload(w, emptyGamePayload(endpoint, userID))
	case "GameLogin", "UpsertUserChargelog", "UpsertClientBookkeeping", "UpsertClientDevelop", "UpsertClientPlayTime", "UpsertClientSetting", "UpsertClientTestmode", "UpsertClientUpload", "UserLogout", "CMLogin", "CMLogout":
		writeResponse(w, endpoint, 1, nil, "")
	default:
		writeResponse(w, endpoint, 1, nil, "")
	}
}

func handleUpsertUserAll(w http.ResponseWriter, request map[string]any, userID int64) {
	if userID == 0 {
		writeResponse(w, "UpsertUserAll", 0, nil, "missing userId")
		return
	}
	all, _ := request["upsertUserAll"].(map[string]any)
	if all == nil {
		writeResponse(w, "UpsertUserAll", 0, nil, "missing upsertUserAll")
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if data := firstObject(all["userData"]); data != nil {
			data["userId"] = userID
			payload, err := json.Marshal(data)
			if err != nil {
				return err
			}
			profile := model.ChuniUser{UserID: userID, UserName: stringValue(data["userName"]), PlayerRating: intValue(data["playerRating"]), ProfileJSON: string(payload)}
			var existing model.ChuniUser
			if err := tx.Where("user_id = ?", userID).First(&existing).Error; err == nil {
				profile.ID = existing.ID
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Save(&profile).Error; err != nil {
				return err
			}
		}
		for _, detail := range objectList(all["userMusicDetailList"]) {
			payload, err := json.Marshal(detail)
			if err != nil {
				return err
			}
			entry := model.ChuniMusicDetail{UserID: userID, MusicID: intValue(detail["musicId"]), Level: intValue(detail["level"]), DetailJSON: string(payload)}
			var existing model.ChuniMusicDetail
			if err := tx.Where("user_id = ? AND music_id = ? AND level = ?", entry.UserID, entry.MusicID, entry.Level).First(&existing).Error; err == nil {
				entry.ID = existing.ID
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Save(&entry).Error; err != nil {
				return err
			}
		}
		for _, playlog := range objectList(all["userPlaylogList"]) {
			payload, err := json.Marshal(playlog)
			if err != nil {
				return err
			}
			if err := tx.Create(&model.ChuniPlaylog{UserID: userID, MusicID: intValue(playlog["musicId"]), Level: intValue(playlog["level"]), PlaylogJSON: string(payload)}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		writeResponse(w, "UpsertUserAll", 0, nil, err.Error())
		return
	}
	writeResponse(w, "UpsertUserAll", 1, nil, "")
}

func handleGetUserData(w http.ResponseWriter, userID int64) {
	var profile model.ChuniUser
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		writePayload(w, map[string]any{"userId": userID, "userData": nil})
		return
	}
	data := map[string]any{}
	_ = json.Unmarshal([]byte(profile.ProfileJSON), &data)
	writePayload(w, map[string]any{"userId": userID, "userData": data})
}

func handleGetUserPreview(w http.ResponseWriter, userID int64) {
	var profile model.ChuniUser
	if err := database.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		writePayload(w, map[string]any{"userId": userID, "isLogin": false})
		return
	}
	writePayload(w, map[string]any{"userId": userID, "userName": profile.UserName, "playerRating": profile.PlayerRating, "isLogin": false, "banState": 0})
}

func handleGetUserMusic(w http.ResponseWriter, userID int64) {
	var details []model.ChuniMusicDetail
	database.DB.Where("user_id = ?", userID).Order("music_id asc, level asc").Find(&details)
	byMusic := map[int][]map[string]any{}
	for _, detail := range details {
		value := map[string]any{}
		_ = json.Unmarshal([]byte(detail.DetailJSON), &value)
		byMusic[detail.MusicID] = append(byMusic[detail.MusicID], value)
	}
	list := make([]map[string]any, 0, len(byMusic))
	for _, values := range byMusic {
		list = append(list, map[string]any{"length": len(values), "userMusicDetailList": values})
	}
	writePayload(w, map[string]any{"userId": userID, "length": len(list), "nextIndex": -1, "userMusicList": list})
}

func gameSetting(r *http.Request, request map[string]any) map[string]any {
	version := stringValue(request["version"])
	if version == "" {
		version = "2.00"
	}
	now := time.Now().UTC()
	base := "http://" + r.Host + "/g/SDHD/" + version + "/ChuniServlet/"
	return map[string]any{"gameSetting": map[string]any{
		"romVersion": version + ".00", "dataVersion": version + ".00", "isMaintenance": false, "requestInterval": 0,
		"rebootStartTime": now.Add(-4 * time.Hour).Format("2006-01-02 15:04:05"), "rebootEndTime": now.Add(-3 * time.Hour).Format("2006-01-02 15:04:05"),
		"isBackgroundDistribute": false, "maxCountCharacter": 300, "maxCountItem": 300, "maxCountMusic": 300,
		"matchStartTime": now.Format("2006-01-02") + " 00:01:00", "matchEndTime": now.Format("2006-01-02") + " 23:59:00", "matchTimeLimit": 10, "matchErrorLimit": 10,
		"matchingUri": base, "matchingUriX": base,
	}, "isDumpUpload": false, "isAou": false}
}

func pagedResponse(endpoint string, userID int64) map[string]any {
	key := map[string]string{"GetUserItem": "userItemList", "GetUserCharacter": "userCharacterList", "GetUserMapArea": "userMapAreaList", "GetUserCourse": "userCourseList", "GetUserCharge": "userChargeList", "GetUserDuel": "userDuelList", "GetUserGacha": "userGachaList", "GetUserActivity": "userActivityList", "GetUserRecentRating": "userRecentRatingList", "GetUserMate": "userMateList", "GetUserRegion": "userRegionList", "GetUserFavoriteItem": "userFavoriteItemList", "GetUserFavoriteCollection": "userFavoriteCollectionList", "GetUserCMission": "userCMissionList", "GetUserCMissionList": "userCMissionList", "GetUserLV": "userLinkedVerseList", "GetUserUC": "userUnlockChallengeList", "GetUserCardPrintError": "userCardPrintErrorList", "GetUserLoginBonus": "userLoginBonusList"}[endpoint]
	return map[string]any{"userId": userID, "length": 0, "nextIndex": -1, key: []any{}}
}

func emptyGamePayload(endpoint string, userID int64) map[string]any {
	key := map[string]string{
		"GetGameEvent":                "gameEventList",
		"GetGameCharge":               "gameChargeList",
		"GetGameGacha":                "gameGachaList",
		"GetGameRanking":              "gameRankingList",
		"GetGameCourseLevel":          "gameCourseLevelList",
		"GetGameUCCondition":          "gameUnlockChallengeConditionList",
		"GetGameLVConditionOpen":      "gameLinkedVerseConditionOpenList",
		"GetGameLVConditionUnlock":    "gameLinkedVerseConditionUnlockList",
		"GetGameMapAreaCondition":     "gameMapAreaConditionList",
		"GetGameIdlist":               "gameIdlistList",
		"GetUserRecMusic":             "userRecMusicList",
		"GetUserRecRating":            "userRecRatingList",
		"GetUserTeam":                 "userTeamList",
		"GetUserNetBattleData":        "userNetBattleDataList",
		"GetUserNetBattleRankingInfo": "userNetBattleRankingInfoList",
		"GetUserRivalData":            "userRivalDataList",
		"GetUserRivalMusic":           "userRivalMusicList",
		"GetUserPrintedCard":          "userPrintedCardList",
		"GetUserCtoCPlay":             "userCtoCPlayList",
		"GetTeamCourseSetting":        "teamCourseSettingList",
		"GetTeamCourseRule":           "teamCourseRuleList",
	}[endpoint]
	return map[string]any{"userId": userID, "length": 0, "nextIndex": -1, key: []any{}}
}

func writePayload(w http.ResponseWriter, payload any) { _ = json.NewEncoder(w).Encode(payload) }

func writeResponse(w http.ResponseWriter, endpoint string, returnCode int, data any, message string) {
	if data != nil && returnCode == 1 {
		writePayload(w, data)
		return
	}
	writePayload(w, map[string]any{"returnCode": returnCode, "apiName": endpoint + "Api", "message": message})
}

func pathTail(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return parts[len(parts)-1]
}

func requestInt64(request map[string]any, key string) int64 {
	switch value := request[key].(type) {
	case float64:
		return int64(value)
	case json.Number:
		parsed, _ := value.Int64()
		return parsed
	case string:
		parsed, _ := strconv.ParseInt(value, 10, 64)
		return parsed
	default:
		return 0
	}
}

func intValue(value any) int {
	switch value := value.(type) {
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := strconv.Atoi(value.String())
		return parsed
	case string:
		parsed, _ := strconv.Atoi(value)
		return parsed
	case int:
		return value
	default:
		return 0
	}
}

func stringValue(value any) string {
	stringValue, _ := value.(string)
	return stringValue
}

func firstObject(value any) map[string]any {
	values := objectList(value)
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func objectList(value any) []map[string]any {
	values, ok := value.([]any)
	if !ok {
		if object, ok := value.(map[string]any); ok {
			return []map[string]any{object}
		}
		return nil
	}
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if object, ok := value.(map[string]any); ok {
			result = append(result, object)
		}
	}
	return result
}
