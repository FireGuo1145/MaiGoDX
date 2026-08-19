package handler

import (
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func TestOngekiPortalDefaultsToCardWithPersistedProfile(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.AutoMigrate(&model.OngekiUser{}); err != nil {
		t.Fatalf("migrate Ongeki profile: %v", err)
	}
	cards := []model.UserCard{
		{UserID: 1, AccessCode: "11111111111111111111"},
		{UserID: 1, AccessCode: "22222222222222222222", GameUserID: 12},
		{UserID: 1, AccessCode: "33333333333333333333", GameUserID: 13},
	}
	if err := database.DB.Create(&cards).Error; err != nil {
		t.Fatalf("create cards: %v", err)
	}
	if err := database.DB.Create(&model.OngekiUser{UserID: 13, UserName: "ONGEKI"}).Error; err != nil {
		t.Fatalf("create Ongeki profile: %v", err)
	}

	selected, found, err := ongekiPortalCard(1, 0)
	if err != nil || !found || selected.ID != cards[2].ID || selected.GameUserID != 13 {
		t.Fatalf("selected=%+v found=%v err=%v", selected, found, err)
	}
}

func TestOngekiPortalHonorsExplicitOwnedCard(t *testing.T) {
	setupMaimaiTestDB(t)
	card := model.UserCard{UserID: 1, AccessCode: "11111111111111111111", GameUserID: 20}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	selected, found, err := ongekiPortalCard(1, card.ID)
	if err != nil || !found || selected.ID != card.ID {
		t.Fatalf("selected=%+v found=%v err=%v", selected, found, err)
	}
	if _, found, err := ongekiPortalCard(2, card.ID); err != nil || found {
		t.Fatalf("unowned card found=%v err=%v", found, err)
	}
}
