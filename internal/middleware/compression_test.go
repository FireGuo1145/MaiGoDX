package middleware

import (
	"bytes"
	"compress/zlib"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGameResponsesAlwaysUseAquaDXDeflateFraming(t *testing.T) {
	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read decoded request: %v", err)
		}
		if string(body) != `{"clientId":"A1668242586"}` {
			t.Fatalf("decoded request=%q", body)
		}
		_, _ = w.Write([]byte(`{"gameSetting":{"requestInterval":10}}`))
	}))

	request := httptest.NewRequest(http.MethodPost, "/g/SDGA/1.60/Maimai2Servlet/GetGameSettingApi", bytes.NewBuffer(compressZlib(t, []byte(`{"clientId":"A1668242586"}`))))
	request.Header.Set("Content-Encoding", "deflate")
	// KanadeDX does not need to send Accept-Encoding. AquaDX still always
	// compresses this game response, which is the compatibility point here.
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if got := response.Header().Get("Content-Encoding"); got != "deflate" {
		t.Fatalf("Content-Encoding=%q, want deflate", got)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type=%q", got)
	}
	if got := string(decompressZlib(t, response.Body.Bytes())); got != `{"gameSetting":{"requestInterval":10}}` {
		t.Fatalf("decoded response=%q", got)
	}
}

func TestNonGameResponsesRemainPlain(t *testing.T) {
	handler := CompressionMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Server running"))
	}))
	request := httptest.NewRequest(http.MethodGet, "/sys/test", nil)
	request.Header.Set("Accept-Encoding", "deflate")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if got := response.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding=%q, want no compression", got)
	}
	if got := response.Body.String(); got != "Server running" {
		t.Fatalf("response=%q", got)
	}
}

func compressZlib(t *testing.T, source []byte) []byte {
	t.Helper()
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(source); err != nil {
		t.Fatalf("compress source: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close compressor: %v", err)
	}
	return compressed.Bytes()
}

func decompressZlib(t *testing.T, source []byte) []byte {
	t.Helper()
	reader, err := zlib.NewReader(bytes.NewReader(source))
	if err != nil {
		t.Fatalf("open response zlib: %v", err)
	}
	result, err := io.ReadAll(reader)
	if err != nil {
		_ = reader.Close()
		t.Fatalf("read response zlib: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close response zlib: %v", err)
	}
	return result
}
