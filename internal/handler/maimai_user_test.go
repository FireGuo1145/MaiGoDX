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
		&model.SystemConfig{}, &model.UserCard{}, &model.UserDetail{}, &model.UserOption{}, &model.UserExtend{}, &model.UserPlaylog{},
		&model.UserCharacter{}, &model.UserItem{}, &model.UserMap{}, &model.UserFavorite{},
		&model.UserMusicDetail{}, &model.UserCourse{}, &model.UserLoginBonus{}, &model.UserGeneralData{},
		&model.UserUdemae{}, &model.UserKaleidx{}, &model.UserIntimate{}, &model.UserActivity{}, &model.UserRegion{},
		&model.UserGameCard{}, &model.UserPrintDetail{}, &model.GameSellingCard{},
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

func TestAquaDXParityEndpointsUseExpectedResponseShapes(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.Create(&model.GameSellingCard{CardID: 7, StartDate: "2026-01-01 00:00:00.0", EndDate: "2027-01-01 00:00:00.0"}).Error; err != nil {
		t.Fatalf("create selling card: %v", err)
	}
	if err := database.DB.Create(&model.UserDetail{UserID: 100, UserName: "Owner"}).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := database.DB.Create(&model.UserDetail{UserID: 200, UserName: "Rival"}).Error; err != nil {
		t.Fatalf("create rival: %v", err)
	}
	var rival model.UserDetail
	if err := database.DB.Where("user_id = ?", 200).First(&rival).Error; err != nil {
		t.Fatalf("load rival: %v", err)
	}
	if err := database.DB.Create(&model.UserMusicDetail{UserID: int64(rival.ID), MusicID: 42, Level: 3, Achievement: 999999, DeluxScoreMax: 1234}).Error; err != nil {
		t.Fatalf("create rival music: %v", err)
	}

	tests := []struct {
		name  string
		path  string
		body  string
		check func(t *testing.T, payload map[string]interface{})
	}{
		{
			name: "selling cards", path: "/g/SDEZ/24000/Maimai2Servlet/CMGetSellingCardApi", body: `{}`,
			check: func(t *testing.T, payload map[string]interface{}) {
				if payload["length"] != float64(1) {
					t.Fatalf("selling-card length=%v", payload["length"])
				}
			},
		},
		{
			name: "rival music", path: "/g/SDEZ/24000/Maimai2Servlet/GetUserRivalMusicApi", body: fmt.Sprintf(`{"userId":100,"rivalId":%d}`, rival.ID),
			check: func(t *testing.T, payload map[string]interface{}) {
				if payload["rivalId"] != float64(rival.ID) {
					t.Fatalf("rivalId=%v", payload["rivalId"])
				}
				list, ok := payload["userRivalMusicList"].([]interface{})
				if !ok || len(list) != 1 {
					t.Fatalf("rival music list=%v", payload["userRivalMusicList"])
				}
			},
		},
		{
			name: "user login region", path: "/g/SDEZ/24000/Maimai2Servlet/UserLoginApi", body: `{"userId":100,"regionId":12}`,
			check: func(t *testing.T, payload map[string]interface{}) {
				if payload["returnCode"] != float64(1) {
					t.Fatalf("login response=%v", payload)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			res := httptest.NewRecorder()
			MaimaiHandler(res, req)
			var payload map[string]interface{}
			if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			test.check(t, payload)
		})
	}

	var region model.UserRegion
	if err := database.DB.Where("user_id = ? AND region_id = ?", 100, 12).First(&region).Error; err != nil {
		t.Fatalf("persist user region: %v", err)
	}
	if region.PlayCount != 0 {
		t.Fatalf("new region playCount=%d, want 0 to match AquaDX", region.PlayCount)
	}
}
