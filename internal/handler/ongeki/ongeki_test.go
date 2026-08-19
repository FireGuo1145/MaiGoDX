package ongeki

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/pbkdf2"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	database.DB = db
	if err := db.AutoMigrate(&model.OngekiUser{}, &model.OngekiMusicDetail{}, &model.OngekiPlaylog{}, &model.OngekiUserRecord{}, &model.SystemConfig{}, &model.GameEvent{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
}

func request(t *testing.T, endpoint, body string) map[string]any {
	t.Helper()
	response := performRequest(endpoint, body)
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode %s response %q: %v", endpoint, response.Body.String(), err)
	}
	return payload
}

func performRequest(endpoint, body string) *httptest.ResponseRecorder {
	return performVersionedRequest("1.50", endpoint, body)
}

func performVersionedRequest(version, endpoint, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	Handler(response, httptest.NewRequest(http.MethodPost, "/g/SDDT/"+version+"/"+endpoint, bytes.NewBufferString(body)))
	return response
}

func TestMissingUserPreviewMatchesOngekiShape(t *testing.T) {
	setupTestDB(t)
	payload := request(t, "GetUserPreviewApi", `{"userId":101}`)
	if payload["userId"] != float64(101) || payload["userName"] != "" || payload["isWarningConfirmed"] != true {
		t.Fatalf("preview=%v", payload)
	}
}

func TestUpsertUserAllRoundTripsProfileMusicAndCollections(t *testing.T) {
	setupTestDB(t)
	payload := request(t, "UpsertUserAllApi", `{
      "userId":101,"regionId":5,
      "upsertUserAll":{
        "userData":[{"userName":"ONGEKI Player","level":20,"point":321,"totalPoint":900,"playerRating":1234,"newPlayerRating":1400,"lastPlayDate":"2026-08-19 20:00:00.0"}],
        "userOption":[{"dispPlayerLv":1,"dispRating":1,"headphone":1}],
        "userMusicDetailList":[{"musicId":42,"level":3,"playCount":2,"techScoreMax":1005000,"platinumScoreMax":900}],
        "userPlaylogList":[{"musicId":42,"level":3,"techScore":1005000,"userPlayDate":"2026-08-19 20:00:00.0"}],
        "userItemList":[{"itemKind":5,"itemId":8,"stock":3,"isValid":true}],
        "userCardList":[{"cardId":1001,"level":20}],
        "userRecentRatingList":[{"musicId":42,"difficultId":3,"score":"1000000","platinumScoreMax":900}]
      }
    }`)
	if payload["returnCode"] != float64(1) || payload["apiName"] != "upsertUserAll" {
		t.Fatalf("upsert=%v", payload)
	}

	var profile model.OngekiUser
	if err := database.DB.Where("user_id = ?", 101).First(&profile).Error; err != nil || profile.UserName != "ONGEKI Player" || profile.Point != 321 || profile.NewPlayerRating != 1400 {
		t.Fatalf("profile=%+v err=%v", profile, err)
	}
	var playCount int64
	database.DB.Model(&model.OngekiPlaylog{}).Where("user_id = ?", 101).Count(&playCount)
	if playCount != 1 {
		t.Fatalf("playCount=%d", playCount)
	}

	data := request(t, "GetUserDataApi", `{"userId":101}`)
	userData := data["userData"].(map[string]any)
	if userData["point"] != float64(321) || userData["totalPoint"] != float64(900) {
		t.Fatalf("userData=%v", userData)
	}
	music := request(t, "GetUserMusicApi", `{"userId":101}`)
	if music["length"] != float64(1) || music["nextIndex"] != float64(-1) {
		t.Fatalf("music=%v", music)
	}
	items := request(t, "GetUserItemApi", `{"userId":101,"nextIndex":50000000000}`)
	if items["itemKind"] != float64(5) || items["length"] != float64(1) {
		t.Fatalf("items=%v", items)
	}
	ratings := request(t, "GetUserRecentRatingApi", `{"userId":101}`)
	if ratings["length"] != float64(1) || ratings["nextIndex"] != float64(0) {
		t.Fatalf("ratings=%v", ratings)
	}
	regions := request(t, "GetUserRegionApi", `{"userId":101}`)
	if regions["length"] != float64(1) {
		t.Fatalf("regions=%v", regions)
	}
}

