package aime

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

func TestAllNetTitleServerHealthEndpointsMatchAquaDX(t *testing.T) {
	for _, test := range []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{name: "ALL.Net self test", handler: HandleAllNetSelfTest, want: "Server running"},
		{name: "Naomi title test", handler: HandleNaomiTest, want: "naomi ok"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			test.handler(response, httptest.NewRequest(http.MethodGet, "/", nil))
			if got := response.Body.String(); got != test.want {
				t.Fatalf("response=%q, want %q", got, test.want)
			}
			if got := response.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
				t.Fatalf("Content-Type=%q", got)
			}
		})
	}
}

func TestTerminalSessionTokenMatchesAquaDXFormat(t *testing.T) {
	token, err := generateTerminalSessionToken()
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	if len(token) != terminalSessionTokenLength {
		t.Fatalf("token length=%d, want %d", len(token), terminalSessionTokenLength)
	}
	for _, value := range token {
		if !strings.ContainsRune(terminalSessionTokenChars, value) {
			t.Fatalf("token contains non-AquaDX character %q: %q", value, token)
		}
	}
}

func TestAllNetPublicHostMatchesAquaDXPrecedence(t *testing.T) {
	setupMaimaiTestDB(t)

	request := httptest.NewRequest(http.MethodPost, "http://proxy.example:8080/sys/servlet/PowerOn", nil)
	request.Header.Set("AllNet-Forwarded-From", "192.168.1.20")
	if got := allNetPublicHost(request); got != "192.168.1.20" {
		t.Fatalf("forwarded host=%q, want 192.168.1.20", got)
	}

	if err := database.DB.Create(&model.SystemConfig{Key: "allnet_public_host", Value: "mai.example"}).Error; err != nil {
		t.Fatalf("create configured host: %v", err)
	}
	if got := allNetPublicHost(request); got != "192.168.1.20" {
		t.Fatalf("forwarded host with configured server=%q, want 192.168.1.20", got)
	}
	if got := allNetResponseHost(request, allNetPublicHost(request)); got != "mai.example" {
		t.Fatalf("PowerOn host field=%q, want mai.example", got)
	}

	request.Header.Del("AllNet-Forwarded-From")
	if got := allNetPublicHost(request); got != "mai.example" {
		t.Fatalf("configured host without forwarded header=%q, want mai.example", got)
	}
}

func TestPowerOnUsesForwardedAddressForTitleURIAndConfiguredHostField(t *testing.T) {
	setupMaimaiTestDB(t)
	if err := database.DB.Create(&model.SystemConfig{Key: "allnet_public_host", Value: "mai.example"}).Error; err != nil {
		t.Fatalf("create configured host: %v", err)
	}
	payload, err := encodeAllNet(map[string]string{
		"serial": "A0000001234", "game_id": "SDEZ", "ver": "1.60", "token": "cabinet-token",
	})
	if err != nil {
		t.Fatalf("encode PowerOn payload: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/sys/servlet/PowerOn", strings.NewReader(string(payload)))
	request.Header.Set("AllNet-Forwarded-From", "192.168.1.20")
	response := httptest.NewRecorder()

	HandleAllNetPowerOn(response, request)
	values, err := url.ParseQuery(strings.TrimSpace(response.Body.String()))
	if err != nil {
		t.Fatalf("parse PowerOn response: %v", err)
	}
	if got := values.Get("uri"); got != "http://192.168.1.20/g/SDEZ/1.60/" {
		t.Fatalf("uri=%q", got)
	}
	if got := values.Get("host"); got != "mai.example" {
		t.Fatalf("host=%q, want mai.example", got)
	}
}
