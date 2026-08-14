package handler

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

func maimaiReadPayload(apiName string, userID int64, body []byte) (interface{}, bool, error) {
	if payload, handled := maimaiCompatibilityPayload(apiName, userID, body); handled {
		return payload, true, nil
	}
	var detail model.UserDetail
	loadProfile := func() error {
		if err := database.DB.Where("user_id = ?", userID).First(&detail).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("player profile not found")
			}
			return err
		}
		return nil
	}

	switch apiName {
	case "CMGetUserPreview":
		if err := loadProfile(); err != nil {
			return nil, true, err
		}
		return map[string]interface{}{
			"userId": userID, "userName": detail.UserName, "rating": detail.Rating,
			"lastDataVersion": detail.LastDataVersion, "isLogin": false, "isExistSellingCard": false,
		}, true, nil
	case "GetUserPreview":
		if err := loadProfile(); err != nil {
			return nil, true, err
		}
		var option model.UserOption
		_ = database.DB.Where("user_id = ?", userID).First(&option).Error
		return map[string]interface{}{
			"userId": userID, "userName": detail.UserName, "isLogin": false,
			"lastGameId": detail.LastGameID, "lastDataVersion": detail.LastDataVersion,
			"lastRomVersion": detail.LastRomVersion, "lastLoginDate": detail.LastPlayDate,
			"lastPlayDate": detail.LastPlayDate, "playerRating": detail.Rating,
			"nameplateId": detail.PlateID, "iconId": detail.IconID, "trophyId": 0,
			"partnerId": detail.PartnerID, "frameId": detail.FrameID, "totalAwake": detail.TotalAwake,
			"isNetMember": detail.IsNetMember, "dailyBonusDate": detail.DailyBonusDate,
			"headPhoneVolume": option.HeadPhoneVolume, "dispRate": option.DispRate,
			"isInherit": false, "banState": 0,
		}, true, nil
	case "GetUserData", "CMGetUserData":
		if err := loadProfile(); err != nil {
			return nil, true, err
		}
		return map[string]interface{}{"userId": userID, "userData": detail, "banState": 0}, true, nil
	case "GetUserExtend":
		var value model.UserExtend
		if err := database.DB.Where("user_id = ?", userID).First(&value).Error; err != nil {
			return nil, true, profileScopedError(err)
		}
		return map[string]interface{}{"userId": userID, "userExtend": value}, true, nil
	case "GetUserOption":
		var value model.UserOption
		if err := database.DB.Where("user_id = ?", userID).First(&value).Error; err != nil {
			return nil, true, profileScopedError(err)
		}
		return map[string]interface{}{"userId": userID, "userOption": value}, true, nil
	case "GetUserCharacter", "CMGetUserCharacter":
		var values []model.UserCharacter
		database.DB.Where("user_id = ?", userID).Find(&values)
		if apiName == "CMGetUserCharacter" {
			return map[string]interface{}{"returnCode": 1, "length": len(values), "userCharacterList": values}, true, nil
		}
		return map[string]interface{}{"userId": userID, "userCharacterList": values}, true, nil
	case "GetUserItem", "CMGetUserItem":
		nextIndex := requestInt64(body, "nextIndex")
		itemKind := int(nextIndex / 10_000_000_000)
		var values []model.UserItem
		database.DB.Where("user_id = ? AND item_kind = ?", userID, itemKind).Order("item_id asc").Find(&values)
		for index := range values {
			values[index].IsValid = true
		}
		return map[string]interface{}{"userId": userID, "nextIndex": 0, "itemKind": itemKind, "userItemList": values}, true, nil
	case "GetUserLoginBonus":
		var values []model.UserLoginBonus
		database.DB.Where("user_id = ?", userID).Find(&values)
		return unpagedMaimaiPayload(userID, "userLoginBonusList", values), true, nil
	case "GetUserMap":
		var values []model.UserMap
		database.DB.Where("user_id = ?", userID).Find(&values)
		return unpagedMaimaiPayload(userID, "userMapList", values), true, nil
	case "GetUserCourse":
		var values []model.UserCourse
		database.DB.Where("user_id = ?", userID).Find(&values)
		return unpagedMaimaiPayload(userID, "userCourseList", values), true, nil
	case "GetUserMusic":
		var values []model.UserMusicDetail
		database.DB.Where("user_id = ?", userID).Find(&values)
		// AquaDX wraps the detail list in one userMusicList entry; returning
		// userMusicDetailList directly leaves the SDGA loader waiting forever.
		return unpagedMaimaiPayload(userID, "userMusicList", []map[string]interface{}{{"userMusicDetailList": values}}), true, nil
	case "GetUserFavorite":
		itemKind := requestInt(body, "itemKind")
		var value model.UserFavorite
		err := database.DB.Where("user_id = ? AND item_kind = ?", userID, itemKind).First(&value).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return map[string]interface{}{"userId": userID, "userFavorite": nil}, true, nil
		}
		if err != nil {
			return nil, true, err
		}
		return map[string]interface{}{"userId": userID, "userFavorite": value}, true, nil
	case "GetUserActivity":
		var values []model.UserActivity
		database.DB.Where("user_id = ?", userID).Order("sort_number asc").Find(&values)
		play, music := make([]model.UserActivity, 0), make([]model.UserActivity, 0)
		for _, value := range values {
			if value.Kind == 1 {
				play = append(play, value)
			} else if value.Kind == 2 {
				music = append(music, value)
			}
		}
		if len(play) > 200 {
			play = play[:200]
		}
		if len(music) > 200 {
			music = music[:200]
		}
		return map[string]interface{}{"userActivity": map[string]interface{}{"playList": play, "musicList": music}}, true, nil
	case "GetUserKaleidxScope":
		if err := loadProfile(); err != nil {
			return nil, true, err
		}
		var values []model.UserKaleidx
		if err := database.DB.Where("user_id = ?", userID).Order("gate_id asc").Find(&values).Error; err != nil {
			return nil, true, err
		}
		gates := make(map[int]model.UserKaleidx, len(values))
		for _, value := range values {
			gates[value.GateID] = value
		}
		unlockGate := func(gateID int) {
			gate, exists := gates[gateID]
			if !exists {
				gate = model.UserKaleidx{UserID: userID, GateID: gateID}
			}
			gate.IsGateFound = true
			gate.IsKeyFound = true
			gates[gateID] = gate
		}
		for gateID := 1; gateID <= 6; gateID++ {
			unlockGate(gateID)
		}
		for gateID := 6; gateID <= 9; gateID++ {
			if gates[gateID].IsClear {
				unlockGate(gateID + 1)
			}
		}
		result := make([]model.UserKaleidx, 0, len(gates))
		for gateID := 1; gateID <= 10; gateID++ {
			if gate, exists := gates[gateID]; exists {
				if gate.ID == 0 {
					if err := database.DB.Create(&gate).Error; err != nil {
						return nil, true, err
					}
				}
				result = append(result, gate)
			}
		}
		return unpagedMaimaiPayload(userID, "userKaleidxScopeList", result), true, nil
	case "GetUserIntimate":
		var values []model.UserIntimate
		database.DB.Where("user_id = ?", userID).Find(&values)
		return unpagedMaimaiPayload(userID, "userIntimateList", values), true, nil
	case "GetUserRating":
		return userRating(userID), true, nil
	case "GetGameRanking":
		var request map[string]json.RawMessage
		_ = json.Unmarshal(body, &request)
		var rankingType interface{}
		_ = json.Unmarshal(request["type"], &rankingType)
		if requestInt(body, "type") != 1 {
			return map[string]interface{}{"type": rankingType, "gameRankingList": []interface{}{}}, true, nil
		}
		type musicRankingRow struct {
			MusicID int
			Weight  int64
		}
		cutoff := time.Now().AddDate(0, 0, -7).Format("2006-01-02 15:04:05")
		var rows []musicRankingRow
		if err := database.DB.Model(&model.UserPlaylog{}).
			Select("music_id, COUNT(DISTINCT user_id) AS weight").
			Where("user_play_date >= ?", cutoff).
			Group("music_id").Order("weight desc, music_id asc").Limit(50).Find(&rows).Error; err != nil {
			return nil, true, err
		}
		ranking := make([]map[string]interface{}, 0, len(rows))
		for _, row := range rows {
			ranking = append(ranking, map[string]interface{}{"id": row.MusicID, "point": row.Weight, "userName": ""})
		}
		return map[string]interface{}{"type": rankingType, "gameRankingList": ranking}, true, nil
	case "GetGameSetting":
		return gameSettingPayload(), true, nil
	case "GetGameEvent":
		return gameEventPayload(), true, nil
	case "GetGameCharge":
		return gameChargePayload(), true, nil
	case "GetUserRecommendRateMusic":
		return recommendedMusicPayload(userID, "maimai_recommend_rate_music_ids", "userRecommendRateMusicIdList"), true, nil
	case "GetUserRecommendSelectMusic":
		return recommendedMusicPayload(userID, "maimai_recommend_select_music_ids", "userRecommendSelectionMusicIdList"), true, nil
	case "GetUserFavoriteItem":
		kind := requestInt(body, "kind")
		propertyKey := ""
		if kind == 1 {
			propertyKey = "favorite_music"
		} else if kind == 2 {
			propertyKey = "favorite_rival"
		}
		items := make([]map[string]int, 0)
		if propertyKey != "" {
			var favorite model.UserGeneralData
			if err := database.DB.Where("user_id = ? AND property_key = ?", userID, propertyKey).First(&favorite).Error; err == nil {
				for orderID, record := range strings.Split(favorite.PropertyValue, ",") {
					if itemID, err := strconv.Atoi(strings.TrimSpace(record)); err == nil {
						items = append(items, map[string]int{"id": itemID, "orderId": orderID})
					}
				}
			}
		}
		return map[string]interface{}{"userId": userID, "kind": kind, "length": len(items), "nextIndex": 0, "userFavoriteItemList": items}, true, nil
	case "GetUserCard", "CMGetUserCard":
		var cards []model.UserGameCard
		database.DB.Where("user_id = ?", userID).Order("card_id asc").Find(&cards)
		return unpagedMaimaiPayload(userID, "userCardList", cards), true, nil
	case "GetUserCharge":
		var charges []model.UserCharge
		database.DB.Where("user_id = ?", userID).Order("charge_id asc").Find(&charges)
		return unpagedMaimaiPayload(userID, "userChargeList", charges), true, nil
	case "GetUserFriendSeasonRanking":
		var rankings []model.UserFriendSeasonRanking
		database.DB.Where("user_id = ?", userID).Order("season_id asc").Find(&rankings)
		return unpagedMaimaiPayload(userID, "userFriendSeasonRankingList", rankings), true, nil
	case "GetUserGhost":
		return unpagedMaimaiPayload(userID, "userGhostList", []interface{}{}), true, nil
	case "GetUserCardPrintError", "CMGetUserCardPrintError":
		return map[string]interface{}{"length": 0, "userPrintDetailList": []interface{}{}}, true, nil
	case "GetGameNgMusicId":
		return map[string]interface{}{"length": 0, "musicIdList": []interface{}{}, "ngMusicDataList": []interface{}{}}, true, nil
	}
	return nil, false, nil
}

