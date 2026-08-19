package chunithm

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

type matchingRoom struct {
	members []map[string]any
	started time.Time
}

var matching = struct {
	sync.Mutex
	nextID int
	rooms  map[int]*matchingRoom
}{rooms: map[int]*matchingRoom{}}

var chuniUserLists = map[string][2]string{
	"GetUserItem":               {"userItemList", "userItemList"},
	"GetUserCharacter":          {"userCharacterList", "userCharacterList"},
	"GetUserMapArea":            {"userMapAreaList", "userMapAreaList"},
	"GetUserCourse":             {"userCourseList", "userCourseList"},
	"GetUserCharge":             {"userChargeList", "userChargeList"},
	"GetUserDuel":               {"userDuelList", "userDuelList"},
	"GetUserGacha":              {"userGachaList", "userGachaList"},
	"GetUserActivity":           {"userActivityList", "userActivityList"},
	"GetUserRecentRating":       {"userRecentRatingList", "userRecentRatingList"},
	"GetUserMate":               {"userMateList", "userMateList"},
	"GetUserRegion":             {"userRegionList", "userRegionList"},
	"GetUserFavoriteItem":       {"userFavoriteItemList", "userFavoriteItemList"},
	"GetUserFavoriteCollection": {"userFavoriteCollectionList", "userFavoriteCollectionList"},
	"GetUserCMission":           {"userCMissionList", "userCMissionList"},
	"GetUserCMissionList":       {"userCMissionList", "userCMissionList"},
	"GetUserLV":                 {"userLinkedVerseList", "userLinkedVerseList"},
	"GetUserUC":                 {"userUnlockChallengeList", "userUnlockChallengeList"},
	"GetUserCardPrintError":     {"userCardPrintStateList", "userCardPrintErrorList"},
	"GetUserLoginBonus":         {"userLoginBonusList", "userLoginBonusList"},
}

func chuniConfig(key, fallback string) string {
	var config model.SystemConfig
	if database.DB.Where("key = ?", key).First(&config).Error != nil || strings.TrimSpace(config.Value) == "" {
		return fallback
	}
	return strings.TrimSpace(config.Value)
}

