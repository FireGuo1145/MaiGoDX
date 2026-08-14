package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// HandleRegister 处理用户注册
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	var req model.RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	// 检查邮箱是否已注册
	var existing model.UserAccount
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "该邮箱已被注册"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "密码加密失败"})
		return
	}

	verificationRequired := emailVerificationRequired()
	verifyToken := ""
	if verificationRequired {
		tokenBytes := make([]byte, 16)
		_, _ = rand.Read(tokenBytes)
		verifyToken = hex.EncodeToString(tokenBytes)
	}

	account := model.UserAccount{
		Email:        req.Email,
		PasswordHash: string(hashed),
		Username:     req.Username,
		IsVerified:   !verificationRequired,
		VerifyToken:  verifyToken,
	}

	if err := database.DB.Create(&account).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "创建账号失败"})
		return
	}

	response := map[string]interface{}{"success": true, "verificationRequired": verificationRequired}
	if verificationRequired {
		response["message"] = "注册成功，请查收验证邮件"
		response["verifyToken"] = verifyToken // 开发与测试便利
	} else {
		response["message"] = "注册成功，现在可以直接登录"
	}
	_ = json.NewEncoder(w).Encode(response)
}

// HandleLogin 处理用户登录
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	var req model.LoginRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	var account model.UserAccount
	if err := database.DB.Where("email = ?", req.Email).First(&account).Error; err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "用户不存在或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "用户不存在或密码错误"})
		return
	}

	if emailVerificationRequired() && !account.IsVerified {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "邮箱尚未验证，请先完成验证"})
		return
	}

	token, err := createSession(&account)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "创建登录会话失败"})
		return
	}
	setSessionCookie(w, token)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true, "message": "登录成功", "username": account.Username,
		"email": account.Email, "isAdmin": account.IsAdmin,
	})
}

// HandleGetPublicSiteSettings exposes only presentation/auth flags needed
// before login; administrative game configuration remains protected.
func HandleGetPublicSiteSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":                  true,
		"siteName":                 authConfigValue("site_name", "MaiGoDX"),
		"requireEmailVerification": emailVerificationRequired(),
	})
}

func emailVerificationRequired() bool {
	value := strings.ToLower(authConfigValue("require_email_verification", "true"))
	return value == "true" || value == "1" || value == "yes"
}

func authConfigValue(key, fallback string) string {
	var config model.SystemConfig
	if err := database.DB.Where("key = ?", key).First(&config).Error; err != nil {
		return fallback
	}
	value := strings.TrimSpace(config.Value)
	if value == "" {
		return fallback
	}
	return value
}

// HandleVerifyEmail 处理邮箱验证
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if account, ok := currentAccount(r); ok {
		clearSession(w, account)
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "已退出登录"})
}

func HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	var req model.VerifyEmailRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Email == "" || req.Token == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "验证参数错误"})
		return
	}

	var account model.UserAccount
	if err := database.DB.Where("email = ? AND verify_token = ?", req.Email, req.Token).First(&account).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "验证令牌无效或邮箱不匹配"})
		return
	}

	account.IsVerified = true
	account.VerifyToken = ""
	database.DB.Save(&account)

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "邮箱验证成功！现在可以登录了",
	})
}
