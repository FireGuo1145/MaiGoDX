package model

import "gorm.io/gorm"

// UserAccount 对应用户登录注册与邮箱验证账号模型
type UserAccount struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	Username     string `json:"username"`
	IsVerified   bool   `gorm:"default:false" json:"isVerified"`
	VerifyToken  string `json:"verifyToken"`
	IsAdmin      bool   `gorm:"default:false" json:"isAdmin"`
}

// UserCard 对应 Aime 卡片绑定模型
type UserCard struct {
	gorm.Model
	UserID        uint   `json:"userId"`
	AccessCode    string `gorm:"uniqueIndex;not null" json:"accessCode"` // 20位卡号
	CardId        string `json:"cardId"`                                // ICCard ID
	CardName      string `json:"cardName"`                              // 卡片备注名
}

// SystemConfig 对应系统下发配置模型（管理员可控）
type SystemConfig struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null" json:"key"`
	Value string `json:"value"`
	Desc  string `json:"desc"`
}

// RegisterRequest 注册请求体
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyEmailRequest 邮箱验证请求体
type VerifyEmailRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

// BindCardRequest 卡片绑定请求体
type BindCardRequest struct {
	Email      string `json:"email"`
	AccessCode string `json:"accessCode"`
	CardName   string `json:"cardName"`
}

// UpdateConfigRequest 更新系统配置请求体
type UpdateConfigRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