func chuniConfigInt(key string, fallback int) int {
	value, err := strconv.Atoi(chuniConfig(key, strconv.Itoa(fallback)))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

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
	case "GetUserData", "CMGetUserData":
		handleGetUserData(w, userID)
	case "GetUserPreview", "CMGetUserPreview":
		handleGetUserPreview(w, userID)
	case "GetUserMusic":
		handleGetUserMusic(w, userID)
	case "GetGameSetting":
		writePayload(w, gameSetting(r, request))
	case "GetUserOption":
		writePayload(w, map[string]any{"userId": userID, "userGameOption": userRecord(userID, "userGameOption")})
	case "GetUserTeam":
		writePayload(w, userTeam(request, userID))
	case "GetUserSymbolChatSetting":
		values := userRecords(userID, "userSymbolChatSettingList")
		writePayload(w, map[string]any{"userId": userID, "length": len(values), "nextIndex": -1, "symbolChatInfoList": values})
	case "GetUserItem", "CMGetUserItem", "GetUserCharacter", "CMGetUserCharacter", "GetUserCourse", "GetUserCharge", "GetUserDuel", "GetUserGacha", "GetUserActivity", "GetUserRecentRating", "GetUserMate", "GetUserRegion", "GetUserFavoriteItem", "GetUserFavoriteCollection", "GetUserCMission", "GetUserCMissionList", "GetUserLV", "GetUserUC", "GetUserCardPrintError", "CMGetUserCardPrintError", "GetUserLoginBonus":
		writePayload(w, pagedResponse(endpoint, userID, request))
	case "GetUserMapArea":
		writePayload(w, userMapArea(userID, request))
	case "UpsertUserChargelog":
		handleChargelog(w, request, userID)
	case "CMUpsertUserGacha":
		handleCardMakerGacha(w, request, userID)
	case "CMUpsertUserPrint", "CMUpsertUserPrintlog", "CMUpsertUserPrintCancel", "CMUpsertUserPrintSubtract":
		writeResponse(w, endpoint, 1, map[string]any{"returnCode": 1, "apiName": endpoint + "Api", "orderId": "0", "serialId": "FAKECARDIMAG12345678"}, "")
	case "BeginMatching":
		writePayload(w, beginMatching(request))
	case "GetMatchingState":
		writePayload(w, matchingState(request))
	case "EndMatching":
		writePayload(w, endMatching(request))
	case "RemoveMatchingMember", "GameLogout", "RemoveToken", "CreateToken", "PrinterLogin", "PrinterLogout", "Ping", "UpsertClientError", "UpsertClientGameStart", "UpsertClientGameEnd":
		writeResponse(w, endpoint, 1, nil, "")
	case "GetGameGachaCardById":
		writePayload(w, map[string]any{"gachaId": intValue(request["gachaId"]), "length": 0, "isPickup": false, "gameGachaCardList": []any{}, "emissionList": []any{}, "afterCalcList": []any{}})
	case "RollGacha":
		writePayload(w, map[string]any{"length": 0, "gameGachaCardList": []any{}})
	case "GetGameEvent", "GetGameCharge":
		writePayload(w, gameStaticPayload(endpoint))
	case "GetGameGacha", "GetGameRanking", "GetGameCourseLevel", "GetGameUCCondition", "GetGameLVConditionOpen", "GetGameLVConditionUnlock", "GetGameMapAreaCondition", "GetGameIdlist", "GetUserRecMusic", "GetUserRecRating", "GetUserNetBattleData", "GetUserNetBattleRankingInfo", "GetUserRivalData", "GetUserRivalMusic", "GetUserPrintedCard", "GetUserCtoCPlay", "GetTeamCourseSetting", "GetTeamCourseRule":
		writePayload(w, emptyGamePayload(endpoint, userID))
	case "GameLogin", "UpsertClientBookkeeping", "UpsertClientDevelop", "UpsertClientPlayTime", "UpsertClientSetting", "UpsertClientTestmode", "UpsertClientUpload", "UserLogout", "CMLogin", "CMLogout":
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
		return persistCollections(tx, userID, all)
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
		writePayload(w, map[string]any{
			"userId": userID, "isLogin": false, "userName": "", "playerRating": 0,
			"level": 0, "exp": 0, "lastGameId": "", "lastRomVersion": "", "lastDataVersion": "",
			"lastPlayDate": "", "lastLoginDate": "", "banState": 0,
		})
		return
	}
	data := map[string]any{}
	_ = json.Unmarshal([]byte(profile.ProfileJSON), &data)
	data["userId"] = userID
	data["userName"] = profile.UserName
	data["playerRating"] = profile.PlayerRating
	data["isLogin"] = false
	if _, ok := data["banState"]; !ok {
		data["banState"] = 0
	}
	writePayload(w, data)
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
	gameID := "SDHD"
	if strings.Contains(strings.ToUpper(r.URL.Path), "/SDGS/") {
		gameID = "SDGS"
	}
	base := "http://" + r.Host + "/g/" + gameID + "/" + version + "/ChuniServlet/"
	matchingURI := chuniConfig("chuni_matching_uri", base)
	reflectorURI := chuniConfig("chuni_reflector_uri", "")
	game := map[string]any{
		"romVersion": version, "dataVersion": version,
		"isMaintenance": strings.EqualFold(chuniConfig("chuni_maintenance_mode", "false"), "true"), "requestInterval": 0,
		"rebootStartTime": now.Add(-4 * time.Hour).Format("2006-01-02 15:04:05"), "rebootEndTime": now.Add(-3 * time.Hour).Format("2006-01-02 15:04:05"),
		"isBackgroundDistribute": false,
		"maxCountCharacter":      chuniConfigInt("chuni_max_count_character", 300),
		"maxCountItem":           chuniConfigInt("chuni_max_count_item", 300),
		"maxCountMusic":          chuniConfigInt("chuni_max_count_music", 300),
		"matchStartTime":         now.Format("2006-01-02") + " 00:01:00", "matchEndTime": now.Format("2006-01-02") + " 23:59:00", "matchTimeLimit": 10, "matchErrorLimit": 10,
		"matchingUri": matchingURI, "matchingUriX": matchingURI,
	}
	if reflectorURI != "" {
		game["reflectorUri"] = reflectorURI
		game["udpHolePunchUri"] = reflectorURI
	}
	return map[string]any{"gameSetting": game, "isDumpUpload": false, "isAou": false}
}

