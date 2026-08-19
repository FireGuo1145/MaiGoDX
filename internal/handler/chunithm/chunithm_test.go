package chunithm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	database.DB = db
	if err := db.AutoMigrate(&model.ChuniUser{}, &model.ChuniMusicDetail{}, &model.ChuniPlaylog{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
}

func TestUpsertUserAllPersistsProfileAndMusic(t *testing.T) {
	setupTestDB(t)
	body := []byte(`{
    "userId": 101,
    "upsertUserAll": {
      "userData": [{"userName": "CHUNI Player", "playerRating": 1234}],
      "userMusicDetailList": [{"musicId": 42, "level": 3, "score": 1009000}],
      "userPlaylogList": [{"musicId": 42, "level": 3, "score": 1009000}]
    }
  }`)
	response := httptest.NewRecorder()
	Handler(response, httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/UpsertUserAllApi", bytes.NewReader(body)))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}

	var profile model.ChuniUser
	if err := database.DB.Where("user_id = ?", 101).First(&profile).Error; err != nil || profile.UserName != "CHUNI Player" || profile.PlayerRating != 1234 {
		t.Fatalf("profile=%+v err=%v", profile, err)
	}
	var music model.ChuniMusicDetail
	if err := database.DB.Where("user_id = ? AND music_id = ?", 101, 42).First(&music).Error; err != nil {
		t.Fatalf("music was not saved: %v", err)
	}
}

func TestGetUserDataAndMusicReturnPersistedPayloads(t *testing.T) {
	setupTestDB(t)
	if err := database.DB.Create(&model.ChuniUser{UserID: 11, UserName: "Player", PlayerRating: 500, ProfileJSON: `{"userName":"Player","playerRating":500}`}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := database.DB.Create(&model.ChuniMusicDetail{UserID: 11, MusicID: 5, Level: 2, DetailJSON: `{"musicId":5,"level":2,"score":1000000}`}).Error; err != nil {
		t.Fatalf("create music: %v", err)
	}

	for _, endpoint := range []string{"GetUserDataApi", "GetUserMusicApi"} {
		response := httptest.NewRecorder()
		Handler(response, httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/"+endpoint, bytes.NewBufferString(`{"userId":11}`)))
		var payload map[string]any
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			t.Fatalf("decode %s: %v", endpoint, err)
		}
		if payload["userId"] != float64(11) {
			t.Fatalf("%s userId=%v", endpoint, payload["userId"])
		}
	}
}

func TestGameSettingReturnsChuniServletAddress(t *testing.T) {
	response := httptest.NewRecorder()
	Handler(response, httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/GetGameSettingApi", bytes.NewBufferString(`{"version":"2.00"}`)))
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	setting := payload["gameSetting"].(map[string]any)
	if setting["matchingUri"] == "" {
		t.Fatal("missing matchingUri")
	}
}

func TestKnownGameListAPIHasAnEmptyListPayload(t *testing.T) {
	response := httptest.NewRecorder()
	Handler(response, httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/GetGameEventApi", bytes.NewBufferString(`{}`)))
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["length"] != float64(0) {
		t.Fatalf("length=%v", payload["length"])
	}
	if _, ok := payload["gameEventList"].([]any); !ok {
		t.Fatalf("gameEventList=%T, want []any", payload["gameEventList"])
	}
}
