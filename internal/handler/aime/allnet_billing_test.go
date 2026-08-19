package aime

import (
	"bytes"
	"compress/zlib"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func encodeBillingTestRequest(t *testing.T, plain string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write([]byte(plain)); err != nil {
		t.Fatalf("write request: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close request: %v", err)
	}
	return compressed.Bytes()
}

func TestBillingRequestCompatibility(t *testing.T) {
	body := encodeBillingTestRequest(t, "keychipid=A0000001234")
	values, err := decodeBilling(body)
	if err != nil {
		t.Fatalf("decode billing request: %v", err)
	}
	if values["keychipid"] != "A0000001234" {
		t.Fatalf("unexpected keychip id: %q", values["keychipid"])
	}

	request := httptest.NewRequest(http.MethodPost, "/request", bytes.NewReader(body))
	response := httptest.NewRecorder()
	HandleBillingRequest(response, request)
	result := response.Result()
	defer result.Body.Close()
	payload, _ := io.ReadAll(result.Body)
	if result.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", result.StatusCode, payload)
	}
	match := regexp.MustCompile(`playlimitsig=([0-9a-f]+)`).FindSubmatch(payload)
	if len(match) != 2 || len(match[1]) != 256 {
		t.Fatalf("expected 1024-bit hexadecimal billing signature, got %q", payload)
	}
}