func pagedResponse(endpoint string, userID int64, request map[string]any) map[string]any {
	if endpoint == "CMGetUserItem" {
		endpoint = "GetUserItem"
	}
	if endpoint == "CMGetUserCharacter" {
		endpoint = "GetUserCharacter"
	}
	if endpoint == "CMGetUserCardPrintError" {
		endpoint = "GetUserCardPrintError"
	}
	kind, key := userListSpec(endpoint)
	values := userRecords(userID, kind)
	if endpoint == "GetUserItem" || endpoint == "GetUserActivity" || endpoint == "GetUserFavoriteItem" {
		requestedKind := intValue(request["kind"])
		if endpoint == "GetUserItem" && requestedKind == 0 {
			requestedKind = int(requestInt64(request, "nextIndex") / 10_000_000_000)
		}
		values = filterByInt(values, "itemKind", requestedKind)
		if endpoint == "GetUserActivity" {
			values = filterByInt(values, "kind", requestedKind)
		}
	}
	if endpoint == "GetUserLoginBonus" && !strings.EqualFold(chuniConfig("chuni_login_bonus_enable", "false"), "true") {
		values = nil
	}
	return map[string]any{"userId": userID, "length": len(values), "nextIndex": -1, key: values}
}

func userMapArea(userID int64, request map[string]any) map[string]any {
	values := userRecords(userID, "userMapAreaList")
	requested := map[int]struct{}{}
	if raw, ok := request["mapAreaIdList"].([]any); ok {
		for _, item := range raw {
			if value, ok := item.(map[string]any); ok {
				requested[intValue(value["mapAreaId"])] = struct{}{}
			}
		}
	}
	if len(requested) > 0 {
		filtered := make([]map[string]any, 0, len(values))
		for _, value := range values {
			if _, ok := requested[intValue(value["mapAreaId"])]; ok {
				filtered = append(filtered, value)
			}
		}
		values = filtered
	}
	return map[string]any{"userId": userID, "userMapAreaList": values}
}

