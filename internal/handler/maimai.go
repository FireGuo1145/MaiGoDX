package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

// MaimaiHandler dispatches the public maimai2 servlet methods. All player-scoped
// responses are selected by the userId carried by the game request.
func MaimaiHandler(w http.ResponseWriter, r *http.Request) {
	rawAPI := pathTail(r.URL.Path)
	apiName := resolveMaimaiAPI(rawAPI, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeGameResponse(w, apiName, 0, nil, "unable to read request body")
		return
	}
	if len(apiName) == 32 {
		log.Printf("[MaiGoDX] encrypted maimai endpoint is not mapped: %s", rawAPI)
		writeGameResponse(w, rawAPI, 1, nil, "")
		return
	}

	switch apiName {
	case "UpsertUserAll":
		var req model.UpsertUserAllRequest
		if err := json.Unmarshal(body, &req); err != nil {
			writeGameResponse(w, apiName, 0, nil, "invalid UpsertUserAll payload")
			return
		}
		if err := UpsertUserAll(req); err != nil {
			writeGameResponse(w, apiName, 0, nil, err.Error())
			return
		}
		writeGameResponse(w, apiName, 1, nil, "")
		return
	case "UploadUserPlaylog":
		handleUploadUserPlaylog(w, apiName, body)
		return
	case "UploadUserPhoto":
		r.Body = io.NopCloser(bytes.NewReader(body))
		HandleUploadUserPhoto(w, r, apiName)
		return
	case "UploadUserPortrait":
		r.Body = io.NopCloser(bytes.NewReader(body))
		HandleUploadUserPortrait(w, r, apiName)
		return
	case "UpsertUserPrint":
		r.Body = io.NopCloser(bytes.NewReader(body))
		HandleUpsertUserPrint(w, r, apiName)
		return
	}

	if isStaticMaimaiAPI(apiName) {
		if apiName == "UpsertUserPlaceCircleRegist" {
			writeGameResponse(w, apiName, 0, nil, "")
			return
		}
		if data, handled, err := maimaiReadPayload(apiName, 0, body); handled {
			if err != nil {
				writeGameResponse(w, apiName, 0, nil, err.Error())
			} else {
				writeGameResponse(w, apiName, 1, data, "")
			}
			return
		}
		writeGameResponse(w, apiName, 1, nil, "")
		return
	}

	userID := requestUserID(body)
	if userID == 0 {
		writeGameResponse(w, apiName, 0, nil, "missing userId")
		return
	}

	if apiName == "GetUserPortrait" {
		HandleGetUserPortrait(w, r, apiName, userID)
		return
	}

	if data, handled, err := maimaiReadPayload(apiName, userID, body); handled {
		if err != nil {
			writeGameResponse(w, apiName, 0, nil, err.Error())
			return
		}
		writeGameResponse(w, apiName, 1, data, "")
		return
	}

	// AquaDX deliberately treats unknown game API calls as no-ops so a newer
	// client does not fail solely because it requested an optional endpoint.
	writeGameResponse(w, apiName, 1, nil, "")
}

func handleUploadUserPlaylog(w http.ResponseWriter, apiName string, body []byte) {
	var request struct {
		UserID  int64             `json:"userId"`
		Playlog model.UserPlaylog `json:"userPlaylog"`
	}
	if err := json.Unmarshal(body, &request); err != nil || request.UserID == 0 {
		writeGameResponse(w, apiName, 0, nil, "invalid UploadUserPlaylog payload")
		return
	}
	var detail model.UserDetail
	if err := database.DB.Where("user_id = ?", request.UserID).First(&detail).Error; err != nil {
		queuePlaylog(request.UserID, request.Playlog)
		writeGameResponse(w, apiName, 1, nil, "")
		return
	}
	if err := SaveUserPlaylog(database.DB, detail.UserID, request.Playlog); err != nil {
		writeGameResponse(w, apiName, 0, nil, err.Error())
		return
	}
	writeGameResponse(w, apiName, 1, nil, "")
}

func requestUserID(body []byte) int64 {
	var request struct {
		UserID int64 `json:"userId"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return 0
	}
	return request.UserID
}

func userRating(userID int64) map[string]interface{} {
	var detail model.UserDetail
	_ = database.DB.Where("user_id = ?", userID).First(&detail).Error
	var udemae model.UserUdemae
	_ = database.DB.Where("user_id = ?", userID).First(&udemae).Error
	return map[string]interface{}{
		"userId": userID,
		"userRating": map[string]interface{}{
			"rating":            detail.Rating,
			"ratingList":        loadRateData(userID, ratingKeyCurrent),
			"newRatingList":     loadRateData(userID, ratingKeyNew),
			"nextRatingList":    loadRateData(userID, ratingKeyNext),
			"nextNewRatingList": loadRateData(userID, ratingKeyNextNew),
			"udemae":            udemae,
		},
	}
}

func loadRateData(userID int64, key string) []model.UserRate {
	var value model.UserGeneralData
	if err := database.DB.Where("user_id = ? AND property_key = ?", userID, key).First(&value).Error; err != nil || value.PropertyValue == "" {
		return []model.UserRate{}
	}
	rates := make([]model.UserRate, 0)
	for _, item := range strings.Split(value.PropertyValue, ",") {
		var rate model.UserRate
		if _, err := fmt.Sscanf(item, "%d:%d:%d:%d", &rate.MusicID, &rate.Level, &rate.RomVersion, &rate.Achievement); err == nil {
			rates = append(rates, rate)
		}
	}
	return rates
}

func isStaticMaimaiAPI(apiName string) bool {
	switch apiName {
	case "CreateToken", "CMGetSellingCard", "CMUpsertUserPrintlog", "GetGameSetting", "GetGameEvent", "GetGameCharge", "GetGameFesta", "GetGameKaleidxScope", "GetGameMusicScore", "GetGameNationalData", "GetGameNgMusicId", "GetGameTournamentInfo", "GetGameWeeklyData", "GetGameMapAreaCondition", "GetPlaceCircleData", "GetUserCardPrintError", "GetUserCircleData", "GetUserCirclePointRanking", "GetUserFesta", "GetUserFriendCheck", "GetUserScoreRanking", "Ping", "RemoveToken", "UserFriendRegist", "UserLogout", "CMLogin", "CMLogout", "CMUpsertBuyCard", "UpsertClientBookkeeping", "UpsertClientPlayTime", "UpsertClientSetting", "UpsertClientTestmode", "UpsertClientUpload", "UpsertUserChargelog", "UpsertUserPlaceCircleRegist":
		return true
	default:
		return false
	}
}

func pathTail(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return parts[len(parts)-1]
}

func writeGameResponse(w http.ResponseWriter, apiName string, returnCode int, data interface{}, message string) {
	// AquaDX serializes successful handler maps/lists directly. Only no-op and
	// error responses use the servlet returnCode/apiName envelope.
	if returnCode == 1 && data != nil {
		_ = json.NewEncoder(w).Encode(data)
		return
	}
	responseAPI := apiName
	if len(apiName) != 32 && apiName != "Ping" && !strings.HasSuffix(apiName, "Api") {
		responseAPI += "Api"
	}
	_ = json.NewEncoder(w).Encode(model.Response{ReturnCode: returnCode, ApiName: "com.sega.maimai2servlet.api." + responseAPI, Message: message})
}

// ComputeEndpointHash returns the protocol MD5 routing hash for callers that need it.
func ComputeEndpointHash(endpoint string, salt string) string {
	hasher := md5.New()
	hasher.Write([]byte(endpoint + salt))
	return hex.EncodeToString(hasher.Sum(nil))
}
