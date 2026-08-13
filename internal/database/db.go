package database

import (
	"log"
	"os"

	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 根据环境变量 DB_TYPE 和 DB_DSN 初始化数据库连接（默认 SQLite），并自动初始化默认管理员与系统配置
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

	// 自动迁移所有模型（含认证、卡片绑定与系统配置）
	err = DB.AutoMigrate(
		&model.UserAccount{},
		&model.UserCard{},
		&model.SystemConfig{},
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

	// 系统初始化：检查是否存在管理员，若无则自动创建默认管理员
	var adminCount int64
	DB.Model(&model.UserAccount{}).Where("is_admin = ?", true).Count(&adminCount)
	if adminCount == 0 {
		hashedPass, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := model.UserAccount{
			Email:        "admin@maigodx.local",
			PasswordHash: string(hashedPass),
			Username:     "SystemAdmin",
			IsVerified:   true,
			IsAdmin:      true,
		}
		DB.Create(&admin)
		log.Println("[MaiGoDX] Default administrator created: admin@maigodx.local / admin123")
	}

	// 系统初始化：检查默认系统配置
	var configCount int64
	DB.Model(&model.SystemConfig{}).Count(&configCount)
	if configCount == 0 {
		defaultConfigs := []model.SystemConfig{
			{Key: "maintenance_mode", Value: "false", Desc: "服务器全局维护模式开关 (true/false)"},
			{Key: "event_mode", Value: "false", Desc: "限时活动模式开关 (true/false)"},
			{Key: "notice_banner", Value: "欢迎来到 MaiGoDX 街机服务器！", Desc: "全服公告横幅内容"},
		}
		DB.Create(&defaultConfigs)
		log.Println("[MaiGoDX] Default system configurations initialized.")
	}

	log.Printf("[MaiGoDX] Database initialized successfully using driver: %s", dbType)
}
