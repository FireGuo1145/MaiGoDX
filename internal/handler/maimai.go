package handler

import (
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

	log.Printf("[MaiGoDX] Handling API call: %s", apiName)

	w.Header().Set("Content-Type", "application/json")

	bodyBytes, _ := io.ReadAll(r.Body)

	var responseData interface{}
	switch apiName {
	case "GetUserPreview":
		var detail model.UserDetail
		// 尝试从数据库查询，若无则返回默认测试用户
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

	case "UpsertUserAll":
		var req model.UpsertUserAllRequest
		if err := json.Unmarshal(bodyBytes, &req); err == nil && len(req.UpsertUserAll.UserData) > 0 {
			userData := req.UpsertUserAll.UserData[0]
			// 保存或更新玩家数据
			database.DB.Save(&userData)

			// 保存成绩记录
			for _, playlog := range req.UpsertUserAll.UserPlaylogList {
				playlog.UserID = userData.UserID
				database.DB.Create(&playlog)
			}
			log.Printf("[MaiGoDX] UpsertUserAll processed successfully for user: %s", userData.UserName)
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
