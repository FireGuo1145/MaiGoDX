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
	"strconv"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

const terminalSessionTTL = 48 * time.Hour

// HandleAllNetSelfTest implements AquaDX's plain ALL.Net health-check
// endpoint. Cabinets expect this exact text rather than the web UI fallback.
func HandleAllNetSelfTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "Server running")
}

// HandleNaomiTest implements the Naomi title-server connectivity check. If it
// falls through to the SPA handler, a cabinet receives HTML and marks the
// title server as bad even though PowerOn itself succeeded.
func HandleNaomiTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "naomi ok")
}

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
	parsedVersion, err := strconv.ParseFloat(version, 64)
	if err != nil || parsedVersion < 1.0 {
		log.Printf("[MaiGoDX] ALL.Net PowerOn rejected: keychip=%s game=%s version=%q is below 1.0", keychip, gameID, version)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(""))
		return
	}
	terminal, err := findTerminalByKeychip(keychip)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn database error: keychip=%s game=%s error=%v", keychip, gameID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	protected := allNetKeychipProtectionEnabled()
	permissive := allNetKeychipPermissive()
	if protected && terminal.ID == 0 && !permissive {
		terminalReject(w, r, "PowerOn", "keychip is not registered: "+keychip)
		return
	}
	if protected && terminal.ID != 0 && !terminal.IsEnabled {
		terminalReject(w, r, "PowerOn", "keychip is disabled: "+terminal.KeychipID)
		return
	}
	if terminal.ID != 0 {
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
	} else {
		log.Printf("[MaiGoDX] ALL.Net PowerOn permissive compatibility: accepting unregistered keychip=%s game=%s", keychip, gameID)
	}
	base := allNetPublicHost(r)
	uriBase := allNetPublicScheme() + "://" + base
	routeMode := "g"
	sessionLog := "disabled"

	// AquaDX only emits /gs/{token}/... when explicit Keychip protection is
	// enabled. Its default PowerOn response uses /g/{game}/{version}/. Keep
	// terminal registration for ownership and auditing, but make protection an
	// opt-in server setting so ordinary clients retain AquaDX-compatible routes.
	if protected {
		token, err := createTerminalSession(&terminal, gameID)
		if err != nil {
			log.Printf("[MaiGoDX] ALL.Net PowerOn session creation failed: keychip=%s error=%v", terminal.KeychipID, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		routeMode = "gs"
		sessionLog = compactTerminalToken(token)
		uriBase += "/gs/" + token
	} else {
		uriBase += "/g"
	}
	uri := uriBase + "/" + gameID + "/" + version + "/"
	fields := aquaPowerOnFields(uri, base, request["token"], request["format_ver"])
	log.Printf("[MaiGoDX] ALL.Net PowerOn accepted: keychip=%s terminal=%d game=%s version=%s route=%s session=%s uri=%s remote=%s", keychip, terminal.ID, gameID, version, routeMode, sessionLog, uri, clientIP(r))
	// AquaDX returns a plain URL-formatted response with a trailing newline.
	// KanadeDX parses this with Split('&') and Split('='), so compressed DFI
	// output or missing base fields causes its PowerOn parser to fail.
	writeAllNetPowerOn(w, fields)
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

	// AquaDX validates /gs/{token}/... and forwards it internally to the
	// controller's canonical /g/... route before dispatching the game API.
	r.URL.Path = "/g/" + strings.Join(parts[2:], "/")
	r.URL.RawPath = ""
	MaimaiHandler(w, r)
}

func allNetKeychipProtectionEnabled() bool {
	return strings.EqualFold(maimaiConfigValue("allnet_check_keychip", "false"), "true")
}

func allNetKeychipPermissive() bool {
	return strings.EqualFold(maimaiConfigValue("allnet_keychip_permissive", "false"), "true")
}

func allNetPublicHost(r *http.Request) string {
	if configured := strings.TrimSpace(maimaiConfigValue("allnet_public_host", "")); configured != "" {
		return configured
	}
	// AquaDX gives this header precedence. It is supplied by an ALL.Net proxy
	// when the Host header points at the proxy rather than the cabinet-reachable
	// game server address.
	if forwarded := strings.TrimSpace(r.Header.Get("AllNet-Forwarded-From")); forwarded != "" {
		return forwarded
	}
	return forwardedHost(r)
}

func allNetPublicScheme() string {
	scheme := strings.ToLower(strings.TrimSpace(maimaiConfigValue("allnet_public_scheme", "http")))
	if scheme == "https" {
		return scheme
	}
	return "http"
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

type allNetField struct {
	key   string
	value string
}

func aquaPowerOnFields(uri, host, clientToken, formatVersion string) []allNetField {
	// Keep this ordered list aligned with AquaDX AllNetProps.map and AllNet.powerOn.
	// Deliberately do not URL-escape or compress: AquaDX writes resp.toUrl()+"\n".
	fields := []allNetField{
		{key: "stat", value: "1"},
		{key: "name", value: maimaiConfigValue("allnet_name", "")},
		{key: "place_id", value: maimaiConfigValue("allnet_place_id", "123")},
		{key: "region0", value: maimaiConfigValue("allnet_region0", "1")},
		{key: "region_name0", value: maimaiConfigValue("allnet_region_name0", "W")},
		{key: "region_name1", value: maimaiConfigValue("allnet_region_name1", "X")},
		{key: "region_name2", value: maimaiConfigValue("allnet_region_name2", "Y")},
		{key: "region_name3", value: maimaiConfigValue("allnet_region_name3", "Z")},
		{key: "country", value: maimaiConfigValue("allnet_country", "JPN")},
		{key: "nickname", value: maimaiConfigValue("allnet_nickname", "")},
		{key: "uri", value: uri},
		{key: "host", value: host},
	}
	if strings.HasPrefix(formatVersion, "2") {
		now := time.Now()
		return append(fields,
			allNetField{key: "year", value: now.Format("2006")},
			allNetField{key: "month", value: now.Format("01")},
			allNetField{key: "day", value: now.Format("02")},
			allNetField{key: "hour", value: now.Format("15")},
			allNetField{key: "minute", value: now.Format("04")},
			allNetField{key: "second", value: now.Format("05")},
			allNetField{key: "setting", value: "1"},
			allNetField{key: "timezone", value: "+09:00"},
			allNetField{key: "res_class", value: "PowerOnResponseV2"},
		)
	}
	return append(fields,
		allNetField{key: "allnet_id", value: "456"},
		allNetField{key: "client_timezone", value: "+0900"},
		allNetField{key: "utc_time", value: time.Now().UTC().Format("2006-01-02T15:04:05Z")},
		allNetField{key: "setting", value: ""},
		allNetField{key: "res_ver", value: "3"},
		allNetField{key: "token", value: clientToken},
	)
}

func writeAllNetPowerOn(w http.ResponseWriter, fields []allNetField) {
	pairs := make([]string, 0, len(fields))
	for _, field := range fields {
		pairs = append(pairs, field.key+"="+field.value)
	}
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(strings.Join(pairs, "&") + "\n"))
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
