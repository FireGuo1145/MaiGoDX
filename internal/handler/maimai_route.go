package handler

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

var maimaiAPINames = []string{
	"GetGameChargeApi", "GetGameEventApi", "GetGameFestaApi", "GetGameMusicScoreApi", "GetGameNgMusicIdApi", "GetGameRankingApi", "GetGameSettingApi", "GetGameWeeklyDataApi", "GetGameMapAreaConditionApi",
	"GetPlaceCircleDataApi", "GetUserActivityApi", "GetUserCardApi", "GetUserCardPrintErrorApi", "GetUserCharacterApi",
	"GetUserCircleChallengeApi", "GetUserCircleDataApi", "GetUserCirclePointDataApi", "GetUserCirclePointRankingApi", "GetUserCourseApi", "GetUserDataApi", "GetUserExtendApi", "GetUserFavoriteApi",
	"GetUserFavoriteItemApi", "GetUserFestaApi", "GetUserGhostApi", "GetUserIntimateApi", "GetUserItemApi",
	"GetUserKaleidxScopeApi", "GetUserLoginBonusApi", "GetUserMapApi", "GetUserMissionDataApi", "GetUserMusicApi",
	"GetUserOptionApi", "GetUserPreviewApi", "GetUserRatingApi", "GetUserRecommendRateMusicApi", "GetUserRecommendSelectMusicApi", "GetUserScoreRankingApi",
	"UploadUserPhotoApi", "UploadUserPlaylogApi", "UploadUserPlaylogListApi", "UpsertUserAllApi", "UpsertUserChargelogApi",
	"UpsertUserPlaceCircleRegistApi", "UpsertUserPrintApi", "UpsertClientBookkeepingApi", "UpsertClientPlayTimeApi", "UpsertClientSettingApi", "UpsertClientTestmodeApi", "UpsertClientUploadApi",
	"CMGetUserCharacterApi", "CMGetUserDataApi", "CMGetUserItemApi", "CMGetUserPreviewApi", "CMLoginApi", "CMLogoutApi", "CMUpsertBuyCardApi", "CMUpsertUserPrintApi", "Ping", "RemoveTokenApi", "UserLoginApi", "UserLogoutApi",
}

func resolveMaimaiAPI(rawAPI, requestPath string) string {
	if len(rawAPI) != 32 || rawAPI != strings.ToLower(rawAPI) {
		return strings.TrimSuffix(rawAPI, "Api")
	}
	var config model.SystemConfig
	if err := database.DB.Where(&model.SystemConfig{Key: "maimai_endpoint_salts"}).First(&config).Error; err != nil || strings.TrimSpace(config.Value) == "" {
		return rawAPI
	}
	suffix := ""
	if strings.Contains(requestPath, "/SDGA/") {
		suffix = "MaimaiExp"
	} else if strings.Contains(requestPath, "/SDGB/") {
		suffix = "MaimaiChn"
	}
	for _, salt := range strings.Split(config.Value, ",") {
		saltBytes, err := hex.DecodeString(strings.TrimSpace(salt))
		if err != nil {
			continue
		}
		for _, endpoint := range maimaiAPINames {
			hash := md5.Sum(append([]byte(endpoint+suffix), saltBytes...))
			if hex.EncodeToString(hash[:]) == rawAPI {
				return strings.TrimSuffix(endpoint, "Api")
			}
		}
	}
	return rawAPI
}