func TestRepeatedUpsertUpdatesMusicAndDoesNotDuplicateCollections(t *testing.T) {
	setupTestDB(t)
	for _, score := range []int{900000, 1007000} {
		request(t, "UpsertUserAllApi", `{"userId":1,"upsertUserAll":{"userData":[{"userName":"P"}],"userMusicDetailList":[{"musicId":7,"level":2,"techScoreMax":`+strconv.Itoa(score)+`}],"userItemList":[{"itemKind":1,"itemId":2,"stock":4}]}}`)
	}
	var music model.OngekiMusicDetail
	if err := database.DB.Where("user_id = ?", 1).First(&music).Error; err != nil || music.TechScoreMax != 1007000 {
		t.Fatalf("music=%+v err=%v", music, err)
	}
	var itemCount int64
	database.DB.Model(&model.OngekiUserRecord{}).Where("user_id = ? AND kind = ?", 1, "userItemList").Count(&itemCount)
	if itemCount != 1 {
		t.Fatalf("itemCount=%d", itemCount)
	}
}

func TestRatingCollectionsKeepUploadOrder(t *testing.T) {
	setupTestDB(t)
	entries := make([]map[string]any, 12)
	for index := range entries {
		entries[index] = map[string]any{"musicId": index + 1, "difficultId": 3, "score": 1_000_000 + index}
	}
	body, err := json.Marshal(map[string]any{"userId": 2, "upsertUserAll": map[string]any{"userData": []any{map[string]any{"userName": "Rating"}}, "userRecentRatingList": entries}})
	if err != nil {
		t.Fatal(err)
	}
	request(t, "UpsertUserAllApi", string(body))
	values := recentRatings(2)
	for index, value := range values {
		if intValue(value["musicId"]) != index+1 {
			t.Fatalf("rating order at %d: %v", index, values)
		}
	}
}

func TestGameSettingUsesRouteVersionAndConfig(t *testing.T) {
	setupTestDB(t)
	if err := database.DB.Create(&[]model.SystemConfig{{Key: "ongeki_maintenance_mode", Value: "true"}, {Key: "ongeki_max_count_music", Value: "88"}}).Error; err != nil {
		t.Fatal(err)
	}
	response := performVersionedRequest("1.52", "GetGameSettingApi", `{}`)
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	setting := payload["gameSetting"].(map[string]any)
	if setting["dataVersion"] != "1.52" || setting["onlineDataVersion"] != "1.52" || setting["isMaintenance"] != true || setting["maxCountMusic"] != float64(88) {
		t.Fatalf("setting=%v", setting)
	}
}

func TestCardMakerUpsertPersistsProfileAndCard(t *testing.T) {
	setupTestDB(t)
	payload := request(t, "CMUpsertUserAllApi", `{"userId":9,"cmUpsertUserAll":{"userData":[{"userName":"CM Player","point":20}],"userCardList":[{"cardId":77,"level":3}]}}`)
	if payload["returnCode"] != float64(1) {
		t.Fatalf("payload=%v", payload)
	}
	card := request(t, "CMGetUserCardApi", `{"userId":9}`)
	if card["length"] != float64(1) {
		t.Fatalf("card=%v", card)
	}
}

func TestCardMakerGachaStateRoundTrips(t *testing.T) {
	setupTestDB(t)
	payload := request(t, "CMUpsertUserGachaApi", `{"userId":9,"gachaId":10,"gachaCnt":5,"selectPoint":20,"cmUpsertUserGacha":{"userData":[{"userName":"Gacha"}]}}`)
	if payload["returnCode"] != float64(1) {
		t.Fatalf("payload=%v", payload)
	}
	state := request(t, "GetUserGachaApi", `{"userId":9}`)
	values := state["userGachaList"].([]any)
	if len(values) != 1 || values[0].(map[string]any)["totalGachaCnt"] != float64(5) || values[0].(map[string]any)["selectPoint"] != float64(20) {
		t.Fatalf("state=%v", state)
	}
	request(t, "CMUpsertUserSelectGachaApi", `{"userId":9,"selectGachaLogList":[{"gachaId":10}],"cmUpsertUserSelectGacha":{"userData":[{"userName":"Gacha"}]}}`)
	state = request(t, "GetUserGachaApi", `{"userId":9}`)
	selected := state["userGachaList"].([]any)[0].(map[string]any)
	if selected["selectPoint"] != float64(0) || selected["useSelectPoint"] != float64(1) {
		t.Fatalf("selected state=%v", state)
	}
}

