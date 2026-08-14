package middleware

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// CompressionMiddleware implements the game-service framing used by AquaDX.
// Game clients expect every /g response to be zlib-compressed regardless of
// Accept-Encoding; Accept-Encoding is an HTTP negotiation header and is not
// used by the Sega game protocol. Returning plain JSON causes a cabinet to
// repeatedly retry GetGameSetting and report Aime as unavailable.
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isCompressedGamePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		isDFI := strings.EqualFold(strings.TrimSpace(r.Header.Get("Pragma")), "DFI")
		if err := decodeGameRequestBody(r, isDFI); err != nil {
			http.Error(w, "Failed to decode request", http.StatusBadRequest)
			return
		}

		response := &gameResponseWriter{
			ResponseWriter: w,
			buf:            new(bytes.Buffer),
			statusCode:     http.StatusOK,
			isDFI:          isDFI,
		}
		next.ServeHTTP(response, r)
		if err := response.Close(); err != nil {
			// The response has not been committed before Close, so a normal HTTP
			// error is still safe here.
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})
}

func isCompressedGamePath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return false
	}
	switch parts[0] {
	case "g":
		// AquaDX excludes WACCA (SDFE) from its game response compressor.
		return !strings.EqualFold(parts[1], "SDFE")
	case "gs":
		// Protected ALL.Net routes are rewritten to /g only after this outer
		// middleware runs, so they must be recognised here as well.
		return len(parts) >= 3 && !strings.EqualFold(parts[2], "SDFE")
	default:
		return false
	}
}

func decodeGameRequestBody(r *http.Request, isDFI bool) error {
	encoded := strings.EqualFold(strings.TrimSpace(r.Header.Get("Content-Encoding")), "deflate")
	if !encoded && !isDFI {
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if isDFI {
		body, err = base64.StdEncoding.DecodeString(strings.TrimSpace(string(body)))
		if err != nil {
			return fmt.Errorf("decode DFI base64: %w", err)
		}
	}
	reader, err := zlib.NewReader(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("open zlib request: %w", err)
	}
	decompressed, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return fmt.Errorf("read zlib request: %w", readErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close zlib request: %w", closeErr)
	}
	r.Body = io.NopCloser(bytes.NewReader(decompressed))
	r.ContentLength = int64(len(decompressed))
	r.Header.Del("Content-Encoding")
	return nil
}

type gameResponseWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
	isDFI      bool
}

func (w *gameResponseWriter) WriteHeader(statusCode int) {
	if w.statusCode == http.StatusOK {
		w.statusCode = statusCode
	}
}

func (w *gameResponseWriter) Write(body []byte) (int, error) {
	return w.buf.Write(body)
}

func (w *gameResponseWriter) Close() error {
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(w.buf.Bytes()); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	payload := compressed.Bytes()
	if w.isDFI {
		w.Header().Set("Pragma", "DFI")
		w.Header().Del("Content-Encoding")
		payload = []byte(base64.StdEncoding.EncodeToString(payload))
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Encoding", "deflate")
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
	w.ResponseWriter.WriteHeader(w.statusCode)
	_, err := w.ResponseWriter.Write(payload)
	return err
}
