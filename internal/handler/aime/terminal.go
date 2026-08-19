package aime

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/handler/chunithm"
	"github.com/FireGuo1145/MaiGoDX/internal/handler/maimai"
	"github.com/FireGuo1145/MaiGoDX/internal/handler/ongeki"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

const (
	terminalSessionTTL         = 48 * time.Hour
	terminalSessionTokenLength = 32
	terminalSessionTokenChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.~"
)

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
	log.Printf("[MaiGoDX] ALL.Net PowerOn received: remote=%s method=%s path=%s host=%s forwarded=%s", clientIP(r), r.Method, r.URL.Path, r.Host, r.Header.Get("AllNet-Forwarded-From"))
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn rejected: cannot read request body: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	request, err := decodeAllNet(payload)
	if err != nil {
		log.Printf("[MaiGoDX] ALL.Net PowerOn rejected: invalid DFI payload remote=%s path=%s bytes=%d error=%v", clientIP(r), r.URL.Path, len(payload), err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	keychip := normalizeKeychip(request["serial"])
	gameID := strings.ToUpper(strings.TrimSpace(request["game_id"]))
	version := strings.TrimSpace(request["ver"])
	log.Printf("[MaiGoDX] ALL.Net PowerOn parsed: remote=%s serial=%s game=%s version=%s format_ver=%s token_present=%t", clientIP(r), keychip, gameID, version, request["format_ver"], strings.TrimSpace(request["token"]) != "")
	if keychip == "" || gameID == "" {
		log.Printf("[MaiGoDX] ALL.Net PowerOn rejected: missing serial or game_id remote=%s request=%v", clientIP(r), request)
		terminalReject(w, r, "PowerOn", "missing serial or game_id")
		return
	}
	parsedVersion, err := strconv.ParseFloat(version, 64)
	if err != nil || parsedVersion < 1.0 {
		log.Printf("[MaiGoDX] ALL.Net PowerOn rejected: keychip=%s game=%s version=%q is below 1.0 remote=%s", keychip, gameID, version, clientIP(r))
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
	// AquaDX deliberately distinguishes the address a cabinet must connect to
	// from the administrative host value returned in the PowerOn payload.  A
	// forwarding ALL.Net proxy supplies AllNet-Forwarded-From for the former;
	// an explicitly configured host remains the latter.
	base := allNetPublicHost(r)
	responseHost := allNetResponseHost(r, base)
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
	fields := aquaPowerOnFields(uri, responseHost, request["token"], request["format_ver"])
	log.Printf("[MaiGoDX] ALL.Net PowerOn accepted: keychip=%s terminal=%d game=%s version=%s route=%s session=%s uri=%s response_host=%s request_host=%s forwarded=%s remote=%s", keychip, terminal.ID, gameID, version, routeMode, sessionLog, uri, responseHost, r.Host, r.Header.Get("AllNet-Forwarded-From"), clientIP(r))
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
	switch strings.ToUpper(session.GameID) {
	case "SDHD", "SDGS":
		chunithm.Handler(w, r)
	case "SDDT":
		ongeki.Handler(w, r)
	default:
		maimai.MaimaiHandler(w, r)
	}
}

func allNetKeychipProtectionEnabled() bool {
	return strings.EqualFold(maimaiConfigValue("allnet_check_keychip", "false"), "true")
}

func allNetKeychipPermissive() bool {
	return strings.EqualFold(maimaiConfigValue("allnet_keychip_permissive", "false"), "true")
}

func allNetPublicHost(r *http.Request) string {
	// AquaDX gives this header precedence over allnet.server.host. It is
	// supplied by an ALL.Net proxy when its Host header is not the address the
	// cabinet should use for its subsequent title/game-server requests.
	if forwarded := strings.TrimSpace(r.Header.Get("AllNet-Forwarded-From")); forwarded != "" {
		return forwarded
	}
	if configured := strings.TrimSpace(maimaiConfigValue("allnet_public_host", "")); configured != "" {
		return configured
	}
	return forwardedHost(r)
}

// allNetResponseHost mirrors AquaDX's `host` PowerOn field. This is the
// configured server host where one exists, not necessarily the forwarded
// cabinet-reachable address used in `uri`.
func allNetResponseHost(r *http.Request, fallback string) string {
	if configured := strings.TrimSpace(maimaiConfigValue("allnet_public_host", "")); configured != "" {
		return configured
	}
	return fallback
}

func allNetPublicScheme() string {
	scheme := strings.ToLower(strings.TrimSpace(maimaiConfigValue("allnet_public_scheme", "http")))
	if scheme == "https" {
		return scheme
	}
	return "http"
}

func createTerminalSession(terminal *model.Terminal, gameID string) (string, error) {
	token, err := generateTerminalSessionToken()
	if err != nil {
		return "", err
	}
	session := model.TerminalSession{Token: token, TerminalID: terminal.ID, GameID: gameID, ExpiresAt: time.Now().Add(terminalSessionTTL)}
	return token, database.DB.Save(&session).Error
}

// generateTerminalSessionToken mirrors AquaDX's KeychipSession token format:
// 32 URL-safe characters selected from [a-zA-Z0-9-_.~].  This exact form is
// used in the PowerOn /gs/{token}/ URL and avoids clients imposing a shorter
// path-segment limit than our old 64-character hexadecimal token.
func generateTerminalSessionToken() (string, error) {
	const unbiasedLimit = 198 // floor(256 / 66) * 66
	token := make([]byte, 0, terminalSessionTokenLength)
	random := make([]byte, terminalSessionTokenLength*2)
	for len(token) < terminalSessionTokenLength {
		if _, err := rand.Read(random); err != nil {
			return "", err
		}
		for _, value := range random {
			if value >= unbiasedLimit {
				continue
			}
			token = append(token, terminalSessionTokenChars[int(value)%len(terminalSessionTokenChars)])
			if len(token) == terminalSessionTokenLength {
				break
			}
		}
	}
	return string(token), nil
}

func findTerminalByKeychip(value string) (model.Terminal, error) {
	var terminals []model.Terminal
	if err := database.DB.Find(&terminals).Error; err != nil {
		return model.Terminal{}, err
	}
	prefix := keychipMatchPrefix(value)
	for _, terminal := range terminals {
		if keychipMatchPrefix(terminal.KeychipID) == prefix {
			return terminal, nil
		}
	}
	return model.Terminal{}, nil
}

func findStoredTerminalByKeychipPrefix(value string) (model.Terminal, bool, error) {
	var terminals []model.Terminal
	if err := database.DB.Unscoped().Find(&terminals).Error; err != nil {
		return model.Terminal{}, false, err
	}
	prefix := keychipMatchPrefix(value)
	for _, terminal := range terminals {
		if keychipMatchPrefix(terminal.KeychipID) == prefix {
			return terminal, true, nil
		}
	}
	return model.Terminal{}, false, nil
}

func normalizeKeychip(value string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), "-", ""))
}

