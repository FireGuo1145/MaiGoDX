package handler

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

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
	case "GetUserPreview", "CMGetUserPreview":
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
		var values []model.UserItem
		database.DB.Where("user_id = ?", userID).Find(&values)
		return map[string]interface{}{"userId": userID, "userItemList": values}, true, nil
	case "GetUserLoginBonus":
		var values []model.UserLoginBonus
		database.DB.Where("user_id = ?", userID).Find(&values)
		return values, true, nil
	case "GetUserMap":
		var values []model.UserMap
		database.DB.Where("user_id = ?", userID).Find(&values)
		return values, true, nil
	case "GetUserCourse":
		var values []model.UserCourse
		database.DB.Where("user_id = ?", userID).Find(&values)
		return values, true, nil
	case "GetUserMusic":
		var values []model.UserMusicDetail
		database.DB.Where("user_id = ?", userID).Find(&values)
		return map[string]interface{}{"userMusicDetailList": values}, true, nil
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
		var values []model.UserKaleidx
		database.DB.Where("user_id = ?", userID).Order("gate_id asc").Find(&values)
		return values, true, nil
	case "GetUserIntimate":
		var values []model.UserIntimate
		database.DB.Where("user_id = ?", userID).Find(&values)
		return values, true, nil
	case "GetUserRating":
		return userRating(userID), true, nil
	case "GetGameRanking":
		var ranking []model.UserDetail
		database.DB.Order("rating desc, max_rating desc, id asc").Limit(100).Find(&ranking)
		return map[string]interface{}{"userId": userID, "gameRankingList": ranking}, true, nil
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
	case "GetUserGhost", "GetUserCard", "GetUserCharge", "GetUserFriendSeasonRanking", "GetUserFavoriteItem", "GetUserNewItemList":
		return []interface{}{}, true, nil
	case "GetUserCardPrintError":
		return map[string]interface{}{"length": 0, "userPrintDetailList": []interface{}{}}, true, nil
	case "GetGameNgMusicId":
		return map[string]interface{}{"length": 0, "musicIdList": []interface{}{}, "ngMusicDataList": []interface{}{}}, true, nil
	}
	return nil, false, nil
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
