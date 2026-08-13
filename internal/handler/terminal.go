package handler

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

const terminalSessionTTL = 48 * time.Hour

// HandleAllNetPowerOn authenticates a cabinet Keychip and returns its game URL.
func HandleAllNetPowerOn(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn rejected: cannot read request body: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	request, err := decodeAllNet(payload)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn rejected: invalid DFI payload (bytes=%d): %v", len(payload), err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	keychip := normalizeKeychip(request["serial"])
	gameID := strings.ToUpper(strings.TrimSpace(request["game_id"]))
	version := strings.TrimSpace(request["ver"])
	if keychip == "" || gameID == "" {
		terminalReject(w, r, "PowerOn", "missing serial or game_id")
		return
	}
	terminal, err := findTerminalByKeychip(keychip)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn database error: keychip=%s game=%s error=%v", keychip, gameID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if terminal.ID == 0 {
		terminalReject(w, r, "PowerOn", "keychip is not registered: "+keychip)
		return
	}
	if !terminal.IsEnabled {
		terminalReject(w, r, "PowerOn", "keychip is disabled: "+terminal.KeychipID)
		return
	}
	// AquaDX authenticates a Keychip independently from the requested game ID.
	// Keep the configured game ID for portal display, but do not reject a valid
	// authenticated Keychip merely because the cabinet reports another game code.
	if terminal.GameID != "" && !strings.EqualFold(terminal.GameID, gameID) {
		log.Printf("[MaiGoDX] ALL.Net PowerOn compatibility: keychip=%s configured_game=%s requested_game=%s; accepting authenticated session", terminal.KeychipID, terminal.GameID, gameID)
	}
	terminal.GameVersion = version
	terminal.LastSeenAt = time.Now()
	terminal.LastSeenIP = clientIP(r)
	if err := database.DB.Save(&terminal).Error; err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn warning: failed to update terminal %d: %v", terminal.ID, err)
	}
	token, err := createTerminalSession(&terminal, gameID)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn session creation failed: keychip=%s error=%v", terminal.KeychipID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	base := forwardedHost(r)
	uri := "http://" + base + "/gs/" + token + "/" + gameID + "/" + version + "/"
	response := map[string]string{
		"uri":             uri,
		"host":            base,
		"region0":         "1",
		"allnet_id":       "456",
		"client_timezone": "+0900",
		"setting":         "",
		"res_ver":         "3",
		"token":           request["token"],
	}
	if strings.HasPrefix(request["format_ver"], "2") {
		now := time.Now()
		response["year"] = now.Format("2006")
		response["month"] = now.Format("01")
		response["day"] = now.Format("02")
		response["hour"] = now.Format("15")
		response["minute"] = now.Format("04")
		response["second"] = now.Format("05")
		response["timezone"] = "+09:00"
		response["res_class"] = "PowerOnResponseV2"
	}
	encoded, err := encodeAllNet(response)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn response encoding failed: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	log.Printf("[MaiGoDX] ALL.Net PowerOn accepted: keychip=%s terminal=%d game=%s version=%s session=%s remote=%s", terminal.KeychipID, terminal.ID, gameID, version, compactTerminalToken(token), clientIP(r))
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Pragma", "DFI")
	_, _ = w.Write(encoded)
}

// HandleTerminalMaimai verifies the PowerOn session before dispatching game APIs.
func HandleTerminalMaimai(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "gs" || strings.TrimSpace(parts[1]) == "" {
		terminalReject(w, r, "SecureGame", "malformed protected path")
		return
	}
	token := parts[1]
	var session model.TerminalSession
	lookup := database.DB.Where("token = ? AND expires_at > ?", token, time.Now()).Limit(1).Find(&session)
	if lookup.Error != nil {
		log.Printf("[MaiGoDX] ALL.Net SecureGame database error: path=%s session=%s error=%v", r.URL.Path, compactTerminalToken(token), lookup.Error)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if lookup.RowsAffected == 0 {
		terminalReject(w, r, "SecureGame", "session token does not exist or has expired: "+compactTerminalToken(token))
		return
	}
	session.ExpiresAt = time.Now().Add(terminalSessionTTL)
	if err := database.DB.Save(&session).Error; err != nil {
		log.Printf("[MaiGoDX] ALL.Net SecureGame warning: failed to renew session=%s: %v", compactTerminalToken(token), err)
	}
	MaimaiHandler(w, r)
}

func createTerminalSession(terminal *model.Terminal, gameID string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	session := model.TerminalSession{Token: token, TerminalID: terminal.ID, GameID: gameID, ExpiresAt: time.Now().Add(terminalSessionTTL)}
	return token, database.DB.Save(&session).Error
}

func findTerminalByKeychip(value string) (model.Terminal, error) {
	keychip := normalizeKeychip(value)
	var terminal model.Terminal
	lookup := database.DB.Where("keychip_id = ?", keychip).Limit(1).Find(&terminal)
	if lookup.Error != nil || lookup.RowsAffected > 0 {
		return terminal, lookup.Error
	}

	// Segatools commonly sends the 11-character Keychip prefix. Try AquaDX's
	// conventional 1337 suffix before accepting another full serial with it.
	if len(keychip) == 11 {
		lookup = database.DB.Where("keychip_id = ?", keychip+"1337").Limit(1).Find(&terminal)
		if lookup.Error != nil || lookup.RowsAffected > 0 {
			return terminal, lookup.Error
		}
		lookup = database.DB.Where("keychip_id LIKE ?", keychip+"%").Order("id asc").Limit(1).Find(&terminal)
		if lookup.Error != nil || lookup.RowsAffected > 0 {
			return terminal, lookup.Error
		}
	}
	return model.Terminal{}, nil
}

func normalizeKeychip(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("AllNet-Forwarded-From"); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func forwardedHost(r *http.Request) string {
	if forwarded := r.Header.Get("Host"); forwarded != "" {
		return forwarded
	}
	return r.Host
}

func terminalReject(w http.ResponseWriter, r *http.Request, stage, reason string) {
	log.Printf("[MaiGoDX] ALL.Net %s rejected: %s | method=%s path=%s remote=%s", stage, reason, r.Method, r.URL.Path, clientIP(r))
	http.Error(w, "", http.StatusForbidden)
}

func compactTerminalToken(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:8] + "..."
}

func decodeAllNet(source []byte) (map[string]string, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(source)))
	if err != nil {
		return nil, err
	}
	reader, err := zlib.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	plain, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, pair := range strings.Split(strings.TrimSpace(string(plain)), "&") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result, nil
}

func encodeAllNet(values map[string]string) ([]byte, error) {
	pairs := make([]string, 0, len(values))
	for key, value := range values {
		pairs = append(pairs, key+"="+value)
	}
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write([]byte(strings.Join(pairs, "&"))); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(compressed.Bytes())), nil
}
