package ongeki

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

var userListEndpoints = map[string][2]string{
	"GetUserTechEvent":         {"userTechEventList", "userTechEventList"},
	"GetUserBoss":              {"userBossList", "userBossList"},
	"GetUserCard":              {"userCardList", "userCardList"},
	"GetUserChapter":           {"userChapterList", "userChapterList"},
	"GetUserMemoryChapter":     {"userMemoryChapterList", "userMemoryChapterList"},
	"GetUserCharacter":         {"userCharacterList", "userCharacterList"},
	"GetUserDeckByKey":         {"userDeckList", "userDeckList"},
	"GetUserEventMusic":        {"userEventMusicList", "userEventMusicList"},
	"GetUserEventPoint":        {"userEventPointList", "userEventPointList"},
	"GetUserKop":               {"userKopList", "userKopList"},
	"GetUserLoginBonus":        {"userLoginBonusList", "userLoginBonusList"},
	"GetUserMissionPoint":      {"userMissionPointList", "userMissionPointList"},
	"GetUserMusicItem":         {"userMusicItemList", "userMusicItemList"},
	"GetUserRival":             {"userRivalList", "userRivalList"},
	"GetUserScenario":          {"userScenarioList", "userScenarioList"},
	"GetUserSkin":              {"userSkinList", "userSkinList"},
	"GetUserStory":             {"userStoryList", "userStoryList"},
	"GetUserTechCount":         {"userTechCountList", "userTechCountList"},
	"GetUserTrainingRoomByKey": {"userTrainingRoomList", "userTrainingRoomList"},
	"GetUserGacha":             {"userGachaList", "userGachaList"},
	"CMGetUserCard":            {"userCardList", "userCardList"},
	"CMGetUserCharacter":       {"userCharacterList", "userCharacterList"},
}

var ratingCollectionKinds = map[string]struct{}{
	"userRecentRatingList": {}, "userBpBaseList": {}, "userRatingBaseBestNewList": {},
	"userRatingBaseBestList": {}, "userRatingBaseHotList": {}, "userRatingBaseNextNewList": {},
	"userRatingBaseNextList": {}, "userRatingBaseHotNextList": {}, "userNewRatingBasePScoreList": {},
	"userNewRatingBaseBestList": {}, "userNewRatingBaseBestNewList": {}, "userNewRatingBaseNextPScoreList": {},
	"userNewRatingBaseNextBestList": {}, "userNewRatingBaseNextBestNewList": {},
}

var knownEndpoints = []string{
	"CMGetUserCardApi", "CMGetUserCharacterApi", "CMGetUserDataApi", "CMGetUserGachaSupplyApi", "CMGetUserItemApi",
	"CMUpsertUserAllApi", "CMUpsertUserGachaApi", "CMUpsertUserPrintApi", "CMUpsertUserPrintlogApi", "CMUpsertUserPrintPlaylogApi", "CMUpsertUserSelectGachaApi",
	"ExtendLockTimeApi", "GameLoginApi", "GameLogoutApi", "GetClientBookkeepingApi", "GetClientTestmodeApi",
	"GetGameEventApi", "GetGameGachaApi", "GetGameGachaCardByIdApi", "GetGameIdlistApi", "GetGameMessageApi", "GetGameMusicReleaseStateApi",
	"GetGamePointApi", "GetGamePresentApi", "GetGameRankingApi", "GetGameRewardApi", "GetGameSettingApi", "GetGameTechMusicApi", "GetGameTheaterApi",
	"GetUserActivityApi", "GetUserBossApi", "GetUserBpBaseApi", "GetUserCardApi", "GetUserChapterApi", "GetUserCharacterApi", "GetUserDataApi",
	"GetUserDeckByKeyApi", "GetUserEventMapApi", "GetUserEventMusicApi", "GetUserEventPointApi", "GetUserEventRankingApi", "GetUserGachaApi",
	"GetUserItemApi", "GetUserKopApi", "GetUserLoginBonusApi", "GetUserMemoryChapterApi", "GetUserMissionPointApi", "GetUserMusicApi", "GetUserMusicItemApi",
	"GetUserOptionApi", "GetUserPreviewApi", "GetUserRatinglogApi", "GetUserRecentRatingApi", "GetUserRegionApi", "GetUserRivalApi", "GetUserRivalDataApi",
	"GetUserRivalMusicApi", "GetUserScenarioApi", "GetUserSkinApi", "GetUserStoryApi", "GetUserTechCountApi", "GetUserTechEventApi", "GetUserTechEventRankingApi",
	"GetUserTradeItemApi", "GetUserTrainingRoomByKeyApi", "PingApi", "PrinterLoginApi", "PrinterLogoutApi", "RegisterPromotionCardApi", "RollGachaApi",
	"UpsertClientBookkeepingApi", "UpsertClientDevelopApi", "UpsertClientErrorApi", "UpsertClientSettingApi", "UpsertClientTestmodeApi", "UpsertUserAllApi", "UpsertUserGplogApi",
}

