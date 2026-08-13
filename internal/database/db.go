package database

import (
	"log"
	"os"

	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 根据环境变量 DB_TYPE 和 DB_DSN 初始化数据库连接（默认 SQLite）
func InitDB() {
	dbType := os.Getenv("DB_TYPE")
	dbDSN := os.Getenv("DB_DSN")

	if dbType == "" {
		dbType = "sqlite"
	}

	var err error
	var dialector gorm.Dialector

	switch dbType {
	case "mysql":
		if dbDSN == "" {
			dbDSN = "root:root@tcp(127.0.0.1:3306)/maigodx?charset=utf8mb4&parseTime=True&loc=Local"
		}
		dialector = mysql.Open(dbDSN)
	case "postgres", "postgresql":
		if dbDSN == "" {
			dbDSN = "host=localhost user=postgres password=postgres dbname=maigodx port=5432 sslmode=disable"
		}
		dialector = postgres.Open(dbDSN)
	case "sqlite":
		fallthrough
	default:
		if dbDSN == "" {
			dbDSN = "maigodx.db"
		}
		dialector = sqlite.Open(dbDSN)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database (%s): %v", dbType, err)
	}

	// 自动迁移所有模型（含认证与用户体系）
	err = DB.AutoMigrate(
		&model.UserAccount{},
		&model.UserDetail{},
		&model.UserOption{},
		&model.UserExtend{},
		&model.UserPlaylog{},
		&model.UserCharacter{},
		&model.UserItem{},
		&model.UserMap{},
		&model.UserFavorite{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	log.Printf("[MaiGoDX] Database initialized successfully using driver: %s", dbType)
}