func TestConfiguredEncryptedEndpointResolves(t *testing.T) {
	setupTestDB(t)
	if err := database.DB.Create(&model.SystemConfig{Key: "ongeki_endpoint_salts", Value: "01020304:10"}).Error; err != nil {
		t.Fatal(err)
	}
	hash := hex.EncodeToString(pbkdf2.Key([]byte("GetGameSettingApi"), []byte{1, 2, 3, 4}, 10, 16, sha1.New))
	if resolved := resolveEndpoint(hash); resolved != "GetGameSettingApi" {
		t.Fatalf("resolved=%q", resolved)
	}
}

func TestAquaDXStaticGameDataDirectory(t *testing.T) {
	setupTestDB(t)
	directory := t.TempDir()
	files := map[string]string{
		"game_event.json":      `[{"id":1234}]`,
		"game_gacha.json":      `[{"gachaId":10,"gachaName":"Test","kind":1},{"gachaId":1112,"gachaName":"Permanent","kind":0}]`,
		"game_gacha_card.json": `[{"gachaId":10,"cardId":20,"rarity":2},{"gachaId":1112,"cardId":30,"rarity":3}]`,
		"game_present.json":    `[{"id":5,"presentName":"Test"}]`,
		"game_encryption.json": `[{"salt":"01020304","iterations":10}]`,
	}
	for name, payload := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(payload), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := database.DB.Create(&model.SystemConfig{Key: "ongeki_game_data_dir", Value: directory}).Error; err != nil {
		t.Fatal(err)
	}
	events := request(t, "GetGameEventApi", `{}`)
	if events["length"] != float64(1) {
		t.Fatalf("events=%v", events)
	}
	gachas := request(t, "GetGameGachaApi", `{}`)
	if gachas["length"] != float64(2) {
		t.Fatalf("gachas=%v", gachas)
	}
	cards := request(t, "GetGameGachaCardByIdApi", `{"gachaId":10}`)
	if cards["length"] != float64(1) {
		t.Fatalf("cards=%v", cards)
	}
	request(t, "UpsertUserAllApi", `{"userId":9,"upsertUserAll":{"userData":[{"userName":"Gacha"}]}}`)
	rolled := request(t, "RollGachaApi", `{"userId":9,"gachaId":10,"times":2}`)
	if rolled["length"] != float64(2) {
		t.Fatalf("rolled=%v", rolled)
	}
	presents := request(t, "GetGamePresentApi", `{}`)
	present := presents["gamePresentList"].([]any)[0].(map[string]any)
	if present["startDate"] != "2000-01-01 05:00:00.0" || present["endDate"] != "2099-01-01 05:00:00.0" {
		t.Fatalf("presents=%v", presents)
	}
	hash := hex.EncodeToString(pbkdf2.Key([]byte("GetGameSettingApi"), []byte{1, 2, 3, 4}, 10, 16, sha1.New))
	if resolved := resolveEndpoint(hash); resolved != "GetGameSettingApi" {
		t.Fatalf("static encryption resolved=%q", resolved)
	}

	request(t, "CMUpsertUserGachaApi", `{"userId":9,"gachaId":10,"gachaCnt":1,"selectPoint":0,"cmUpsertUserGacha":{"userData":[{"userName":"Gacha"}]}}`)
	guaranteed := request(t, "RollGachaApi", `{"userId":9,"gachaId":10,"times":5}`)
	first := guaranteed["gameGachaCardList"].([]any)[0].(map[string]any)
	if first["rarity"] != float64(3) || first["gachaId"] != float64(1112) {
		t.Fatalf("first five pull did not use permanent SR guarantee: %v", guaranteed)
	}
}