// Handler implements AquaDX-compatible Ongeki JSON endpoints on SDDT. Ongeki
// uses MaimaiServlet in cabinet URLs despite being a separate game protocol.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rawEndpoint := pathTail(r.URL.Path)
	endpoint := resolveEndpoint(rawEndpoint)
	if endpoint == "" {
		writePayload(w, defaultNoOpPayload(rawEndpoint))
		return
	}
	endpoint = strings.TrimSuffix(endpoint, "Api")

	request := map[string]any{}
	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil && !errors.Is(err, io.EOF) {
		// Some cabinet ping/logout calls carry an empty body.
		writePayload(w, map[string]any{"returnCode": 0, "apiName": endpoint + "Api"})
		return
	}
	if version := pathGameVersion(r.URL.Path); version != "" {
		request["version"] = version
	}
	userID := int64Value(request["userId"])

	switch endpoint {
	case "UpsertUserAll":
		handleUpsertUserAll(w, request, userID)
	case "CMUpsertUserAll", "CMUpsertUserGacha", "CMUpsertUserSelectGacha":
		handleCardMakerUpsert(w, endpoint, request, userID)
	case "GetUserData":
		handleGetUserData(w, userID)
	case "CMGetUserData":
		handleCMGetUserData(w, userID)
	case "GetUserPreview":
		handleGetUserPreview(w, userID)
	case "GetUserOption":
		writePayload(w, map[string]any{"userId": userID, "userOption": firstRecord(userID, "userOption")})
	case "GetUserEventMap":
		writePayload(w, map[string]any{"userId": userID, "userEventMap": firstRecord(userID, "userEventMap")})
	case "GetUserMusic":
		handleGetUserMusic(w, userID)
	case "GetUserItem", "CMGetUserItem":
		handleGetUserItem(w, userID, request)
	case "GetUserActivity":
		handleGetUserActivity(w, userID, request)
	case "GetUserRecentRating":
		writeUnpaged(w, userID, "userRecentRatingList", recentRatings(userID), nil)
	case "GetUserTradeItem":
		handleGetUserTradeItem(w, userID, request)
	case "GetUserTechEventRanking":
		handleTechEventRanking(w, userID)
	case "GetUserEventRanking":
		handleEventRanking(w, userID)
	case "GetUserRegion":
		writeUnpaged(w, userID, "userRegionList", records(userID, "userRegionList"), nil)
	case "GetUserBpBase", "GetUserRatinglog":
		key := lowerFirst(strings.TrimPrefix(endpoint, "Get")) + "List"
		writeUnpaged(w, userID, key, []map[string]any{}, nil)
	case "GetUserRivalData":
		handleGetUserRivalData(w, request)
	case "GetUserRivalMusic":
		handleGetUserRivalMusic(w, userID, request)
	case "GetGameSetting":
		writePayload(w, gameSetting(request))
	case "GetGameEvent":
		writePayload(w, gameEvents())
	case "GetGamePoint":
		writePayload(w, gamePoints())
	case "GetGamePresent":
		writePayload(w, gamePresents())
	case "GetGameReward":
		writePayload(w, staticGameListPayload("game_reward.json", "gameRewardList"))
	case "GetGameTechMusic":
		writePayload(w, map[string]any{"length": 0, "gameTechMusicList": []any{}})
	case "GetGameMessage":
		writePayload(w, map[string]any{"type": intValue(request["type"]), "length": 0, "gameMessageList": []any{}})
	case "GetGameMusicReleaseState":
		writePayload(w, map[string]any{"techScore": 0, "cardNum": 0})
	case "GetGameIdlist":
		writePayload(w, map[string]any{"type": intValue(request["type"]), "length": 0, "gameIdlistList": []any{}})
	case "GetGameRanking":
		writePayload(w, map[string]any{"type": intValue(request["type"]), "length": 0, "gameRankingList": []any{}})
	case "GetClientBookkeeping":
		writePayload(w, map[string]any{"placeId": request["placeId"], "length": 0, "clientBookkeepingList": []any{}})
	case "GetClientTestmode":
		writePayload(w, map[string]any{"placeId": request["placeId"], "length": 0, "clientTestmodeList": []any{}})
	case "GetGameGacha":
		writePayload(w, gameGachas())
	case "GetGameGachaCardById":
		writePayload(w, gameGachaCards(int64Value(request["gachaId"])))
	case "RollGacha":
		if !ongekiUserExists(userID) {
			writeOngekiError(w, http.StatusBadRequest, "User not found")
			return
		}
		if _, exists := findGacha(int64Value(request["gachaId"])); !exists {
			writeOngekiError(w, http.StatusNotFound, "Gacha not found")
			return
		}
		writePayload(w, rollGacha(userID, request))
	case "CMGetUserGachaSupply":
		writePayload(w, map[string]any{"supplyId": 0, "length": 0, "supplyCardList": []any{}})
	case "GetGameTheater":
		writePayload(w, map[string]any{"length": 0, "gameTheaterList": []any{}, "registIdList": []any{}})
	case "GameLogin":
		writePayload(w, map[string]any{"returnCode": "1"})
	case "PrinterLogin", "PrinterLogout":
		writePayload(w, map[string]any{"returnCode": 1})
	case "CMUpsertUserPrint", "CMUpsertUserPrintlog", "CMUpsertUserPrintPlaylog":
		writePayload(w, defaultNoOpPayload(endpoint))
	case "ExtendLockTime", "GameLogout", "RegisterPromotionCard", "UpsertClientBookkeeping", "UpsertClientDevelop", "UpsertClientError", "UpsertClientSetting", "UpsertClientTestmode", "UpsertUserGplog":
		writePayload(w, map[string]any{"returnCode": 1, "apiName": endpoint + "Api"})
	case "Ping":
		writePayload(w, map[string]any{"returnCode": 1, "apiName": "Ping"})
	default:
		if spec, ok := userListEndpoints[endpoint]; ok {
			values := records(userID, spec[0])
			if endpoint == "CMGetUserCard" {
				for _, value := range values {
					value["printCount"] = 99
				}
			}
			writeUnpaged(w, userID, spec[1], values, nil)
			return
		}
		writePayload(w, defaultNoOpPayload(endpoint))
	}
}

