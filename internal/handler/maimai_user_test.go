package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupMaimaiTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	database.DB = db
	if err := db.AutoMigrate(
		&model.SystemConfig{}, &model.UserDetail{}, &model.UserOption{}, &model.UserExtend{}, &model.UserPlaylog{},
		&model.UserCharacter{}, &model.UserItem{}, &model.UserMap{}, &model.UserFavorite{},
		&model.UserMusicDetail{}, &model.UserCourse{}, &model.UserLoginBonus{}, &model.UserGeneralData{},
		&model.UserUdemae{}, &model.UserKaleidx{}, &model.UserIntimate{}, &model.UserActivity{},
		&model.UserGameCard{}, &model.UserPrintDetail{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
}

func TestUpsertUserAllPersistsAndDeduplicates(t *testing.T) {
	setupMaimaiTestDB(t)
	req := model.UpsertUserAllRequest{UserID: 42}
	req.UpsertUserAll.UserData = []model.UserDetail{{UserName: "Player", Rating: 12345, MaxRating: 12400}}
	req.UpsertUserAll.UserOption = []model.UserOption{{NoteSpeed: 6}}
	req.UpsertUserAll.UserCharacterList = []model.UserCharacter{{CharacterID: 1, Level: 10}, {CharacterID: 1, Level: 11}}
	req.UpsertUserAll.UserMusicDetailList = []model.UserMusicDetail{{MusicID: 100, Level: 3, Achievement: 1000000}}
	req.UpsertUserAll.UserCourseList = []model.UserCourse{{CourseID: 7, PlayCount: 1}}
	req.UpsertUserAll.UserLoginBonusList = []model.UserLoginBonus{{BonusID: 5, IsCurrent: true}}
	req.UpsertUserAll.UserRatingList = []model.UserRatingPayload{{RatingList: []model.UserRate{{MusicID: 100, Level: 3, RomVersion: 24000, Achievement: 1000000}}}}
	req.UpsertUserAll.UserPlaylogList = []model.UserPlaylog{{MusicID: 100, Level: 3, Achievement: 1000000, UserPlayDate: "2026-08-13 12:00:00.0", AfterRating: 12345}}

	if err := UpsertUserAll(req); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := UpsertUserAll(req); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	assertRows(t, &model.UserDetail{}, 1)
	assertRows(t, &model.UserOption{}, 1)
	assertRows(t, &model.UserCharacter{}, 1)
	assertRows(t, &model.UserMusicDetail{}, 1)
	assertRows(t, &model.UserCourse{}, 1)
	assertRows(t, &model.UserLoginBonus{}, 1)
	assertRows(t, &model.UserPlaylog{}, 1)
	assertRows(t, &model.UserGeneralData{}, 4)

	var character model.UserCharacter
	if err := database.DB.Where("user_id = ? AND character_id = ?", 42, 1).First(&character).Error; err != nil {
		t.Fatalf("load character: %v", err)
	}
	if character.Level != 11 {
		t.Fatalf("latest duplicate character was not retained: got %d", character.Level)
	}
	var rating model.UserGeneralData
	if err := database.DB.Where("user_id = ? AND property_key = ?", 42, ratingKeyCurrent).First(&rating).Error; err != nil {
		t.Fatalf("load rating: %v", err)
	}
	if rating.PropertyValue != "100:3:24000:1000000" {
		t.Fatalf("unexpected rating format: %q", rating.PropertyValue)
	}
}

func TestBackloggedPlaylogIsPersistedWithFirstUpsert(t *testing.T) {
	setupMaimaiTestDB(t)
	queuePlaylog(88, model.UserPlaylog{MusicID: 9, Level: 2, UserPlayDate: "2026-08-13 13:00:00.0"})
	req := model.UpsertUserAllRequest{UserID: 88}
	req.UpsertUserAll.UserData = []model.UserDetail{{UserName: "FirstPlay"}}
	if err := UpsertUserAll(req); err != nil {
		t.Fatalf("upsert with backlog: %v", err)
	}
	assertRows(t, &model.UserPlaylog{}, 1)
}

func assertRows(t *testing.T, value interface{}, want int64) {
	t.Helper()
	var got int64
	if err := database.DB.Model(value).Count(&got).Error; err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if got != want {
		t.Fatalf("row count mismatch for %T: got %d want %d", value, got, want)
	}
}

func TestGetGameSettingDoesNotRequireUserID(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.Create(&model.SystemConfig{Key: "maintenance_mode", Value: "true"}).Error; err != nil {
		t.Fatalf("create test config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/g/SDEZ/24000/Maimai2Servlet/GetGameSettingApi", strings.NewReader("{}"))
	res := httptest.NewRecorder()
	MaimaiHandler(res, req)

	var payload map[string]interface{}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	setting, ok := payload["gameSetting"].(map[string]interface{})
	if !ok {
		t.Fatalf("GetGameSetting payload missing gameSetting: %v", payload)
	}
	if setting["isMaintenance"] != true {
		t.Fatalf("maintenance config not applied: %v", setting["isMaintenance"])
	}
}

func TestUpsertUserPrintPersistsCardAndReceipt(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.Create(&model.UserDetail{UserID: 77, UserName: "Printer"}).Error; err != nil {
		t.Fatalf("create player: %v", err)
	}
	body := `{"userId":77,"userPrintDetail":{"orderId":1,"printNumber":2,"userCard":{"cardId":4,"cardTypeId":5,"charaId":6,"mapId":7}}}`
	req := httptest.NewRequest(http.MethodPost, "/g/SDEZ/24000/Maimai2Servlet/UpsertUserPrintApi", strings.NewReader(body))
	res := httptest.NewRecorder()
	MaimaiHandler(res, req)

	var payload map[string]interface{}
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["returnCode"] != float64(1) {
		t.Fatalf("UpsertUserPrint returnCode = %v", payload["returnCode"])
	}
	if serial, _ := payload["serialId"].(string); len(serial) != 20 {
		t.Fatalf("unexpected print serial: %q", serial)
	}
	assertRows(t, &model.UserGameCard{}, 1)
	assertRows(t, &model.UserPrintDetail{}, 1)
}
