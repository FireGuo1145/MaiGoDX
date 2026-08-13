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