func handleUpsertUserAll(w http.ResponseWriter, request map[string]any, userID int64) {
	all, _ := request["upsertUserAll"].(map[string]any)
	if userID <= 0 || all == nil {
		writePayload(w, map[string]any{"returnCode": 0, "apiName": "UpsertUserAllApi"})
		return
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := saveProfile(tx, userID, firstObject(all["userData"])); err != nil {
			return err
		}
		for _, detail := range objectList(all["userMusicDetailList"]) {
			if err := saveMusicDetail(tx, userID, detail); err != nil {
				return err
			}
		}
		for _, playlog := range objectList(all["userPlaylogList"]) {
			if err := createPlaylog(tx, userID, playlog); err != nil {
				return err
			}
		}
		if err := persistAllCollections(tx, userID, all); err != nil {
			return err
		}
		regionID := intValue(request["regionId"])
		if regionID > 0 && len(objectList(all["userPlaylogList"])) > 0 {
			return incrementRegion(tx, userID, regionID)
		}
		return nil
	})
	if err != nil {
		writePayload(w, map[string]any{"returnCode": 0, "apiName": "UpsertUserAllApi", "message": err.Error()})
		return
	}
	writePayload(w, defaultNoOpPayload("UpsertUserAll"))
}

func handleCardMakerUpsert(w http.ResponseWriter, endpoint string, request map[string]any, userID int64) {
	key := map[string]string{"CMUpsertUserAll": "cmUpsertUserAll", "CMUpsertUserGacha": "cmUpsertUserGacha", "CMUpsertUserSelectGacha": "cmUpsertUserSelectGacha"}[endpoint]
	all, _ := request[key].(map[string]any)
	if userID <= 0 || all == nil {
		writePayload(w, map[string]any{"returnCode": 0, "apiName": endpoint + "Api"})
		return
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := saveProfile(tx, userID, firstObject(all["userData"])); err != nil {
			return err
		}
		if err := persistAllCollections(tx, userID, all); err != nil {
			return err
		}
		switch endpoint {
		case "CMUpsertUserGacha":
			return updateUserGacha(tx, userID, request)
		case "CMUpsertUserSelectGacha":
			return useUserGachaSelection(tx, userID, request)
		default:
			return nil
		}
	})
	if err != nil {
		writePayload(w, map[string]any{"returnCode": 0, "apiName": endpoint + "Api", "message": err.Error()})
		return
	}
	writePayload(w, defaultNoOpPayload(endpoint))
}

func updateUserGacha(tx *gorm.DB, userID int64, request map[string]any) error {
	gachaID := int64Value(request["gachaId"])
	if gachaID <= 0 {
		return nil
	}
	pullCount := intValue(request["gachaCnt"])
	selectPoint := intValue(request["selectPoint"])
	key := strconv.FormatInt(gachaID, 10)
	state := existingRecord(tx, userID, "userGachaList", key)
	isNew := len(state) == 0
	now := time.Now()
	if isNew {
		state = map[string]any{
			"gachaId": gachaID, "totalGachaCnt": 0, "selectPoint": 0, "useSelectPoint": 0,
			"ceilingGachaCnt": 0, "fiveGachaCnt": 0, "elevenGachaCnt": 0, "dailyGachaCnt": 0,
		}
	}
	state["totalGachaCnt"] = intValue(state["totalGachaCnt"]) + pullCount
	if sameCalendarDay(stringValue(state["dailyGachaDate"]), now) {
		state["dailyGachaCnt"] = intValue(state["dailyGachaCnt"]) + pullCount
	} else {
		state["dailyGachaCnt"] = pullCount
	}
	state["dailyGachaDate"] = now.Format("2006-01-02 15:04:05")
	if selectPoint > 0 {
		state["selectPoint"] = intValue(state["selectPoint"]) + selectPoint
		if isNew {
			state["ceilingGachaCnt"] = 1
		} else {
			state["ceilingGachaCnt"] = intValue(state["ceilingGachaCnt"]) + pullCount
		}
	}
	if pullCount == 5 {
		state["fiveGachaCnt"] = intValue(state["fiveGachaCnt"]) + 1
	}
	if pullCount == 11 {
		state["elevenGachaCnt"] = intValue(state["elevenGachaCnt"]) + 1
	}
	return saveRecord(tx, userID, "userGachaList", key, state)
}

