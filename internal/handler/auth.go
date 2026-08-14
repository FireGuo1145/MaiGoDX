package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
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
	if verificationRequired {
		if err := sendVerificationEmail(req.Email, verifyToken); err != nil {
			// Do not leave an account that cannot receive a verification token.
			database.DB.Unscoped().Delete(&account)
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "验证邮件发送失败：" + err.Error()})
			return
		}
	}

	response := map[string]interface{}{"success": true, "verificationRequired": verificationRequired}
	if verificationRequired {
		response["message"] = "注册成功，请查收验证邮件"
		if emailVerificationDelivery() == "development" {
			response["verifyToken"] = verifyToken // 本地开发模式的测试便利
		}
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

func emailVerificationDelivery() string {
	return strings.ToLower(authConfigValue("email_verification_delivery", "development"))
}

func sendVerificationEmail(recipient, token string) error {
	if emailVerificationDelivery() == "development" {
		return nil
	}
	if emailVerificationDelivery() != "smtp" {
		return fmt.Errorf("不支持的邮件发送方式 %q", emailVerificationDelivery())
	}

	host := authConfigValue("email_smtp_host", "")
	port := authConfigValue("email_smtp_port", "587")
	from := authConfigValue("email_smtp_from", "")
	if host == "" || from == "" {
		return fmt.Errorf("SMTP 主机和发件人地址必须配置")
	}
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return fmt.Errorf("SMTP 端口无效")
	}
	parsedFrom, err := mail.ParseAddress(from)
	if err != nil || parsedFrom.Address != from {
		return fmt.Errorf("SMTP 发件人地址无效")
	}

	username := authConfigValue("email_smtp_username", "")
	password := authConfigValue("email_smtp_password", "")
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	siteName := authConfigValue("site_name", "MaiGoDX")
	message := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s 邮箱验证\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n你的验证令牌是：%s\r\n\r\n请回到 %s 并输入此令牌完成邮箱验证。\r\n", recipient, from, siteName, token, siteName)
	if err := smtp.SendMail(net.JoinHostPort(host, port), auth, parsedFrom.Address, []string{recipient}, []byte(message)); err != nil {
		return err
	}
	return nil
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
