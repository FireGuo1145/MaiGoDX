package handler

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

// HandleUploadUserPhoto 处理玩家成绩结算截图上传
func HandleUploadUserPhoto(w http.ResponseWriter, r *http.Request, apiName string) {
	bodyBytes, _ := io.ReadAll(r.Body)
	var req struct {
		UserID    int64  `json:"userId"`
		TrackNo   int    `json:"trackNo"`
		DivNumber int    `json:"divNumber"`
		DivLength int    `json:"divLength"`
		DivData   string `json:"divData"`
	}

	if err := json.Unmarshal(bodyBytes, &req); err == nil && req.DivData != "" {
		uploadDir := filepath.Join("data", "upload", "mai2", "plays")
		_ = os.MkdirAll(uploadDir, 0755)

		decoded, err := base64.StdEncoding.DecodeString(req.DivData)
		if err == nil {
			filePath := filepath.Join(uploadDir, string(req.UserID)+".jpg")
			f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				_, _ = f.Write(decoded)
				_ = f.Close()
			}
		}
		log.Printf("[MaiGoDX] User photo uploaded for userId: %d, part %d/%d", req.UserID, req.DivNumber+1, req.DivLength)
	}

	resp := model.Response{
		ReturnCode: 1,
		ApiName:    "com.sega.maimai2servlet.api." + apiName,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// HandleUpsertUserPrint 处理玩家打印机/实体卡凭证上传
func HandleUpsertUserPrint(w http.ResponseWriter, r *http.Request, apiName string) {
	log.Printf("[MaiGoDX] UpsertUserPrint processed.")
	resp := model.Response{
		ReturnCode: 1,
		ApiName:    "com.sega.maimai2servlet.api." + apiName,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
