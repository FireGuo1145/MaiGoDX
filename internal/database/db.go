package database

import (
	"log"

	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化 SQLite 数据库及自动迁移
func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("maigodx.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 自动迁移所有模型
	err = DB.AutoMigrate(
		&model.UserDetail{},
		&model.UserOption{},
		&model.UserPlaylog{},
		&model.UserCharacter{},
		&model.UserItem{},
		&model.UserMap{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	log.Println("[MaiGoDX] Database initialized and migrated successfully.")
}
