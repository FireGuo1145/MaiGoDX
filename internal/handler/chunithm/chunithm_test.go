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
	if err := db.AutoMigrate(&model.ChuniUser{}, &model.ChuniMusicDetail{}, &model.ChuniPlaylog{}, &model.ChuniUserRecord{}, &model.GameEvent{}, &model.GameCharge{}, &model.SystemConfig{}); err != nil {
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
	setupTestDB(t)
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

func TestGameSettingUsesChuniConfig(t *testing.T) {
	setupTestDB(t)
	configs := []model.SystemConfig{
		{Key: "chuni_maintenance_mode", Value: "true"},
		{Key: "chuni_matching_uri", Value: "https://matching.example/"},
		{Key: "chuni_reflector_uri", Value: "https://reflector.example/"},
		{Key: "chuni_max_count_music", Value: "480"},
	}
	if err := database.DB.Create(&configs).Error; err != nil {
		t.Fatalf("create configs: %v", err)
	}

	response := httptest.NewRecorder()
	Handler(response, httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/GetGameSettingApi", bytes.NewBufferString(`{"version":"2.00"}`)))
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	setting := payload["gameSetting"].(map[string]any)
	if setting["isMaintenance"] != true || setting["matchingUri"] != "https://matching.example/" || setting["matchingUriX"] != "https://matching.example/" || setting["reflectorUri"] != "https://reflector.example/" || setting["maxCountMusic"] != float64(480) {
		t.Fatalf("unexpected game setting: %v", setting)
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

func TestUpsertUserAllRoundTripsGenericCollections(t *testing.T) {
	setupTestDB(t)
	body := []byte(`{
    "userId": 102,
    "upsertUserAll": {
      "userData": [{"userName": "Collections"}],
      "userGameOption": [{"playerLevel": 12, "headphone": 7}],
      "userItemList": [
        {"itemKind": 5, "itemId": 8000, "stock": 3, "isValid": true},
        {"itemKind": 6, "itemId": 10, "stock": 1, "isValid": true}
      ],
      "userCharacterList": [{"characterId": 100, "level": 4}]
    }
  }`)
	Handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/UpsertUserAllApi", bytes.NewReader(body)))

	optionResponse := httptest.NewRecorder()
	Handler(optionResponse, httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/GetUserOptionApi", bytes.NewBufferString(`{"userId":102}`)))
	var option map[string]any
	if err := json.NewDecoder(optionResponse.Body).Decode(&option); err != nil {
		t.Fatalf("decode option: %v", err)
	}
	if option["userGameOption"].(map[string]any)["playerLevel"] != float64(12) {
		t.Fatalf("option=%v", option)
	}

	itemsResponse := httptest.NewRecorder()
	Handler(itemsResponse, httptest.NewRequest(http.MethodPost, "/g/SDHD/2.00/ChuniServlet/GetUserItemApi", bytes.NewBufferString(`{"userId":102,"kind":5}`)))
	var items map[string]any
	if err := json.NewDecoder(itemsResponse.Body).Decode(&items); err != nil {
		t.Fatalf("decode items: %v", err)
	}
	values := items["userItemList"].([]any)
	if len(values) != 1 || values[0].(map[string]any)["itemId"] != float64(8000) {
		t.Fatalf("items=%v", items)
	}
}

func TestMatchingRoomLifecycle(t *testing.T) {
	matching.Lock()
	matching.nextID = 0
	matching.rooms = map[int]*matchingRoom{}
	matching.Unlock()

	first := beginMatching(map[string]any{"matchingMemberInfo": map[string]any{"userId": 1}})
	roomID := first["roomId"].(int)
	beginMatching(map[string]any{"matchingMemberInfo": map[string]any{"userId": 2}})
	state := matchingState(map[string]any{"roomId": roomID})
	if state["matchingWaitState"].(map[string]any)["matchingMemberCount"] != 2 {
		t.Fatalf("state=%v", state)
	}
	result := endMatching(map[string]any{"roomId": roomID})
	if result["matchingResult"] != 1 {
		t.Fatalf("result=%v", result)
	}
}
