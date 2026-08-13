package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

const sessionCookieName = "maigodx_session"

func createSession(account *model.UserAccount) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	account.SessionToken = hex.EncodeToString(bytes)
	account.SessionExpiresAt = time.Now().Add(24 * time.Hour)
	return account.SessionToken, database.DB.Save(account).Error
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: token, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: int((24 * time.Hour).Seconds()),
	})
}

func currentAccount(r *http.Request) (*model.UserAccount, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil, false
	}
	var account model.UserAccount
	if err := database.DB.Where("session_token = ? AND session_expires_at > ?", cookie.Value, time.Now()).First(&account).Error; err != nil {
		return nil, false
	}
	return &account, true
}

func requireAccount(w http.ResponseWriter, r *http.Request) (*model.UserAccount, bool) {
	account, ok := currentAccount(r)
	if ok {
		return account, true
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "登录会话无效或已过期"})
	return nil, false
}

func requireAdmin(w http.ResponseWriter, r *http.Request) (*model.UserAccount, bool) {
	account, ok := requireAccount(w, r)
	if !ok {
		return nil, false
	}
	if account.IsAdmin {
		return account, true
	}
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "需要管理员权限"})
	return nil, false
}

func clearSession(w http.ResponseWriter, account *model.UserAccount) {
	account.SessionToken = ""
	account.SessionExpiresAt = time.Time{}
	_ = database.DB.Save(account).Error
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
}