func sameCalendarDay(value string, now time.Time) bool {
	for _, layout := range []string{"2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05", time.RFC3339Nano} {
		parsed, err := time.ParseInLocation(layout, value, now.Location())
		if err == nil {
			year, month, day := parsed.Date()
			nowYear, nowMonth, nowDay := now.Date()
			return year == nowYear && month == nowMonth && day == nowDay
		}
	}
	return false
}

func useUserGachaSelection(tx *gorm.DB, userID int64, request map[string]any) error {
	logs := objectList(request["selectGachaLogList"])
	if len(logs) == 0 {
		return nil
	}
	gachaID := int64Value(logs[0]["gachaId"])
	if gachaID <= 0 {
		return nil
	}
	key := strconv.FormatInt(gachaID, 10)
	state := existingRecord(tx, userID, "userGachaList", key)
	if len(state) == 0 {
		state = map[string]any{
			"gachaId": gachaID, "totalGachaCnt": 0, "selectPoint": 0, "useSelectPoint": 0,
			"ceilingGachaCnt": 0, "fiveGachaCnt": 0, "elevenGachaCnt": 0, "dailyGachaCnt": 0,
		}
	}
	state["selectPoint"] = 0
	state["useSelectPoint"] = 1
	return saveRecord(tx, userID, "userGachaList", key, state)
}

func existingRecord(tx *gorm.DB, userID int64, kind, key string) map[string]any {
	var record model.OngekiUserRecord
	if tx.Where("user_id = ? AND kind = ? AND record_key = ?", userID, kind, key).First(&record).Error != nil {
		return nil
	}
	return decodeJSON(record.DataJSON)
}

func saveProfile(tx *gorm.DB, userID int64, data map[string]any) error {
	if data == nil {
		return nil
	}
	var existing model.OngekiUser
	result := tx.Where("user_id = ?", userID).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if existing.ID != 0 {
		if previous := decodeJSON(existing.ProfileJSON); stringValue(data["eventWatchedDate"]) == "" {
			data["eventWatchedDate"] = stringValue(previous["lastPlayDate"])
		}
		if stringValue(data["cmEventWatchedDate"]) == "" {
			data["cmEventWatchedDate"] = stringValue(decodeJSON(existing.ProfileJSON)["lastPlayDate"])
		}
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	profile := model.OngekiUser{
		UserID: userID, UserName: stringValue(data["userName"]), Level: intValue(data["level"]),
		ReincarnationNum: intValue(data["reincarnationNum"]), Point: int64Value(data["point"]), TotalPoint: int64Value(data["totalPoint"]),
		PlayCount: intValue(data["playCount"]), PlayerRating: intValue(data["playerRating"]), HighestRating: intValue(data["highestRating"]),
		NewPlayerRating: intValue(data["newPlayerRating"]), NewHighestRating: intValue(data["newHighestRating"]), ProfileJSON: string(payload),
	}
	if existing.ID != 0 {
		profile.Model = existing.Model
	}
	return tx.Save(&profile).Error
}

func saveMusicDetail(tx *gorm.DB, userID int64, data map[string]any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	entry := model.OngekiMusicDetail{UserID: userID, MusicID: intValue(data["musicId"]), Level: intValue(data["level"]), TechScoreMax: intValue(data["techScoreMax"]), DetailJSON: string(payload)}
	var existing model.OngekiMusicDetail
	result := tx.Where("user_id = ? AND music_id = ? AND level = ?", entry.UserID, entry.MusicID, entry.Level).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if existing.ID != 0 {
		entry.Model = existing.Model
	}
	return tx.Save(&entry).Error
}

func createPlaylog(tx *gorm.DB, userID int64, data map[string]any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return tx.Create(&model.OngekiPlaylog{UserID: userID, MusicID: intValue(data["musicId"]), Level: intValue(data["level"]), TechScore: intValue(data["techScore"]), PlayDate: stringValue(data["userPlayDate"]), PlaylogJSON: string(payload)}).Error
}

func persistAllCollections(tx *gorm.DB, userID int64, all map[string]any) error {
	for kind, raw := range all {
		if kind == "userData" || kind == "userMusicDetailList" || kind == "userPlaylogList" || strings.HasPrefix(kind, "isNew") || kind == "clientSystemInfo" {
			continue
		}
		if _, ok := ratingCollectionKinds[kind]; ok {
			if err := replaceCollection(tx, userID, kind, objectList(raw)); err != nil {
				return err
			}
			continue
		}
		values := objectList(raw)
		if value, ok := raw.(map[string]any); ok {
			values = []map[string]any{value}
		}
		for index, value := range values {
			if err := saveRecord(tx, userID, kind, recordKey(value, index), value); err != nil {
				return err
			}
		}
	}
	return nil
}

func replaceCollection(tx *gorm.DB, userID int64, kind string, values []map[string]any) error {
	if err := tx.Unscoped().Where("user_id = ? AND kind = ?", userID, kind).Delete(&model.OngekiUserRecord{}).Error; err != nil {
		return err
	}
	for index, value := range values {
		if err := saveRecord(tx, userID, kind, orderedIndexKey(index), value); err != nil {
			return err
		}
	}
	return nil
}

func saveRecord(tx *gorm.DB, userID int64, kind, key string, data map[string]any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	record := model.OngekiUserRecord{UserID: userID, Kind: kind, RecordKey: key, DataJSON: string(payload)}
	var existing model.OngekiUserRecord
	result := tx.Where("user_id = ? AND kind = ? AND record_key = ?", userID, kind, key).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if existing.ID != 0 {
		record.Model = existing.Model
	}
	return tx.Save(&record).Error
}

func incrementRegion(tx *gorm.DB, userID int64, regionID int) error {
	key := strconv.Itoa(regionID)
	value := map[string]any{"regionId": regionID, "playCount": 1, "created": time.Now().Format("2006-01-02")}
	var record model.OngekiUserRecord
	result := tx.Where("user_id = ? AND kind = ? AND record_key = ?", userID, "userRegionList", key).Limit(1).Find(&record)
	if result.Error != nil {
		return result.Error
	}
	if record.ID != 0 {
		value = decodeJSON(record.DataJSON)
		value["playCount"] = intValue(value["playCount"]) + 1
	}
	return saveRecord(tx, userID, "userRegionList", key, value)
}

func handleGetUserData(w http.ResponseWriter, userID int64) {
	var profile model.OngekiUser
	result := database.DB.Where("user_id = ?", userID).Limit(1).Find(&profile)
	if result.Error != nil || result.RowsAffected == 0 {
		writePayload(w, map[string]any{"userId": userID, "userData": nil})
		return
	}
	writePayload(w, map[string]any{"userId": userID, "userData": decodeJSON(profile.ProfileJSON)})
}

func handleCMGetUserData(w http.ResponseWriter, userID int64) {
	var profile model.OngekiUser
	result := database.DB.Where("user_id = ?", userID).Limit(1).Find(&profile)
	if result.Error != nil || result.RowsAffected == 0 {
		writeOngekiError(w, http.StatusBadRequest, "User not found")
		return
	}
	writePayload(w, map[string]any{"userId": userID, "userData": decodeJSON(profile.ProfileJSON)})
}

func ongekiUserExists(userID int64) bool {
	if userID <= 0 {
		return false
	}
	var count int64
	return database.DB.Model(&model.OngekiUser{}).Where("user_id = ?", userID).Count(&count).Error == nil && count > 0
}

func writeOngekiError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	writePayload(w, map[string]any{"message": message})
}

