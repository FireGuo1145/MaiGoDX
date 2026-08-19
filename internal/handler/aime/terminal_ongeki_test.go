package aime

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func TestProtectedSDDTRouteDispatchesToOngeki(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.AutoMigrate(&model.TerminalSession{}, &model.OngekiUser{}, &model.OngekiMusicDetail{}, &model.OngekiPlaylog{}, &model.OngekiUserRecord{}); err != nil {
		t.Fatalf("migrate Ongeki session tables: %v", err)
	}
	terminal := model.Terminal{KeychipID: "A0000001234", GameID: "SDDT", IsEnabled: true}
	if err := database.DB.Create(&terminal).Error; err != nil {
		t.Fatalf("create terminal: %v", err)
	}
	session := model.TerminalSession{Token: "ongeki-session", TerminalID: terminal.ID, GameID: "SDDT", ExpiresAt: time.Now().Add(time.Hour)}
	if err := database.DB.Create(&session).Error; err != nil {
		t.Fatalf("create session: %v", err)
	}

	body := `{"userId":88,"upsertUserAll":{"userData":[{"userName":"Protected Ongeki","playerRating":1200}]}}`
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/gs/ongeki-session/SDDT/1.50/MaimaiServlet/UpsertUserAllApi", bytes.NewBufferString(body))
	HandleTerminalMaimai(response, request)

	var profile model.OngekiUser
	if err := database.DB.Where("user_id = ?", 88).First(&profile).Error; err != nil || profile.UserName != "Protected Ongeki" {
		t.Fatalf("profile=%+v err=%v response=%s", profile, err, response.Body.String())
	}
}
