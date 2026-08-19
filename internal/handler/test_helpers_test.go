package handler

import (
	"fmt"
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
		&model.Terminal{},
		&model.UserCharacter{}, &model.UserItem{}, &model.UserMap{}, &model.UserFavorite{},
		&model.UserMusicDetail{}, &model.UserCharge{}, &model.UserFriendSeasonRanking{}, &model.UserCourse{}, &model.UserLoginBonus{}, &model.UserGeneralData{},
		&model.UserUdemae{}, &model.UserKaleidx{}, &model.UserIntimate{}, &model.UserActivity{}, &model.UserRegion{},
		&model.UserGameCard{}, &model.UserPrintDetail{}, &model.GameSellingCard{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
}
