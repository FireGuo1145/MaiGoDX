package handler

import (
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
	detail := model.UserDetail{UserID: 11, PartnerID: 1}
	if err := database.DB.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}
	if err := database.DB.Create(&model.UserItem{UserID: 11, ItemKind: 1, ItemID: 99, Stock: 1}).Error; err != nil {
		t.Fatalf("create unrelated item: %v", err)
	}
	request := portalProfileUpdateRequest{
		PartnerID:       8,
		TravelPartners:  []portalTravelPartner{{PartnerID: 3, IntimateLevel: 4, IntimateCountRewarded: 2}},
		FunctionTickets: []portalFunctionTicket{{ItemID: 7, Stock: 5}},
		Regions:         []portalRegion{{RegionID: 2, PlayCount: 9}},
	}
	if err := savePortalProfile(detail, request); err != nil {
		t.Fatalf("save portal profile: %v", err)
	}
	var saved model.UserDetail
	if err := database.DB.Where("user_id = ?", 11).First(&saved).Error; err != nil || saved.PartnerID != 8 {
		t.Fatalf("saved partner=%d err=%v", saved.PartnerID, err)
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
