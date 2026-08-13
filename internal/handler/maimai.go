package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

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

	log.Printf("[MaiGoDX] Handling API call: %s", apiName)

	w.Header().Set("Content-Type", "application/json")

	var responseData interface{}
	switch apiName {
	case "GetUserPreview":
		responseData = model.UserPreviewData{
			UserID:   114514,
			UserName: "杂鱼大哥哥",
			IsLogin:  true,
			LastData: "2026-08-13 12:00:00",
		}
	case "GetUserData":
		responseData = model.UserDetailData{
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
	case "UpsertUserAll":
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
