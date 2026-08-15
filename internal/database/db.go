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
		&model.SiteMetadata{},
		&model.GameEvent{},
		&model.GameCharge{},
		&model.GameSellingCard{},
		&model.Terminal{},
		&model.TerminalSession{},
		&model.UserDetail{},
		&model.UserOption{},
		&model.UserExtend{},
		&model.UserPlaylog{},
		&model.UserCharacter{},
		&model.UserItem{}, &model.UserMap{}, &model.UserFavorite{}, &model.UserMusicDetail{},
		&model.UserCharge{}, &model.UserFriendSeasonRanking{}, &model.UserCourse{}, &model.UserLoginBonus{}, &model.UserGeneralData{}, &model.UserUdemae{},

		&model.UserKaleidx{},
		&model.UserIntimate{},
		&model.UserActivity{},
		&model.UserRegion{},
		&model.UserGameCard{},
		&model.UserPrintDetail{},
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

	// 系统初始化：以键为粒度补齐默认配置，保留管理员已经修改过的值。
	defaultConfigs := []model.SystemConfig{
		{Key: "site_name", Value: "MaiGoDX", Desc: "管理门户显示的站点名称与浏览器页面标题后缀"},
		{Key: "require_email_verification", Value: "true", Desc: "注册后是否必须完成邮箱验证才允许登录 (true/false)"},
		{Key: "email_verification_delivery", Value: "development", Desc: "邮箱验证令牌发送方式（development 或 smtp）"},
		{Key: "email_smtp_host", Value: "", Desc: "SMTP 服务器地址"},
		{Key: "email_smtp_port", Value: "587", Desc: "SMTP 服务器端口（STARTTLS 通常使用 587）"},
		{Key: "email_smtp_username", Value: "", Desc: "SMTP 登录用户名"},
		{Key: "email_smtp_password", Value: "", Desc: "SMTP 登录密码或应用专用密码"},
		{Key: "email_smtp_from", Value: "", Desc: "验证邮件发件人地址"},
		{Key: "maintenance_mode", Value: "false", Desc: "服务器全局维护模式开关 (true/false)"},
		{Key: "event_mode", Value: "false", Desc: "限时活动模式开关 (true/false)"},
		{Key: "notice_banner", Value: "欢迎来到 MaiGoDX 街机服务器！", Desc: "全服公告横幅内容"},
		{Key: "maimai_endpoint_salts", Value: "", Desc: "maimai 加密端点盐值；多个十六进制盐值以逗号分隔"},
		{Key: "maimai_bearer_token", Value: "", Desc: "maimai CreateToken/UserLogin 下发的 Bearer 令牌；留空表示不启用额外令牌值"},
		{Key: "allnet_check_keychip", Value: "false", Desc: "ALL.Net Keychip 会话保护开关（true 时 PowerOn 下发 /gs/{token}/，false 时按 AquaDX 默认下发 /g/）"},
		{Key: "allnet_keychip_permissive", Value: "false", Desc: "ALL.Net Keychip 宽松测试模式；仅在保护开启时允许未登记 Keychip 继续获取会话"},
		{Key: "allnet_public_host", Value: "", Desc: "ALL.Net PowerOn 下发的公开主机（留空时使用请求 Host）"},
		{Key: "allnet_public_scheme", Value: "http", Desc: "ALL.Net PowerOn 下发地址协议（http 或 https）"},
		{Key: "allnet_name", Value: "", Desc: "ALL.Net 场所名称字段"},
		{Key: "allnet_place_id", Value: "123", Desc: "ALL.Net 场所 ID"},
		{Key: "allnet_region0", Value: "1", Desc: "ALL.Net 一级区域 ID"},
		{Key: "allnet_region_name0", Value: "W", Desc: "ALL.Net 一级区域名称"},
		{Key: "allnet_region_name1", Value: "X", Desc: "ALL.Net 二级区域名称"},
		{Key: "allnet_region_name2", Value: "Y", Desc: "ALL.Net 三级区域名称"},
		{Key: "allnet_region_name3", Value: "Z", Desc: "ALL.Net 四级区域名称"},
		{Key: "allnet_country", Value: "JPN", Desc: "ALL.Net 国家代码"},
		{Key: "allnet_nickname", Value: "", Desc: "ALL.Net 场所昵称"},
		{Key: "maimai_recommend_select_music_ids", Value: "", Desc: "maimai 推荐选曲乐曲 ID；多个数字以逗号分隔"},
		{Key: "maimai_recommend_rate_music_ids", Value: "", Desc: "maimai 推荐评级乐曲 ID；多个数字以逗号分隔"},
		{Key: "maimai_weekly_mission_category", Value: "0", Desc: "maimai 周任务分类"},
		{Key: "maimai_weekly_update_date", Value: "2024-01-01 00:00:00.0", Desc: "maimai 周任务开始时间"},
		{Key: "maimai_weekly_before_date", Value: "2077-01-01 00:00:00.0", Desc: "maimai 周任务结束时间"},
		{Key: "maimai_festa_event_id", Value: "0", Desc: "maimai 嘉年华活动 ID"},
		{Key: "maimai_festa_rally_period", Value: "false", Desc: "maimai 嘉年华是否处于拉力期"},
		{Key: "maimai_festa_circle_join_not_allowed", Value: "false", Desc: "maimai 嘉年华是否禁止加入圆环"},
		{Key: "maimai_festa_jacking_side_id", Value: "-1", Desc: "maimai 嘉年华劫持阵营 ID（-1 为随机）"},
		{Key: "maimai_circle_name", Value: "一緒に歌おう！", Desc: "maimai 圆环默认名称"},
		{Key: "card_print_expiration_days", Value: "15", Desc: "maimai 卡片打印凭证有效天数"},
		{Key: "game_reboot_start_time", Value: "2020-01-01 23:59:00.0", Desc: "maimai 机台重启开始时间"},
		{Key: "game_reboot_end_time", Value: "2020-01-01 23:59:00.0", Desc: "maimai 机台重启结束时间"},
		{Key: "game_reboot_interval", Value: "0", Desc: "maimai 机台重启间隔"},
		{Key: "game_request_interval", Value: "10", Desc: "maimai 请求间隔"},
		{Key: "game_movie_upload_limit", Value: "0", Desc: "maimai 视频上传限制"},
		{Key: "game_movie_status", Value: "0", Desc: "maimai 视频服务状态"},
		{Key: "game_movie_server_uri", Value: "", Desc: "maimai 视频服务器地址"},
		{Key: "game_deliver_server_uri", Value: "", Desc: "maimai 配送服务器地址"},
		{Key: "game_old_server_uri", Value: "", Desc: "maimai 旧服务器地址"},
		{Key: "game_usb_download_server_uri", Value: "", Desc: "maimai USB 下载服务器地址"},
		{Key: "game_ping_disable", Value: "true", Desc: "maimai 是否禁用 Ping"},
		{Key: "game_packet_timeout", Value: "20000", Desc: "maimai 数据包超时毫秒数"},
		{Key: "game_packet_timeout_long", Value: "60000", Desc: "maimai 长数据包超时毫秒数"},
		{Key: "game_packet_retry_count", Value: "5", Desc: "maimai 数据包重试次数"},
		{Key: "game_user_data_download_error_timeout", Value: "300000", Desc: "maimai 用户数据下载错误超时毫秒数"},
		{Key: "game_user_data_download_error_retry_count", Value: "5", Desc: "maimai 用户数据下载错误重试次数"},
		{Key: "game_user_data_download_same_packet_retry_count", Value: "5", Desc: "maimai 用户数据下载同包重试次数"},
		{Key: "game_user_data_upload_skip_timeout", Value: "0", Desc: "maimai 用户数据上传跳过超时毫秒数"},
		{Key: "game_user_data_upload_skip_retry_count", Value: "0", Desc: "maimai 用户数据上传跳过重试次数"},
		{Key: "game_icon_photo_disable", Value: "true", Desc: "maimai 是否禁用头像照片"},
		{Key: "game_upload_photo_disable", Value: "false", Desc: "maimai 是否禁用成绩照片上传"},
		{Key: "game_max_count_music", Value: "0", Desc: "maimai 音乐下发最大数量"},
		{Key: "game_max_count_item", Value: "0", Desc: "maimai 道具下发最大数量"},
	}
	for _, config := range defaultConfigs {
		if err := DB.Where(&model.SystemConfig{Key: config.Key}).FirstOrCreate(&config).Error; err != nil {
			log.Fatalf("Failed to initialize system config %s: %v", config.Key, err)
		}
	}
	log.Println("[MaiGoDX] Default system configurations initialized.")

	log.Printf("[MaiGoDX] Database initialized successfully using driver: %s", dbType)
}
