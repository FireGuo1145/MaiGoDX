package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

// MaimaiHandler 处理所有 /Maimai2Servlet/ 下的请求
func MaimaiHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	apiName := "Unknown"
	if len(parts) > 0 {
		apiName = parts[len(parts)-1]
	}

	if len(apiName) == 32 {
		log.Printf("[MaiGoDX] Encrypted endpoint hash detected: %s", apiName)
	}

	log.Printf("[MaiGoDX] Handling API call: %s", apiName)

	w.Header().Set("Content-Type", "application/json")

	bodyBytes, _ := io.ReadAll(r.Body)

	var responseData interface{} = []interface{}{}

	switch apiName {
	case "GetUserPreview":
		var detail model.UserDetail
		if err := database.DB.First(&detail).Error; err != nil {
			detail = model.UserDetail{
				UserID:   114514,
				UserName: "杂鱼大哥哥",
				Rating:   15000,
			}
		}
		responseData = detail

	case "GetUserData":
		var detail model.UserDetail
		if err := database.DB.First(&detail).Error; err != nil {
			detail = model.UserDetail{
				UserID:            114514,
				UserName:          "杂鱼大哥哥",
				EquipGlassesID:    1,
				EquipBackGroundID: 2,
				EquipNamePlateID:  3,
				EquipFrameID:      4,
				EquipIconID:       5,
				Rating:            15000,
				MaxRating:         16000,
				TotalPoint:        999999,
			}
		}
		responseData = detail

	case "GetUserCharacter":
		var chars []model.UserCharacter
		database.DB.Find(&chars)
		responseData = map[string]interface{}{
			"userId":            114514,
			"userCharacterList": chars,
		}

	case "GetUserItem":
		var items []model.UserItem
		database.DB.Find(&items)
		responseData = map[string]interface{}{
			"userId":       114514,
			"userItemList": items,
		}

	case "GetUserRating":
		responseData = map[string]interface{}{
			"userId":    114514,
			"rating":    15000,
			"maxRating": 16000,
		}

	case "GetGameRanking":
		responseData = []interface{}{}

	case "UploadUserPhoto":
		HandleUploadUserPhoto(w, r, apiName)
		return

	case "UpsertUserPrint":
		HandleUpsertUserPrint(w, r, apiName)
		return

	case "UpsertUserAll":
		var req model.UpsertUserAllRequest
		if err := json.Unmarshal(bodyBytes, &req); err == nil && len(req.UpsertUserAll.UserData) > 0 {
			userData := req.UpsertUserAll.UserData[0]
			database.DB.Save(&userData)

			for _, playlog := range req.UpsertUserAll.UserPlaylogList {
				playlog.UserID = userData.UserID
				database.DB.Create(&playlog)
			}
			for _, char := range req.UpsertUserAll.UserCharacterList {
				char.UserID = userData.UserID
				database.DB.Save(&char)
			}
			for _, item := range req.UpsertUserAll.UserItemList {
				item.UserID = userData.UserID
				database.DB.Save(&item)
			}
			for _, m := range req.UpsertUserAll.UserMapList {
				m.UserID = userData.UserID
				database.DB.Save(&m)
			}
			log.Printf("[MaiGoDX] UpsertUserAll successfully processed for: %s", userData.UserName)
		}

		resp := model.Response{
			ReturnCode: 1,
			ApiName:    "com.sega.maimai2servlet.api." + apiName,
		}
		_ = json.NewEncoder(w).Encode(resp)
		return

	default:
		resp := model.Response{
			ReturnCode: 1,
			ApiName:    "com.sega.maimai2servlet.api." + apiName,
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	resp := model.Response{
		ReturnCode: 1,
		ApiName:    "com.sega.maimai2servlet.api." + apiName,
		Data:       responseData,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// HandleGetStats 提供给前端展示的统计数据
func HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var totalUsers int64
	database.DB.Model(&model.UserAccount{}).Count(&totalUsers)

	var totalPlays int64
	database.DB.Model(&model.UserPlaylog{}).Count(&totalPlays)

	var recentPlays []model.UserPlaylog
	database.DB.Order("id desc").Limit(10).Find(&recentPlays)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"totalUsers":  totalUsers,
		"totalPlays":  totalPlays,
		"recentPlays": recentPlays,
	})
}

// ComputeEndpointHash 模拟 MD5 哈希
func ComputeEndpointHash(endpoint string, salt string) string {
	hasher := md5.New()
	hasher.Write([]byte(endpoint + salt))
	return hex.EncodeToString(hasher.Sum(nil))
}