func TestRivalDataAndMusicMatchAquaDXShape(t *testing.T) {
	setupTestDB(t)
	request(t, "UpsertUserAllApi", `{"userId":101,"upsertUserAll":{"userData":[{"userName":"Alpha"}],"userMusicDetailList":[{"musicId":7,"level":2,"techScoreMax":1000000},{"musicId":7,"level":3,"techScoreMax":1005000},{"musicId":8,"level":1,"techScoreMax":990000}]}}`)
	request(t, "UpsertUserAllApi", `{"userId":102,"upsertUserAll":{"userData":[{"userName":"Beta"}]}}`)

	response := performRequest("GetUserRivalDataApi", `{"userId":1,"userRivalList":[{"rivalUserId":102},{"rivalUserId":999},{"rivalUserId":101}]}`)
	var rivals []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&rivals); err != nil {
		t.Fatal(err)
	}
	if len(rivals) != 2 || rivals[0]["rivalUserName"] != "Beta" || rivals[1]["rivalUserName"] != "Alpha" {
		t.Fatalf("rivals=%v", rivals)
	}

	music := request(t, "GetUserRivalMusicApi", `{"userId":1,"rivalUserId":101}`)
	groups := music["userRivalMusicList"].([]any)
	if music["rivalUserId"] != float64(101) || len(groups) != 2 || groups[0].(map[string]any)["length"] != float64(2) {
		t.Fatalf("rival music=%v", music)
	}
}

func TestEventRankingCountsUsersWithHigherPoints(t *testing.T) {
	setupTestDB(t)
	for _, entry := range []struct {
		userID int
		point  int
	}{{1, 300}, {2, 900}, {3, 500}, {4, 300}} {
		request(t, "UpsertUserAllApi", `{"userId":`+strconv.Itoa(entry.userID)+`,"upsertUserAll":{"userData":[{"userName":"Rank"}],"userEventPointList":[{"eventId":77,"point":`+strconv.Itoa(entry.point)+`}]}}`)
	}
	ranking := request(t, "GetUserEventRankingApi", `{"userId":3}`)
	value := ranking["userEventRankingList"].([]any)[0].(map[string]any)
	if value["rank"] != float64(2) || value["point"] != float64(500) {
		t.Fatalf("ranking=%v", ranking)
	}
}

func TestDailyGachaCountResetsAcrossCalendarDays(t *testing.T) {
	setupTestDB(t)
	if err := saveRecord(database.DB, 9, "userGachaList", "10", map[string]any{
		"gachaId": 10, "totalGachaCnt": 10, "dailyGachaCnt": 10, "dailyGachaDate": "2000-01-01 00:00:00",
	}); err != nil {
		t.Fatal(err)
	}
	request(t, "CMUpsertUserGachaApi", `{"userId":9,"gachaId":10,"gachaCnt":2,"selectPoint":0,"cmUpsertUserGacha":{"userData":[{"userName":"Gacha"}]}}`)
	state := request(t, "GetUserGachaApi", `{"userId":9}`)["userGachaList"].([]any)[0].(map[string]any)
	if state["dailyGachaCnt"] != float64(2) || state["totalGachaCnt"] != float64(12) {
		t.Fatalf("state=%v", state)
	}
}

func TestCardMakerMissingUserReturnsBadRequest(t *testing.T) {
	setupTestDB(t)
	response := performRequest("CMGetUserDataApi", `{"userId":404}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestRollGachaRejectsMissingUserAndUnknownGacha(t *testing.T) {
	setupTestDB(t)
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "game_gacha.json"), []byte(`[{"gachaId":10,"kind":0}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&model.SystemConfig{Key: "ongeki_game_data_dir", Value: directory}).Error; err != nil {
		t.Fatal(err)
	}
	if response := performRequest("RollGachaApi", `{"userId":404,"gachaId":10,"times":1}`); response.Code != http.StatusBadRequest {
		t.Fatalf("missing user status=%d body=%s", response.Code, response.Body.String())
	}
	request(t, "UpsertUserAllApi", `{"userId":9,"upsertUserAll":{"userData":[{"userName":"Gacha"}]}}`)
	if response := performRequest("RollGachaApi", `{"userId":9,"gachaId":999,"times":1}`); response.Code != http.StatusNotFound {
		t.Fatalf("unknown gacha status=%d body=%s", response.Code, response.Body.String())
	}
}
