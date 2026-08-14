package handler

import (
	"net/http"
	"net/http/httptest"
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
	if got := allNetPublicHost(request); got != "mai.example" {
		t.Fatalf("configured host=%q, want mai.example", got)
	}
}
