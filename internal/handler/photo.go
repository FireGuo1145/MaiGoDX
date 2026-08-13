package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/gorm"
)

// HandleUploadUserPhoto stores multipart score-image chunks under the owning game user ID.
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
		_ = os.MkdirAll(uploadDir, 0o755)
		if decoded, err := base64.StdEncoding.DecodeString(req.DivData); err == nil {
			filePath := filepath.Join(uploadDir, strconv.FormatInt(req.UserID, 10)+".jpg")
			if file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
				_, _ = file.Write(decoded)
				_ = file.Close()
			}
		}
		log.Printf("[MaiGoDX] user photo uploaded for userId=%d part=%d/%d", req.UserID, req.DivNumber+1, req.DivLength)
	}
	writeGameResponse(w, apiName, 1, nil, "")
}

// HandleUpsertUserPrint persists the card-maker card and its print receipt.
func HandleUpsertUserPrint(w http.ResponseWriter, r *http.Request, apiName string) {
	var req struct {
		UserID          int64 `json:"userId"`
		UserPrintDetail struct {
			model.UserPrintDetail
			UserCard *model.UserGameCard `json:"userCard"`
		} `json:"userPrintDetail"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == 0 || req.UserPrintDetail.UserCard == nil {
		writeGameResponse(w, apiName, 0, nil, "invalid UpsertUserPrint payload")
		return
	}

	var profile model.UserDetail
	if err := database.DB.Where("user_id = ?", req.UserID).First(&profile).Error; err != nil {
		writeGameResponse(w, apiName, 0, nil, "player profile not found")
		return
	}

	location, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(location)
	expirationDays := cardExpirationDays()
	card := *req.UserPrintDetail.UserCard
	card.ID, card.UserID = 0, req.UserID
	card.StartDate = now.Format("2006-01-02 15:04:05.000000")
	card.EndDate = now.AddDate(0, 0, expirationDays).Format("2006-01-02 15:04:05.000000")

	serialID := ""
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var existing model.UserGameCard
		if err := tx.Where("user_id = ? AND card_id = ?", card.UserID, card.CardID).First(&existing).Error; err == nil {
			card.ID = existing.ID
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err := tx.Save(&card).Error; err != nil {
			return err
		}

		printDetail := req.UserPrintDetail.UserPrintDetail
		printDetail.ID, printDetail.UserID, printDetail.UserGameCardID = 0, req.UserID, card.ID
		serial, err := nextPrintSerial()
		if err != nil {
			return err
		}
		printDetail.SerialID = serial
		serialID = serial
		return tx.Create(&printDetail).Error
	}); err != nil {
		writeGameResponse(w, apiName, 0, nil, "failed to persist print detail")
		return
	}

	writeGameResponse(w, apiName, 1, map[string]interface{}{
		"returnCode": 1, "orderId": 0, "serialId": serialID,
		"startDate": card.StartDate, "endDate": card.EndDate,
	}, "")
}

func cardExpirationDays() int {
	var config model.SystemConfig
	if err := database.DB.Where(&model.SystemConfig{Key: "card_print_expiration_days"}).First(&config).Error; err != nil {
		return 15
	}
	value, err := strconv.Atoi(config.Value)
	if err != nil || value < 1 {
		return 15
	}
	return value
}

func nextPrintSerial() (string, error) {
	const upper = int64(10_000_000_000)
	left, err := rand.Int(rand.Reader, big.NewInt(upper))
	if err != nil {
		return "", err
	}
	right, err := rand.Int(rand.Reader, big.NewInt(upper))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%010d%010d", left.Int64(), right.Int64()), nil
}
