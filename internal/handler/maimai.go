package handler

import (
	"encoding/json"
	"io"
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

	// 读取请求体（支持解析复杂的 JSON 请求，如 UpsertUserAll）
	var reqBodyMap map[string]interface{}
	if r.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil && len(bodyBytes) > 0 {
			_ = json.Unmarshal(bodyBytes, &reqBodyMap)
		}
	}

	var responseData interface{}
	switch apiName {
	case "GetUserPreview":
		responseData = model.UserDetail{
			UserID:   114514,
			UserName: "杂鱼大哥哥",
			Rating:   15000,
		}
	case "GetUserData":
		responseData = model.UserDetail{
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
		// 解析全量上传数据结构
		log.Printf("[MaiGoDX] Processing UpsertUserAll for user...")
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