func userTeam(request map[string]any, userID int64) map[string]any {
	teamName := chuniConfig("chuni_team_name", "")
	if strings.TrimSpace(teamName) == "" {
		return map[string]any{"userId": userID, "teamId": 0}
	}
	return map[string]any{
		"userId": userID, "teamId": 1, "teamRank": 1, "teamName": teamName,
		"userTeamPoint": map[string]any{"userId": userID, "teamId": 1, "orderId": 1, "teamPoint": 1, "aggrDate": stringValue(request["playDate"])},
	}
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

func gameStaticPayload(endpoint string) map[string]any {
	switch endpoint {
	case "GetGameEvent":
		events := []model.GameEvent{}
		database.DB.Order("id asc").Find(&events)
		return map[string]any{"type": 1, "length": len(events), "gameEventList": events}
	case "GetGameCharge":
		charges := []model.GameCharge{}
		database.DB.Order("charge_id asc").Find(&charges)
		return map[string]any{"length": len(charges), "gameChargeList": charges}
	default:
		return emptyGamePayload(endpoint, 0)
	}
}

func persistCollections(tx *gorm.DB, userID int64, all map[string]any) error {
	for kind, raw := range all {
		if kind == "userData" || kind == "userMusicDetailList" || kind == "userPlaylogList" {
			continue
		}
		for index, value := range objectList(raw) {
			if err := saveUserRecord(tx, userID, kind, recordKey(value, index), value); err != nil {
				return err
			}
		}
	}
	return nil
}

func handleChargelog(w http.ResponseWriter, request map[string]any, userID int64) {
	if userID == 0 {
		writeResponse(w, "UpsertUserChargelog", 0, nil, "missing userId")
		return
	}
	charge, _ := request["userCharge"].(map[string]any)
	if charge != nil {
		if err := saveUserRecord(database.DB, userID, "userChargeList", recordKey(charge, 0), charge); err != nil {
			writeResponse(w, "UpsertUserChargelog", 0, nil, err.Error())
			return
		}
	}
	writeResponse(w, "UpsertUserChargelog", 1, nil, "")
}

func handleCardMakerGacha(w http.ResponseWriter, request map[string]any, userID int64) {
	if userID == 0 {
		writeResponse(w, "CMUpsertUserGacha", 0, nil, "missing userId")
		return
	}
	payload, _ := request["cmUpsertUserGacha"].(map[string]any)
	if payload != nil {
		if err := persistCollections(database.DB, userID, payload); err != nil {
			writeResponse(w, "CMUpsertUserGacha", 0, nil, err.Error())
			return
		}
		for index, state := range objectList(payload["gameGachaCardList"]) {
			if err := saveUserRecord(database.DB, userID, "userCardPrintStateList", recordKey(state, index), state); err != nil {
				writeResponse(w, "CMUpsertUserGacha", 0, nil, err.Error())
				return
			}
		}
	}
	writePayload(w, map[string]any{"returnCode": 1, "apiName": "CMUpsertUserGachaApi", "userCardPrintStateList": []any{}})
}

func saveUserRecord(db *gorm.DB, userID int64, kind, key string, data map[string]any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	record := model.ChuniUserRecord{UserID: userID, Kind: kind, RecordKey: key, DataJSON: string(payload)}
	var existing model.ChuniUserRecord
	if err := db.Where("user_id = ? AND kind = ? AND record_key = ?", userID, kind, key).First(&existing).Error; err == nil {
		record.ID = existing.ID
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Save(&record).Error
}

func userRecord(userID int64, kind string) map[string]any {
	values := userRecords(userID, kind)
	if len(values) == 0 {
		return map[string]any{}
	}
	return values[0]
}

func userRecords(userID int64, kind string) []map[string]any {
	var records []model.ChuniUserRecord
	database.DB.Where("user_id = ? AND kind = ?", userID, kind).Order("record_key asc").Find(&records)
	values := make([]map[string]any, 0, len(records))
	for _, record := range records {
		value := map[string]any{}
		if json.Unmarshal([]byte(record.DataJSON), &value) == nil {
			values = append(values, value)
		}
	}
	return values
}

func userListSpec(endpoint string) (string, string) {
	if endpoint == "CMGetUserItem" {
		endpoint = "GetUserItem"
	}
	if endpoint == "CMGetUserCharacter" {
		endpoint = "GetUserCharacter"
	}
	if endpoint == "CMGetUserCardPrintError" {
		endpoint = "GetUserCardPrintError"
	}
	spec := chuniUserLists[endpoint]
	return spec[0], spec[1]
}

func recordKey(value map[string]any, index int) string {
	for _, field := range []string{"characterId", "itemId", "mapAreaId", "activityId", "chargeId", "courseId", "duelId", "gachaId", "unlockChallengeId", "linkedVerseId", "mateId", "regionId", "musicId", "missionId", "cardId", "id"} {
		if raw, ok := value[field]; ok {
			if field == "itemId" || field == "activityId" {
				return strconv.Itoa(intValue(value["kind"])) + ":" + strconv.Itoa(intValue(raw))
			}
			return strconv.Itoa(intValue(raw))
		}
	}
	return strconv.Itoa(index)
}

func filterByInt(values []map[string]any, key string, wanted int) []map[string]any {
	if wanted == 0 {
		return values
	}
	filtered := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if intValue(value[key]) == wanted {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func beginMatching(request map[string]any) map[string]any {
	member, _ := request["matchingMemberInfo"].(map[string]any)
	if member == nil {
		member = map[string]any{}
	}
	matching.Lock()
	defer matching.Unlock()
	for id, room := range matching.rooms {
		if len(room.members) < 4 && time.Since(room.started) < 120*time.Second {
			room.members = append(room.members, member)
			return map[string]any{"roomId": id, "matchingWaitState": waitState(room)}
		}
	}
	matching.nextID++
	room := &matchingRoom{members: []map[string]any{member}, started: time.Now()}
	matching.rooms[matching.nextID] = room
	return map[string]any{"roomId": matching.nextID, "matchingWaitState": waitState(room)}
}

func matchingState(request map[string]any) map[string]any {
	matching.Lock()
	defer matching.Unlock()
	room := matching.rooms[intValue(request["roomId"])]
	if room == nil {
		return map[string]any{"roomId": intValue(request["roomId"]), "matchingWaitState": waitState(nil)}
	}
	return map[string]any{"roomId": intValue(request["roomId"]), "matchingWaitState": waitState(room)}
}

func endMatching(request map[string]any) map[string]any {
	matching.Lock()
	defer matching.Unlock()
	roomID := intValue(request["roomId"])
	room := matching.rooms[roomID]
	delete(matching.rooms, roomID)
	if room == nil {
		return map[string]any{"matchingMemberInfoList": []any{}, "matchingMemberRoleList": []any{}, "matchingResult": 0}
	}
	roles := make([]map[string]any, len(room.members))
	for index := range room.members {
		roles[index] = map[string]any{"role": index}
	}
	return map[string]any{"matchingMemberInfoList": room.members, "matchingMemberRoleList": roles, "matchingResult": 1}
}

func waitState(room *matchingRoom) map[string]any {
	if room == nil {
		return map[string]any{"matchingMemberInfoList": []any{}, "matchingMemberCount": 0, "matchingWaitTime": 0, "state": 0}
	}
	remaining := int(120 - time.Since(room.started).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return map[string]any{"matchingMemberInfoList": room.members, "matchingMemberCount": len(room.members), "matchingWaitTime": remaining, "state": 1}
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
