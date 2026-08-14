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
	"strconv"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
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
		req, err := decodeUpsertUserAll(body)
		if err != nil {
			log.Printf("[MaiGoDX] UpsertUserAll rejected: invalid payload bytes=%d error=%v", len(body), err)
			writeGameResponse(w, apiName, 0, nil, "invalid UpsertUserAll payload")
			return
		}
		if err := UpsertUserAll(req); err != nil {
			log.Printf("[MaiGoDX] UpsertUserAll rejected: userId=%d userData=%d playlogs=%d error=%v", req.UserID, len(req.UpsertUserAll.UserData), len(req.UserPlaylogList)+len(req.UpsertUserAll.UserPlaylogList), err)
			writeGameResponse(w, apiName, 0, nil, err.Error())
			return
		}
		log.Printf("[MaiGoDX] UpsertUserAll saved: userId=%d userData=%d playlogs=%d", req.UserID, len(req.UpsertUserAll.UserData), len(req.UserPlaylogList)+len(req.UpsertUserAll.UserPlaylogList))
		writeGameResponse(w, apiName, 1, nil, "")
		return
	case "UploadUserPlaylog":
		handleUploadUserPlaylog(w, apiName, body)
		return
	case "UploadUserPlaylogList":
		handleUploadUserPlaylogList(w, apiName, body)
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

// handleUploadUserPlaylogList is used by SDGA 1.60 during logout. AquaDX
// processes every entry through the ordinary playlog handler so first-play
// scores are retained until UpsertUserAll creates the player profile.
func handleUploadUserPlaylogList(w http.ResponseWriter, apiName string, body []byte) {
	var request struct {
		UserID   int64               `json:"userId"`
		Playlogs []model.UserPlaylog `json:"userPlaylogList"`
	}
	if err := json.Unmarshal(body, &request); err != nil || request.UserID == 0 {
		writeGameResponse(w, apiName, 0, nil, "invalid UploadUserPlaylogList payload")
		return
	}
	var detail model.UserDetail
	if err := database.DB.Where("user_id = ?", request.UserID).First(&detail).Error; err != nil {
		for _, playlog := range request.Playlogs {
			queuePlaylog(request.UserID, playlog)
		}
		log.Printf("[MaiGoDX] UploadUserPlaylogList queued: userId=%d entries=%d", request.UserID, len(request.Playlogs))
		writeGameResponse(w, apiName, 1, nil, "")
		return
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, playlog := range request.Playlogs {
			if err := SaveUserPlaylog(tx, detail.UserID, playlog); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Printf("[MaiGoDX] UploadUserPlaylogList rejected: userId=%d entries=%d error=%v", request.UserID, len(request.Playlogs), err)
		writeGameResponse(w, apiName, 0, nil, err.Error())
		return
	}
	log.Printf("[MaiGoDX] UploadUserPlaylogList saved: userId=%d entries=%d", request.UserID, len(request.Playlogs))
	writeGameResponse(w, apiName, 1, nil, "")
}

// decodeUpsertUserAll accepts AquaDX's list-shaped payload and the SDGA 1.60
// single-object userData variant. The upstream Jackson mapper accepts the
// latter, while encoding/json rejects it before a new profile can be created.
func decodeUpsertUserAll(body []byte) (model.UpsertUserAllRequest, error) {
	var request model.UpsertUserAllRequest
	initialErr := json.Unmarshal(body, &request)
	if initialErr == nil {
		return request, nil
	}

	var wire struct {
		UserID        int64                      `json:"userId"`
		UpsertUserAll map[string]json.RawMessage `json:"upsertUserAll"`
		UserPlaylogs  json.RawMessage            `json:"userPlaylogList"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		return model.UpsertUserAllRequest{}, err
	}
	if wire.UpsertUserAll == nil {
		return model.UpsertUserAllRequest{}, fmt.Errorf("missing upsertUserAll: %w", initialErr)
	}
	// SDGA 1.60 sends one object instead of an array for several optional
	// collection fields. AquaDX's mapper accepts these singleton collections;
	// normalise every known list rather than only userData, otherwise another
	// singleton field still rejects the whole logout save.
	for _, key := range []string{
		"userData", "userOption", "userExtend", "userCharacterList", "userGhost",
		"userMapList", "userLoginBonusList", "userRatingList", "userItemList",
		"userMusicDetailList", "userCourseList", "userFriendSeasonRankingList",
		"userChargeList", "userFavoriteList", "userActivityList", "userGamePlaylogList",
		"userFavoritemusicList", "userKaleidxScopeList", "userIntimateList", "userPlaylogList",
	} {
		value, ok := wire.UpsertUserAll[key]
		if !ok {
			continue
		}
		trimmed := bytes.TrimSpace(value)
		if len(trimmed) > 0 && trimmed[0] == '{' {
			wire.UpsertUserAll[key] = append(append([]byte{'['}, trimmed...), ']')
		}
	}
	rootPlaylogs := bytes.TrimSpace(wire.UserPlaylogs)
	if len(rootPlaylogs) > 0 && rootPlaylogs[0] == '{' {
		rootPlaylogs = append(append([]byte{'['}, rootPlaylogs...), ']')
	}
	if len(rootPlaylogs) == 0 {
		rootPlaylogs = []byte("[]")
	}
	normalized, err := json.Marshal(map[string]interface{}{
		"userId":          wire.UserID,
		"upsertUserAll":   wire.UpsertUserAll,
		"userPlaylogList": json.RawMessage(rootPlaylogs),
	})
	if err != nil {
		return model.UpsertUserAllRequest{}, err
	}
	if err := json.Unmarshal(normalized, &request); err != nil {
		return model.UpsertUserAllRequest{}, fmt.Errorf("decode normalized payload: %w (initial error: %v)", err, initialErr)
	}
	return request, nil
}

func requestUserID(body []byte) int64 {
	var request map[string]interface{}
	if err := json.Unmarshal(body, &request); err != nil {
		return 0
	}
	val, ok := request["userId"]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int64(v)
	case string:
		id, _ := strconv.ParseInt(v, 10, 64)
		return id
	case json.Number:
		id, _ := v.Int64()
		return id
	default:
		return 0
	}
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
	case "CreateToken", "CMGetSellingCard", "CMUpsertUserPrintlog", "GetGameSetting", "GetGameEvent", "GetGameCharge", "GetGameFesta", "GetGameKaleidxScope", "GetGameMusicScore", "GetGameNationalData", "GetGameNgMusicId", "GetGameRanking", "GetGameTournamentInfo", "GetGameWeeklyData", "GetGameMapAreaCondition", "GetPlaceCircleData", "GetUserCardPrintError", "CMGetUserCardPrintError", "GetUserCircleData", "GetUserCirclePointRanking", "GetUserFesta", "GetUserFriendCheck", "GetUserScoreRanking", "Ping", "RemoveToken", "UserFriendRegist", "UserLogout", "CMLogin", "CMLogout", "CMUpsertBuyCard", "UpsertClientBookkeeping", "UpsertClientPlayTime", "UpsertClientSetting", "UpsertClientTestmode", "UpsertClientUpload", "UpsertUserChargelog", "UpsertUserPlaceCircleRegist":
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