func handleGetUserPreview(w http.ResponseWriter, userID int64) {
	var profile model.OngekiUser
	result := database.DB.Where("user_id = ?", userID).Limit(1).Find(&profile)
	if result.Error != nil || result.RowsAffected == 0 {
		writePayload(w, emptyPreview(userID))
		return
	}
	data := decodeJSON(profile.ProfileJSON)
	preview := map[string]any{
		"userId": userID, "isLogin": false, "userName": profile.UserName, "reincarnationNum": profile.ReincarnationNum,
		"level": profile.Level, "exp": int64Value(data["exp"]), "playerRating": profile.PlayerRating, "newPlayerRating": profile.NewPlayerRating,
		"lastGameId": stringValue(data["lastGameId"]), "lastRomVersion": stringValue(data["lastRomVersion"]), "lastDataVersion": stringValue(data["lastDataVersion"]),
		"lastPlayDate": stringValue(data["lastPlayDate"]), "lastLoginDate": stringValue(data["lastPlayDate"]), "nameplateId": intValue(data["nameplateId"]),
		"trophyId": intValue(data["trophyId"]), "cardId": intValue(data["cardId"]), "lastEmoneyBrand": 4, "lastEmoneyCredit": 10000,
		"dispPlayerLv": 1, "dispRating": 1, "dispBP": 1, "headphone": 0, "banStatus": 0, "isWarningConfirmed": false,
	}
	if option := firstRecord(userID, "userOption"); len(option) > 0 {
		for _, key := range []string{"dispPlayerLv", "dispRating", "dispBP", "headphone"} {
			preview[key] = intValue(option[key])
		}
	}
	writePayload(w, preview)
}

func emptyPreview(userID int64) map[string]any {
	return map[string]any{
		"userId": userID, "isLogin": false, "lastLoginDate": "0000-00-00 00:00:00", "userName": "", "reincarnationNum": 0,
		"level": 0, "exp": 0, "playerRating": 0, "lastGameId": "", "lastRomVersion": "", "lastDataVersion": "", "lastPlayDate": "",
		"nameplateId": 0, "trophyId": 0, "cardId": 0, "dispPlayerLv": 0, "dispRating": 0, "dispBP": 0, "headphone": 0,
		"banStatus": 0, "isWarningConfirmed": true,
	}
}

