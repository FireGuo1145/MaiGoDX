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

const mai2PhotoChunkSize = 10 * 1024

type mai2ImageChunk struct {
	OrderID    int64  `json:"orderId"`
	UserID     int64  `json:"userId"`
	DivNumber  int    `json:"divNumber"`
	DivLength  int    `json:"divLength"`
	DivData    string `json:"divData"`
	PlaceID    int    `json:"placeId"`
	ClientID   string `json:"clientId"`
	UploadDate string `json:"uploadDate"`
	PlaylogID  int64  `json:"playlogId"`
	TrackNo    int    `json:"trackNo"`
	FileName   string `json:"fileName"`
}

// HandleUploadUserPhoto implements AquaDX's nested userPhoto chunk protocol.
// Chunks are assembled in a temporary file and only become a visible score photo
// after the final part arrives, preventing partial images from being published.
func HandleUploadUserPhoto(w http.ResponseWriter, r *http.Request, apiName string) {
	bodyBytes, _ := io.ReadAll(r.Body)
	var envelope struct {
		UserPhoto mai2ImageChunk `json:"userPhoto"`
	}
	if err := json.Unmarshal(bodyBytes, &envelope); err != nil {
		writeGameResponse(w, apiName, 0, nil, "invalid UploadUserPhoto payload")
		return
	}
	chunk := envelope.UserPhoto
	if chunk.UserID == 0 {
		// Accept the historical flat form as a migration compatibility path.
		_ = json.Unmarshal(bodyBytes, &chunk)
	}
	if err := appendMai2ImageChunk(chunk, "plays", fmt.Sprintf("%d-%d", chunk.UserID, chunk.TrackNo)); err != nil {
		writeGameResponse(w, apiName, 0, nil, err.Error())
		return
	}
	writeGameResponse(w, apiName, 1, nil, "")
}

// HandleUploadUserPortrait stores an AquaDX-compatible profile portrait.
func HandleUploadUserPortrait(w http.ResponseWriter, r *http.Request, apiName string) {
	var envelope struct {
		UserPortrait mai2ImageChunk `json:"userPortrait"`
	}
	if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil || envelope.UserPortrait.UserID == 0 {
		writeGameResponse(w, apiName, 0, nil, "invalid UploadUserPortrait payload")
		return
	}
	if err := appendMai2ImageChunk(envelope.UserPortrait, "portraits", fmt.Sprintf("%d-up", envelope.UserPortrait.UserID)); err != nil {
		writeGameResponse(w, apiName, 0, nil, err.Error())
		return
	}
	writeGameResponse(w, apiName, 1, nil, "")
}

// HandleGetUserPortrait returns the stored portrait as the chunked Base64 shape
// that AquaDX provides. A missing portrait is a normal empty success response.
func HandleGetUserPortrait(w http.ResponseWriter, r *http.Request, apiName string, userID int64) {
	filePath := filepath.Join("data", "upload", "mai2", "portraits", fmt.Sprintf("%d-up.jpg", userID))
	contents, err := os.ReadFile(filePath)
	if err != nil || len(contents) == 0 {
		writeGameResponse(w, apiName, 1, map[string]interface{}{"length": 0, "userPortraitList": []interface{}{}}, "")
		return
	}
	count := (len(contents) + mai2PhotoChunkSize - 1) / mai2PhotoChunkSize
	portraitList := make([]map[string]interface{}, 0, count)
	for offset, divNumber := 0, 0; offset < len(contents); offset, divNumber = offset+mai2PhotoChunkSize, divNumber+1 {
		end := offset + mai2PhotoChunkSize
		if end > len(contents) {
			end = len(contents)
		}
		portraitList = append(portraitList, map[string]interface{}{
			"userId": userID, "divData": base64.StdEncoding.EncodeToString(contents[offset:end]),
			"divNumber": divNumber, "divLength": count,
		})
	}
	writeGameResponse(w, apiName, 1, map[string]interface{}{"length": len(portraitList), "userPortraitList": portraitList}, "")
}

func appendMai2ImageChunk(chunk mai2ImageChunk, category, stem string) error {
	if chunk.UserID == 0 || chunk.DivLength < 1 || chunk.DivNumber < 0 || chunk.DivNumber >= chunk.DivLength || chunk.DivData == "" {
		return fmt.Errorf("invalid image chunk metadata")
	}
	decoded, err := base64.StdEncoding.DecodeString(chunk.DivData)
	if err != nil {
		return fmt.Errorf("invalid image chunk data: %w", err)
	}
	tmpDir := filepath.Join("data", "tmp")
	targetDir := filepath.Join("data", "upload", "mai2", category)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}
	tmpPath := filepath.Join(tmpDir, stem+".tmp")
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	if chunk.DivNumber == 0 {
		flags |= os.O_TRUNC
	}
	file, err := os.OpenFile(tmpPath, flags, 0o644)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(decoded)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if chunk.DivNumber+1 < chunk.DivLength {
		log.Printf("[MaiGoDX] %s upload chunk userId=%d part=%d/%d", category, chunk.UserID, chunk.DivNumber+1, chunk.DivLength)
		return nil
	}
	fileName := stem + ".jpg"
	if category == "plays" {
		fileName = fmt.Sprintf("%d-%d.jpg", chunk.UserID, time.Now().UnixMilli())
	}
	if err := os.Rename(tmpPath, filepath.Join(targetDir, fileName)); err != nil {
		return err
	}
	if category == "portraits" {
		metadata := chunk
		metadata.DivData = ""
		encoded, _ := json.Marshal(metadata)
		_ = os.WriteFile(filepath.Join(targetDir, stem+".json"), encoded, 0o644)
	}
	log.Printf("[MaiGoDX] %s upload complete userId=%d parts=%d", category, chunk.UserID, chunk.DivLength)
	return nil
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
