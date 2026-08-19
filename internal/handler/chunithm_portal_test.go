package handler

import (
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func TestChuniPortalDefaultsToCardWithPersistedProfile(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.AutoMigrate(&model.ChuniUser{}); err != nil {
		t.Fatalf("migrate CHUNITHM profile: %v", err)
	}
	cards := []model.UserCard{
		{UserID: 1, AccessCode: "11111111111111111111"},
		{UserID: 1, AccessCode: "22222222222222222222"},
		{UserID: 1, AccessCode: "33333333333333333333", GameUserID: 3},
	}
	if err := database.DB.Create(&cards).Error; err != nil {
		t.Fatalf("create cards: %v", err)
	}
	if err := database.DB.Create(&model.ChuniUser{UserID: 3, UserName: "CHUNI"}).Error; err != nil {
		t.Fatalf("create CHUNITHM profile: %v", err)
	}

	selected, found, err := chuniPortalCard(1, 0)
	if err != nil || !found || selected.ID != cards[2].ID || selected.GameUserID != 3 {
		t.Fatalf("selected=%+v found=%v err=%v", selected, found, err)
	}
}

func TestChuniPortalHonorsExplicitOwnedCard(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.AutoMigrate(&model.ChuniUser{}); err != nil {
		t.Fatalf("migrate CHUNITHM profile: %v", err)
	}
	card := model.UserCard{UserID: 1, AccessCode: "11111111111111111111"}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}

	selected, found, err := chuniPortalCard(1, card.ID)
	if err != nil || !found || selected.ID != card.ID {
		t.Fatalf("selected=%+v found=%v err=%v", selected, found, err)
	}
	if _, found, err := chuniPortalCard(2, card.ID); err != nil || found {
		t.Fatalf("unowned card found=%v err=%v", found, err)
	}
}