func handleGetUserMusic(w http.ResponseWriter, userID int64) {
	var details []model.OngekiMusicDetail
	database.DB.Where("user_id = ?", userID).Order("music_id asc, level asc").Find(&details)
	groups := make([]map[string]any, 0)
	currentID := -1
	var current []map[string]any
	flush := func() {
		if current != nil {
			groups = append(groups, map[string]any{"length": len(current), "userMusicDetailList": current})
		}
	}
	for _, detail := range details {
		if detail.MusicID != currentID {
			flush()
			currentID = detail.MusicID
			current = []map[string]any{}
		}
		current = append(current, decodeJSON(detail.DetailJSON))
	}
	flush()
	writePayload(w, map[string]any{"userId": userID, "length": len(groups), "nextIndex": -1, "userMusicList": groups})
}

func handleGetUserItem(w http.ResponseWriter, userID int64, request map[string]any) {
	kind := int(int64Value(request["nextIndex"]) / 10_000_000_000)
	values := filterInt(records(userID, "userItemList"), "itemKind", kind)
	writePayload(w, map[string]any{"userId": userID, "length": len(values), "nextIndex": -1, "itemKind": kind, "userItemList": values})
}

func handleGetUserActivity(w http.ResponseWriter, userID int64, request map[string]any) {
	kind := intValue(request["kind"])
	values := filterInt(records(userID, "userActivityList"), "kind", kind)
	sort.SliceStable(values, func(i, j int) bool { return intValue(values[i]["sortNumber"]) > intValue(values[j]["sortNumber"]) })
	limit := 10
	if kind == 1 {
		limit = 15
	}
	if len(values) > limit {
		values = values[:limit]
	}
	writeUnpaged(w, userID, "userActivityList", values, map[string]any{"kind": kind})
}

func handleGetUserTradeItem(w http.ResponseWriter, userID int64, request map[string]any) {
	start, end := intValue(request["startChapterId"]), intValue(request["endChapterId"])
	values := records(userID, "userTradeItemList")
	filtered := make([]map[string]any, 0, len(values))
	for _, value := range values {
		chapterID := intValue(value["chapterId"])
		if chapterID >= start && chapterID <= end {
			filtered = append(filtered, value)
		}
	}
	writeUnpaged(w, userID, "userTradeItemList", filtered, nil)
}

func handleTechEventRanking(w http.ResponseWriter, userID int64) {
	now := time.Now().Format("2006-01-02 15:04:05.0")
	values := records(userID, "userTechEventList")
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		result = append(result, map[string]any{"eventId": value["eventId"], "date": now, "rank": 1, "totalTechScore": value["totalTechScore"], "totalPlatinumScore": value["totalPlatinumScore"]})
	}
	writeUnpaged(w, userID, "userTechEventRankingList", result, nil)
}

func handleEventRanking(w http.ResponseWriter, userID int64) {
	now := time.Now().Format("2006-01-02 15:04:05.0")
	values := records(userID, "userEventPointList")
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		eventID := int64Value(value["eventId"])
		point := int64Value(value["point"])
		result = append(result, map[string]any{"eventId": value["eventId"], "type": 1, "date": now, "rank": eventRank(eventID, point), "point": value["point"]})
	}
	writeUnpaged(w, userID, "userEventRankingList", result, nil)
}

func eventRank(eventID, point int64) int {
	var rows []model.OngekiUserRecord
	database.DB.Where("kind = ?", "userEventPointList").Find(&rows)
	pointsByUser := make(map[int64]int64)
	for _, row := range rows {
		value := decodeJSON(row.DataJSON)
		if int64Value(value["eventId"]) != eventID {
			continue
		}
		candidate := int64Value(value["point"])
		if previous, ok := pointsByUser[row.UserID]; !ok || candidate > previous {
			pointsByUser[row.UserID] = candidate
		}
	}
	rank := 1
	for _, candidate := range pointsByUser {
		if candidate > point {
			rank++
		}
	}
	return rank
}

func handleGetUserRivalData(w http.ResponseWriter, request map[string]any) {
	requested := objectList(request["userRivalList"])
	ids := make([]int64, 0, len(requested))
	seen := make(map[int64]struct{}, len(requested))
	for _, rival := range requested {
		id := int64Value(rival["rivalUserId"])
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		writePayload(w, []any{})
		return
	}
	var profiles []model.OngekiUser
	database.DB.Where("user_id IN ?", ids).Find(&profiles)
	byID := make(map[int64]model.OngekiUser, len(profiles))
	for _, profile := range profiles {
		byID[profile.UserID] = profile
	}
	result := make([]map[string]any, 0, len(profiles))
	for _, id := range ids {
		if profile, ok := byID[id]; ok {
			result = append(result, map[string]any{"rivalUserId": profile.UserID, "rivalUserName": profile.UserName})
		}
	}
	writePayload(w, result)
}

