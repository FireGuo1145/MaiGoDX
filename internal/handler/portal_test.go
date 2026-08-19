package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func TestPortalProfileCollectionsUsePersistedGameData(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.Create(&[]model.UserIntimate{{UserID: 10, PartnerID: 2, IntimateLevel: 7, IntimateCountRewarded: 3}}).Error; err != nil {
		t.Fatalf("create travel partner: %v", err)
	}
	if err := database.DB.Create(&[]model.UserItem{{UserID: 10, ItemKind: 12, ItemID: 9, Stock: 4}, {UserID: 10, ItemKind: 1, ItemID: 99, Stock: 1}}).Error; err != nil {
		t.Fatalf("create items: %v", err)
	}
	if err := database.DB.Create(&model.UserRegion{UserID: 10, RegionID: 5, PlayCount: 12}).Error; err != nil {
		t.Fatalf("create region: %v", err)
	}

	partners := portalTravelPartners(10)
	tickets := portalFunctionTickets(10)
	regions := portalRegions(10)
	if len(partners) != 1 || partners[0].PartnerID != 2 || partners[0].IntimateLevel != 7 {
		t.Fatalf("travel partners=%v", partners)
	}
	if len(tickets) != 1 || tickets[0].ItemID != 9 || tickets[0].Stock != 4 {
		t.Fatalf("function tickets=%v", tickets)
	}
	if len(regions) != 1 || regions[0].RegionID != 5 || regions[0].PlayCount != 12 {
		t.Fatalf("regions=%v", regions)
	}
}

func TestSavePortalProfileReplacesEditableCollections(t *testing.T) {
	setupMaimaiTestDB(t)
	detail := model.UserDetail{UserID: 11, PartnerID: 1, TotalPoint: 99999}
	if err := database.DB.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}
	if err := database.DB.Create(&model.UserItem{UserID: 11, ItemKind: 1, ItemID: 99, Stock: 1}).Error; err != nil {
		t.Fatalf("create unrelated item: %v", err)
	}
	partnerID, iconID, mapStock := 8, 101, 25
	point := int64(12345)
	request := portalProfileUpdateRequest{
		PartnerID:       &partnerID,
		Point:           &point,
		IconID:          &iconID,
		MapStock:        &mapStock,
		TravelPartners:  []portalTravelPartner{{PartnerID: 3, IntimateLevel: 4, IntimateCountRewarded: 2}},
		FunctionTickets: []portalFunctionTicket{{ItemID: 7, Stock: 5}},
		Regions:         []portalRegion{{RegionID: 2, PlayCount: 9}},
	}
	if err := savePortalProfile(detail, request); err != nil {
		t.Fatalf("save portal profile: %v", err)
	}
	var saved model.UserDetail
	if err := database.DB.Where("user_id = ?", 11).First(&saved).Error; err != nil || saved.PartnerID != 8 || saved.Point != 12345 || saved.TotalPoint != 99999 || saved.IconID != 101 || saved.MapStock != 25 {
		t.Fatalf("saved detail=%+v err=%v", saved, err)
	}
	if partners := portalTravelPartners(11); len(partners) != 1 || partners[0].PartnerID != 3 {
		t.Fatalf("saved travel partners=%v", partners)
	}
	if tickets := portalFunctionTickets(11); len(tickets) != 1 || tickets[0].ItemID != 7 {
		t.Fatalf("saved tickets=%v", tickets)
	}
	if regions := portalRegions(11); len(regions) != 1 || regions[0].RegionID != 2 {
		t.Fatalf("saved regions=%v", regions)
	}
	var unrelated model.UserItem
	if err := database.DB.Where("user_id = ? AND item_kind = ?", 11, 1).First(&unrelated).Error; err != nil {
		t.Fatalf("unrelated item was removed: %v", err)
	}
}

func TestSavePortalProfilePreservesOmittedCollections(t *testing.T) {
	setupMaimaiTestDB(t)
	detail := model.UserDetail{UserID: 12, Point: 10, TotalPoint: 500}
	if err := database.DB.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}
	if err := database.DB.Create(&model.UserIntimate{UserID: 12, PartnerID: 4, IntimateLevel: 7}).Error; err != nil {
		t.Fatalf("create travel partner: %v", err)
	}
	if err := database.DB.Create(&model.UserItem{UserID: 12, ItemKind: 12, ItemID: 11001, Stock: 3}).Error; err != nil {
		t.Fatalf("create function ticket: %v", err)
	}
	if err := database.DB.Create(&model.UserRegion{UserID: 12, RegionID: 2, PlayCount: 6}).Error; err != nil {
		t.Fatalf("create region: %v", err)
	}

	point := int64(250)
	if err := savePortalProfile(detail, portalProfileUpdateRequest{Point: &point}); err != nil {
		t.Fatalf("save partial profile: %v", err)
	}
	var saved model.UserDetail
	if err := database.DB.Where("user_id = ?", 12).First(&saved).Error; err != nil || saved.Point != 250 || saved.TotalPoint != 500 {
		t.Fatalf("saved detail=%+v err=%v", saved, err)
	}
	if len(portalTravelPartners(12)) != 1 || len(portalFunctionTickets(12)) != 1 || len(portalRegions(12)) != 1 {
		t.Fatal("partial profile update removed an omitted collection")
	}
}