func formatKeychip(value string) string {
	keychip := normalizeKeychip(value)
	if len(keychip) > 4 {
		return keychip[:4] + "-" + keychip[4:]
	}
	return keychip
}

func keychipMatchPrefix(value string) string {
	keychip := normalizeKeychip(value)
	if len(keychip) > 11 {
		return keychip[:len(keychip)-4]
	}
	return keychip
}

func isKeychipRegistrationFormat(value string) bool {
	keychip := strings.ToUpper(strings.TrimSpace(value))
	if len(keychip) != 16 || keychip[4] != '-' || keychip[0] != 'A' {
		return false
	}
	for index, char := range keychip {
		if index == 4 {
			continue
		}
		if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'Z')) {
			return false
		}
	}
	return true
}

// IsKeychipRegistrationFormat validates the Keychip format used by portal
// terminal management handlers.
func IsKeychipRegistrationFormat(value string) bool {
	return isKeychipRegistrationFormat(value)
}

// FormatKeychip returns the canonical hyphenated Keychip representation.
func FormatKeychip(value string) string {
	return formatKeychip(value)
}

// FindStoredTerminalByKeychipPrefix includes soft-deleted terminals so the
// portal can safely restore a previous registration.
func FindStoredTerminalByKeychipPrefix(value string) (model.Terminal, bool, error) {
	return findStoredTerminalByKeychipPrefix(value)
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