func handleGetUserRivalMusic(w http.ResponseWriter, userID int64, request map[string]any) {
	rivalUserID := int64Value(request["rivalUserId"])
	var details []model.OngekiMusicDetail
	database.DB.Where("user_id = ?", rivalUserID).Order("music_id asc, level asc").Find(&details)
	groups := make([]map[string]any, 0)
	currentMusicID := -1
	var current []map[string]any
	flush := func() {
		if current != nil {
			groups = append(groups, map[string]any{"userRivalMusicDetailList": current, "length": len(current)})
		}
	}
	for _, detail := range details {
		if detail.MusicID != currentMusicID {
			flush()
			currentMusicID = detail.MusicID
			current = []map[string]any{}
		}
		current = append(current, decodeJSON(detail.DetailJSON))
	}
	flush()
	writeUnpaged(w, userID, "userRivalMusicList", groups, map[string]any{"rivalUserId": rivalUserID})
}

func gameSetting(request map[string]any) map[string]any {
	version := stringValue(request["version"])
	if version == "" {
		version = "1.50.00"
	}
	return map[string]any{"isAou": false, "isDumpUpload": false, "gameSetting": map[string]any{
		"dataVersion": version, "onlineDataVersion": version, "isMaintenance": configBool("ongeki_maintenance_mode", false),
		"requestInterval": configInt("ongeki_request_interval", 10), "rebootStartTime": "2020-01-01 23:59:00.0", "rebootEndTime": "2020-01-01 23:59:00.0",
		"isBackgroundDistribute": false, "maxCountCharacter": configInt("ongeki_max_count_character", 50), "maxCountCard": configInt("ongeki_max_count_card", 300),
		"maxCountItem": configInt("ongeki_max_count_item", 300), "maxCountMusic": configInt("ongeki_max_count_music", 50),
		"maxCountMusicItem": configInt("ongeki_max_count_music_item", 300), "maxCountRivalMusic": configInt("ongeki_max_count_rival_music", 300),
	}}
}

func gameEvents() map[string]any {
	events := loadGameData("game_event.json")
	values := make([]map[string]any, 0, len(events))
	for _, event := range events {
		values = append(values, map[string]any{"id": event["id"], "type": 1, "startDate": "2005-01-01 00:00:00.0", "endDate": "2099-01-01 05:00:00.0"})
	}
	return map[string]any{"type": 1, "length": len(values), "gameEventList": values}
}

func gamePoints() map[string]any {
	if configured := loadGameData("game_point.json"); len(configured) > 0 {
		for _, point := range configured {
			point["startDate"] = "2000-01-01 05:00:00.0"
			point["endDate"] = "2099-01-01 05:00:00.0"
		}
		return map[string]any{"length": len(configured), "gamePointList": configured}
	}
	values := make([]map[string]any, 0, 6)
	for index, cost := range []int{100, 200, 300, 333, 666, 999} {
		values = append(values, map[string]any{"id": index + 1, "type": index, "cost": cost, "startDate": "2000-01-01 05:00:00.0", "endDate": "2099-01-01 05:00:00.0"})
	}
	return map[string]any{"length": len(values), "gamePointList": values}
}

func writeUnpaged(w http.ResponseWriter, userID int64, key string, values []map[string]any, extra map[string]any) {
	payload := map[string]any{"userId": userID, "nextIndex": 0, "length": len(values), key: values}
	for key, value := range extra {
		payload[key] = value
	}
	writePayload(w, payload)
}

func records(userID int64, kind string) []map[string]any {
	var rows []model.OngekiUserRecord
	database.DB.Where("user_id = ? AND kind = ?", userID, kind).Order("record_key asc").Find(&rows)
	values := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		values = append(values, decodeJSON(row.DataJSON))
	}
	return values
}

func recentRatings(userID int64) []map[string]any {
	values := records(userID, "userRecentRatingList")
	if len(values) > 0 {
		return values
	}
	var plays []model.OngekiPlaylog
	database.DB.Where("user_id = ?", userID).Order("id desc").Limit(30).Find(&plays)
	values = make([]map[string]any, 0, len(plays))
	for _, play := range plays {
		values = append(values, map[string]any{"musicId": play.MusicID, "difficultId": play.Level, "romVersionCode": "1000000", "score": play.TechScore})
	}
	return values
}