func TestLegacyMaimileRequestUpdatesPointOnly(t *testing.T) {
	setupMaimaiTestDB(t)
	detail := model.UserDetail{UserID: 13, TotalPoint: 700}
	if err := database.DB.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}
	legacy := int64(88)
	if err := savePortalProfile(detail, portalProfileUpdateRequest{LegacyMaimile: &legacy}); err != nil {
		t.Fatalf("save legacy request: %v", err)
	}
	var saved model.UserDetail
	if err := database.DB.Where("user_id = ?", 13).First(&saved).Error; err != nil || saved.Point != 88 || saved.TotalPoint != 700 {
		t.Fatalf("saved detail=%+v err=%v", saved, err)
	}
}

func TestSavePortalProfileUpdatesCurrentAndTotalMaimileIndependently(t *testing.T) {
	setupMaimaiTestDB(t)
	detail := model.UserDetail{UserID: 15, Point: 10, TotalPoint: 700}
	if err := database.DB.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}

	totalPoint := int64(900)
	if err := savePortalProfile(detail, portalProfileUpdateRequest{TotalPoint: &totalPoint}); err != nil {
		t.Fatalf("save total maimile: %v", err)
	}
	var saved model.UserDetail
	if err := database.DB.Where("user_id = ?", 15).First(&saved).Error; err != nil || saved.Point != 10 || saved.TotalPoint != 900 {
		t.Fatalf("saved detail after total update=%+v err=%v", saved, err)
	}

	point := int64(25)
	if err := savePortalProfile(detail, portalProfileUpdateRequest{Point: &point}); err != nil {
		t.Fatalf("save current maimile: %v", err)
	}
	if err := database.DB.Where("user_id = ?", 15).First(&saved).Error; err != nil || saved.Point != 25 || saved.TotalPoint != 900 {
		t.Fatalf("saved detail after current update=%+v err=%v", saved, err)
	}
}

func TestPortalProfileChangesReachMaimaiUserData(t *testing.T) {
	setupMaimaiTestDB(t)
	detail := model.UserDetail{UserID: 14, TotalPoint: 900}
	if err := database.DB.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}
	point := int64(321)
	partnerID, iconID, plateID := 17, 27, 37
	request := portalProfileUpdateRequest{Point: &point, PartnerID: &partnerID, IconID: &iconID, PlateID: &plateID}
	if err := savePortalProfile(detail, request); err != nil {
		t.Fatalf("save portal profile: %v", err)
	}

	response := httptest.NewRecorder()
	MaimaiHandler(response, httptest.NewRequest(http.MethodPost, "/g/SDEZ/24000/Maimai2Servlet/GetUserDataApi", strings.NewReader(`{"userId":14}`)))
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode GetUserData: %v", err)
	}
	userData, ok := payload["userData"].(map[string]any)
	if !ok {
		t.Fatalf("userData=%T payload=%v", payload["userData"], payload)
	}
	if userData["point"] != float64(321) || userData["totalPoint"] != float64(900) || userData["partnerId"] != float64(17) || userData["iconId"] != float64(27) || userData["plateId"] != float64(37) {
		t.Fatalf("userData=%v", userData)
	}
}

func TestPortalProfileSelectionIsScopedToBoundCard(t *testing.T) {
	setupMaimaiTestDB(t)
	accountID := uint(1)
	cards := []model.UserCard{{UserID: accountID, AccessCode: "11111111111111111111", GameUserID: 701}, {UserID: accountID, AccessCode: "22222222222222222222", GameUserID: 702}}
	if err := database.DB.Create(&cards).Error; err != nil {
		t.Fatalf("create cards: %v", err)
	}
	if err := database.DB.Create(&[]model.UserDetail{{UserID: 701, UserName: "First"}, {UserID: 702, UserName: "Second"}}).Error; err != nil {
		t.Fatalf("create profiles: %v", err)
	}
	selected, detail, err := portalProfileForAccount(accountID, cards[1].ID)
	if err != nil || selected.ID != cards[1].ID || detail.UserName != "Second" {
		t.Fatalf("selected=%+v detail=%+v err=%v", selected, detail, err)
	}
	if _, _, err := portalProfileForAccount(accountID, cards[1].ID+100); err == nil {
		t.Fatal("unowned card selection should fail")
	}
}

func TestPortalProfileDefaultsToCardWithPersistedMaimaiData(t *testing.T) {
	setupMaimaiTestDB(t)
	cards := []model.UserCard{
		{UserID: 1, AccessCode: "11111111111111111111", GameUserID: 700},
		{UserID: 1, AccessCode: "22222222222222222222", GameUserID: 701},
	}
	if err := database.DB.Create(&cards).Error; err != nil {
		t.Fatalf("create cards: %v", err)
	}
	if err := database.DB.Create(&model.UserDetail{UserID: 701, UserName: "Persisted"}).Error; err != nil {
		t.Fatalf("create profile: %v", err)
	}
	selected, detail, err := portalProfileForAccount(1, 0)
	if err != nil || selected.ID != cards[1].ID || detail.UserID != 701 {
		t.Fatalf("selected=%+v detail=%+v err=%v", selected, detail, err)
	}
}

func TestAdjustPortalFunctionTicketCreatesAndIncrements(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := adjustPortalFunctionTicket(12, 4, 5); err != nil {
		t.Fatalf("create ticket: %v", err)
	}
	if err := adjustPortalFunctionTicket(12, 4, 10); err != nil {
		t.Fatalf("increment ticket: %v", err)
	}
	tickets := portalFunctionTickets(12)
	if len(tickets) != 1 || tickets[0].ItemID != 4 || tickets[0].Stock != 15 {
		t.Fatalf("tickets=%v", tickets)
	}
}
