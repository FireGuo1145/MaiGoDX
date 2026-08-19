package aime

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	_ "embed"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

//go:embed billing.der
var billingPrivateKeyDER []byte

var (
	billingKeyOnce sync.Once
	billingKey     *rsa.PrivateKey
	billingKeyErr  error
)

// HandleAllNetDownloadOrder implements the ALL.Net order-download acknowledgement.
// Unlike PowerOn, its response remains plain URL-formatted text.
func HandleAllNetDownloadOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	request, err := decodeAllNet(body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	serial := strings.TrimSpace(request["serial"])
	if serial == "" {
		serial = "A69E01A8888"
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintf(w, "stat=1&serial=%s\n", serial)
}

// HandleBillingRequest implements the public billing /request response used by
// Sega cabinets. Billing requests use raw DEFLATE rather than ALL.Net's Base64+
// zlib wrapper.
func HandleBillingRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	request, err := decodeBilling(body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	keychipID := request["keychipid"]
	playLimitSignature := billingSignature(keychipID, 1024)
	nearFullSignature := billingSignature(keychipID, 66048)
	response := []string{
		"result=0",
		"waittime=100",
		"linelimit=1",
		"message=",
		"playlimit=1024",
		"playlimitsig=" + playLimitSignature,
		"protocolver=1.000",
		"nearfull=66048",
		"nearfullsig=" + nearFullSignature,
		"fixlogcnt=0",
		"fixinterval=5",
		"playhistory=000000/0:000000/0:000000/0",
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, strings.Join(response, "&")+"\n")
}

func decodeBilling(source []byte) (map[string]string, error) {
	reader := flate.NewReader(bytes.NewReader(source))
	plain, err := io.ReadAll(reader)
	_ = reader.Close()
	if err == nil {
		return decodeProtocolValues(string(plain)), nil
	}

	// The official protocol uses raw DEFLATE. Accept a zlib wrapper as well so
	// diagnostic and older runtime clients are not rejected solely on framing.
	zlibReader, zlibErr := zlib.NewReader(bytes.NewReader(source))
	if zlibErr != nil {
		return nil, err
	}
	defer zlibReader.Close()
	plain, zlibErr = io.ReadAll(zlibReader)
	if zlibErr != nil {
		return nil, zlibErr
	}
	return decodeProtocolValues(string(plain)), nil
}

func billingSignature(keychipID string, value uint32) string {
	key, err := loadBillingKey()
	if err != nil {
		return ""
	}
	payload := make([]byte, 15)
	binary.LittleEndian.PutUint32(payload, value)
	copy(payload[4:], []byte(keychipID))
	digest := sha1.Sum(payload)
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, digest[:])
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", signature)
}

func loadBillingKey() (*rsa.PrivateKey, error) {
	billingKeyOnce.Do(func() {
		parsed, err := x509.ParsePKCS8PrivateKey(billingPrivateKeyDER)
		if err != nil {
			billingKeyErr = err
			return
		}
		var ok bool
		billingKey, ok = parsed.(*rsa.PrivateKey)
		if !ok {
			billingKeyErr = fmt.Errorf("billing key is not RSA")
		}
	})
	return billingKey, billingKeyErr
}

func decodeProtocolValues(value string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(strings.TrimSpace(value), "&") {
		key, entry, found := strings.Cut(pair, "=")
		if found {
			result[key] = entry
		}
	}
	return result
}