func firstRecord(userID int64, kind string) map[string]any {
	values := records(userID, kind)
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func firstRecordByKey(userID int64, kind, key string) map[string]any {
	var row model.OngekiUserRecord
	if database.DB.Where("user_id = ? AND kind = ? AND record_key = ?", userID, kind, key).Limit(1).Find(&row).RowsAffected == 0 {
		return nil
	}
	return decodeJSON(row.DataJSON)
}

func recordKey(value map[string]any, index int) string {
	if value["activityId"] != nil {
		return strconv.Itoa(intValue(value["kind"])) + ":" + strconv.Itoa(intValue(value["activityId"]))
	}
	if value["itemId"] != nil {
		return strconv.Itoa(intValue(value["itemKind"])) + ":" + strconv.Itoa(intValue(value["itemId"]))
	}
	if value["eventId"] != nil && value["musicId"] != nil {
		return strconv.Itoa(intValue(value["eventId"])) + ":" + strconv.Itoa(intValue(value["type"])) + ":" + strconv.Itoa(intValue(value["musicId"]))
	}
	if value["chapterId"] != nil && value["tradeItemId"] != nil {
		return strconv.Itoa(intValue(value["chapterId"])) + ":" + strconv.Itoa(intValue(value["tradeItemId"]))
	}
	if value["kopId"] != nil && value["areaId"] != nil {
		return strconv.Itoa(intValue(value["kopId"])) + ":" + strconv.Itoa(intValue(value["areaId"]))
	}
	fields := []string{"activityId", "musicId", "characterId", "cardId", "deckId", "roomId", "storyId", "chapterId", "itemId", "bonusId", "eventId", "levelId", "scenarioId", "tradeItemId", "kopId", "gachaId", "id"}
	parts := make([]string, 0, 3)
	for _, field := range fields {
		if raw, ok := value[field]; ok {
			if field == "musicId" && value["level"] != nil {
				parts = append(parts, strconv.Itoa(intValue(value["level"])))
			}
			if field == "eventId" && value["type"] != nil {
				parts = append(parts, strconv.Itoa(intValue(value["type"])))
			}
			parts = append(parts, strconv.FormatInt(int64Value(raw), 10))
			return strings.Join(parts, ":")
		}
	}
	return orderedIndexKey(index)
}

func orderedIndexKey(index int) string { return fmt.Sprintf("%09d", index) }

func resolveEndpoint(raw string) string {
	for _, endpoint := range knownEndpoints {
		if raw == endpoint || raw == strings.TrimSuffix(endpoint, "Api") {
			return endpoint
		}
	}
	if len(raw) != 32 {
		return ""
	}
	for _, spec := range endpointSaltSpecs() {
		for _, endpoint := range knownEndpoints {
			digest := pbkdf2.Key([]byte(endpoint), spec.salt, spec.iterations, 16, sha1.New)
			if strings.EqualFold(raw, hex.EncodeToString(digest)) {
				return endpoint
			}
		}
	}
	return ""
}

type endpointSalt struct {
	salt       []byte
	iterations int
}

func endpointSaltSpecs() []endpointSalt {
	value := config("ongeki_endpoint_salts", "")
	result := []endpointSalt{}
	for _, entry := range strings.Split(value, ",") {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 2 {
			continue
		}
		salt, err := hex.DecodeString(parts[0])
		iterations, iterErr := strconv.Atoi(parts[1])
		if err == nil && iterErr == nil && iterations > 0 {
			result = append(result, endpointSalt{salt: salt, iterations: iterations})
		}
	}
	if strings.TrimSpace(value) == "" {
		for _, entry := range loadGameData("game_encryption.json") {
			salt, err := hex.DecodeString(stringValue(entry["salt"]))
			iterations := intValue(entry["iterations"])
			if err == nil && len(salt) > 0 && iterations > 0 {
				result = append(result, endpointSalt{salt: salt, iterations: iterations})
			}
		}
	}
	return result
}

func config(key, fallback string) string {
	var entry model.SystemConfig
	if database.DB == nil || database.DB.Where("key = ?", key).First(&entry).Error != nil || strings.TrimSpace(entry.Value) == "" {
		return fallback
	}
	return strings.TrimSpace(entry.Value)
}

func configInt(key string, fallback int) int {
	value, err := strconv.Atoi(config(key, strconv.Itoa(fallback)))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func configBool(key string, fallback bool) bool {
	value := config(key, strconv.FormatBool(fallback))
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func objectList(raw any) []map[string]any {
	values, ok := raw.([]any)
	if !ok {
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

func firstObject(raw any) map[string]any {
	values := objectList(raw)
	if len(values) == 0 {
		return nil
	}
	return values[0]
}

func filterInt(values []map[string]any, key string, wanted int) []map[string]any {
	filtered := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if intValue(value[key]) == wanted {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func decodeJSON(raw string) map[string]any {
	value := map[string]any{}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func int64Value(value any) int64 {
	switch value := value.(type) {
	case json.Number:
		parsed, _ := value.Int64()
		return parsed
	case float64:
		return int64(value)
	case int:
		return int64(value)
	case int64:
		return value
	case string:
		parsed, _ := strconv.ParseInt(value, 10, 64)
		return parsed
	default:
		return 0
	}
}

func intValue(value any) int { return int(int64Value(value)) }

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(toJSON(value)), "\"", ""))
}

func toJSON(value any) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

func pathTail(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return parts[len(parts)-1]
}

func pathGameVersion(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for index, part := range parts {
		if strings.EqualFold(part, "SDDT") && index+1 < len(parts) && looksLikeVersion(parts[index+1]) {
			return strings.TrimSpace(parts[index+1])
		}
	}
	return ""
}

func looksLikeVersion(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}
	return true
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func defaultNoOpPayload(endpoint string) map[string]any {
	return map[string]any{"returnCode": 1, "apiName": lowerFirst(strings.TrimSuffix(endpoint, "Api"))}
}

func writePayload(w http.ResponseWriter, payload any) { _ = json.NewEncoder(w).Encode(payload) }