// unpagedMaimaiPayload mirrors AquaDX's BaseHandler.unpaged response framing.
// SDGA does not accept a bare JSON array for these endpoints.
func unpagedMaimaiPayload(userID int64, listKey string, values interface{}) map[string]interface{} {
	length := 0
	switch value := values.(type) {
	case []model.UserLoginBonus:
		length = len(value)
	case []model.UserMap:
		length = len(value)
	case []model.UserCourse:
		length = len(value)
	case []model.UserKaleidx:
		length = len(value)
	case []model.UserIntimate:
		length = len(value)
	case []model.UserGameCard:
		length = len(value)
	case []model.UserCharge:
		length = len(value)
	case []model.UserFriendSeasonRanking:
		length = len(value)
	case []map[string]interface{}:
		length = len(value)
	case []interface{}:
		length = len(value)
	}
	return map[string]interface{}{"userId": userID, "nextIndex": 0, "length": length, listKey: values}
}

func profileScopedError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("player profile not found")
	}
	return err
}

func requestInt(body []byte, key string) int {
	var request map[string]json.RawMessage
	if err := json.Unmarshal(body, &request); err != nil {
		return 0
	}
	var value int
	_ = json.Unmarshal(request[key], &value)
	return value
}

func gameSettingPayload() map[string]interface{} {
	configs := map[string]string{}
	var values []model.SystemConfig
	database.DB.Find(&values)
	for _, value := range values {
		configs[value.Key] = value.Value
	}
	return map[string]interface{}{
		"isAouAccession": true,
		"gameSetting": map[string]interface{}{
			"rebootStartTime":                   configNonEmptyValue(configs, "game_reboot_start_time", "2020-01-01 23:59:00.0"),
			"rebootEndTime":                     configNonEmptyValue(configs, "game_reboot_end_time", "2020-01-01 23:59:00.0"),
			"rebootInterval":                    configInt(configs, "game_reboot_interval", 0),
			"isMaintenance":                     strings.EqualFold(configs["maintenance_mode"], "true"),
			"requestInterval":                   configInt(configs, "game_request_interval", 10),
			"movieUploadLimit":                  configInt(configs, "game_movie_upload_limit", 0),
			"movieStatus":                       configInt(configs, "game_movie_status", 0),
			"movieServerUri":                    configs["game_movie_server_uri"],
			"deliverServerUri":                  configs["game_deliver_server_uri"],
			"oldServerUri":                      configs["game_old_server_uri"],
			"usbDlServerUri":                    configs["game_usb_download_server_uri"],
			"pingDisable":                       strings.EqualFold(configValue(configs, "game_ping_disable", "true"), "true"),
			"packetTimeout":                     configInt(configs, "game_packet_timeout", 20000),
			"packetTimeoutLong":                 configInt(configs, "game_packet_timeout_long", 60000),
			"packetRetryCount":                  configInt(configs, "game_packet_retry_count", 5),
			"userDataDlErrTimeout":              configInt(configs, "game_user_data_download_error_timeout", 300000),
			"userDataDlErrRetryCount":           configInt(configs, "game_user_data_download_error_retry_count", 5),
			"userDataDlErrSamePacketRetryCount": configInt(configs, "game_user_data_download_same_packet_retry_count", 5),
			"userDataUpSkipTimeout":             configInt(configs, "game_user_data_upload_skip_timeout", 0),
			"userDataUpSkipRetryCount":          configInt(configs, "game_user_data_upload_skip_retry_count", 0),
			"iconPhotoDisable":                  strings.EqualFold(configValue(configs, "game_icon_photo_disable", "true"), "true"),
			"uploadPhotoDisable":                strings.EqualFold(configValue(configs, "game_upload_photo_disable", "false"), "true"),
			"maxCountMusic":                     configInt(configs, "game_max_count_music", 0),
			"maxCountItem":                      configInt(configs, "game_max_count_item", 0),
		},
	}
}

func configValue(values map[string]string, key, fallback string) string {
	if value, ok := values[key]; ok {
		return value
	}
	return fallback
}
func configNonEmptyValue(values map[string]string, key, fallback string) string {
	if value, ok := values[key]; ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func configInt(values map[string]string, key string, fallback int) int {
	value, err := strconv.Atoi(values[key])
	if err != nil {
		return fallback
	}
	return value
}
