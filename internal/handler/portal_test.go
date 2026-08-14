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
