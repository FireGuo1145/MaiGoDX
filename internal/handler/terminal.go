package handler

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

const terminalSessionTTL = 24 * time.Hour

// HandleAllNetPowerOn authenticates a cabinet Keychip and returns its game URL.
func HandleAllNetPowerOn(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	request, err := decodeAllNet(payload)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	keychip := strings.TrimSpace(request["serial"])
	gameID := strings.TrimSpace(request["game_id"])
	version := strings.TrimSpace(request["ver"])
	if keychip == "" || gameID == "" {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	terminal, err := findTerminalByKeychip(keychip)
	if err != nil || !terminal.IsEnabled {
		http.Error(w, "", http.StatusForbidden)
		return
	}
	if terminal.GameID != "" && !strings.EqualFold(terminal.GameID, gameID) {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	terminal.GameVersion = version
	terminal.LastSeenAt = time.Now()
	terminal.LastSeenIP = clientIP(r)
	_ = database.DB.Save(&terminal).Error

	token, err := createTerminalSession(&terminal, gameID)
	if err != nil {
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
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Pragma", "DFI")
	_, _ = w.Write(encoded)
}

// HandleTerminalMaimai verifies the PowerOn session before dispatching game APIs.
func HandleTerminalMaimai(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 || parts[0] != "gs" {
		http.Error(w, "", http.StatusForbidden)
		return
	}

	var session model.TerminalSession
	if lookup := database.DB.Where("token = ? AND expires_at > ?", parts[1], time.Now()).Limit(1).Find(&session); lookup.Error != nil || lookup.RowsAffected == 0 {
		http.Error(w, "", http.StatusForbidden)
		return
	}
	var terminal model.Terminal
	if lookup := database.DB.Where("id = ?", session.TerminalID).Limit(1).Find(&terminal); lookup.Error != nil || lookup.RowsAffected == 0 || !terminal.IsEnabled || !strings.EqualFold(terminal.GameID, parts[2]) {
		http.Error(w, "", http.StatusForbidden)
		return
	}
	terminal.LastSeenAt = time.Now()
	terminal.LastSeenIP = clientIP(r)
	_ = database.DB.Save(&terminal).Error
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
